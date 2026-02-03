#!/bin/bash
# Kubernetes Integration Test Runner (Docker + KinD)
# Reference: integration-testing Skill
# Runs K8s integration tests using KinD cluster
# Usage: ./docker-k8s-integration-test.sh [test-pattern]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
PROJECT_ROOT="$(cd "${SKILL_DIR}/../../.." && pwd)"
CICD_SKILL_DIR="${SKILL_DIR}/../cicd-pipeline-optimization"

TEST_PATTERN="${1:-.}"
CLUSTER_NAME="platform-test"
KEEP_CLUSTER="${KEEP_CLUSTER:-0}"

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

log_info "Kubernetes Integration Tests (KinD)"
log_info "Project: ${PROJECT_ROOT}"
log_info "Cluster: ${CLUSTER_NAME}"
log_info "Test pattern: ${TEST_PATTERN}"
log_info ""

cd "$PROJECT_ROOT"

# Cleanup function
cleanup() {
    local exit_code=$?
    
    if [ $exit_code -eq 0 ]; then
        log_success "Kubernetes tests completed successfully!"
    else
        log_error "Kubernetes tests failed with exit code: $exit_code"
        
        # Export logs on failure
        if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
            log_info "Exporting KinD cluster logs..."
            kind export logs /tmp/kind-logs --name "${CLUSTER_NAME}" 2>/dev/null || true
            log_info "Logs exported to /tmp/kind-logs"
        fi
    fi
    
    if [ "$KEEP_CLUSTER" != "1" ]; then
        log_info "Destroying KinD cluster..."
        bash "${CICD_SKILL_DIR}/scripts/kind-cleanup.sh" "${CLUSTER_NAME}" 2>/dev/null || true
    else
        log_warning "Keeping cluster '${CLUSTER_NAME}' running (KEEP_CLUSTER=1)"
        log_info "To manually cleanup: kind delete cluster --name ${CLUSTER_NAME}"
    fi
    
    return $exit_code
}

trap cleanup EXIT

# Check if Docker is running
if ! docker info &> /dev/null; then
    log_error "Docker daemon is not running. Please start Docker."
    exit 1
fi

# Check if kind is installed
if ! command -v kind &> /dev/null; then
    log_error "KinD is not installed. Installing KinD..."
    
    # Install kind
    curl -sLo /tmp/kind "https://kind.sigs.k8s.io/dl/v0.20.0/kind-$(uname)-amd64"
    chmod +x /tmp/kind
    sudo mv /tmp/kind /usr/local/bin/kind || mv /tmp/kind ~/bin/kind
    
    if ! command -v kind &> /dev/null; then
        log_error "Failed to install KinD"
        exit 1
    fi
    
    log_success "KinD installed successfully"
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    log_error "kubectl is not installed. Installing kubectl..."
    
    # Install kubectl
    KUBECTL_VERSION=$(curl -Ls https://dl.k8s.io/release/stable.txt)
    curl -sLo /tmp/kubectl "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl"
    chmod +x /tmp/kubectl
    sudo mv /tmp/kubectl /usr/local/bin/kubectl || mv /tmp/kubectl ~/bin/kubectl
    
    if ! command -v kubectl &> /dev/null; then
        log_error "Failed to install kubectl"
        exit 1
    fi
    
    log_success "kubectl installed successfully"
fi

# Setup KinD cluster
log_info "Setting up KinD cluster..."

# Check if kind config exists
KIND_CONFIG="${PROJECT_ROOT}/test/integration/scripts/kind-config.yaml"
if [ ! -f "$KIND_CONFIG" ]; then
    log_warning "KinD config not found at ${KIND_CONFIG}, using default config"
    KIND_CONFIG=""
fi

chmod +x "${CICD_SKILL_DIR}/scripts/kind-setup.sh"

if bash "${CICD_SKILL_DIR}/scripts/kind-setup.sh" "${CLUSTER_NAME}" "${KIND_CONFIG}"; then
    log_success "KinD cluster '${CLUSTER_NAME}' is ready"
else
    log_error "Failed to setup KinD cluster"
    exit 1
fi

# Verify cluster is accessible
log_info "Verifying cluster access..."
if kubectl cluster-info --context "kind-${CLUSTER_NAME}" &> /dev/null; then
    log_success "kubectl can access the cluster"
else
    log_error "kubectl cannot access the cluster"
    kubectl cluster-info --context "kind-${CLUSTER_NAME}"
    exit 1
fi

# Run K8s integration tests
log_info "Running Kubernetes integration tests..."
log_info "Command: go test -v -timeout 10m -run ${TEST_PATTERN} ./test/integration/k8s_basic_test.go"
echo ""

# Set KUBECONFIG for tests
export KUBECONFIG="$(kind get kubeconfig --name "${CLUSTER_NAME}" --internal 2>/dev/null || kind get kubeconfig --name "${CLUSTER_NAME}")"

# Run tests
if go test -v -timeout 10m -run "${TEST_PATTERN}" ./test/integration/k8s_basic_test.go; then
    log_success "All Kubernetes integration tests PASSED!"
else
    log_error "Kubernetes integration tests FAILED!"
    exit 1
fi
