# API Standards

Standard conventions for API design, response formats, and error handling in the platform-go project.

## Table of Contents

- [Response Format](#response-format)
- [Error Handling](#error-handling)
- [HTTP Status Codes](#http-status-codes)
- [Pagination](#pagination)
- [Authentication](#authentication)
- [Best Practices](#best-practices)

## Response Format

All API responses follow a consistent JSON structure.

### Success Response

```json
{
  "data": {
    "id": 1,
    "name": "example"
  }
}
```

### Error Response

```json
{
  "error": "Resource not found",
  "code": "NOT_FOUND"
}
```

### List Response

```json
{
  "data": [
    {"id": 1, "name": "item1"},
    {"id": 2, "name": "item2"}
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100
  }
}
```

## Error Handling

Errors are returned with appropriate HTTP status codes and descriptive messages.

### Error Structure

- `error`: Human-readable error message
- `code`: Machine-readable error code
- `details`: Optional additional context

### Common Error Codes

| Code | Description |
|------|-------------|
| `VALIDATION_ERROR` | Input validation failed |
| `NOT_FOUND` | Resource does not exist |
| `UNAUTHORIZED` | Authentication required |
| `FORBIDDEN` | Insufficient permissions |
| `CONFLICT` | Resource already exists |
| `INTERNAL_ERROR` | Server error |

## HTTP Status Codes

Standard HTTP status codes used throughout the API.

### Success Codes

- `200 OK` - Request successful
- `201 Created` - Resource created
- `204 No Content` - Delete successful

### Client Error Codes

- `400 Bad Request` - Invalid input
- `401 Unauthorized` - Authentication failed
- `403 Forbidden` - Permission denied
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict

### Server Error Codes

- `500 Internal Server Error` - Unexpected error
- `503 Service Unavailable` - Service down

## Pagination

List endpoints support pagination with query parameters.

### Parameters

- `page` - Page number (default: 1)
- `limit` - Items per page (default: 20, max: 100)

### Example Request

```bash
GET /api/users?page=2&limit=50
```

### Example Response

```json
{
  "data": [...],
  "pagination": {
    "page": 2,
    "limit": 50,
    "total": 150,
    "pages": 3
  }
}
```

## Authentication

API uses token-based authentication.

### Request Header

```
Authorization: Bearer <token>
```

### Token Expiry

- Default token lifetime: 24 hours
- Refresh endpoint: `/api/auth/refresh`

## Best Practices

### Request Guidelines

- Use appropriate HTTP methods (GET, POST, PUT, DELETE)
- Include `Content-Type: application/json` header
- Validate input before sending requests
- Handle rate limiting (429 status code)

### Response Guidelines

- Always check HTTP status code
- Parse error messages for user feedback
- Handle pagination for large datasets
- Implement retry logic for 5xx errors

### Security

- Never expose tokens in URLs
- Use HTTPS in production
- Validate and sanitize all inputs
- Implement request timeouts

## Related Documentation

- [K8s Architecture Analysis](K8S_ARCHITECTURE_ANALYSIS.md)
- [Main README](../README.md)
