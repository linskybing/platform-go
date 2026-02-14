# Frontend Integration Guide

This guide provides the necessary technical details for frontend developers to integrate with the Platform Go API.

## Unified Response Structure

All API responses follow a consistent format:

```json
{
  "code": 200,
  "message": "Operation successful",
  "data": { ... }
}
```

- **code**: HTTP status code (e.g., 200, 201, 400, 401, 403, 404, 500).
- **message**: Human-readable status or error message.
- **data**: The actual payload. It can be an object, an array, or `null`.

## Authentication

The platform uses JWT for authentication.

- **Login**: `POST /auth/login` (accepts `username` and `password` as form-data).
- **Token Storage**: The server sets a `token` cookie (HttpOnly, Secure in production). The response body also includes the token in `data.token` for flexible usage.
- **Request Header**: For subsequent requests, include the token in the header:
  `Authorization: Bearer <your_token>`

## Resource Identification (UUID)

All entities use **UUID v4** as their primary identifier.
- `user_id`, `group_id`, `project_id`, `job_id`, etc., are all 36-character UUID strings (e.g., `550e8400-e29b-41d4-a716-446655440000`).
- **Recommendation**: Always treat IDs as opaque strings. Do not try to parse them as integers.

## Hierarchical Projects

Projects are organized in a tree structure using PostgreSQL `ltree`.

- **Path**: Each project has a `path` field (e.g., `550e8400_e29b...`). This represents its location in the hierarchy.
- **Separators**: Components are separated by dots (`.`).
- **Sanitization**: UUID hyphens are replaced with underscores (`_`) in the `ltree` path to comply with database syntax rules.
- **Interpretation**: A project with path `A.B` is a sub-project of project `A`.

## Configuration & Versioning (Git-like)

We use Content-Addressable Storage for configuration files (YAML).

1. **Save Config**: `POST /configfiles`
   - Send `raw_yaml` and `project_id`.
   - The backend deduplicates content. If the YAML already exists, it creates a new `commit` pointing to the existing `blob`.
   - Returns a `config_commit` object containing the new `id`.
2. **Submit Job**: `POST /jobs/submit`
   - You must provide a `config_commit_id`. This ensures the job runs with a specific, immutable version of the configuration.

## Resource Planning (Schedule Windows)

Resource allocations can be constrained by time windows.

- **Week Window**: Represented as a recurring weekly range (0 to 604800 seconds).
- **Validation**: The backend automatically rejects job submissions if the project is outside its allowed time window or if quotas are exceeded.

## Job Lifecycle

### Statuses
- `PENDING`: Job is in the queue, waiting for resources.
- `RUNNING`: Job is currently executing on a node.
- `COMPLETED`: Job finished successfully.
- `FAILED`: Job encountered an error.
- `PREEMPTED`: Job was stopped by the system to free up resources for a higher-priority task.

### Monitoring
- **List Jobs**: `GET /jobs` (Filters available via query params).
- **Cancel**: `POST /jobs/:id/cancel`.

## Error Handling

Frontend should check the `code` field in the response.
- `400`: Validation error or business logic violation (e.g., "max concurrent jobs exceeded").
- `401`: Token expired or missing.
- `403`: Permission denied (RBAC).
- `404`: Resource not found.
- `500`: Internal server error.
- **Database Conflicts**: If the backend returns an error related to "exclusion violation", it usually means a schedule conflict occurred.
