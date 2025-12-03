#!/bin/bash

# Navigate to the project directory
cd "$(dirname "$0")/.."

# Build the Docker image
docker build -t k8s-project:latest .

# Optionally, you can push the image to a Docker registry
# docker push k8s-project:latest

echo "Docker image built successfully: k8s-project:latest"