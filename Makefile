# Storage Control Plane - Go Monolith Makefile

.PHONY: help dev build test clean fmt vet deps air-install

# Default target
help:
	@echo "🚀 Storage Control Plane - Available Commands:"
	@echo ""
	@echo "  dev          - Start development server with hot reload (air)"
	@echo "  build        - Build the application binary"
	@echo "  run          - Run the application directly"
	@echo "  test         - Run all tests"
	@echo "  test-verbose - Run tests with verbose output"
	@echo "  clean        - Clean build artifacts"
	@echo "  fmt          - Format all Go files"
	@echo "  vet          - Run Go vet (static analysis)"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  air-install  - Install Air for hot reload"
	@echo "  setup        - Full development setup"
	@echo ""

# Development server with hot reload
dev:
	@echo "🔥 Starting development server with hot reload..."
	@air

# Build the application
build:
	@echo "🏗️ Building Storage Control Plane..."
	@go build -o storage-control-plane .

# Run without hot reload
run:
	@echo "🚀 Starting Storage Control Plane..."
	@go run .

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test ./...

# Run tests with verbose output
test-verbose:
	@echo "🧪 Running tests (verbose)..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf tmp/
	@rm -f storage-control-plane
	@rm -f storage-control-plane.exe
	@rm -f build-errors.log

# Format Go code
fmt:
	@echo "🎨 Formatting Go code..."
	@go fmt ./...

# Run static analysis
vet:
	@echo "🔍 Running static analysis..."
	@go vet ./...

# Download and tidy dependencies
deps:
	@echo "📦 Managing dependencies..."
	@go mod download
	@go mod tidy

# Install Air for hot reload
air-install:
	@echo "💨 Installing Air for hot reload..."
	@go install github.com/air-verse/air@latest

# Full development setup
setup: air-install deps
	@echo "🛠️ Setting up development environment..."
	@cp .env.example .env 2>/dev/null || echo ".env already exists"
	@echo "✅ Setup complete! Run 'make dev' to start development server"

# Production build
build-prod:
	@echo "🏭 Building for production..."
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o storage-control-plane .

# Health check all services
health-check:
	@echo "🏥 Checking service health..."
	@curl -s http://localhost:8080/health | jq . || echo "Auth Gateway (8080) not responding"
	@curl -s http://localhost:8000/health | jq . || echo "Tenant Node (8000) not responding"
	@curl -s http://localhost:8081/health | jq . || echo "Operation Node (8081) not responding"
	@curl -s http://localhost:8082/health | jq . || echo "CBO Engine (8082) not responding"
	@curl -s http://localhost:8083/health | jq . || echo "Metadata Catalog (8083) not responding"
	@curl -s http://localhost:8084/health | jq . || echo "Monitoring (8084) not responding"
	@curl -s http://localhost:8085/health | jq . || echo "Query Interpreter (8085) not responding"

# Demo API calls
demo:
	@echo "🎭 Running API demos..."
	@echo "1. Auth login:"
	@curl -X POST http://localhost:8080/auth/login -H "Content-Type: application/json" -d '{"username":"admin","password":"password"}' | jq .
	@echo ""
	@echo "2. Execute distributed query:"
	@curl -X POST http://localhost:8081/query/execute -H "Content-Type: application/json" -d '{"query":"SELECT * FROM orders LIMIT 10"}' | jq .
	@echo ""
	@echo "3. Parse SQL query:"
	@curl -X POST http://localhost:8085/parse/sql -H "Content-Type: application/json" -d '{"query":"SELECT customer_id, SUM(amount) FROM orders GROUP BY customer_id"}' | jq .
