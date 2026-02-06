#!/bin/bash
# KinD Cluster Setup Script
# Reference: cicd-pipeline-optimization Skill
# Creates Kubernetes in Docker cluster for integration testing
# Usage: ./kind-setup.sh [cluster-name] [config-file]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../../.." && pwd)"

CLUSTER_NAME="${1:-platform-test}"
CONFIG_FILE="${2:-${PROJECT_ROOT}/test/integration/scripts/kind-config.yaml}"
K8S_VERSION="${K8S_VERSION:-v1.28.0}"

log_info() {
    echo "[INFO] $*"
}

log_error() {
    echo "[ERROR] $*" >&2
}

log_info "Setting up KinD cluster: $CLUSTER_NAME"

# Check if kind is installed
if ! command -v kind &> /dev/null; then
    log_info "Installing kind..."
    curl -sLo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
    chmod +x ./kind
    sudo mv ./kind /usr/local/bin/kind
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    log_info "Installing kubectl..."
    curl -sLO "https://dl.k8s.io/release/$(curl -Ls https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
    chmod +x kubectl
    sudo mv kubectl /usr/local/bin/kubectl
fi

# Delete existing cluster if it exists
if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
    log_info "Deleting existing cluster: $CLUSTER_NAME"
    kind delete cluster --name "$CLUSTER_NAME" || true
fi

# Create new cluster
log_info "Creating KinD cluster with image kindest/node:${K8S_VERSION}"
log_info "Config file: $CONFIG_FILE"

if [ ! -f "$CONFIG_FILE" ]; then
    log_error "Config file not found: $CONFIG_FILE"
    exit 1
fi

kind create cluster --name "$CLUSTER_NAME" --config "$CONFIG_FILE" --image "kindest/node:${K8S_VERSION}" || {
    log_error "Failed to create KinD cluster"
    exit 1
}

# Wait for cluster to be ready
log_info "Waiting for cluster to be ready..."

# First wait for API server to be responsive
for i in {1..60}; do
    if kubectl cluster-info 2>/dev/null | grep -q "control plane"; then
        log_info "API server is responsive"
        break
    fi
    log_info "Waiting for API server... ($i/60)"
    sleep 1
done

# Then wait for all nodes to be ready
for i in {1..30}; do
    ready_nodes=$(kubectl get nodes --no-headers 2>/dev/null | grep -c "Ready" || echo "0")
    total_nodes=$(kubectl get nodes --no-headers 2>/dev/null | wc -l | tr -d '[:space:]' || echo "0")
    
    # Remove any whitespace/newlines from variables
    ready_nodes=$(echo "$ready_nodes" | tr -d '[:space:]')
    total_nodes=$(echo "$total_nodes" | tr -d '[:space:]')
    
    if [ "$total_nodes" -gt 0 ] 2>/dev/null && [ "$ready_nodes" -eq "$total_nodes" ] 2>/dev/null; then
        log_info "All $total_nodes nodes are ready!"
        break
    fi
    log_info "Waiting for nodes... ($i/30) - Ready: $ready_nodes/$total_nodes"
    sleep 2
done

# Create test namespace
log_info "Creating test namespace..."
kubectl create namespace platform-test --dry-run=client -o yaml | kubectl apply --validate=false -f - || {
    log_error "Failed to create test namespace"
    exit 1
}

# Apply RBAC for tests
log_info "Applying RBAC..."
cat <<EOF | kubectl apply --validate=false -f - || {
    log_error "Failed to apply RBAC"
    exit 1
}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: platform-test-sa
  namespace: platform-test
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: platform-test-role
rules:
- apiGroups: [""]
  resources: ["pods", "persistentvolumeclaims", "namespaces", "services"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["apps"]
  resources: ["deployments", "statefulsets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["batch"]
  resources: ["jobs"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: platform-test-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: platform-test-role
subjects:
- kind: ServiceAccount
  name: platform-test-sa
  namespace: platform-test
EOF
}

log_info "KinD cluster setup complete!"
log_info "Cluster name: $CLUSTER_NAME"
log_info "Kubernetes version: $K8S_VERSION"
kubectl cluster-info --context "kind-${CLUSTER_NAME}" || true
