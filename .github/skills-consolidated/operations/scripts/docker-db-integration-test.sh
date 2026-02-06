#!/bin/bash
# Docker-based Database Integration Test Runner
# Reference: cicd-pipeline-optimization Skill
# Runs database integration tests using docker-compose
# Usage: ./docker-db-integration-test.sh [test-pattern] [timeout]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../../.." && pwd)"

TEST_PATTERN="${1:-.}"
TIMEOUT="${2:-10m}"

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
    log_info "Stopping Docker services..."
    docker-compose -f docker-compose.integration.yml down || true
    return $exit_code
}

trap cleanup EXIT

log_info "Database Integration Tests (Docker Compose)"
log_info "Project: $PROJECT_ROOT"
log_info "Test pattern: $TEST_PATTERN"
log_info "Timeout: $TIMEOUT"
log_info ""

cd "$PROJECT_ROOT"

# Check if Docker and docker-compose are available
if ! command -v docker &> /dev/null; then
    log_error "Docker is not installed or not in PATH"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    log_error "docker-compose is not installed or not in PATH"
    exit 1
fi

# Start services and run tests
log_info "Starting PostgreSQL and Redis services..."
if ! docker-compose -f docker-compose.integration.yml up -d postgres redis; then
    log_error "Failed to start database services"
    exit 1
fi

log_success "Database services started"

# Wait for services to be ready
log_info "Waiting for services to be ready..."
sleep 10

# Run tests
log_info "Running database integration tests..."
if docker-compose -f docker-compose.integration.yml run \
    --rm \
    -e "TEST_PATTERN=$TEST_PATTERN" \
    -e "TIMEOUT=$TIMEOUT" \
    db-integration-tests; then
    log_success "All database integration tests passed!"
else
    log_error "Database integration tests failed"
    
    # Show logs on failure
    log_info "PostgreSQL logs:"
    docker-compose -f docker-compose.integration.yml logs postgres || true
    
    log_info "Redis logs:"
    docker-compose -f docker-compose.integration.yml logs redis || true
    
    exit 1
fi
