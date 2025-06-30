# Makefile for Storage Control Plane

.PHONY: test test-unit test-integration test-e2e build run clean help setup install-tools dev-linux

# Default target
help:
	@echo "Available commands:"
	@echo "  make test          - Run unit + integration tests"
	@echo "  make test-unit     - Run unit tests (fast, no dependencies)"
	@echo "  make test-integration - Run integration tests (requires services)"
	@echo "  make test-e2e      - Run end-to-end tests"
	@echo "  make test-all      - Run all tests including E2E"
	@echo "  make build         - Build the application"
	@echo "  make run           - Run the application"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make dev           - Run with hot reload (air)"
	@echo "  make dev-linux     - Linux development setup"
	@echo "  make setup         - Setup development environment"
	@echo "  make install-tools - Install development tools"

# Build the application
build:
	@echo "🔨 Building Storage Control Plane..."
	@mkdir -p bin
	go build -o bin/storage-control-plane ./cmd/api

# Build for multiple platforms
build-all:
	@echo "🔨 Building for multiple platforms..."
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/storage-control-plane-linux-amd64 ./cmd/api
	GOOS=windows GOARCH=amd64 go build -o bin/storage-control-plane-windows-amd64.exe ./cmd/api
	GOOS=darwin GOARCH=amd64 go build -o bin/storage-control-plane-darwin-amd64 ./cmd/api
	GOOS=darwin GOARCH=arm64 go build -o bin/storage-control-plane-darwin-arm64 ./cmd/api
	@echo "✅ Built for Linux, Windows, macOS (Intel & Apple Silicon)"

# Run unit tests (fast, no external dependencies)
test-unit:
	@echo "🧪 Running unit tests..."
	go test -v ./test/unit/...

# Run integration tests (requires external services)
test-integration:
	@echo "🔗 Running integration tests..."
	@echo "⚠️  Make sure ClickHouse and other services are running"
	go test -v ./test/integration/...

# Run end-to-end tests
test-e2e:
	@echo "🚀 Running end-to-end tests..."
ifeq ($(OS),Windows_NT)
	powershell -ExecutionPolicy Bypass -File ./test/e2e/test_e2e.ps1
else
	./test/e2e/test_e2e.sh
endif

# Run all tests
test: test-unit test-integration
	@echo "✅ All tests completed"

# Run all tests including E2E
test-all: test-unit test-integration test-e2e
	@echo "🎉 All tests (including E2E) completed"

# Run end-to-end tests (requires running server)
test-e2e:
	@echo "🌐 Running end-to-end tests..."
	@echo "⚠️  Make sure the server is running on :8081"
	@if command -v pwsh &> /dev/null; then \
		pwsh -File test_e2e.ps1; \
	elif [ -f "test_e2e.sh" ]; then \
		chmod +x test_e2e.sh && ./test_e2e.sh; \
	else \
		echo "❌ No E2E test script found"; \
	fi

# Linux/Unix development workflow
dev-linux:
	@echo "🐧 Setting up Linux development environment..."
	@if [ -f "dev.sh" ]; then \
		chmod +x dev.sh && ./dev.sh; \
	else \
		echo "❌ dev.sh not found"; \
	fi

# Setup development environment
setup:
	@echo "🔧 Setting up development environment..."
	@mkdir -p data/rocksdb data/parquet data/wal tmp bin
	@if [ ! -f ".env" ]; then \
		cp .env.example .env; \
		echo "📄 Created .env from .env.example"; \
	fi
	go mod download
	@echo "✅ Development environment ready"

# Install development tools
install-tools:
	@echo "🛠️  Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "✅ Development tools installed"

# Run the application
run:
	@echo "🚀 Starting Storage Control Plane..."
	go run ./cmd/api

# Run with hot reload
dev:
	@echo "🔥 Starting with hot reload..."
	air

# Clean build artifacts and test data
clean:
	@echo "🧹 Cleaning up..."
	rm -rf bin/
	rm -rf tmp/
	rm -rf test_data/
	rm -f *.log
	rm -f coverage.out coverage.html
	go clean -cache -testcache

# Setup test environment
setup-test:
	@echo "🔧 Setting up test environment..."
	mkdir -p test_data/rocksdb
	mkdir -p test_data/parquet
	cp .env.test .env

# Performance tests
test-perf:
	@echo "⚡ Running performance tests..."
	go test -bench=. -benchmem ./internal/...

# Coverage report
test-coverage:
	@echo "📊 Generating coverage report..."
	go test -coverprofile=coverage.out ./internal/... ./pkg/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "📋 Coverage report generated: coverage.html"

# Lint code
lint:
	@echo "🔍 Linting code..."
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Format code
fmt:
	@echo "💅 Formatting code..."
	go fmt ./...
	goimports -w .

# Check for security issues
security:
	@echo "🔒 Checking for security issues..."
	@if command -v gosec &> /dev/null; then \
		gosec ./...; \
	else \
		echo "⚠️  gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi
