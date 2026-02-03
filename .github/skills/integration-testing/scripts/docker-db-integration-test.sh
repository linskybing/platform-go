#!/bin/bash
# Database Integration Test Runner (Docker)
# Reference: integration-testing Skill
# Runs all database integration tests in Docker containers
# Usage: ./docker-db-integration-test.sh [test-pattern]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
PROJECT_ROOT="$(cd "${SKILL_DIR}/../../.." && pwd)"
DOCKER_DIR="${SKILL_DIR}/docker"
DOCKER_COMPOSE_FILE="${DOCKER_DIR}/docker-compose.test.yml"

TEST_PATTERN="${1:-.}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

log_info "Database Integration Tests (Docker)"
log_info "Project: ${PROJECT_ROOT}"
log_info "Test pattern: ${TEST_PATTERN}"
log_info ""

cd "$PROJECT_ROOT"

log_info "Building test image..."
docker compose -f "${DOCKER_COMPOSE_FILE}" build db-tests

# Cleanup function
cleanup() {
    local exit_code=$?
    
    log_info "Stopping test containers..."
    docker compose -f "${DOCKER_COMPOSE_FILE}" down -v 2>/dev/null || true
    
    if [ $exit_code -eq 0 ]; then
        log_success "Database tests completed successfully!"
    else
        log_error "Database tests failed with exit code: $exit_code"
        log_info "View logs with:"
        log_info "  docker compose -f ${DOCKER_COMPOSE_FILE} logs"
    fi
    
    return $exit_code
}

trap cleanup EXIT

# Start PostgreSQL and Redis
log_info "Starting PostgreSQL and Redis..."
docker compose -f "${DOCKER_COMPOSE_FILE}" up -d postgres redis

# Wait for health checks
log_info "Waiting for services to be healthy..."

MAX_WAIT=60
WAITED=0

while [ $WAITED -lt $MAX_WAIT ]; do
    POSTGRES_HEALTHY=$(docker compose -f "${DOCKER_COMPOSE_FILE}" ps -q postgres | xargs docker inspect -f '{{.State.Health.Status}}' 2>/dev/null || echo "starting")
    REDIS_HEALTHY=$(docker compose -f "${DOCKER_COMPOSE_FILE}" ps -q redis | xargs docker inspect -f '{{.State.Health.Status}}' 2>/dev/null || echo "starting")
    
    if [ "$POSTGRES_HEALTHY" = "healthy" ] && [ "$REDIS_HEALTHY" = "healthy" ]; then
        log_success "All services are healthy!"
        break
    fi
    
    log_info "PostgreSQL: $POSTGRES_HEALTHY, Redis: $REDIS_HEALTHY (${WAITED}s/${MAX_WAIT}s)"
    sleep 5
    WAITED=$((WAITED + 5))
done

if [ $WAITED -ge $MAX_WAIT ]; then
    log_error "Services failed to become healthy within ${MAX_WAIT} seconds"
    log_error "PostgreSQL logs:"
    docker compose -f "${DOCKER_COMPOSE_FILE}" logs postgres
    log_error "Redis logs:"
    docker compose -f "${DOCKER_COMPOSE_FILE}" logs redis
    exit 1
fi

# Run database integration tests
log_info "Running database integration tests..."
log_info "Command: go test -v -timeout 10m -tags=integration -run ${TEST_PATTERN} ./test/integration/..."
echo ""

# Build and run test container
docker compose -f "${DOCKER_COMPOSE_FILE}" run --rm \
    -e DATABASE_URL="postgres://testuser:testpass@postgres:5432/testdb?sslmode=disable" \
    -e REDIS_URL="redis://redis:6379" \
    -e TEST_ENV="integration" \
    db-tests \
    bash -c "
        set -e
        echo '[INFO] Verifying database connection...'
        until pg_isready -h postgres -p 5432 -U testuser 2>/dev/null; do
            echo '[INFO] Waiting for PostgreSQL...'
            sleep 2
        done
        echo '[SUCCESS] PostgreSQL is ready'
        
        echo '[INFO] Verifying Redis connection...'
        until redis-cli -h redis -p 6379 ping 2>/dev/null | grep -q PONG; do
            echo '[INFO] Waiting for Redis...'
            sleep 2
        done
        echo '[SUCCESS] Redis is ready'
        
        echo ''
        echo '[INFO] Running integration tests...'
        go test -v -timeout 10m -tags=integration -run '${TEST_PATTERN}' ./test/integration/...
    "

TEST_EXIT=$?

if [ $TEST_EXIT -eq 0 ]; then
    log_success "All database integration tests PASSED!"
else
    log_error "Database integration tests FAILED!"
    exit $TEST_EXIT
fi
