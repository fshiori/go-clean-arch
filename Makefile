.PHONY: help build build-all run-api run-worker run-cron test lint clean docker-build

# Variables
APP_NAME=go-clean-arch
BUILD_DIR=bin
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-w -s"

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the application for current platform
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./cmd/app

build-all: ## Build for all platforms (Linux, macOS, Windows)
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 ./cmd/app
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 ./cmd/app
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 ./cmd/app
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe ./cmd/app

run-api: ## Run the API server
	$(GO) run ./cmd/app/main.go --mode=api

run-worker: ## Run the worker
	$(GO) run ./cmd/app/main.go --mode=worker

run-cron: ## Run the cron scheduler
	$(GO) run ./cmd/app/main.go --mode=cron

test: ## Run tests
	$(GO) test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linter
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Installing..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run --timeout=5m

fmt: ## Format code
	$(GO) fmt ./...
	goimports -w .

vet: ## Run go vet
	$(GO) vet ./...

mod-download: ## Download dependencies
	$(GO) mod download

mod-tidy: ## Tidy dependencies
	$(GO) mod tidy

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

docker-build: ## Build Docker image
	docker build -t $(APP_NAME):latest .

docker-run-api: ## Run API in Docker
	docker run -p 8080:8080 --env-file .env $(APP_NAME):latest --mode=api

docker-run-worker: ## Run worker in Docker
	docker run --env-file .env $(APP_NAME):latest --mode=worker

docker-run-cron: ## Run cron in Docker
	docker run --env-file .env $(APP_NAME):latest --mode=cron

migrate-up: ## Run database migrations
	@echo "Running migrations..."
	# Add your migration tool command here

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	# Add your migration tool command here

.DEFAULT_GOAL := help
