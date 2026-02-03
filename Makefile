.PHONY: test test-unit test-integration test-verbose test-coverage test-race fmt lint vet build help \
         skills-lint skills-compile skills-migration skills-validate skills-test-docker \
         skills-test-k8s skills-coverage skills-test-race skills-validate-push \
         ci-extended production-check docker-integration

# Project paths
PROJECT_ROOT := $(shell pwd)
SKILLS_DIR := $(PROJECT_ROOT)/.github/skills

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
CYAN := \033[0;36m
NC := \033[0m # No Color

help:
	@echo "$(GREEN)Platform-Go Make Commands:$(NC)"
	@echo ""
	@echo "$(CYAN)=== Basic Testing ===$(NC)"
	@echo "  make test              - Run all tests"
	@echo "  make test-unit         - Run unit tests only"
	@echo "  make test-verbose      - Run tests with verbose output"
	@echo "  make test-coverage     - Run tests with coverage report"
	@echo "  make test-race         - Run tests with race detector"
	@echo "  make coverage-html     - Generate HTML coverage report"
	@echo ""
	@echo "$(CYAN)=== Code Quality ===$(NC)"
	@echo "  make fmt               - Format code with gofmt"
	@echo "  make lint              - Run linter (golangci-lint)"
	@echo "  make vet               - Run go vet"
	@echo ""
	@echo "$(CYAN)=== Build & Deploy ===$(NC)"
	@echo "  make build             - Build API and scheduler binaries"
	@echo "  make build-api         - Build API binary only"
	@echo "  make build-scheduler   - Build scheduler binary only"
	@echo "  make clean             - Remove build artifacts"
	@echo "  make deps              - Download and verify dependencies"
	@echo ""
	@echo "$(CYAN)=== Kubernetes ===$(NC)"
	@echo "  make k8s-deploy        - Deploy all resources to Kubernetes"
	@echo "  make k8s-delete        - Delete all resources from Kubernetes"
	@echo "  make k8s-status        - Check Kubernetes resource status"
	@echo "  make k8s-logs-api      - Stream API server logs"
	@echo "  make k8s-logs-scheduler - Stream scheduler logs"
	@echo ""
	@echo "$(CYAN)=== Integration Tests ===$(NC)"
	@echo "  make test-integration  - Run integration tests"
	@echo "  make test-integration-quick - Run quick integration tests"
	@echo "  make test-clean        - Clean test environment"
	@echo "  make docker-integration - Run Docker-based integration tests (Skills)"
	@echo ""
	@echo "$(CYAN)=== Skills-Based Commands ===$(NC)"
	@echo "  make skills-lint       - Production standards lint check (golang-production-standards)"
	@echo "  make skills-compile    - Compile check (golang-production-standards)"
	@echo "  make skills-migration  - Database migration validation (database-best-practices)"
	@echo "  make skills-validate   - Pre-commit validation (code-validation-standards)"
	@echo "  make skills-validate-push - Push validation (code-validation-standards)"
	@echo "  make skills-coverage   - Code coverage report (testing-best-practices)"
	@echo "  make skills-test-race  - Race detector test (testing-best-practices)"
	@echo "  make skills-test-docker - Docker integration tests (integration-testing)"
	@echo "  make skills-test-k8s   - Kubernetes integration tests (integration-testing)"
	@echo ""
	@echo "$(CYAN)=== Combined Pipelines ===$(NC)"
	@echo "  make ci                - Run CI pipeline (format check, lint, vet, test, build)"
	@echo "  make ci-extended       - Extended CI (includes skills validation and tests)"
	@echo "  make production-check  - Full production readiness check"
	@echo "  make local-test        - Run local tests with coverage report"
	@echo "  make all               - Run full pipeline (CI + K8s deploy)"

## Testing targets
test:
	@echo "$(YELLOW)Running all tests...$(NC)"
	@go test ./... -v

test-unit:
	@echo "$(YELLOW)Running unit tests (excluding integration)...$(NC)"
	@go test ./pkg/... ./internal/... -v -short

test-verbose:
	@echo "$(YELLOW)Running tests with verbose output...$(NC)"
	@go test ./pkg/... ./internal/... -v -count=1

test-coverage:
	@echo "$(YELLOW)Running tests with coverage...$(NC)"
	@go test ./... -v -coverprofile=coverage.out -covermode=atomic
	@echo "$(GREEN)Coverage report: coverage.out$(NC)"
	@go tool cover -func=coverage.out | tail -1

coverage-html: test-coverage
	@echo "$(YELLOW)Generating HTML coverage report...$(NC)"
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Open coverage.html in your browser$(NC)"

test-race:
	@echo "$(YELLOW)Running tests with race detector...$(NC)"
	@go test ./... -v -race

test-integration:
	@echo "$(YELLOW)Running integration tests...$(NC)"
	@echo "$(YELLOW)Note: Requires PostgreSQL and Kubernetes cluster$(NC)"
	@cd test/integration && cp .env.test ../../.env || true
	@go test -v -timeout 30m ./test/integration/...
	@echo "$(GREEN)Integration tests complete$(NC)"

test-integration-quick:
	@echo "$(YELLOW)Running quick integration tests (skipping slow tests)...$(NC)"
	@cd test/integration && cp .env.test ../../.env || true
	@go test -v -timeout 15m -short ./test/integration/...
	@echo "$(GREEN)Quick integration tests complete$(NC)"

test-integration-k8s:
	@echo "$(YELLOW)Running K8s integration tests only...$(NC)"
	@cd test/integration && cp .env.test ../../.env || true
	@go test -v -timeout 20m ./test/integration/ -run K8s
	@echo "$(GREEN)K8s integration tests complete$(NC)"

test-clean:
	@echo "$(YELLOW)Cleaning test environment...$(NC)"
	@kubectl get ns | grep test-integration | awk '{print $$1}' | xargs -r kubectl delete ns || true
	@dropdb platform_test 2>/dev/null || true
	@createdb platform_test 2>/dev/null || true
	@echo "$(GREEN)Test environment cleaned$(NC)"

## Code quality targets
fmt:
	@echo "$(YELLOW)Formatting code...$(NC)"
	@gofmt -w .
	@echo "$(GREEN)Code formatted$(NC)"

fmt-check:
	@echo "$(YELLOW)Checking code format...$(NC)"
	@if gofmt -l . | grep -q .; then \
		echo "$(RED)Format issues found:$(NC)"; \
		gofmt -l .; \
		exit 1; \
	else \
		echo "$(GREEN)Code is properly formatted$(NC)"; \
	fi

lint:
	@echo "$(YELLOW)Running golangci-lint...$(NC)"
	@if command -v golangci-lint > /dev/null 2>&1; then \
		golangci-lint run ./... --timeout=5m; \
	elif [ -f $(HOME)/go/bin/golangci-lint ]; then \
		$(HOME)/go/bin/golangci-lint run ./... --timeout=5m; \
	elif [ -f $(HOME)/bin/golangci-lint ]; then \
		$(HOME)/bin/golangci-lint run ./... --timeout=5m; \
	else \
		echo "$(RED)golangci-lint not found. Installing...$(NC)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		$(HOME)/go/bin/golangci-lint run ./... --timeout=5m; \
	fi

vet:
	@echo "$(YELLOW)Running go vet...$(NC)"
	@go vet ./...

## Build targets
build: build-api build-scheduler
	@echo "$(GREEN)Build complete$(NC)"

build-api:
	@echo "$(YELLOW)Building API server...$(NC)"
	@go build -o platform-api ./cmd/api
	@echo "$(GREEN)Built: platform-api$(NC)"

build-scheduler:
	@echo "$(YELLOW)Building scheduler...$(NC)"
	@go build -o platform-scheduler ./cmd/scheduler
	@echo "$(GREEN)Built: platform-scheduler$(NC)"

## Dependency targets
deps:
	@echo "$(YELLOW)Downloading dependencies...$(NC)"
	@go mod download
	@go mod verify
	@echo "$(GREEN)Dependencies verified$(NC)"

## Cleanup
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -f platform-api platform-scheduler
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)Clean complete$(NC)"

## Kubernetes targets
k8s-deploy:
	@echo "$(YELLOW)Deploying to Kubernetes...$(NC)"
	@kubectl apply -f k8s/secret.yaml
	@kubectl apply -f k8s/postgres.yaml
	@kubectl apply -f k8s/ca.yaml
	@kubectl apply -f k8s/go-api.yaml
	@kubectl apply -f k8s/go-scheduler.yaml
	@echo "$(GREEN)Kubernetes deployment complete$(NC)"

k8s-delete:
	@echo "$(YELLOW)Deleting Kubernetes resources...$(NC)"
	@kubectl delete -f k8s/go-scheduler.yaml || true
	@kubectl delete -f k8s/go-api.yaml || true
	@kubectl delete -f k8s/ca.yaml || true
	@kubectl delete -f k8s/postgres.yaml || true
	@kubectl delete -f k8s/storage.yaml || true
	@kubectl delete -f k8s/secret.yaml || true
	@echo "$(GREEN)Kubernetes resources deleted$(NC)"

k8s-status:
	@echo "$(YELLOW)Checking Kubernetes resources...$(NC)"
	@kubectl get deployments
	@echo ""
	@kubectl get pods
	@echo ""
	@kubectl get svc

k8s-logs-api:
	@kubectl logs -f deployment/go-api --tail=100

k8s-logs-scheduler:
	@kubectl logs -f deployment/go-scheduler --tail=100

## Combined targets
ci: fmt-check lint vet test-unit build
	@echo "$(GREEN)CI checks passed$(NC)"

local-test: clean deps test coverage-html
	@echo "$(GREEN)Local testing complete. See coverage.html for details$(NC)"

all: ci k8s-deploy
	@echo "$(GREEN)Full build and deploy pipeline complete$(NC)"
## Skills-Based Targets (from .github/skills)

### Code Quality & Standards (golang-production-standards)
skills-lint:
	@echo "$(CYAN)Running production standards lint check...$(NC)"
	@bash $(SKILLS_DIR)/golang-production-standards/scripts/lint-check.sh

skills-compile:
	@echo "$(CYAN)Running compile check...$(NC)"
	@bash $(SKILLS_DIR)/golang-production-standards/scripts/compile-check.sh

### Database (database-best-practices)
skills-migration:
	@echo "$(CYAN)Validating database migrations...$(NC)"
	@bash $(SKILLS_DIR)/database-best-practices/scripts/migration-check.sh

### Code Validation (code-validation-standards)
skills-validate:
	@echo "$(CYAN)Running pre-commit validation...$(NC)"
	@bash $(SKILLS_DIR)/code-validation-standards/scripts/pre-commit-validate.sh

skills-validate-push:
	@echo "$(CYAN)Running push validation...$(NC)"
	@bash $(SKILLS_DIR)/code-validation-standards/scripts/validate-before-push.sh

### Testing (testing-best-practices)
skills-coverage:
	@echo "$(CYAN)Generating code coverage report...$(NC)"
	@bash $(SKILLS_DIR)/testing-best-practices/scripts/coverage-html.sh

skills-test-race:
	@echo "$(CYAN)Running race detector tests...$(NC)"
	@bash $(SKILLS_DIR)/testing-best-practices/scripts/test-with-race.sh

### Integration Testing (integration-testing)
skills-test-docker:
	@echo "$(CYAN)Running Docker-based integration tests...$(NC)"
	@bash $(SKILLS_DIR)/integration-testing/scripts/docker-db-integration-test.sh

skills-test-k8s:
	@echo "$(CYAN)Running Kubernetes integration tests...$(NC)"
	@bash $(SKILLS_DIR)/integration-testing/scripts/docker-k8s-integration-test.sh

docker-integration: skills-test-docker
	@echo "$(GREEN)Docker integration tests complete$(NC)"

### CI/CD Pipeline (cicd-pipeline-optimization)
skills-test-all:
	@echo "$(CYAN)Running all CI/CD integration tests...$(NC)"
	@bash $(SKILLS_DIR)/cicd-pipeline-optimization/scripts/run-integration-tests.sh

## Extended Pipelines

# Extended CI that includes skills validation and Docker tests
ci-extended: fmt-check lint vet skills-compile skills-validate test-unit build skills-test-docker
	@echo "$(GREEN)Extended CI pipeline passed$(NC)"

# Full production readiness check
production-check: skills-lint skills-compile skills-migration skills-validate-push skills-coverage test coverage-html build
	@echo "$(GREEN)Production readiness check complete!$(NC)"
	@echo "$(GREEN)Coverage report: coverage.html$(NC)"

.PHONY: skills-lint skills-compile skills-migration skills-validate skills-validate-push \
        skills-coverage skills-test-race skills-test-docker skills-test-k8s skills-test-all \
        docker-integration ci-extended production-check