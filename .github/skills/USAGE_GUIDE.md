---
name: skills-usage-guide
description: How to use Agent Skills and their validation scripts in platform-go project
---

# Agent Skills Usage Guide

This guide explains how to use the platform-go Agent Skills and their associated validation scripts for daily development.

## Directory Structure

Each Agent Skill is organized with its own `scripts/` subdirectory:

```
.github/skills/
├── code-validation-standards/
│   ├── SKILL.md
│   └── scripts/
│       ├── pre-commit-validate.sh
│       └── validate-before-push.sh
├── testing-best-practices/
│   ├── SKILL.md
│   └── scripts/
│       ├── test-coverage.sh
│       ├── test-with-race.sh
│       └── coverage-html.sh
├── api-design-patterns/
│   ├── SKILL.md
│   └── scripts/
│       ├── swagger-generate.sh
│       └── swagger-validate.sh
├── database-best-practices/
│   ├── SKILL.md
│   └── scripts/
│       └── migration-check.sh
├── golang-production-standards/
│   ├── SKILL.md
│   └── scripts/
│       ├── lint-check.sh
│       └── compile-check.sh
└── [other skills]/
    ├── SKILL.md
    └── scripts/
        └── .gitkeep (ready for future scripts)
```

## Using Skills in VS Code

### Enable Agent Skills
1. Open VS Code Settings (Cmd+,)
2. Search for "Agent Skills"
3. Enable: `Chat: Use Agent Skills` 
4. Enable: `Chat: Display Agent Skills In Chat`

### Discover Available Skills
In VS Code Chat, type:
```
@skills
```

This shows all available skills with quick navigation.

### Get Guidance for a Task
Ask Copilot with skill reference:
```
@api-design-patterns How do I create a new API endpoint with Swagger docs?
@testing-best-practices How do I write proper unit tests?
@code-validation-standards What checks should I run before committing?
```

## Running Validation Scripts

### Before Every Code Change

**1. Check Code Validation Standards**
```bash
bash .github/skills/code-validation-standards/scripts/validate-before-push.sh
```

This runs comprehensive checks:
- Code format (gofmt)
- Static analysis (go vet)
- Compilation
- Linting (golangci-lint)
- Test coverage
- Race condition detection
- Security checks
- Dependency validation

**2. Check Production Standards**
```bash
bash .github/skills/golang-production-standards/scripts/lint-check.sh
bash .github/skills/golang-production-standards/scripts/compile-check.sh
```

### When Writing Tests
```bash
bash .github/skills/testing-best-practices/scripts/test-coverage.sh
bash .github/skills/testing-best-practices/scripts/test-with-race.sh
bash .github/skills/testing-best-practices/scripts/coverage-html.sh
```

### When Creating API Endpoints
```bash
bash .github/skills/api-design-patterns/scripts/swagger-generate.sh
bash .github/skills/api-design-patterns/scripts/swagger-validate.sh
```

### When Writing Database Code
```bash
bash .github/skills/database-best-practices/scripts/migration-check.sh
```

## Daily Workflow

### Before Writing Code
1. Consult relevant skill documentation
   ```bash
   cat .github/skills/[skill-name]/SKILL.md
   ```
2. Review pre-code validation checklist
3. Plan your implementation

### While Writing Code
1. Use IDE settings for auto-formatting
2. Keep files under 200 lines (file-structure-guidelines)
3. Follow code organization patterns from golang-production-standards
4. Use examples from relevant skills

### Before Committing
1. Run code validation scripts
   ```bash
   bash .github/skills/code-validation-standards/scripts/pre-commit-validate.sh
   ```
2. Fix any issues reported
3. Run tests with coverage
   ```bash
   bash .github/skills/testing-best-practices/scripts/test-coverage.sh
   ```

### Before Pushing
1. Run comprehensive validation
   ```bash
   bash .github/skills/code-validation-standards/scripts/validate-before-push.sh
   ```
2. Review any warnings
3. Only push if all checks pass

## Skill Quick Reference by Task

### Writing a New Feature
1. **Plan Structure**: file-structure-guidelines
2. **Write Code**: golang-production-standards
3. **Write Tests**: testing-best-practices
4. **Handle Errors**: error-handling-guide
5. **Validate**: code-validation-standards
6. Scripts:
   ```bash
   bash .github/skills/golang-production-standards/scripts/compile-check.sh
   bash .github/skills/testing-best-practices/scripts/test-coverage.sh
   ```

