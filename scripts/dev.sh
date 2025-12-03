#!/bin/bash

# Exit on error
set -e

echo "Starting Development Environment..."

# 1. Apply Kubernetes Manifests
echo "Applying K8s manifests..."
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/postgres.yaml
kubectl apply -f k8s/go-api.yaml

# 2. Wait for Pods to be Ready
echo "Waiting for go-api deployment to be ready..."
kubectl rollout status deployment/go-api --timeout=120s

# 3. Get the Pod Name
POD_NAME=$(kubectl get pods -l app=go-api -o jsonpath="{.items[0].metadata.name}")
echo "Found pod: $POD_NAME"

# 4. Run the Application
echo "Starting Go application inside the pod..."
echo "   (Press Ctrl+C to stop the application, the pod will remain running)"
echo "   (Source code is mounted at /go/web-go)"

# We use 'exec -it' to keep the terminal interactive so you can see logs and stop it with Ctrl+C
kubectl exec -it $POD_NAME -- /bin/bash -c "tmux"