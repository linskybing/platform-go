# internal/priority/

Resource priority and preemption management.

## Table of Contents

1. [Overview](#overview)
2. [Structure](#structure)
3. [Preemption Flow](#preemption-flow)
4. [Priority Levels](#priority-levels)

---

## Overview

Implements priority-based resource allocation where course workloads have higher priority than batch jobs. When resource contention occurs, lower-priority batch jobs are preempted to free resources for higher-priority course workloads.

## Structure

```
internal/priority/
├─ monitor/   - Resource usage monitoring
└─ preemptor/ - Job preemption logic
```

## Preemption Flow

1. Monitor detects resource shortage
2. Course workload requires resources
3. Preemptor selects low-priority jobs to terminate
4. Graceful termination with 120s grace period
5. Create checkpoint if job supports it
6. Clean up resources
7. Allow course workload to start

## Priority Levels

- Course (Pods): Priority 1000
- Batch Jobs: Priority 100
