# Contributing Guide

Thank you for your interest in contributing to the Platform API!

## Development Environment

### Prerequisites
- **Go**: 1.21 or later
- **PostgreSQL**: 13 or later (with `ltree`, `btree_gist`, `uuid-ossp` enabled)
- **Docker**: For running integration tests

### Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/linskybing/platform-go.git
   cd platform-go
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Database**:
   Ensure you have a PostgreSQL instance running. Create a database (e.g., `platform_db`) and update your `.env` file accordingly.

## Project Structure

- `cmd/`: Application entry points (API server).
- `internal/`: Private application code.
    - `api/`: HTTP handlers and routing.
    - `application/`: Business logic services.
    - `domain/`: Core entities and interfaces.
    - `repository/`: Data access implementations.
- `pkg/`: Public shared libraries (Utils, K8s helpers).
- `test/integration/`: Integration tests.

## Testing

We prioritize **Integration Tests** to ensure the correctness of our complex database interactions and API contracts.

### Running Tests

Use the provided helper script to run the full integration suite in a clean Docker environment:

```bash
bash scripts/run-integration-tests.sh
```

This script will:
1. Spin up PostgreSQL and Redis containers.
2. Wait for them to be ready.
3. Run `go test ./test/integration/...`.
4. Tear down the containers.

**Note**: Do not use `go test ./...` for integration tests unless you have a local DB running and configured exactly as the test suite expects.

## Code Standards

- **Formatting**: Use `gofmt`.
- **Linting**: We recommend `golangci-lint`.
- **Naming**: Follow standard Go conventions (CamelCase).
- **Architecture**: Respect the separation of concerns (Handler -> Service -> Repository).

## Submitting a Pull Request

1. Create a new branch: `git checkout -b feature/my-feature`.
2. Implement your changes.
3. **Add Tests**: Ensure your changes are covered by integration tests.
4. Run the test suite: `bash scripts/run-integration-tests.sh`.
5. Commit your changes with a clear message.
6. Push to your fork and submit a PR.
