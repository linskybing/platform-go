#!/bin/bash
# Quick Integration Test Runner
# Reference: integration-testing Skill
# Simplified script to run tests in Docker

set -euo pipefail

cd "$(dirname "$0")/../../.."

DOCKER_COMPOSE_FILE=".github/skills/integration-testing/docker/docker-compose.test.yml"

echo "=========================================="
echo "Integration Test Quick Runner"
echo "=========================================="
echo ""

# Cleanup first
echo "[1/4] Cleaning up existing containers..."
docker compose -f "$DOCKER_COMPOSE_FILE" down -v 2>/dev/null || true

# Build
echo "[2/4] Building test image..."
docker compose -f "$DOCKER_COMPOSE_FILE" build db-tests

# Start services
echo "[3/4] Starting PostgreSQL and Redis..."
docker compose -f "$DOCKER_COMPOSE_FILE" up -d postgres redis

# Wait for health
echo "[4/4] Waiting for services to be healthy..."
sleep 20

# Run tests
echo ""
echo "==================== RUNNING TESTS ===================="
docker compose -f "$DOCKER_COMPOSE_FILE" run --rm db-tests

# Cleanup
echo ""
echo "==================== CLEANUP ===================="
docker compose -f "$DOCKER_COMPOSE_FILE" down -v

echo ""
echo "=========================================="
echo "Tests Complete!"
echo "=========================================="
