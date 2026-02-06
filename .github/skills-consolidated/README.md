---
name: Platform-Go Skills Framework
description: Consolidated skill sets for platform-go development
version: "1.0"
date: "2024-01-15"
---

# Platform-Go Consolidated Skills Framework

This directory contains the complete skill sets for platform-go development, consolidated from 17 individual skills into 5 integrated, cohesive skill categories.

## Skills Overview

### 1. **Code Quality** - `code-quality/`
Foundation for production-grade code development and testing.

**Focus:**
- Golang production standards and best practices
- Comprehensive testing (unit, integration, 70%+ coverage)
- Robust error handling and context propagation
- Code validation and pre-commit checks
- Clean code organization (200-line file limit)

**When to Use:**
- Writing new Go code
- Implementing features
- Testing strategy
- Code reviews
- Error handling design

**Key Files:**
- `SKILL.md` - Complete reference
- `scripts/validate-code.sh` - Code quality checks
- `scripts/run-tests.sh` - Test execution with coverage

---

### 2. **Architecture** - `architecture/`
System design patterns, API design, database optimization, and scalability.

**Focus:**
- RESTful API design patterns (Gin framework)
- Clean layered architecture (domain → application → API)
- Database design and optimization (GORM, PostgreSQL)
- Production readiness checklists
- Scalability patterns

**When to Use:**
- Designing new features
- API endpoint planning
- Database schema design
- System architecture decisions
- Production deployment planning

**Key Files:**
- `SKILL.md` - Complete reference
- `scripts/validate-architecture.sh` - Architecture validation
- `scripts/validate-migrations.sh` - Database migration checks

---

### 3. **Operations** - `operations/`
CI/CD automation, Kubernetes integration, caching, and monitoring.

**Focus:**
- GitHub Actions CI/CD pipelines
- Kubernetes deployment and management
- Redis caching strategies
- Production monitoring and observability
- Health checks and metrics

**When to Use:**
- Setting up CI/CD pipelines
- Kubernetes deployment
- Caching implementation
- Monitoring setup
- Production operations

**Key Files:**
- `SKILL.md` - Complete reference
- `scripts/deploy.sh` - Kubernetes deployment
- `scripts/health-check.sh` - Cluster health verification
- `scripts/monitor-cache.sh` - Redis monitoring

---

### 4. **Security & Compliance** - `security-compliance/`
Authentication, authorization, API key management, and secure coding.

**Focus:**
- Unified authentication (JWT + API Key)
- Role-based access control (RBAC)
- API key management and validation
- Secure coding practices
- Default admin initialization

**When to Use:**
- Implementing authentication
- Setting up access control
- API key lifecycle management
- Security audits
- Compliance verification

**Key Files:**
- `SKILL.md` - Complete reference
- `scripts/security-scan.sh` - Vulnerability scanning
- `scripts/check-secrets.sh` - Secret detection
- `scripts/audit-keys.sh` - API key auditing

---

### 5. **Documentation** - `documentation/`
Markdown standards, documentation structure, and content guidelines.

**Focus:**
- Markdown formatting standards
- Documentation organization
- API documentation templates
- README best practices
- Architecture documentation

**When to Use:**
- Writing documentation
- API reference creation
- Architecture documentation
- README updates
- User guides

**Key Files:**
- `SKILL.md` - Complete reference
- `scripts/lint-markdown.sh` - Markdown validation
- `scripts/generate-toc.sh` - Table of contents generation
- `scripts/validate-links.sh` - Link validation

---

## Consolidation Mapping

This 5-skill structure consolidates the original 17 skills:

| Consolidated Skill | Original Skills |
|-------------------|-----------------|
| **code-quality** | golang-production-standards, testing-best-practices, error-handling-guide, code-validation-standards, file-structure-guidelines |
| **architecture** | api-design-patterns, file-structure-guidelines, database-best-practices, production-readiness-checklist |
| **operations** | cicd-pipeline-optimization, kubernetes-integration, redis-caching, monitoring-observability |
| **security-compliance** | security-best-practices, access-control-best-practices, unified-authentication-strategy, automigration-apikey-initialization |
| **documentation** | markdown-documentation-standards |

## Quick Start

### For Developers

1. **Writing New Code?** → `code-quality/`
2. **Designing Features?** → `architecture/`
3. **Need Secure Auth?** → `security-compliance/`
4. **Setting Up Testing?** → `code-quality/`
5. **Writing Docs?** → `documentation/`

### For DevOps/Operations

1. **CI/CD Pipeline?** → `operations/`
2. **Kubernetes Deployment?** → `operations/`
3. **Monitoring Setup?** → `operations/`
4. **Caching Strategy?** → `operations/`

