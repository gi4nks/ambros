# Build parameters
BINARY_NAME=ambros
VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_DIR=bin
GOARCH=amd64
CGO_ENABLED=0

# Detect current OS for default build
GOOS ?= $(shell go env GOOS)

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Source files - Updated to correct path
MAIN_FILE=cmd/main.go
MIGRATE_FILE=cmd/migrate/main.go
SRC_DIRS=./...

# Test flags
TEST_FLAGS=-race -v
BENCH_FLAGS=-bench=. -benchmem

# Build flags
LDFLAGS=-ldflags "-X main.Version=${VERSION} -s -w"
BUILD_FLAGS=-trimpath $(LDFLAGS)

# Docker parameters
DOCKER_IMAGE=ambros
DOCKER_TAG=$(VERSION)

# Coverage parameters
COVERAGE_DIR=coverage
MIN_COVERAGE=80

# Migration parameters
OLD_DB?=$(HOME)/.ambros/ambros.db
NEW_DB?=$(HOME)/.ambros/ambros_new.db

.PHONY: all build clean test coverage deps tidy lint run help migrate docker

all: clean deps lint test build ## Run the most common development targets

build: ## Build the binary
	@printf "Building $(BINARY_NAME) for $(GOOS)...\n"
	@mkdir -p $(BUILD_DIR)
	@if [ ! -f "$(MAIN_FILE)" ]; then \
		printf "Error: Main source file not found at $(MAIN_FILE)\n"; \
		exit 1; \
	fi
	CGO_ENABLED=$(CGO_ENABLED) GOARCH=$(GOARCH) GOOS=$(GOOS) $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@printf "Build complete: $(BUILD_DIR)/$(BINARY_NAME)\n"

build-all: ## Build for all platforms (darwin, linux, windows)
	@printf "Building for all platforms...\n"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOARCH=$(GOARCH) GOOS=darwin $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin $(MAIN_FILE)
	CGO_ENABLED=$(CGO_ENABLED) GOARCH=$(GOARCH) GOOS=linux $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(MAIN_FILE)
	CGO_ENABLED=$(CGO_ENABLED) GOARCH=$(GOARCH) GOOS=windows $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows.exe $(MAIN_FILE)
	@printf "Multi-platform build complete\n"

clean: ## Clean build artifacts
	@printf "Cleaning...\n"
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	@rm -f coverage.out
	@printf "Clean complete\n"

test: ## Run unit tests
	@printf "Running unit tests...\n"
	@$(GOTEST) $(TEST_FLAGS) -short $(SRC_DIRS)

test-integration: ## Run integration tests
	@printf "Running integration tests...\n"
	@$(GOTEST) $(TEST_FLAGS) -run Integration $(SRC_DIRS)

test-all: test test-integration ## Run all tests

benchmark: ## Run benchmarks
	@printf "Running benchmarks...\n"
	@$(GOTEST) $(BENCH_FLAGS) $(SRC_DIRS)

coverage: ## Run tests with coverage
	@printf "Running tests with coverage...\n"
	@mkdir -p $(COVERAGE_DIR)
	@$(GOTEST) $(TEST_FLAGS) -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic $(SRC_DIRS)
	@$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@$(GOCMD) tool cover -func=$(COVERAGE_DIR)/coverage.out | tee $(COVERAGE_DIR)/coverage.txt
	@printf "Checking minimum coverage threshold ($(MIN_COVERAGE)%%)...\n"
	@coverage=$$(go tool cover -func=$(COVERAGE_DIR)/coverage.out | grep total | awk '{print $$3}' | tr -d '%'); \
	if [ $${coverage%.*} -lt $(MIN_COVERAGE) ]; then \
		printf "Coverage $$coverage%% is below minimum $(MIN_COVERAGE)%%\n"; \
		exit 1; \
	fi

deps: ## Download dependencies
	@printf "Downloading dependencies...\n"
	@$(GOGET) -v -t -d $(SRC_DIRS)

tidy: ## Tidy up module dependencies
	@printf "Tidying up modules...\n"
	@$(GOMOD) tidy

lint: ## Run linters
	@printf "Running linters...\n"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		printf "Warning: golangci-lint is not installed. Skipping linting.\n"; \
		printf "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest\n"; \
	fi

# Docker targets
docker-build: ## Build Docker image
	@printf "Building Docker image...\n"
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: docker-build ## Run Docker container
	@printf "Running Docker container...\n"
	docker run --rm -it $(DOCKER_IMAGE):$(DOCKER_TAG)

