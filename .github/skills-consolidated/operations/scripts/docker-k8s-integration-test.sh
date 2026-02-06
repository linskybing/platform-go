#!/bin/bash
# Docker-based Integration Test Runner
# Reference: cicd-pipeline-optimization Skill
# Runs Kubernetes integration tests in Docker container
# Usage: ./docker-k8s-integration-test.sh [test-pattern] [timeout]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../../.." && pwd)"

TEST_PATTERN="${1:-.}"
TIMEOUT="${2:-5m}"
IMAGE_NAME="platform-go-k8s-integration"
CONTAINER_NAME="platform-go-k8s-test-$$"

log_info() {
    echo "[INFO] $*"
}

log_error() {
    echo "[ERROR] $*" >&2
}

log_success() {
    echo "[SUCCESS] $*"
}

cleanup() {
    local exit_code=$?
    log_info "Cleaning up Docker container..."
    docker rm -f "$CONTAINER_NAME" 2>/dev/null || true
    return $exit_code
}

trap cleanup EXIT

log_info "Kubernetes Integration Tests (Docker)"
log_info "Project: $PROJECT_ROOT"
log_info "Test pattern: $TEST_PATTERN"
log_info "Timeout: $TIMEOUT"
log_info "Image: $IMAGE_NAME"
log_info ""

cd "$PROJECT_ROOT"

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    log_error "Docker is not installed or not in PATH"
    exit 1
fi

# Build Docker image
log_info "Building Docker image for K8s integration tests..."
if ! docker build \
    -f Dockerfile.integration \
    --target k8s-test-final \
    -t "$IMAGE_NAME" \
    .; then
    log_error "Failed to build Docker image"
    exit 1
fi

log_success "Docker image built successfully"

# Run tests in container with Docker socket mounted for KinD
log_info "Starting K8s integration tests in Docker container..."
log_info "Container name: $CONTAINER_NAME"

if docker run \
    --name "$CONTAINER_NAME" \
    --privileged \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -e "DOCKER_HOST=unix:///var/run/docker.sock" \
    -e "TEST_PATTERN=$TEST_PATTERN" \
    -e "TIMEOUT=$TIMEOUT" \
    "$IMAGE_NAME" \
    bash -c "
        set -e
        echo '[INFO] Container started, running tests...'
        
        # Run test script
        bash -c '
            TEST_PATTERN=\"${TEST_PATTERN}\" \
            TIMEOUT=\"${TIMEOUT}\" \
            .github/skills/cicd-pipeline-optimization/scripts/k8s-integration-test.sh
        '
    "; then
    log_success "All K8s integration tests passed!"
else
    log_error "K8s integration tests failed"
    
    # Try to export logs if they exist
    if docker cp "$CONTAINER_NAME":/tmp/kind-logs /tmp/kind-logs-$(date +%s) 2>/dev/null; then
        log_info "KinD logs exported"
    fi
    exit 1
fi
