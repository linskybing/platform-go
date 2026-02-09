---
title: Code-Quality Flows
description: Developer workflows for PR, testing, and release quality checks with references to repository locations.
---

# Code Quality Flows

## Q1. Pull Request -> Review -> Merge Flow

1. Developer opens PR with feature branch. Ensure `gofmt` and `golangci-lint` run locally.
2. CI runs tests and linters (`go test ./...`, `golangci-lint run`).
3. Reviewer checks `internal/*` changes for file size (<200 lines), error wrapping, context-first signatures, and test coverage.
4. Merge only after: unit tests pass, integration tests either pass or are marked/skipped, and coverage threshold met.

Reference files:
- lint scripts under `.github/scripts` or `scripts/format-skills.sh`
- tests under `internal/*_test.go` and `test/integration/`

## Q2. Test & Validation Flow for Services

1. Unit tests: write table-driven tests for service methods in `internal/application/*_test.go` using mocks (gomock).
2. Integration tests: use `test/integration/` harness with test DB and K8s fixtures (use `t.Skip()` when infra not present).
3. Coverage: collect `coverage.out` and fail PR if below threshold (CI config).

Reference examples:
- `internal/application/project_service_test.go`
- `test/integration/project_handler_test.go`
