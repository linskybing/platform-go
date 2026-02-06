#!/bin/bash
# Kubernetes Integration Test Runner
# Reference: cicd-pipeline-optimization Skill
# Runs K8s integration tests using KinD cluster
# Usage: ./k8s-integration-test.sh [test-pattern] [timeout]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../../.." && pwd)"
SKILL_DIR="${SCRIPT_DIR}/.."

TEST_PATTERN="${1:-.}"
TIMEOUT="${2:-5m}"
CLUSTER_NAME="platform-test"

log_info() {
    echo "[INFO] $*"
}

log_error() {
    echo "[ERROR] $*" >&2
}

log_success() {
    echo "[SUCCESS] $*"
}

cleanup_on_exit() {
    local exit_code=$?
    if [ $exit_code -ne 0 ]; then
        log_error "Tests failed with exit code $exit_code"
        
        # Export logs on failure
        if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
            log_info "Exporting KinD logs..."
            kind export logs /tmp/kind-logs || true
        fi
    fi
    
    # Always cleanup
    log_info "Cleaning up cluster..."
    bash "${SKILL_DIR}/scripts/kind-cleanup.sh" "$CLUSTER_NAME" || true
    
    return $exit_code
}

trap cleanup_on_exit EXIT

log_info "Kubernetes Integration Tests"
log_info "Project: $PROJECT_ROOT"
log_info "Cluster: $CLUSTER_NAME"
log_info "Test pattern: $TEST_PATTERN"
log_info "Timeout: $TIMEOUT"
log_info ""

# Setup cluster
log_info "Setting up KinD cluster..."
bash "${SKILL_DIR}/scripts/kind-setup.sh" "$CLUSTER_NAME" "${PROJECT_ROOT}/test/integration/scripts/kind-config.yaml" || {
    log_error "Failed to setup KinD cluster"
    exit 1
}

log_success "KinD cluster is ready"

# Run tests
log_info "Running K8s integration tests..."
cd "$PROJECT_ROOT"

if go test -v -timeout "$TIMEOUT" -run "$TEST_PATTERN" ./test/integration/k8s_basic_test.go; then
    log_success "All tests passed!"
else
    log_error "Tests failed"
    exit 1
fi
