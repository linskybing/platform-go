---
name: integration-testing
description: Integration testing framework using Docker, PostgreSQL, Redis, and KinD for platform-go
---

# Integration Testing

This skill defines the Docker based integration testing setup for platform-go.

## When to Use

- Running integration tests that require PostgreSQL and Redis
- Running Kubernetes integration tests with KinD
- Debugging integration test failures in isolated containers
- Running integration tests in CI pipelines

## Quick Start

```bash
# Run database integration tests
bash .github/skills/integration-testing/scripts/docker-db-integration-test.sh

# Run all integration tests
bash .github/skills/integration-testing/scripts/run-all-integration-tests.sh

# Quick runner
bash .github/skills/integration-testing/scripts/quick-test.sh

# Cleanup
bash .github/skills/integration-testing/scripts/cleanup.sh
```

## Docker Files

- .github/skills/integration-testing/docker/Dockerfile.integration
- .github/skills/integration-testing/docker/docker-compose.test.yml

## Scripts

- .github/skills/integration-testing/scripts/run-all-integration-tests.sh
- .github/skills/integration-testing/scripts/docker-db-integration-test.sh
- .github/skills/integration-testing/scripts/docker-k8s-integration-test.sh
- .github/skills/integration-testing/scripts/cleanup.sh
- .github/skills/integration-testing/scripts/quick-test.sh

## Environment Variables

```bash
DATABASE_URL=postgres://testuser:testpass@postgres:5432/testdb?sslmode=disable
REDIS_URL=redis://redis:6379
TEST_ENV=integration
```

## Notes

- Docker must be installed and running on the host.
- Tests use Docker networks and service names for connectivity.
- Use the cleanup script after failures to remove test resources.
