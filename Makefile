# Variables
APP_NAME := highperf-api
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := gofmt
GOLINT := golangci-lint

# Build parameters
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"
BUILD_DIR := bin
BINARY_NAME := $(BUILD_DIR)/$(APP_NAME)

# Docker parameters
DOCKER_IMAGE := $(APP_NAME):$(VERSION)
DOCKER_IMAGE_LATEST := $(APP_NAME):latest

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) cmd/api/main.go
	@echo "Build complete: $(BINARY_NAME)"

.PHONY: build-linux
build-linux: ## Build for Linux (useful for Docker)
	@echo "Building $(APP_NAME) for Linux..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) cmd/api/main.go

.PHONY: run
run: ## Run the application
	$(GOCMD) run cmd/api/main.go

.PHONY: test
test: ## Run tests
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

.PHONY: test-coverage
test-coverage: test ## Run tests and show coverage
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: benchmark
benchmark: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

.PHONY: clean
clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

.PHONY: deps
deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

.PHONY: format
format: ## Format Go code
	$(GOFMT) -s -w .

.PHONY: lint
lint: ## Run linter
	$(GOLINT) run

.PHONY: vet
vet: ## Run go vet
	$(GOCMD) vet ./...

.PHONY: check
check: format vet lint test ## Run all checks (format, vet, lint, test)

.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) -t $(DOCKER_IMAGE_LATEST) .

.PHONY: docker-run
docker-run: ## Run Docker container
	docker run --rm -p 8080:8080 $(DOCKER_IMAGE_LATEST)

.PHONY: docker-compose-up
docker-compose-up: ## Start all services with docker-compose
	docker-compose up -d

.PHONY: docker-compose-down
docker-compose-down: ## Stop all services
	docker-compose down

.PHONY: docker-compose-logs
docker-compose-logs: ## Show logs from all services
	docker-compose logs -f

.PHONY: migrate-up
migrate-up: ## Run database migrations (up)
	@echo "Running database migrations..."
	# Add your migration command here
	# Example: migrate -path ./migrations -database "postgres://user:pass@localhost:5432/db?sslmode=disable" up

.PHONY: migrate-down
migrate-down: ## Run database migrations (down)
	@echo "Rolling back database migrations..."
	# Add your migration rollback command here

.PHONY: seed
seed: ## Seed database with test data
	@echo "Seeding database..."
	# Add your seeding command here

.PHONY: install-tools
install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

.PHONY: security-scan
security-scan: ## Run security scan
	go install github.com/securecodewarrior/github-action-add-sarif@latest
	gosec -fmt sarif -out gosec.sarif ./...

.PHONY: api-docs
api-docs: ## Generate API documentation
	@echo "Generating API documentation..."
	# Add your API docs generation command here
	# Example: swag init -g cmd/api/main.go

.PHONY: dev-setup
dev-setup: deps install-tools ## Setup development environment
	@echo "Development environment setup complete!"

.PHONY: release
release: clean check build ## Prepare release build
	@echo "Release build complete: $(BINARY_NAME)"

.PHONY: deploy-staging
deploy-staging: ## Deploy to staging
	@echo "Deploying to staging..."
	# Add your staging deployment commands here

.PHONY: deploy-production
deploy-production: ## Deploy to production
	@echo "Deploying to production..."
	# Add your production deployment commands here

# Development shortcuts
.PHONY: dev
dev: ## Run in development mode with hot reload (requires air)
	air

.PHONY: mock
mock: ## Generate mocks
	@echo "Generating mocks..."
	go generate ./...