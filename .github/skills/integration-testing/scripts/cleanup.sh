#!/bin/bash
# Cleanup Integration Test Resources
# Reference: integration-testing Skill
# Cleans up all Docker containers, networks, and KinD clusters from integration tests
# Usage: ./cleanup.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
PROJECT_ROOT="$(cd "${SKILL_DIR}/../../.." && pwd)"
DOCKER_DIR="${SKILL_DIR}/docker"
DOCKER_COMPOSE_FILE="${DOCKER_DIR}/docker-compose.test.yml"
CICD_SKILL_DIR="${SKILL_DIR}/../cicd-pipeline-optimization"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
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

log_info "Cleaning up integration test resources..."
echo ""

# Stop docker compose services
log_info "Stopping docker compose services..."
if [ -f "${DOCKER_COMPOSE_FILE}" ]; then
    docker compose -f "${DOCKER_COMPOSE_FILE}" down -v 2>/dev/null || true
    log_success "Docker Compose services stopped"
else
    log_info "docker-compose.test.yml not found, skipping"
fi

# Cleanup KinD clusters
log_info "Cleaning up KinD clusters..."
if command -v kind &> /dev/null; then
    # Get all KinD clusters
    CLUSTERS=$(kind get clusters 2>/dev/null || echo "")
    
    if [ -n "$CLUSTERS" ]; then
        echo "$CLUSTERS" | while read -r cluster; do
            if [[ "$cluster" == platform-* ]]; then
                log_info "Deleting KinD cluster: $cluster"
                kind delete cluster --name "$cluster" 2>/dev/null || true
            fi
        done
        log_success "KinD clusters cleaned up"
    else
        log_info "No KinD clusters found"
    fi
else
    log_info "KinD not installed, skipping cluster cleanup"
fi

# Cleanup docker networks
log_info "Cleaning up Docker networks..."
NETWORKS=$(docker network ls --filter name=platform-go --format "{{.Name}}" 2>/dev/null || echo "")
if [ -n "$NETWORKS" ]; then
    echo "$NETWORKS" | while read -r network; do
        log_info "Removing Docker network: $network"
        docker network rm "$network" 2>/dev/null || true
    done
    log_success "Docker networks cleaned up"
else
    log_info "No matching Docker networks found"
fi

# Cleanup any dangling volumes
log_info "Cleaning up dangling Docker volumes..."
docker volume prune -f &>/dev/null || true
log_success "Dangling volumes cleaned up"

# Cleanup test artifacts
log_info "Cleaning up test artifacts..."
rm -rf /tmp/kind-logs 2>/dev/null || true
rm -f /tmp/kubeconfig* 2>/dev/null || true
rm -f "${PROJECT_ROOT}/.env.tmp" 2>/dev/null || true
log_success "Test artifacts cleaned up"

echo ""
log_success "All integration test resources cleaned up!"
