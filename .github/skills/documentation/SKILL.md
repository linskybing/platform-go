---
name: documentation
description: Markdown documentation standards, file organization, and documentation best practices for platform-go
license: Proprietary
metadata:
  author: platform-go
  version: "1.0"
  consolidated_from:
    - markdown-documentation-standards
---

# Documentation Excellence

Comprehensive guidelines for creating clear, maintainable, and user-friendly documentation in Markdown format.

## Documentation Structure

### Project Documentation Hierarchy
```
docs/
├── README.md                 (Overview & quick start)
├── ARCHITECTURE.md           (System design & diagrams)
├── API.md                    (API reference & examples)
├── INSTALLATION.md           (Setup instructions)
├── DEPLOYMENT.md             (Production deployment)
├── TROUBLESHOOTING.md        (Common issues & solutions)
├── CONTRIBUTING.md           (Developer guidelines)
├── CHANGELOG.md              (Version history)
└── guides/
    ├── authentication.md     (Auth setup & usage)
    ├── caching.md            (Redis caching guide)
    ├── kubernetes.md         (K8s deployment)
    └── monitoring.md         (Observability setup)
```

## Markdown Standards

### File Naming
- Use lowercase with hyphens: `quick-start.md`, not `QuickStart.md`
- Descriptive names: `api-authentication.md` not `doc.md`
- Version-specific docs: `deployment-v2.md` if multiple versions
- Maximum 80 characters for filename

### Heading Hierarchy
```markdown
# Main Title (H1 - Only one per document)

## Section (H2)

### Subsection (H3)

#### Details (H4)

# Not Recommended (H5+)
```

### Code Blocks
```markdown
# For specific languages
\`\`\`go
package main

func main() {
    fmt.Println("Hello, World!")
}
\`\`\`

# For shell commands
\`\`\`bash
docker build -t myapp .
docker run myapp
\`\`\`

# For configuration files
\`\`\`yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: config
\`\`\`
```

### Lists and Tables
```markdown
## Unordered Lists
- Item 1
- Item 2
  - Nested item
  - Another nested

## Ordered Lists
1. First step
2. Second step
3. Final step

## Tables
| Feature | Status | Notes |
|---------|--------|-------|
| Auth    | ✓      | JWT + API Key |
| Caching | ✓      | Redis |
| K8s     | ✓      | 1.24+ |
```

### Links and References
```markdown
# Internal links
See [Installation Guide](./INSTALLATION.md) for setup steps
View [Architecture Diagram](../diagrams/architecture.png)

# External links
Check [Kubernetes Docs](https://kubernetes.io/docs/)
Reference [Go Specification](https://golang.org/ref/spec)

# Link to GitHub code
See implementation in [handlers/user.go](../../internal/api/handlers/user.go#L42)
```

### Emphasis and Special Formatting
```markdown
**Important**: Do not skip validation
*Optional* configuration available
~~Deprecated~~ feature removed
`inline code` for variables and commands

> Important note or warning
> Multiple lines supported
```

## README Standards

### README Template
```markdown
# Project Name

Brief description of what the project does (1-2 sentences).

## Features

- Feature 1
- Feature 2
- Feature 3

## Quick Start

### Prerequisites
- Go 1.20+
- Docker
- PostgreSQL 13+

### Installation

1. Clone repository
\`\`\`bash
git clone <repo-url>
cd platform-go
\`\`\`

2. Configure environment
\`\`\`bash
cp .env.example .env
# Edit .env with your settings
\`\`\`

3. Build and run
\`\`\`bash
make run
\`\`\`

## Architecture

[Brief overview + link to detailed docs]

## API Documentation

[Quick API reference + link to full docs]

## Configuration

### Environment Variables
| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | Server port |
| DATABASE_URL | - | PostgreSQL connection |
| REDIS_URL | - | Redis connection |

## Development

### Running Tests
\`\`\`bash
go test ./... -timeout 5m
\`\`\`

### Code Style
- Follow [Go Code Review Comments](...)
- Use `gofmt` for formatting
- Maintain 70%+ test coverage

## Deployment

See [DEPLOYMENT.md](docs/DEPLOYMENT.md) for production setup.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[License type]

## Support

- Issues: GitHub Issues
- Discussions: GitHub Discussions
- Email: support@example.com
```

## API Documentation

