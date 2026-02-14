.PHONY: help test test-unit test-coverage test-race test-verbose fmt fmt-check lint vet \
         build build-api clean deps \
         integration-test integration-test-db integration-test-k8s \
         ci ci-extended production-check \
         k8s-deploy k8s-delete k8s-status k8s-logs-api

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
	@echo "$(CYAN)=== Testing ===$(NC)"
	@echo "  make test              - Run all unit tests"
	@echo "  make test-unit         - Run unit tests only (faster)"
	@echo "  make test-coverage     - Run tests with coverage report"
	@echo "  make test-race         - Run tests with race detector"
	@echo ""
	@echo "$(CYAN)=== Integration Tests ===$(NC)"
	@echo "  make integration-test           - Run all integration tests (local)"
	@echo "  make integration-test-db        - Run database integration tests"
	@echo "  make integration-test-k8s       - Run Kubernetes integration tests"
	@echo "  make integration-test-docker    - Run integration tests in Docker containers"
	@echo "  make test-quick [TYPE]          - Quick test runner (user/group/project/etc)"
	@echo "                                    Examples: make test-quick TYPE=user"
	@echo ""
	@echo "$(CYAN)=== Code Quality ===$(NC)"
	@echo "  make fmt               - Format code with gofmt"
	@echo "  make fmt-check         - Check code formatting"
	@echo "  make lint              - Run golangci-lint"
	@echo "  make vet               - Run go vet"
	@echo ""
	@echo "$(CYAN)=== Build ===$(NC)"
	@echo "  make build             - Build API binary"
	@echo "  make clean             - Remove build artifacts"
	@echo "  make deps              - Download and verify dependencies"
	@echo ""
	@echo "$(CYAN)=== Kubernetes ===$(NC)"
	@echo "  make k8s-deploy        - Deploy all resources to Kubernetes"
	@echo "  make k8s-delete        - Delete all resources from Kubernetes"
	@echo "  make k8s-status        - Check Kubernetes resource status"
	@echo "  make k8s-logs-api      - Stream API server logs"
	@echo ""
	@echo "$(CYAN)=== CI/CD Pipelines ===$(NC)"
	@echo "  make ci                - Run CI pipeline (format, lint, vet, unit tests, build)"
	@echo "  make ci-extended       - Extended CI (includes code quality checks)"
	@echo "  make production-check  - Full production readiness check"

## Testing targets
test:
	@echo "$(YELLOW)Running all tests...$(NC)"
	@go test ./... -v

test-unit:
	@echo "$(YELLOW)Running unit tests (excluding integration)...$(NC)"
	@go test ./pkg/... ./internal/... -v -short

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

## Integration Tests targets
integration-test:
	@echo "$(YELLOW)Running all integration tests locally...$(NC)"
	@bash ./scripts/run-integration-tests.sh all local

integration-test-db:
	@echo "$(YELLOW)Running database integration tests...$(NC)"
	@bash ./scripts/run-integration-tests.sh db local

integration-test-k8s:
	@echo "$(YELLOW)Running Kubernetes integration tests...$(NC)"
	@bash ./scripts/run-integration-tests.sh k8s local

integration-test-docker:
	@echo "$(YELLOW)Running integration tests in Docker...$(NC)"
	@bash ./scripts/run-integration-tests.sh all docker

test-quick:
	@echo "$(YELLOW)Running quick integration test...$(NC)"
	@bash ./scripts/quick-test.sh $(TYPE)

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
build: build-api
	@echo "$(GREEN)Build complete$(NC)"

build-api:
	@echo "$(YELLOW)Building API server...$(NC)"
	@go build -o platform-api ./cmd/api
	@echo "$(GREEN)Built: platform-api$(NC)"

## Dependency targets
deps:
	@echo "$(YELLOW)Downloading dependencies...$(NC)"
	@go mod download
	@go mod verify
	@echo "$(GREEN)Dependencies verified$(NC)"

## Cleanup
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)Clean complete$(NC)"

## Kubernetes targets
k8s-deploy:
	@echo "$(YELLOW)Deploying to Kubernetes...$(NC)"
	@kubectl apply -f k8s/secret.yaml
	@kubectl apply -f k8s/postgres.yaml
	@kubectl apply -f k8s/ca.yaml
	@kubectl apply -f k8s/go-api.yaml
	@echo "$(GREEN)Kubernetes deployment complete$(NC)"

k8s-delete:
	@echo "$(YELLOW)Deleting Kubernetes resources...$(NC)"
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

## Combined targets
ci: fmt-check lint vet test-unit build
	@echo "$(GREEN)CI checks passed$(NC)"

ci-extended: fmt-check lint vet test-unit integration-test-docker build
	@echo "$(GREEN)Extended CI pipeline passed$(NC)"

production-check: fmt-check lint vet test-coverage build integration-test-docker
	@echo "$(GREEN)Production readiness check complete!$(NC)"
	@echo "$(GREEN)Coverage report: coverage.html$(NC)"