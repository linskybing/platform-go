# internal/domain/

Domain entities and business logic organized by aggregate.

## Table of Contents

1. [Overview](#overview)
2. [Structure](#structure)
3. [Aggregates](#aggregates)
4. [File Guidelines](#file-guidelines)
5. [Design Principles](#design-principles)

---

## Overview

Domain-driven design layer containing core business entities and logic. Each aggregate is self-contained and has clear responsibilities.

## Structure

```
internal/domain/
├─ audit/      - Audit log entities
├─ configfile/ - Configuration file entities
├─ form/       - Form entities
├─ group/      - User group entities
├─ image/      - Container image entities
├─ job/        - Job entities and execution logic
├─ project/    - Project entities and management
├─ resource/   - Resource quota entities
├─ storage/    - Storage and PVC entities
├─ user/       - User entities and authentication
└─ view/       - Database view models
```

## Aggregates

### Job

Job entities, types, and domain logic.

- Job execution models
- Job status and lifecycle
- Job-related domain logic

### Project

Project entities and membership management.

- Project definition and configuration
- User-project relationships
- Project quotas and limits

### User

User entities and authentication.

- User account information
- Password and credential management
- User roles and permissions

### Group

User group entities and management.

- Group definition
- Group membership
- Group permissions

### Resource

Resource entities and quotas.

- Resource types (CPU, Memory, GPU, Storage)
- Quota definitions
- Resource allocation tracking

---

## File Guidelines

Each domain folder contains:

- `model.go` - Entity definitions
- `types.go` - Enums and type definitions
- `repository.go` - Repository interface
- `service.go` - Domain service interface (optional)

Standards:

- Keep files under 100 lines each
- Use English comments only
- No infrastructure-specific code
- No direct database queries
- Use repository pattern for data access

## Design Principles

- **Pure domain logic** - No HTTP, database, or framework code
- **Encapsulation** - Hide implementation details
- **Clear boundaries** - Each aggregate has explicit relationships
- **Immutability** - Prefer immutable value objects
- **Explicit types** - Use domain-specific types
- **No side effects** - Pure functions where possible