### Creating API Endpoints
1. **Design**: api-design-patterns
2. **Add Swagger**: api-design-patterns
3. **Handle Errors**: error-handling-guide
4. **Security**: security-best-practices
5. Scripts:
   ```bash
   bash .github/skills/api-design-patterns/scripts/swagger-generate.sh
   bash .github/skills/api-design-patterns/scripts/swagger-validate.sh
   ```

### Database Operations
1. **Design**: database-best-practices
2. **Write Migrations**: database-best-practices
3. **Optimize**: database-best-practices
4. Scripts:
   ```bash
   bash .github/skills/database-best-practices/scripts/migration-check.sh
   ```

### Working with Kubernetes
1. **Design**: kubernetes-integration
2. **Implementation**: kubernetes-integration
3. **Testing**: testing-best-practices
4. Scripts: (future additions)

### Pre-Deployment
1. **Review**: production-readiness-checklist
2. **Security**: security-best-practices
3. **Monitoring**: monitoring-observability
4. **CI/CD**: cicd-pipeline-optimization
5. Scripts:
   ```bash
   bash .github/skills/code-validation-standards/scripts/validate-before-push.sh
   ```

## Script Output Examples

### Successful Validation
```
========================================
Comprehensive Pre-Push Validation
========================================

[1/8] Format check...
✓ Format OK

[2/8] Running go vet...
✓ Vet OK

[3/8] Compile check...
✓ API compiles
✓ Scheduler compiles

[4/8] Lint check...
✓ Lint OK

[5/8] Test coverage check...
✓ Tests passed
  Coverage: 75.3% (target: 70%)

[6/8] Race condition detection...
✓ No race conditions

[7/8] Security checks...
✓ No SQL injection patterns

[8/8] Dependency check...
✓ Dependencies OK

========================================
Validation Summary
========================================
✓ All validation checks passed!
Ready to push.
```

### Failed Validation
```
✗ Format issues found:
  - internal/api/handler.go
  - pkg/utils/helper.go

Fix with: make fmt
```

## Troubleshooting

### Script Permissions Error
```bash
chmod +x .github/skills/[skill]/scripts/*.sh
```

### Script Not Found
Make sure you're running from the project root:
```bash
cd /path/to/platform-go
bash .github/skills/code-validation-standards/scripts/validate-before-push.sh
```

### Script Fails but Seems Wrong
1. Check script output carefully
2. Review the specific skill documentation
3. Try running individual checks manually:
   ```bash
   go fmt ./...
   go vet ./...
   go build ./...
   ```

## Integration with Git Hooks

Set up pre-commit hook:
```bash
ln -s ../../.github/skills/code-validation-standards/scripts/pre-commit-validate.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

Now validation runs automatically on `git commit`.

## IDE Configuration

### VS Code
Add to `.vscode/settings.json`:
```json
{
  "go.lintOnSave": "package",
  "go.lintTool": "golangci-lint",
  "[go]": {
    "editor.formatOnSave": true,
    "editor.defaultFormatter": "golang.go"
  }
}
```

### GoLand/IntelliJ
1. Settings → Go → Code Style → Formatter: gofmt
2. Settings → Go → Tools → Linter: golangci-lint
3. Settings → Build, Execution, Deployment → Build Tools → Go Modules

## Best Practices

1. **Always run validation before committing**
   - Use pre-commit hook for automatic checks
   - Run full validation before pushing

2. **Read skill documentation**
   - Each skill has examples and patterns
   - Review checklist at end of each skill

3. **Fix issues immediately**
   - Don't ignore warnings
   - Address test failures before moving on

4. **Keep scripts updated**
   - Scripts evolve with skill improvements
   - Pull latest version before major changes

5. **Use scripts for CI/CD**
   - GitHub Actions can run same scripts
   - Ensures local and CI consistency

## Related Resources

- [Skills Index](./README.md)
- [Individual Skill Guides](./)
- [Project README](../../README.md)
- [Testing Documentation](../../test/integration/README.md)

---

**Last Updated**: 2026-02-02
**Skills Version**: 12 (11 active)
**Total Scripts**: 10

