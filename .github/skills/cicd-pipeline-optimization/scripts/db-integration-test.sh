#!/bin/bash
# Database Integration Test Runner
# Reference: cicd-pipeline-optimization Skill
# Runs database integration tests with PostgreSQL and Redis
# Usage: ./db-integration-test.sh [test-pattern] [timeout]

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

log_info "Database Integration Tests"
log_info "Project: $PROJECT_ROOT"
log_info "Test pattern: $TEST_PATTERN"
log_info "Timeout: $TIMEOUT"
log_info ""

cd "$PROJECT_ROOT"

# Check if docker-compose services are running
if ! command -v docker-compose &> /dev/null; then
    log_error "docker-compose not found. Required for DB integration tests."
    exit 1
fi

log_info "Checking PostgreSQL connection..."
until pg_isready -h localhost -p 5432 -U testuser &> /dev/null; do
    log_info "Waiting for PostgreSQL..."
    sleep 2
done
log_success "PostgreSQL is ready"

log_info "Checking Redis connection..."
until redis-cli -h localhost -p 6379 ping &> /dev/null; do
    log_info "Waiting for Redis..."
    sleep 2
done
log_success "Redis is ready"

# Run tests
log_info "Running database integration tests..."
if go test -v -timeout "$TIMEOUT" -tags=integration -run "$TEST_PATTERN" ./test/integration/...
then
    log_success "All tests passed!"
else
    log_error "Tests failed"
    exit 1
fi
