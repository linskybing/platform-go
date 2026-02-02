---
name: cicd-pipeline-optimization
description: GitHub Actions CI/CD configuration, build optimization, automated testing, and deployment strategies for platform-go
---

# CI/CD Pipeline Optimization

This skill provides guidelines for GitHub Actions configuration, build optimization, and automated testing pipelines.

## When to Use

Apply this skill when:
- Setting up or updating GitHub Actions workflows
- Optimizing build and test times
- Adding new tests to CI pipeline
- Implementing automated deployments
- Configuring dependency caching
- Setting up code quality checks
- Implementing security scanning
- Optimizing Docker builds

## Workflow Structure

### 1. Test Workflow

```yaml
# .github/workflows/test.yml
name: Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: testuser
          POSTGRES_PASSWORD: testpass
          POSTGRES_DB: testdb
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'
          cache: true        # Cache Go modules
          cache-dependency-path: go.sum

      - name: Run unit tests
        run: go test -v -race -coverprofile=coverage.out ./...
        env:
          DATABASE_URL: postgres://testuser:testpass@localhost:5432/testdb

      - name: Upload coverage reports
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
          flags: unittests
          name: codecov-umbrella

      - name: Check coverage threshold
        run: |
          total=$(go tool cover -func=coverage.out | grep total | grep -oP '(?<=\t)\d+\.\d+')
          if (( $(echo "$total < 70" | bc -l) )); then
            echo "Coverage $total% is below 70% threshold"
            exit 1
          fi
```

### 2. Build Workflow

```yaml
# .github/workflows/build.yml
name: Build

on:
  push:
    branches: [ main ]
    tags: [ 'v*' ]

jobs:
  build:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'
          cache: true

      - name: Lint with golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          args: --timeout 5m

      - name: Build API server
        run: go build -v -o bin/api ./cmd/api

      - name: Build Scheduler
        run: go build -v -o bin/scheduler ./cmd/scheduler

      - name: Run go vet
        run: go vet ./...

      - name: Check formatting
        run: |
          if gofmt -l . | grep -q .; then
            echo "Code formatting issues found"
            gofmt -l .
            exit 1
          fi

      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: binaries
          path: bin/
          retention-days: 5
```

### 3. Integration Tests Workflow

```yaml
# .github/workflows/integration-test.yml
name: Integration Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  integration:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: testuser
          POSTGRES_PASSWORD: testpass
          POSTGRES_DB: integrationdb
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'
          cache: true

      - name: Run integration tests
        run: go test -v -timeout 30m -tags=integration ./test/integration/...
        env:
          DATABASE_URL: postgres://testuser:testpass@localhost:5432/integrationdb
          REDIS_URL: redis://localhost:6379
          TEST_ENV: integration

      - name: Run race detector
        run: go test -race -short ./...
```

### 4. Security Scanning Workflow

```yaml
# .github/workflows/security.yml
name: Security Scan

on:
  push:
    branches: [ main ]
  schedule:
    - cron: '0 2 * * 0'  # Weekly

jobs:
  security:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v4

      - name: Run Gosec Security Scanner
        uses: securego/gosec@master
        with:
          args: '-no-fail -fmt sarif -out gosec-results.sarif ./...'

      - name: Upload Gosec results to GitHub Security tab
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: gosec-results.sarif

      - name: Check for hardcoded secrets
        uses: trufflesecurity/trufflehog@main
        with:
          path: ./
          base: ${{ github.event.repository.default_branch }}
          head: HEAD
          extra_args: --debug --only-verified
```

