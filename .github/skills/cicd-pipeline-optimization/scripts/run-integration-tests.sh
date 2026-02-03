#!/bin/bash
# Unified Integration Test Runner
# Reference: cicd-pipeline-optimization Skill
# Runs integration tests in Docker or locally
# Usage: ./run-integration-tests.sh [k8s|db|all] [docker|local]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../../.." && pwd)"
SKILL_DIR="${PROJECT_ROOT}/.github/skills/cicd-pipeline-optimization"

TEST_TYPE="${1:-all}"  # k8s, db, or all
RUNNER="${2:-docker}"  # docker or local

log_info() {
    echo "[INFO] $*"
}

log_error() {
    echo "[ERROR] $*" >&2
}

log_success() {
    echo "[SUCCESS] $*"
}

print_usage() {
    echo "Usage: $0 [test-type] [runner]"
    echo ""
    echo "test-type:"
    echo "  k8s     - Run Kubernetes integration tests only"
    echo "  db      - Run database integration tests only"
    echo "  all     - Run all integration tests (default)"
    echo ""
    echo "runner:"
    echo "  docker  - Run in Docker containers (default, no host pollution)"
    echo "  local   - Run directly on host machine"
    echo ""
    echo "Examples:"
    echo "  $0 k8s docker        - Run K8s tests in Docker"
    echo "  $0 db local          - Run DB tests locally"
    echo "  $0 all docker        - Run all tests in Docker"
}

# Validate inputs
if [ "$TEST_TYPE" != "k8s" ] && [ "$TEST_TYPE" != "db" ] && [ "$TEST_TYPE" != "all" ]; then
    log_error "Invalid test type: $TEST_TYPE"
    print_usage
    exit 1
fi

if [ "$RUNNER" != "docker" ] && [ "$RUNNER" != "local" ]; then
    log_error "Invalid runner: $RUNNER"
    print_usage
    exit 1
fi

log_info "Integration Test Runner"
log_info "Test type: $TEST_TYPE"
log_info "Runner: $RUNNER"
log_info "Project: $PROJECT_ROOT"
log_info ""

cd "$PROJECT_ROOT"

run_k8s_tests() {
    if [ "$RUNNER" = "docker" ]; then
        log_info "Running K8s tests in Docker container..."
        chmod +x "$SKILL_DIR/scripts/docker-k8s-integration-test.sh"
        "$SKILL_DIR/scripts/docker-k8s-integration-test.sh"
    else
        log_info "Running K8s tests locally..."
        chmod +x "$SKILL_DIR/scripts/k8s-integration-test.sh"
        "$SKILL_DIR/scripts/k8s-integration-test.sh"
    fi
}

run_db_tests() {
    if [ "$RUNNER" = "docker" ]; then
        log_info "Running database tests with Docker Compose..."
        chmod +x "$SKILL_DIR/scripts/docker-db-integration-test.sh"
        "$SKILL_DIR/scripts/docker-db-integration-test.sh"
    else
        log_info "Running database tests locally..."
        chmod +x "$SKILL_DIR/scripts/db-integration-test.sh"
        "$SKILL_DIR/scripts/db-integration-test.sh"
    fi
}

# Run tests
case "$TEST_TYPE" in
    k8s)
        run_k8s_tests
        ;;
    db)
        run_db_tests
        ;;
    all)
        run_db_tests
        log_info ""
        log_info "=========================================="
        log_info ""
        run_k8s_tests
        ;;
esac

log_success "All requested integration tests completed!"
