# internal/

Private application code that cannot be imported by external projects.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Structure](#structure)
4. [Layers](#layers)

---

## Overview

Contains all core business logic organized by feature and layer. This code is private to the platform-go project.

## Architecture

Follows layered architecture with clear separation of concerns:

```
API Layer (HTTP)
    ↓
Application Layer (Business Logic)
    ↓
Domain Layer (Entities & Business Rules)
    ↓
Repository Layer (Data Access)
    ↓
Database (PostgreSQL)
```

## Structure

```
internal/
├─ api/            - HTTP API layer (routes, handlers, middleware)
├─ application/    - Business logic services
├─ domain/         - Business entities and domain models
├─ repository/     - Data access layer (GORM)
├─ config/         - Configuration management
├─ constants/      - Application constants
├─ cron/           - Scheduled jobs and cron tasks
├─ priority/       - Resource priority and preemption logic
└─ scheduler/      - Job scheduling implementation
```

## Layers

### API Layer

HTTP request handling and routing.

- Route definitions
- Request handlers
- Middleware (authentication, logging, recovery)
- Request validation
- Response formatting

### Application Layer

Business logic and use cases.

- Service implementations
- Business rule enforcement
- Cross-domain orchestration
- Redis cache integration

### Domain Layer

Business entities and core business logic.

- Entity definitions
- Value objects
- Domain-specific types
- Repository interfaces

### Repository Layer

Data persistence abstraction.

- GORM-based database queries
- Entity mapping
- Query abstractions
- Transaction management