### 5. Release Workflow

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Build for multiple platforms
        run: |
          mkdir -p bin
          
          # Linux
          GOOS=linux GOARCH=amd64 go build -o bin/api-linux-amd64 ./cmd/api
          GOOS=linux GOARCH=arm64 go build -o bin/api-linux-arm64 ./cmd/api
          
          # macOS
          GOOS=darwin GOARCH=amd64 go build -o bin/api-darwin-amd64 ./cmd/api
          GOOS=darwin GOARCH=arm64 go build -o bin/api-darwin-arm64 ./cmd/api

      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: bin/*
          draft: false
          prerelease: false
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Build Optimization

### 6. Docker Build Optimization

```dockerfile
# Multi-stage build for optimized Docker image
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy only go files needed for dependency resolution
COPY go.mod go.sum ./

# Download dependencies (cached layer)
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api ./cmd/api

# Final stage - minimal image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/api .

EXPOSE 8080

CMD ["./api"]
```

### 7. Build Caching Strategy

```yaml
# Optimize build times with caching
steps:
  - uses: actions/checkout@v4
  
  - name: Set up Go with module cache
    uses: actions/setup-go@v4
    with:
      go-version: '1.25'
      cache: true                    # Caches Go modules
      cache-dependency-path: go.sum

  - name: Cache build output
    uses: actions/cache@v3
    with:
      path: |
        ~/.cache/go-build
        ~/go/pkg/mod
      key: go-build-${{ runner.os }}-${{ hashFiles('**/go.sum') }}
      restore-keys: |
        go-build-${{ runner.os }}-

  - name: Build
    run: go build -v ./...
```

## Performance Optimization

### 8. Workflow Performance Tips

```yaml
# Run independent jobs in parallel
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: go test -short ./...

  lint:
    runs-on: ubuntu-latest
    steps:
      - run: golangci-lint run

  build:
    runs-on: ubuntu-latest
    steps:
      - run: go build ./...

# Avoid unnecessary checks in PRs
on:
  push:
    branches: [ main ]  # Full checks on main
  pull_request:
    branches: [ main ]  # Faster checks on PR

# Use conditional steps
steps:
  - name: Upload coverage only on main
    if: github.ref == 'refs/heads/main'
    run: codecov upload

  - name: Deploy only on tagged release
    if: startsWith(github.ref, 'refs/tags/')
    run: ./scripts/deploy.sh
```

## Secrets Management

### 9. GitHub Secrets

```yaml
# Securely store and use secrets
jobs:
  deploy:
    runs-on: ubuntu-latest
    environment:
      name: production
      url: https://api.example.com
    
    steps:
      - uses: actions/checkout@v4

      - name: Configure credentials
        env:
          DOCKER_USERNAME: ${{ secrets.DOCKER_USERNAME }}
          DOCKER_PASSWORD: ${{ secrets.DOCKER_PASSWORD }}
          DATABASE_URL: ${{ secrets.DATABASE_URL }}
        run: |
          echo $DOCKER_PASSWORD | docker login -u $DOCKER_USERNAME --password-stdin
          # Use DATABASE_URL in deployment
```

## Notification & Status

### 10. Workflow Status Checks

```yaml
# Notify team of failures
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: go test ./...

  notify:
    needs: [test, lint, build]
    runs-on: ubuntu-latest
    if: failure()
    
    steps:
      - name: Notify Slack on failure
        uses: 8398a7/action-slack@v3
        with:
          status: ${{ job.status }}
          text: 'Pipeline failed for ${{ github.event.head_commit.message }}'
          webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

## CI/CD Configuration Checklist

- [ ] Unit tests run with coverage reporting (target 70%+)
- [ ] Integration tests run on separate schedule
- [ ] Code linting with golangci-lint configured
- [ ] Race detector enabled for concurrent code testing
- [ ] Go modules cached to speed up builds
- [ ] Security scanning enabled (gosec, trufflehog)
- [ ] Docker image built with multi-stage for optimization
- [ ] Artifacts uploaded for failed builds (for debugging)
- [ ] Deployment automated for tagged releases
- [ ] Environment secrets properly managed
- [ ] Workflow status badges added to README
- [ ] Build times monitored (target <5 minutes)
- [ ] Dependency vulnerabilities scanned (nancy, dependabot)
- [ ] Code formatting check enforced (gofmt)
- [ ] Performance benchmarks tracked over time

## Performance Guidelines

- Unit tests should complete in <2 minutes
- Integration tests should complete in <10 minutes
- Full build pipeline should complete in <5 minutes
- Docker image build should complete in <3 minutes
- Coverage reports should show trends
- Benchmark results should be published for tracking

## Common Workflow Commands

```bash
# Run tests locally like CI
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Build Docker image like CI
docker build -t platform-go:local .

# Run linter locally
golangci-lint run

# Check formatting
gofmt -l .

# Run security scanner
gosec ./...

# Check for dependencies vulnerabilities
nancy sleuth

# Build for release
GOOS=linux GOARCH=amd64 go build -o bin/api ./cmd/api
```
