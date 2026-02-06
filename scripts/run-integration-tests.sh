#!/bin/bash
# Optimized Integration Test Runner for platform-go
# Supports database and Kubernetes testing with proper cleanup
# Usage: ./scripts/run-integration-tests.sh [db|k8s|all] [local|docker]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
SKILLS_DIR="${PROJECT_ROOT}/.github/skills"

TEST_TYPE="${1:-all}"      # db, k8s, or all
RUNNER="${2:-docker}"      # docker or local
TIMEOUT="${3:-30m}"        # Test timeout

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log_info() {
    echo -e "${CYAN}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

log_section() {
    echo ""
    echo -e "${CYAN}════════════════════════════════════════${NC}"
    echo -e "${CYAN}$*${NC}"
    echo -e "${CYAN}════════════════════════════════════════${NC}"
    echo ""
}

print_usage() {
    cat << EOF
${CYAN}Platform-Go Integration Test Runner${NC}

Usage: $0 [TEST_TYPE] [RUNNER] [TIMEOUT]

TEST_TYPE (default: all):
  ${YELLOW}db${NC}     - Run database integration tests only
  ${YELLOW}k8s${NC}    - Run Kubernetes integration tests only
  ${YELLOW}all${NC}    - Run all integration tests

RUNNER (default: docker):
  ${YELLOW}docker${NC}  - Run in Docker containers (isolated environment)
  ${YELLOW}local${NC}   - Run directly on host machine

TIMEOUT (default: 30m):
  Go test timeout duration (e.g., 15m, 1h)

Examples:
  $0 db local          # Run DB tests locally
  $0 k8s docker        # Run K8s tests in Docker
  $0 all docker 1h     # Run all tests in Docker with 1 hour timeout
EOF
}

# Validate inputs
validate_inputs() {
    if [[ ! "$TEST_TYPE" =~ ^(db|k8s|all)$ ]]; then
        log_error "Invalid TEST_TYPE: $TEST_TYPE"
        print_usage
        exit 1
    fi

    if [[ ! "$RUNNER" =~ ^(docker|local)$ ]]; then
        log_error "Invalid RUNNER: $RUNNER"
        print_usage
        exit 1
    fi
}

# Setup database for local testing
setup_postgres_local() {
    log_info "Setting up PostgreSQL for local testing..."
    
    if ! command -v psql &> /dev/null; then
        log_error "PostgreSQL client not found. Install it or use 'docker' runner."
        exit 1
    fi

    DB_NAME="platform_test"
    DB_USER="${DB_USER:-postgres}"
    DB_PASSWORD="${DB_PASSWORD:-}"
    
    # Drop and create fresh test database
    if [ -n "$DB_PASSWORD" ]; then
        PGPASSWORD="$DB_PASSWORD" psql -U "$DB_USER" -tc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" | grep -q 1 && \
            PGPASSWORD="$DB_PASSWORD" psql -U "$DB_USER" -c "DROP DATABASE \"$DB_NAME\";" || true
        
        PGPASSWORD="$DB_PASSWORD" psql -U "$DB_USER" -c "CREATE DATABASE \"$DB_NAME\";"
    else
        psql -U "$DB_USER" -tc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" | grep -q 1 && \
            psql -U "$DB_USER" -c "DROP DATABASE \"$DB_NAME\";" || true
        
        psql -U "$DB_USER" -c "CREATE DATABASE \"$DB_NAME\";"
    fi
    
    log_success "PostgreSQL test database created: $DB_NAME"
}

# Setup test environment variables
setup_test_env() {
    if [ "$RUNNER" = "local" ]; then
        DB_PASSWORD="${DB_PASSWORD:-postgres}"
        export DATABASE_URL="postgres://${DB_USER:-postgres}:${DB_PASSWORD}@localhost:5432/platform_test"
        export REDIS_URL="redis://localhost:6379/0"
    else
        # Docker will provide these via compose
        export DATABASE_URL="postgres://testuser:testpass@postgres:5432/platform_test"
        export REDIS_URL="redis://redis:6379/0"
    fi
    
    export ENVIRONMENT="test"
}

# Run database integration tests (local)
run_db_tests_local() {
    log_section "Running Database Integration Tests (Local)"
    
    setup_postgres_local
    setup_test_env
    
    cd "$PROJECT_ROOT"
    go test -v -timeout "$TIMEOUT" -tags=integration ./test/integration/... \
        -run "^Test(ConfigFile|Group|Project|User)" 2>&1 | tee integration-test.log
    
    log_success "Database tests completed"
}

# Run database integration tests (Docker)
run_db_tests_docker() {
    log_section "Running Database Integration Tests (Docker)"
    
    COMPOSE_FILE="$PROJECT_ROOT/docker-compose.integration.yml"
    
    if [ ! -f "$COMPOSE_FILE" ]; then
        log_error "Docker Compose file not found: $COMPOSE_FILE"
        exit 1
    fi
    
    # Check if docker and docker compose are available
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi
    
    # Stop any existing containers
    log_info "Cleaning up existing containers..."
    docker compose -f "$COMPOSE_FILE" down -v 2>/dev/null || true
    
    # Start services
    log_info "Starting PostgreSQL and Redis services..."
    if ! docker compose -f "$COMPOSE_FILE" up -d 2>&1; then
        log_error "Failed to start Docker services"
        exit 1
    fi
    
    # Wait for services to be ready
    log_info "Waiting for services to be ready..."
    sleep 10
    
    # Verify services are running
    if ! docker compose -f "$COMPOSE_FILE" ps | grep -q "Up"; then
        log_error "Services failed to start"
        docker compose -f "$COMPOSE_FILE" logs
        docker compose -f "$COMPOSE_FILE" down -v || true
        exit 1
    fi
    
    # Run tests
    cd "$PROJECT_ROOT"
    export DATABASE_URL="postgres://testuser:testpass@localhost:5433/platform_test"
    export REDIS_URL="redis://localhost:6380/0"
    export ENVIRONMENT="test"
    
    log_info "Running Go integration tests..."
    go test -v -timeout "$TIMEOUT" -tags=integration ./test/integration/... 2>&1 | tee integration-test.log
    TEST_RESULT=$?
    
    # Cleanup
    log_info "Stopping and removing Docker services..."
    docker compose -f "$COMPOSE_FILE" down -v || true
    
    if [ $TEST_RESULT -ne 0 ]; then
        log_error "Tests failed with exit code $TEST_RESULT"
        exit 1
    fi
    
    log_success "Database tests completed"
}

# Run Kubernetes integration tests (local)
run_k8s_tests_local() {
    log_section "Running Kubernetes Integration Tests (Local)"
    
    setup_test_env
    
    if ! command -v kind &> /dev/null && ! command -v kubectl &> /dev/null; then
        log_error "KinD or kubectl not found. Install them or use 'docker' runner."
        exit 1
    fi
    
    cd "$PROJECT_ROOT"
    go test -v -timeout "$TIMEOUT" -tags=integration ./test/integration/... \
        -run "^TestK8s" 2>&1 | tee integration-k8s-test.log
    
    log_success "Kubernetes tests completed"
}

# Run Kubernetes integration tests (Docker)
run_k8s_tests_docker() {
    log_section "Running Kubernetes Integration Tests (Docker)"
    
    if [ -x "$SKILLS_DIR/cicd-pipeline-optimization/scripts/docker-k8s-integration-test.sh" ]; then
        bash "$SKILLS_DIR/cicd-pipeline-optimization/scripts/docker-k8s-integration-test.sh"
    else
        log_error "Docker K8s integration test script not found"
        exit 1
    fi
    
    log_success "Kubernetes tests completed"
}

# Cleanup test artifacts
cleanup() {
    log_info "Cleaning up test artifacts..."
    rm -f integration-test.log integration-k8s-test.log
}

# Main execution
main() {
    log_section "Platform-Go Integration Test Runner"
    log_info "Test Type: $TEST_TYPE"
    log_info "Runner: $RUNNER"
    log_info "Timeout: $TIMEOUT"
    log_info "Project Root: $PROJECT_ROOT"
    
    validate_inputs
    
    trap cleanup EXIT
    
    case "$TEST_TYPE" in
        db)
            if [ "$RUNNER" = "local" ]; then
                run_db_tests_local
            else
                run_db_tests_docker
            fi
            ;;
        k8s)
            if [ "$RUNNER" = "local" ]; then
                run_k8s_tests_local
            else
                run_k8s_tests_docker
            fi
            ;;
        all)
            if [ "$RUNNER" = "local" ]; then
                run_db_tests_local
                run_k8s_tests_local
            else
                run_db_tests_docker
                run_k8s_tests_docker
            fi
            ;;
    esac
    
    log_section "Integration Tests Complete"
    log_success "All tests passed!"
}

# Show help if requested
if [ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ]; then
    print_usage
    exit 0
fi

main "$@"
