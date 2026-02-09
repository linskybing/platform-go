# pkg/

Public library code that can be imported by external projects.

## Table of Contents

1. [Overview](#overview)
2. [Structure](#structure)
3. [Packages](#packages)
   - [K8s Package](#k8s-package)
   - [Logger Package](#logger-package)
   - [Cache Package](#cache-package)
4. [GPU Quota / Notes](#gpu-quota--notes)

---

## Overview

Contains reusable utilities and helpers that could be shared with other projects. These are infrastructure-level utilities, not business logic.

## Structure

```
pkg/
├─ cache/      - Redis caching utilities
├─ k8s/        - Kubernetes client utilities
├─ logger/     - Logging utilities
├─ mps/        - (deprecated) MPS GPU sharing helpers (see notes)
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

### GPU / MPS Notes

Historically this repo included an `mps` package for managing NVIDIA MPS (Multi-Process Service) GPU sharing. The codebase has been refactored to remove operational MPS control logic while preserving GPU quota concepts and helpers used by higher-level services.

- GPU quota and conversion helpers remain (see `pkg/gpu` or related helpers).
- Operational MPS control (starting/stopping MPS on nodes) has been deprecated/removed from the core codebase.
- The `pkg/mps` folder may contain legacy helpers and documentation; treat it as deprecated.

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

## GPU Quota / Notes

This project models GPU allocation at the project level. Where MPS concepts were used historically, the system now prefers a simpler GPU quota approach:

- GPU quotas are represented as numeric GPU counts on projects.
- Optional helper functions map quota values to equivalent MPS-like units for display or compatibility.
- If you need to reintroduce operational MPS control, consider implementing it as an external operator or separate service rather than embedding node-level MPS control here.
