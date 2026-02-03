# Platform-Go Agent Skills Index

Complete set of production-grade Agent Skills for platform-go project in VS Code.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Skills Overview](#skills-overview)
   - [Foundation Skills](#foundation-skills)
   - [Implementation Skills](#implementation-skills)
   - [Production Skills](#production-skills)
   - [Documentation Skills](#documentation-skills)
3. [Quick Reference](#quick-reference)

---

## Quick Start

### Using Skills & Scripts

Each skill includes its own validation scripts in a `scripts/` subdirectory:

```bash
# Code Validation
bash .github/skills/code-validation-standards/scripts/validate-before-push.sh
bash .github/skills/code-validation-standards/scripts/pre-commit-validate.sh

# Testing
bash .github/skills/testing-best-practices/scripts/test-coverage.sh
bash .github/skills/testing-best-practices/scripts/test-with-race.sh
bash .github/skills/testing-best-practices/scripts/coverage-html.sh

# API Design
bash .github/skills/api-design-patterns/scripts/swagger-generate.sh
bash .github/skills/api-design-patterns/scripts/swagger-validate.sh

# Database
bash .github/skills/database-best-practices/scripts/migration-check.sh

# Golang Standards
bash .github/skills/golang-production-standards/scripts/lint-check.sh
bash .github/skills/golang-production-standards/scripts/compile-check.sh
```

### Workflow Recommendation

1. Write your code
2. Run validation scripts from the relevant skill
3. Fix any issues found
4. Commit when all checks pass

---

## Skills Overview

This directory contains comprehensive skills that enforce production standards, Golang best practices, and project-specific patterns for the platform-go project.

Skills are organized into four categories:

### Foundation Skills

Core standards for all development:

- **golang-production-standards** - Production-grade Go code standards
- **file-structure-guidelines** - Code organization and modular design
- **code-validation-standards** - Pre-commit checks and validation

### Implementation Skills

Feature development and architecture:

- **api-design-patterns** - RESTful API design and Gin framework patterns
- **database-best-practices** - PostgreSQL, GORM, and query optimization
- **kubernetes-integration** - Client-go usage and resource management
- error-handling-guide

**Quality Skills** (Testing & Deployment):
- testing-best-practices
- security-best-practices
- monitoring-observability
- cicd-pipeline-optimization
- production-readiness-checklist

### 1. golang-production-standards
**Focus**: Core Golang coding standards and best practices
**When to use**: When writing any Go code for platform-go
**Key topics**:
- Code organization with clean architecture
- Error handling with fmt.Errorf and custom error types
- Context propagation and timeouts
- Concurrency patterns (WaitGroup, semaphores, resource pooling)
- Testing requirements (70% coverage, table-driven tests)
- Database transaction handling
- Kubernetes client safety
- Security best practices
- Performance optimization

**Related skills**: database-best-practices, error-handling-guide, package-organization

---

### 2. api-design-patterns
**Focus**: RESTful API design with Gin framework
**When to use**: When creating HTTP endpoints or handlers
**Key topics**:
- RESTful resource naming conventions
- Request/Response DTOs (never expose domain models)
- Standardized response format
- Input validation with binding tags
- Error handling and HTTP status codes
- Middleware patterns (auth, logging, rate limiting)
- Query parameter handling
- File upload handling
- Context and timeout management
- API versioning strategies

**Related skills**: error-handling-guide, security-best-practices

---

### 3. kubernetes-integration
**Focus**: Kubernetes API integration and resource management
**When to use**: When working with Kubernetes resources or K8s client
**Key topics**:
- Client initialization with nil checks
- Namespace management and safe naming
- Label selectors for efficient queries
- Resource specifications (requests/limits, priority classes)
- GPU resource management (dedicated/shared, MPS)
- PVC lifecycle management
- Service and networking patterns
- Error handling and retries with exponential backoff
- Resource cleanup and deletion
- Monitoring and observability
- Testing with fake Kubernetes client
- Common patterns (GetOrCreate, pagination)

**Related skills**: golang-production-standards, error-handling-guide

---

### 4. testing-best-practices
**Focus**: Comprehensive testing strategies and patterns
**When to use**: When writing unit tests, integration tests, or benchmarks
**Key topics**:
- Test structure (AAA pattern)
- Table-driven tests (preferred pattern)
- Mocking external dependencies with interfaces
- Testing with context (timeout, cancellation)
- Integration tests with database setup/teardown
- Test fixtures and helper functions
- Testing concurrency and race conditions
- Benchmarking performance-critical code
- Comprehensive error case testing
- 70% coverage enforcement with HTML reports
- Pre-commit testing checklist

**Related skills**: golang-production-standards, database-best-practices

---

### 5. database-best-practices
**Focus**: Database design, ORM usage with GORM, and query optimization
**When to use**: When creating tables, writing queries, or optimizing performance
**Key topics**:
- Schema design with proper indexes
- GORM query patterns (prepared statements, eager loading)
- Complex query building
- Database transactions and savepoints
- Query optimization and pagination
- Bulk operations and batching
- Connection pool management
- Error handling for database errors
- Migration versioning
- Performance targets (<100ms for queries)

**Related skills**: golang-production-standards, testing-best-practices

---

### 6. security-best-practices
**Focus**: Authentication, authorization, and secure coding practices
**When to use**: When handling user input, authentication, or sensitive data
**Key topics**:
- JWT token management
- Password hashing with bcrypt
- Role-based access control (RBAC)
- Input validation and sanitization
- SQL injection prevention
- Secrets and credentials management
- File upload security
- Security headers and CORS
- Rate limiting
- API security patterns
- 15-point security checklist

**Related skills**: api-design-patterns, error-handling-guide

---

### 7. error-handling-guide
**Focus**: Comprehensive error handling patterns and recovery
**When to use**: When handling errors, logging, or implementing retry logic
**Key topics**:
- Custom error types definition
- Error wrapping with context (fmt.Errorf, %w)
- Structured logging with slog
- Error recovery and defer cleanup
- Panic recovery in goroutines
- Retry strategies with exponential backoff
- Error to HTTP status code mapping
- Error handling in concurrent code
- 10-point error handling checklist

**Related skills**: golang-production-standards, api-design-patterns

---

### 8. package-organization
**Focus**: Clean architecture, package structure, and code organization
**When to use**: When designing packages, refactoring code, or planning new features
**Key topics**:
- Clean architecture layers (handler → service → repository → domain)
- Project directory structure
- Package naming conventions
- Package boundaries and dependencies
- Domain models and business logic
- Service layer (application logic)
- Repository pattern for data access
- Handler layer (HTTP)
- Dependency injection
- Import organization
- Circular dependency prevention
- Public vs private visibility rules

**Related skills**: golang-production-standards, api-design-patterns

---

### 9. cicd-pipeline-optimization
**Focus**: GitHub Actions workflows, build optimization, and automated testing
**When to use**: When setting up CI/CD, optimizing builds, or adding tests to pipeline
**Key topics**:
- Test workflow configuration
- Build workflow with linting
- Integration tests in separate workflow
- Security scanning (gosec, trufflehog)
- Release workflow for tagged versions
- Docker build optimization (multi-stage)
- Build caching strategies
- Workflow performance optimization
- GitHub Secrets management
- Workflow status notifications

**Related skills**: testing-best-practices, security-best-practices

---

### 10. monitoring-observability
**Focus**: Logging, metrics, tracing, and alerting
**When to use**: When implementing logging, metrics collection, or health checks
**Key topics**:
- Structured JSON logging with slog
- Logging best practices (levels, context)
- Contextual logging with request IDs
- Prometheus metrics collection
- Counter, histogram, and gauge metrics
- Health check endpoints
- Kubernetes probes (startup, readiness, liveness)
- Distributed tracing (OpenTelemetry)
- Performance benchmarking
- Alerting guidelines

**Related skills**: golang-production-standards, error-handling-guide

---

### 11. file-structure-guidelines
**Focus**: Code file organization, modular design, 200-line file limit
**When to use**: When creating new features, splitting large files, or designing directory structure
**Key topics**:
- 200-line file limit rationale and benefits
- Feature-based directory organization
- Service layer splitting strategies
- Repository layer splitting patterns
- Handler layer splitting by operation
- File naming conventions
- Directory structure for maintainability
- Large file refactoring strategies
- Production quality checklist

**Related skills**: package-organization, golang-production-standards

---

### 12. production-readiness-checklist
**Focus**: Production readiness verification, quality gates, deployment checks
**When to use**: When preparing code for deployment or conducting pre-release verification
**Key topics**:
- Pre-commit code quality checklist (10 items)
- Error handling verification (8 items)
- Testing coverage requirements (7 items)
- Security compliance (8 items)
- API design verification (6 items)
- Database readiness (7 items)
- Concurrency safety (5 items)
- Configuration management (8 items)
- Performance requirements (6 items)
- Deployment checklist (20+ items)
- Post-deployment verification (10 items)
- Emergency rollback procedures

**Related skills**: All other skills (comprehensive quality gate)

---

## Quick Start

### Enable Agent Skills

1. Open VS Code settings (Ctrl+, or Cmd+,)
2. Search for "Agent Skills"
3. Enable `chat.useAgentSkills` setting

Or add to settings.json:
```json
{
  "chat.useAgentSkills": true
}
```

4. Restart VS Code

### Using Skills with GitHub Copilot

- Skills are automatically discovered from `.github/skills/` directory
- Copilot will suggest patterns from relevant skills based on code context
- Type `@`  in the Copilot chat to see available skills
- Use `/` in the Copilot chat to see skill commands

### Skill Dependencies

Skills build on each other:
```
golang-production-standards (foundation)
├── error-handling-guide
├── package-organization
├── database-best-practices
├── security-best-practices
├── testing-best-practices
├── kubernetes-integration
├── api-design-patterns
├── monitoring-observability
└── cicd-pipeline-optimization
```

## Statistics

Total Skills: 12
Total Documentation Lines: 5,935
Average Skill Size: 495 lines

Complete breakdown:
- Foundation Skills: 3 (golang-production-standards, package-organization, file-structure-guidelines)
- Implementation Skills: 4 (api-design-patterns, database-best-practices, kubernetes-integration, error-handling-guide)  
- Quality Skills: 5 (testing-best-practices, security-best-practices, monitoring-observability, cicd-pipeline-optimization, production-readiness-checklist)

All skills:
- 100% English documentation
- Production-ready quality
- 250+ code examples
- 120+ checklists
- 60+ anti-patterns
- 6+ performance baselines

---

| Component | Target | Skill |
|-----------|--------|-------|
| API Response | <200ms p95 | golang-production-standards |
| Database Query | <100ms | database-best-practices |
| Kubernetes API | <500ms | kubernetes-integration |
| Unit Tests | <2 minutes | testing-best-practices |
| Full Build | <5 minutes | cicd-pipeline-optimization |
| Code Coverage | ≥70% | testing-best-practices |

---

## Quality Checklist

Run through these checklists before committing:

### Before Code Review
- [ ] Golang Production Standards checklist
- [ ] Error Handling checklist
- [ ] Testing checklist
- [ ] Security checklist (if handling user data)

### Before Merge
- [ ] All tests pass (unit + integration)
- [ ] Coverage ≥70%
- [ ] Linting passes (golangci-lint)
- [ ] No race conditions detected
- [ ] Security scan passes
- [ ] Code formatted with gofmt

### Before Release
- [ ] CI/CD pipeline passes
- [ ] Release workflow creates artifacts
- [ ] Docker image builds successfully
- [ ] All health checks operational
- [ ] Monitoring configured and alerts active

---

## Common Workflows

### Writing a New API Endpoint
1. Read: api-design-patterns (RESTful resource naming)
2. Read: error-handling-guide (error responses)
3. Read: security-best-practices (input validation)
4. Create handler following DTO pattern
5. Run tests with testing-best-practices checklist
6. Verify monitoring with monitoring-observability

### Implementing New Database Feature
1. Read: database-best-practices (schema design)
2. Read: golang-production-standards (context usage)
3. Create domain model in internal/domain/
4. Create repository in internal/repository/
5. Create service in internal/application/
6. Write tests following testing-best-practices
7. Verify performance <100ms

### Adding Kubernetes Resource Management
1. Read: kubernetes-integration (client safety)
2. Read: error-handling-guide (retry logic)
3. Implement with nil checks for test env
4. Add label selectors for efficient queries
5. Implement retry with exponential backoff
6. Test with fake Kubernetes client

### Optimizing Slow Code
1. Read: golang-production-standards (performance tips)
2. Profile with pprof
3. Add benchmarks
4. Optimize following checklist
5. Measure improvement
6. Commit with benchmark results

---

## Integration with Development Tools

### Pre-commit Hook
```bash
#!/bin/bash
go test -race -cover ./...
go fmt ./...
golangci-lint run
```

### IDE Settings
- Format on save: enabled
- Go linter: golangci-lint
- Test on save: recommended
- Coverage reporting: enabled

### GitHub Actions
All workflows automatically check skills compliance:
- Test: 70% coverage required
- Build: linting required
- Security: gosec + trufflehog
- Release: multi-platform builds

---

## Contributing to Skills

When updating a skill:
1. Maintain consistent format with "When to Use" section
2. Include code examples for all patterns
3. Add checklist at end of skill
4. Update anti-patterns section
5. Keep skills focused (one responsibility)
6. Link related skills at top
7. Test examples in context of platform-go

---

## License

These skills are part of platform-go project and follow project guidelines.

---

## Support

For questions about skills usage:
1. Check skill documentation
2. Review code examples in skill
3. Check related skills
4. Consult project README.md
5. Ask in team Slack channel

---

Last Updated: 2026-02-02
Total Skills: 11
Total Documentation: 5,900+ lines
Code Examples: 250+
Quality Checklists: 120+
