#!/bin/bash
# KinD Cluster Cleanup Script
# Reference: cicd-pipeline-optimization Skill
# Destroys Kubernetes in Docker cluster and cleans up resources
# Usage: ./kind-cleanup.sh [cluster-name]

set -euo pipefail

CLUSTER_NAME="${1:-platform-test}"

log_info() {
    echo "[INFO] $*"
}

log_error() {
    echo "[ERROR] $*" >&2
}

log_info "Cleaning up KinD cluster: $CLUSTER_NAME"

# Delete cluster if exists
if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
    log_info "Deleting cluster..."
    kind delete cluster --name "$CLUSTER_NAME" || {
        log_error "Failed to delete cluster"
        exit 1
    }
    log_info "Cluster deleted successfully"
else
    log_info "Cluster $CLUSTER_NAME not found"
fi

# Clean up Docker containers
log_info "Cleaning up Docker resources..."
if docker ps -a --filter "name=kind-${CLUSTER_NAME}" --format '{{.ID}}' | grep -q .; then
    docker ps -a --filter "name=kind-${CLUSTER_NAME}" --format '{{.ID}}' | xargs -r docker rm -f || true
    log_info "Docker containers cleaned"
fi

log_info "Cleanup complete!"
