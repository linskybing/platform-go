# internal/scheduler/

Job scheduling system with support for multiple job types.

## Table of Contents

1. [Overview](#overview)
2. [Structure](#structure)
3. [Features](#features)
4. [Job Types](#job-types)
5. [Priority Levels](#priority-levels)
6. [Job Lifecycle](#job-lifecycle)

---

## Overview

Implements a priority-based job scheduling system supporting multiple job types with resource availability checking and lifecycle management.

## Structure

```
internal/scheduler/
├─ queue/    - Priority queue management
├─ executor/ - Job execution logic for different job types
└─ mpi/      - MPI-specific job handling
```

## Features

- Priority-based scheduling (interactive > batch)
- Extensible job type system (Normal, MPI, GPU)
- Resource availability checking
- Job lifecycle management
- MPI job coordination with OpenMPI

## Job Types

### Normal

Standard containerized jobs.

- Single container execution
- Standard resource allocation
- Container-based isolation

### MPI

Distributed computing jobs using OpenMPI.

- Multi-process execution
- Inter-process communication
- Distributed resource coordination

### GPU

Jobs requiring dedicated GPU resources.

- GPU resource allocation
- NVIDIA runtime integration
- GPU memory management

## Priority Levels

- Interactive workloads: 1000 (higher priority)
- Batch jobs: 100 (can be preempted)

## Job Lifecycle

```
Queued → Scheduling → Running → Completed/Failed
```