### For Security Review

1. **Access Control?** → `security-compliance/`
2. **API Keys?** → `security-compliance/`
3. **Secure Coding?** → `security-compliance/`
4. **Code Review?** → `code-quality/`

## Using Skills in Your Project

### 1. Reference in Code Comments
```go
// Security best practices: https://github.com/.../security-compliance/SKILL.md
func validateAPIKey(key string) error {
    // Implementation following security-compliance guidelines
}
```

### 2. PR Checklist Template
```markdown
## Pre-Submission Checklist

Code Quality:
- [ ] Review: code-quality/SKILL.md
- [ ] Tests: 70%+ coverage
- [ ] Linting: scripts/validate-code.sh passed

Architecture:
- [ ] API design follows: architecture/SKILL.md
- [ ] Database schema validated

Security:
- [ ] Review: security-compliance/SKILL.md
- [ ] No hardcoded secrets

Documentation:
- [ ] Updated: documentation/SKILL.md standards
```

### 3. Validation Scripts
```bash
# Run all skill validations
bash code-quality/scripts/validate-code.sh
bash architecture/scripts/validate-architecture.sh
bash security-compliance/scripts/security-scan.sh
bash documentation/scripts/lint-markdown.sh
```

## Directory Structure

```
.github/skills-consolidated/
├── code-quality/
│   ├── SKILL.md
│   ├── scripts/
│   │   ├── validate-code.sh
│   │   ├── run-tests.sh
│   │   └── format-check.sh
│   └── examples/
│
├── architecture/
│   ├── SKILL.md
│   ├── scripts/
│   │   ├── validate-architecture.sh
│   │   └── validate-migrations.sh
│   └── examples/
│
├── operations/
│   ├── SKILL.md
│   ├── scripts/
│   │   ├── deploy.sh
│   │   ├── health-check.sh
│   │   └── monitor-cache.sh
│   └── examples/
│
├── security-compliance/
│   ├── SKILL.md
│   ├── scripts/
│   │   ├── security-scan.sh
│   │   ├── check-secrets.sh
│   │   └── audit-keys.sh
│   └── examples/
│
├── documentation/
│   ├── SKILL.md
│   ├── scripts/
│   │   ├── lint-markdown.sh
│   │   ├── generate-toc.sh
│   │   └── validate-links.sh
│   └── examples/
│
└── README.md (this file)
```

## Best Practices

### Learning Path
1. Start with `code-quality/SKILL.md` for fundamentals
2. Move to `architecture/` for system design
3. Add `security-compliance/` for secure implementation
4. Use `operations/` for deployment
5. Follow `documentation/` for knowledge sharing

### Code Review Process
1. Check `code-quality/` checklist
2. Verify `architecture/` compliance
3. Audit `security-compliance/` requirements
4. Validate `documentation/` standards

### Feature Development Workflow
```
Design (architecture/) → Implement (code-quality/) → 
Secure (security-compliance/) → Deploy (operations/) → 
Document (documentation/)
```

## Contributing to Skills

### Updating Skills
Each consolidated skill in `SKILL.md` includes:
- Principles and patterns
- Code examples
- Validation checklist
- Tools and scripts
- References

To update a skill:
1. Edit the relevant `SKILL.md`
2. Update associated scripts
3. Add examples if needed
4. Submit PR with documentation

### Adding New Examples
Create examples in `examples/` directory:
```
examples/
├── example1.md
├── example2.md
└── code/
    ├── handler.go
    └── service.go
```

## References

### External Resources
- Go Best Practices: https://golang.org/doc/effective_go
- Kubernetes: https://kubernetes.io/docs/
- PostgreSQL: https://www.postgresql.org/docs/
- Redis: https://redis.io/documentation
- GitHub Actions: https://docs.github.com/en/actions

### Internal Documentation
- Project README: [../../README.md](../../README.md)
- API Standards: [../../docs/API_STANDARDS.md](../../docs/API_STANDARDS.md)
- K8s Architecture: [../../docs/K8S_ARCHITECTURE_ANALYSIS.md](../../docs/K8S_ARCHITECTURE_ANALYSIS.md)

## Support

### Questions?
1. Check the relevant skill's `SKILL.md`
2. Review examples in `examples/` directory
3. Run validation scripts in `scripts/` directory
4. Refer to external references in the skill documentation

### Report Issues
If you find outdated information:
1. Create an issue with the skill name
2. Include specific section and page
3. Provide reference or suggested correction

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2024-01-15 | Initial consolidation from 17 skills into 5 integrated sets |

## License

Proprietary - platform-go project

---

**Last Updated**: 2024-01-15
**Maintainer**: platform-go team