### API Documentation Template
```markdown
# API Reference

## Base URL
\`\`\`
https://api.example.com/api/v1
\`\`\`

## Authentication

### JWT Token (Web Clients)
\`\`\`bash
curl -H "Authorization: Bearer <token>" \\
  https://api.example.com/api/v1/users
\`\`\`

### API Key (Services)
\`\`\`bash
curl -H "X-API-Key: <api_key>" \\
  https://api.example.com/api/v1/jobs
\`\`\`

## Endpoints

### Create User
\`\`\`
POST /users
\`\`\`

#### Request Body
\`\`\`json
{
  "name": "Alice",
  "email": "alice@example.com",
  "password": "secure_password"
}
\`\`\`

#### Response (201 Created)
\`\`\`json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "Alice",
    "email": "alice@example.com",
    "created_at": "2024-01-15T10:30:00Z"
  },
  "message": "User created"
}
\`\`\`

#### Error Response (400 Bad Request)
\`\`\`json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Email is invalid"
  }
}
\`\`\`

#### Status Codes
| Code | Meaning |
|------|---------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 500 | Server Error |

### List Users
\`\`\`
GET /users?page=1&limit=10&sort=created_at
\`\`\`

#### Query Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| page | int | Page number (1-indexed) |
| limit | int | Results per page (max 100) |
| sort | string | Sort field (default: created_at) |

#### Response
\`\`\`json
{
  "success": true,
  "data": [
    { "id": 1, "name": "Alice", ... },
    { "id": 2, "name": "Bob", ... }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 42,
    "total_pages": 5
  }
}
\`\`\`

### Get User
\`\`\`
GET /users/{id}
\`\`\`

#### Response (200 OK)
\`\`\`json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "Alice",
    "email": "alice@example.com",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
\`\`\`
```

## Architecture Documentation

### Architecture Document Template
```markdown
# System Architecture

## Overview

[High-level description of system]

## Architecture Diagram

[ASCII diagram or image reference]

```
┌──────────────┐
│   API Layer  │
└──────┬───────┘
       │
┌──────▼──────────┐
│  Application    │
│   Services      │
└──────┬──────────┘
       │
┌──────▼──────────┐
│   Domain        │
│   Models        │
└──────┬──────────┘
       │
┌──────▼──────────┐
│ Infrastructure  │
│ (DB, K8s, etc)  │
└─────────────────┘
\`\`\`

## Components

### API Layer
- Handles HTTP requests
- Input validation
- Response formatting

### Application Layer
- Business logic
- Orchestration
- Service coordination

### Domain Layer
- Core entities
- Business rules
- Domain interfaces

### Infrastructure Layer
- Database access
- External service integration
- Implementation details

## Data Flow

[Sequence diagrams or flow descriptions]

## Technologies

- **Language**: Go 1.20+
- **Framework**: Gin Web Framework
- **Database**: PostgreSQL 13+
- **Cache**: Redis
- **Container**: Docker
- **Orchestration**: Kubernetes
```

## Writing Best Practices

### Clarity
- Use clear, simple language
- Avoid jargon without explanation
- One idea per paragraph
- Short sentences (< 20 words)

### Completeness
- Include code examples
- Provide step-by-step instructions
- Cover common use cases
- Document limitations

### Consistency
- Use consistent terminology
- Keep tone professional
- Maintain consistent formatting
- Update docs with code changes

### Examples

**Good Example:**
> To authenticate with JWT, include your token in the Authorization header:
> \`\`\`bash
> curl -H "Authorization: Bearer YOUR_TOKEN" https://api.example.com/users
> \`\`\`

**Poor Example:**
> Authorization uses Bearer tokens in the header.

**Good Example:**
> Create a new project by sending a POST request with the project name and description:
> \`\`\`json
> { "name": "My Project", "description": "..." }
> \`\`\`

**Poor Example:**
> POST creates stuff.

## Maintenance

### Version Control
- Keep docs in Git with code
- Update docs in same PR as code changes
- Document breaking changes clearly

### Outdated Content
- Mark outdated sections with `⚠️ Outdated` banner
- Provide link to current version
- Set review date on complex docs

### Review Checklist
- [ ] Links are correct and functional
- [ ] Code examples are tested
- [ ] Instructions follow step-by-step
- [ ] Technical accuracy verified
- [ ] Spelling and grammar correct
- [ ] Formatting consistent
- [ ] Images/diagrams current
- [ ] Examples work with latest version

## Tools & Automation

### Markdown Linting
```bash
# Check markdown formatting
bash .github/skills-consolidated/documentation/scripts/lint-markdown.sh

# Generate table of contents
bash .github/skills-consolidated/documentation/scripts/generate-toc.sh

# Validate links
bash .github/skills-consolidated/documentation/scripts/validate-links.sh
```

## Documentation Examples

See `/docs/` directory for real examples:
- [API Documentation](../../docs/API_STANDARDS.md)
- [Architecture Analysis](../../docs/K8S_ARCHITECTURE_ANALYSIS.md)

## References
- Markdown Guide: https://www.markdownguide.org/
- Google Style Guide: https://google.github.io/styleguide/
- CommonMark Spec: https://spec.commonmark.org/