# Development tools
dev-deps: ## Install development dependencies
	@printf "Installing development dependencies...\n"
	@$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GOGET) github.com/golang/mock/mockgen@latest
	@$(GOGET) github.com/vektra/mockery/v2@latest
	@$(GOGET) golang.org/x/tools/cmd/goimports@latest

generate-mocks: ## Generate mocks for testing
	@printf "Generating mocks...\n"
	@if ! command -v mockgen >/dev/null 2>&1; then \
		printf "Error: mockgen is not installed. Install with: go install github.com/golang/mock/mockgen@latest\n"; \
		exit 1; \
	fi
	@mkdir -p internal/repos/mocks internal/scheduler/mocks internal/chain/mocks
	@if [ -f "internal/repos/repository.go" ]; then \
		mockgen -source=internal/repos/repository.go -destination=internal/repos/mocks/mock_repository.go -package=mocks; \
	else \
		printf "Warning: internal/repos/repository.go not found, skipping mock generation\n"; \
	fi
	@if [ -f "internal/scheduler/scheduler.go" ]; then \
		mockgen -source=internal/scheduler/scheduler.go -destination=internal/scheduler/mocks/mock_scheduler.go -package=mocks; \
	else \
		printf "Warning: internal/scheduler/scheduler.go not found, skipping mock generation\n"; \
	fi
	@if [ -f "internal/chain/chain.go" ]; then \
		mockgen -source=internal/chain/chain.go -destination=internal/chain/mocks/mock_chain.go -package=mocks; \
	else \
		printf "Warning: internal/chain/chain.go not found, skipping mock generation\n"; \
	fi

fmt: ## Format code
	@printf "Formatting code...\n"
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		gofmt -w .; \
		printf "Warning: goimports not found, used gofmt instead\n"; \
	fi

# Component builds
build-components: ## Build all components that exist
	@printf "Building available components...\n"
	@$(MAKE) -s api || true
	@$(MAKE) -s templates || true
	@$(MAKE) -s scheduler || true
	@$(MAKE) -s analytics || true

api: ## Build API server
	@if [ -f "cmd/api/main.go" ]; then \
		printf "Building API server...\n"; \
		mkdir -p $(BUILD_DIR); \
		$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-api cmd/api/main.go; \
	else \
		printf "Warning: cmd/api/main.go not found, skipping API build\n"; \
	fi

templates: ## Build template generator
	@if [ -f "cmd/template/main.go" ]; then \
		printf "Building template generator...\n"; \
		mkdir -p $(BUILD_DIR); \
		$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-template cmd/template/main.go; \
	else \
		printf "Warning: cmd/template/main.go not found, skipping template build\n"; \
	fi

scheduler: ## Build scheduler
	@if [ -f "cmd/scheduler/main.go" ]; then \
		printf "Building scheduler...\n"; \
		mkdir -p $(BUILD_DIR); \
		$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-scheduler cmd/scheduler/main.go; \
	else \
		printf "Warning: cmd/scheduler/main.go not found, skipping scheduler build\n"; \
	fi

analytics: ## Build analytics tool
	@if [ -f "cmd/analytics/main.go" ]; then \
		printf "Building analytics tool...\n"; \
		mkdir -p $(BUILD_DIR); \
		$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-analytics cmd/analytics/main.go; \
	else \
		printf "Warning: cmd/analytics/main.go not found, skipping analytics build\n"; \
	fi

# Migration targets
build-migrate: ## Build migration tool
	@printf "Building migration tool...\n"
	@if [ ! -f "$(MIGRATE_FILE)" ]; then \
		printf "Error: Migration source file not found at $(MIGRATE_FILE)\n"; \
		exit 1; \
	fi
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/migrate $(MIGRATE_FILE)

migrate: build-migrate ## Run database migration
	@printf "Migrating database from $(OLD_DB) to $(NEW_DB)...\n"
	@if [ ! -f "$(OLD_DB)" ]; then \
		printf "Error: Source database not found at $(OLD_DB)\n"; \
		exit 1; \
	fi
	@mkdir -p $$(dirname "$(NEW_DB)")
	@$(BUILD_DIR)/migrate -src "$(OLD_DB)" -dst "$(NEW_DB)"

migrate-backup: ## Backup database before migration
	@printf "Creating backup of old database...\n"
	@if [ -f "$(OLD_DB)" ]; then \
		cp "$(OLD_DB)" "$(OLD_DB).backup-$$(date +%Y%m%d_%H%M%S)"; \
		printf "Backup created successfully\n"; \
	else \
		printf "No existing database found to backup\n"; \
	fi

migrate-all: migrate-backup migrate ## Backup and migrate database

# Versioning
version: ## Display version information
	@printf "Version: $(VERSION)\n"

# Help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help