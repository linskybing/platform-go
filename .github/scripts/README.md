# Platform-Go Scripts Documentation

This directory contains scripts for building, deploying, and managing the platform-go project.
All scripts are organized according to **golang-production-standards** from `.github/skills`.

## Table of Contents

1. [Directory Structure](#directory-structure)
2. [Scripts Overview](#scripts-overview)
   - [Build Scripts](#build-scripts-build)
   - [Deployment Scripts](#deployment-scripts-deploy)
   - [Development Scripts](#development-scripts-dev)
   - [Tool Scripts](#tool-scripts-tools)
3. [Script Standards](#script-standards)
4. [Running Scripts](#running-scripts)
5. [Maintenance Guidelines](#maintenance-guidelines)
6. [Troubleshooting](#troubleshooting)

---

## Directory Structure

```
.github/scripts/
├── build/          # Build-related scripts
├── deploy/         # Deployment-related scripts
├── dev/            # Development environment scripts
└── tools/          # Utility scripts and tools
```

## Scripts Overview

### Build Scripts (`build/`)

#### `images.sh`

- **Purpose**: Builds Go API and Postgres images, pushes to private Harbor registry
- **Usage**: `./build/images.sh`
- **Configuration**: Edit HARBOR_HOST, PROJECT_NAME, TAG as needed
- **Output**: Images pushed to Harbor
- **Production Standard**: Automated build process with clear logging

### Deployment Scripts (`deploy/`)

#### `redeploy.sh`

- **Purpose**: Redeploys application to Kubernetes cluster
- **Usage**: `./deploy/redeploy.sh`
- **Prerequisites**: Kubernetes cluster configured, kubectl available
- **Production Standard**: Safe redeployment with rollout verification

### Development Scripts (`dev/`)

#### `startup.sh`

- **Purpose**: Starts development environment with Kubernetes manifests
- **Usage**: `./dev/startup.sh`
- **Prerequisites**: Local Kubernetes cluster (minikube/kind)
- **Features**: Auto pod discovery, hot reloading, log streaming
- **Production Standard**: Development environment follows production patterns

#### `cleanup.sh`

- **Purpose**: Cleans up development environment
- **Usage**: `./dev/cleanup.sh`
- **Effect**: Removes deployments from local cluster

#### `setup_k8s.sh`

- **Purpose**: Sets up fake/local Kubernetes environment
- **Usage**: `./dev/setup_k8s.sh`
- **Installs**: Minikube and required tools
- **Production Standard**: Reproducible local development setup

### Tool Scripts (`tools/`)

#### `genmock.sh`

- **Purpose**: Installs mockgen for generating test mocks
- **Usage**: `./tools/genmock.sh`
- **Prerequisite**: Go development environment
- **Production Standard**: Automated test setup

#### `create_gpu_pod.py`

- **Purpose**: Creates Kubernetes pods with GPU support (Python)
- **Usage**: `python tools/create_gpu_pod.py`
- **Prerequisites**: Python 3, Kubernetes Python client
- **Production Standard**: Declarative GPU resource management

---

## Script Standards

All scripts follow these production standards:

### 1. Error Handling
- All scripts use `set -e` for fail-fast behavior
- Clear error messages for debugging
- Exit codes properly set

### 2. Logging
- Descriptive progress messages
- Clear section separation
- Timestamps for operations (optional)

### 3. Documentation
- Inline comments explaining complex steps
- Usage examples in headers
- Configuration variables clearly marked

### 4. Modularity
- Each script has single responsibility
- Reusable components extracted
- Minimal dependencies

### 5. Compatibility
- Bash 4.0+ for bash scripts
- Python 3.8+ for Python scripts
- Clear prerequisite listing

## Running Scripts

### From Project Root
```bash
# Build images
bash .github/scripts/build/images.sh

# Deploy application
bash .github/scripts/deploy/redeploy.sh

# Development setup
bash .github/scripts/dev/startup.sh
bash .github/scripts/dev/cleanup.sh
bash .github/scripts/dev/setup_k8s.sh

# Tools
bash .github/scripts/tools/genmock.sh
python .github/scripts/tools/create_gpu_pod.py
```

### From .github/scripts Directory
```bash
cd .github/scripts
bash build/images.sh
bash deploy/redeploy.sh
# etc.
```

## Maintenance Guidelines

1. **Regular Review**: Review scripts quarterly for obsolete dependencies
2. **Documentation**: Keep this README updated with new scripts
3. **Standards**: All new scripts must follow the standards listed above
4. **Testing**: Test scripts in isolated environment before merging
5. **Version Control**: Track script changes in git with clear commit messages

## Related Documentation

- **Build Standards**: See `.github/skills/golang-production-standards/`
- **Production Readiness**: See `.github/skills/production-readiness-checklist/`
- **CI/CD Optimization**: See `.github/skills/cicd-pipeline-optimization/`

## Troubleshooting

### Scripts Not Executable
```bash
chmod +x .github/scripts/*/*.sh
chmod +x .github/scripts/*/*/*.sh
```

### Permission Denied
Ensure you have execute permissions:
```bash
ls -la .github/scripts/build/  # Check permissions
sudo chmod u+x <script-name>   # Add permissions if needed
```

### Missing Dependencies
Each script header lists required tools. Install:
```bash
# For Kubernetes tools
brew install kubectl
brew install minikube

# For Python dependencies
pip install kubernetes

# For Go tools
go install github.com/golang/mock/mockgen@latest
```

## Adding New Scripts

When adding new scripts:

1. Choose appropriate subdirectory (build, deploy, dev, tools)
2. Follow naming convention: `descriptive-name.sh` (kebab-case)
3. Add shebang and error handling at top
4. Document purpose and usage in header
5. Update this README.md with script details
6. Test thoroughly before committing
7. Ensure executable permissions: `chmod +x script.sh`

## Standards Reference

All scripts are organized according to:
- **File Structure Guidelines**: `.github/skills/file-structure-guidelines/`
- **Golang Production Standards**: `.github/skills/golang-production-standards/`
- **CICD Pipeline Optimization**: `.github/skills/cicd-pipeline-optimization/`
- **Production Readiness**: `.github/skills/production-readiness-checklist/`
