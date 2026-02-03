# pkg/

Public library code that can be imported by external projects.

## Table of Contents

1. [Overview](#overview)
2. [Structure](#structure)
3. [Packages](#packages)
   - [K8s Package](#k8s-package)
   - [MPS Package](#mps-package)
   - [Logger Package](#logger-package)
   - [Cache Package](#cache-package)
4. [MPS GPU Sharing](#mps-gpu-sharing)

---

## Overview

Contains reusable utilities and helpers that could be shared with other projects. These are infrastructure-level utilities, not business logic.

## Structure

```
pkg/
├─ cache/      - Redis caching utilities
├─ k8s/        - Kubernetes client utilities
├─ logger/     - Logging utilities
├─ mps/        - MPS GPU sharing management
├─ middleware/ - HTTP middleware
├─ response/   - Response formatting
├─ storage/    - Storage client utilities
├─ types/      - Type definitions
├─ utils/      - Utility functions
└─ errors/     - Error handling
```

## Packages

### K8s Package

Kubernetes client utilities and helpers.

- Client initialization and connection management
- Namespace operations
- Pod management
- PVC (Persistent Volume Claim) operations
- Volume management
- WebSocket support for pod logs

### MPS Package

MPS (Multi-Process Service) GPU sharing management.

- GPU resource calculation
- MPS quota management
- Configuration validation
- Error handling for GPU operations

### Logger Package

Centralized logging utilities.

- Structured logging
- Log level configuration
- Output formatting

### Cache Package

Redis-based distributed caching.

- Async worker pool for non-blocking writes
- Singleflight pattern for cache stampede prevention
- Distributed locking with Lua atomicity
- Key and prefix-based cache invalidation
- Graceful degradation when Redis unavailable

---

## MPS GPU Sharing

MPS (Multi-Process Service) enables efficient GPU sharing:
- 1 dedicated GPU = 10 MPS units
- Project-level MPS limits
- Validation against quotas
- Dynamic allocation tracking
