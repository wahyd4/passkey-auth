.PHONY: help build run dev test clean docker-build docker-run k8s-deploy k8s-undeploy

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the Go binary
	@echo "🔨 Building..."
	@go build -o bin/passkey-auth .

run: build ## Run the application locally
	@echo "🚀 Running..."
	@./bin/passkey-auth

dev: ## Run in development mode with hot reload
	@echo "🔧 Starting development server..."
	@./scripts/dev.sh

test: ## Run tests
	@echo "🧪 Running tests..."
	@go test -v ./...

clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	@rm -rf bin/
	@rm -f *.db

docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	@docker build -t passkey-auth:latest .

docker-run: docker-build ## Run with Docker Compose
	@echo "🐳 Starting with Docker Compose..."
	@mkdir -p data
	@docker-compose up -d
	@echo "Access the application at http://localhost:8080"

deps: ## Download dependencies
	@echo "📦 Downloading dependencies..."
	@go mod download
	@go mod tidy

fmt: ## Format code
	@echo "🎨 Formatting code..."
	@go fmt ./...

lint: ## Run linter with auto-fix (excluding test files)
	@echo "🔍 Running linter with auto-fix (excluding test files)..."
	@golangci-lint run --fix --tests=false

security: ## Run security scan
	@echo "🔒 Running security scan..."
	@gosec ./...
