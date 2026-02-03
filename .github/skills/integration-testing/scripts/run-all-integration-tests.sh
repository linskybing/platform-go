#!/bin/bash
# Master Integration Test Runner
# Reference: integration-testing Skill
# Runs all integration tests in Docker containers
# Usage: ./run-all-integration-tests.sh [all|db|k8s]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
PROJECT_ROOT="$(cd "${SKILL_DIR}/../../.." && pwd)"

TEST_SUITE="${1:-all}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $*"
}

print_banner() {
    echo "======================================================================"
    echo "  Integration Test Suite - platform-go"
    echo "======================================================================"
    echo ""
}

print_usage() {
    echo "Usage: $0 [test-suite]"
    echo ""
    echo "test-suite:"
    echo "  all  - Run all integration tests (default)"
    echo "  db   - Run database integration tests only"
    echo "  k8s  - Run Kubernetes integration tests only"
    echo ""
    echo "Examples:"
    echo "  $0           # Run all tests"
    echo "  $0 db        # Run only database tests"
    echo "  $0 k8s       # Run only K8s tests"
    echo ""
}

# Validate input
if [ "$TEST_SUITE" != "all" ] && [ "$TEST_SUITE" != "db" ] && [ "$TEST_SUITE" != "k8s" ]; then
    log_error "Invalid test suite: $TEST_SUITE"
    print_usage
    exit 1
fi

# Check Docker is available
if ! command -v docker &> /dev/null; then
    log_error "Docker is not installed. Please install Docker first."
    exit 1
fi

if ! command -v docker compose &> /dev/null; then
    log_error "docker compose is not installed. Please install docker compose first."
    exit 1
fi

# Check Docker daemon is running
if ! docker info &> /dev/null; then
    log_error "Docker daemon is not running. Please start Docker."
    exit 1
fi

print_banner

log_info "Test suite: ${TEST_SUITE}"
log_info "Project root: ${PROJECT_ROOT}"
log_info "Skill directory: ${SKILL_DIR}"
log_info ""

# Change to project root
cd "$PROJECT_ROOT"

# Cleanup function
cleanup() {
    local exit_code=$?
    
    if [ $exit_code -eq 0 ]; then
        log_success "All tests completed successfully!"
    else
        log_error "Tests failed with exit code: $exit_code"
    fi
    
    log_info "Cleaning up test resources..."
    bash "${SKILL_DIR}/scripts/cleanup.sh" || true
    
    return $exit_code
}

trap cleanup EXIT

# Run database integration tests
run_db_tests() {
    log_info "========================================"
    log_info "Starting Database Integration Tests"
    log_info "========================================"
    echo ""
    
    chmod +x "${SKILL_DIR}/scripts/docker-db-integration-test.sh"
    
    if bash "${SKILL_DIR}/scripts/docker-db-integration-test.sh"; then
        log_success "Database integration tests PASSED"
        return 0
    else
        log_error "Database integration tests FAILED"
        return 1
    fi
}

# Run Kubernetes integration tests
run_k8s_tests() {
    log_info "========================================"
    log_info "Starting Kubernetes Integration Tests"
    log_info "========================================"
    echo ""
    
    chmod +x "${SKILL_DIR}/scripts/docker-k8s-integration-test.sh"
    
    if bash "${SKILL_DIR}/scripts/docker-k8s-integration-test.sh"; then
        log_success "Kubernetes integration tests PASSED"
        return 0
    else
        log_error "Kubernetes integration tests FAILED"
        return 1
    fi
}

# Execute tests based on suite selection
case "$TEST_SUITE" in
    all)
        DB_EXIT=0
        K8S_EXIT=0
        
        run_db_tests || DB_EXIT=$?
        echo ""
        
        run_k8s_tests || K8S_EXIT=$?
        echo ""
        
        if [ $DB_EXIT -ne 0 ] || [ $K8S_EXIT -ne 0 ]; then
            log_error "Some tests failed:"
            [ $DB_EXIT -ne 0 ] && log_error "  - Database tests: FAILED"
            [ $K8S_EXIT -ne 0 ] && log_error "  - Kubernetes tests: FAILED"
            exit 1
        fi
        ;;
    
    db)
        run_db_tests
        ;;
    
    k8s)
        run_k8s_tests
        ;;
esac

log_success "Test suite '${TEST_SUITE}' completed successfully!"
