.PHONY: help lint test build clean install

# Binary name
BINARY_NAME=stacksenv

# Build info
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT_SHA?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# LDFLAGS for build
LDFLAGS=-s -w -X github.com/stacksenv/cli/version.Version=$(VERSION) -X github.com/stacksenv/cli/version.CommitSHA=$(COMMIT_SHA)

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

lint-install: ## Install golangci-lint
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

test: ## Run tests with race detection
	@echo "Running tests..."
	@go test --race ./...

test-verbose: ## Run tests with verbose output
	@echo "Running tests (verbose)..."
	@go test -v --race ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) .

build-linux: ## Build binary for Linux
	@echo "Building $(BINARY_NAME) for Linux..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)-linux-amd64 .

build-darwin: ## Build binary for macOS
	@echo "Building $(BINARY_NAME) for macOS..."
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)-darwin-amd64 .
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)-darwin-arm64 .

build-windows: ## Build binary for Windows
	@echo "Building $(BINARY_NAME) for Windows..."
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)-windows-amd64.exe .

build-all: build-linux build-darwin build-windows ## Build binaries for all platforms

clean: ## Remove build artifacts
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -f $(BINARY_NAME)-*
	@rm -f coverage.out coverage.html

install: build ## Build and install the binary
	@echo "Installing $(BINARY_NAME)..."
	@go install -ldflags "$(LDFLAGS)" .

