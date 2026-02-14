# Platform Go

High-concurrency, multi-tenant platform for managing cloud resources and jobs on Kubernetes.

## Core Features

- **Multi-Tenant Architecture**: Robust isolation using PostgreSQL Class Table Inheritance and Kubernetes namespaces.
- **Hierarchical Projects**: Flexible project topology managed with PostgreSQL `ltree`.
- **Git-like Configuration**: Deduplicated, version-controlled configuration storage.
- **Advanced Scheduling**: Time-window based resource planning with PostgreSQL exclusion constraints.
- **Preemptive Job Queue**: High-performance job queue with SQL-based victim selection for resource preemption.
