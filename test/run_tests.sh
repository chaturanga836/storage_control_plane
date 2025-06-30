#!/bin/bash

# Test Runner Script for Storage Control Plane
# This script provides an easy way to run different types of tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test directories
UNIT_DIR="./test/unit"
INTEGRATION_DIR="./test/integration"
E2E_DIR="./test/e2e"

# Functions
print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Load test environment
load_test_env() {
    if [ -f ".env.test" ]; then
        export $(cat .env.test | grep -v '^#' | xargs)
        print_success "Loaded test environment from .env.test"
    else
        print_warning "No .env.test file found, using defaults"
    fi
}

# Check if services are running
check_clickhouse() {
    if command -v clickhouse-client &> /dev/null; then
        if clickhouse-client --query "SELECT 1" &> /dev/null; then
            print_success "ClickHouse is running"
            return 0
        fi
    fi
    print_error "ClickHouse is not running or not accessible"
    return 1
}

# Run unit tests
run_unit_tests() {
    print_header "Running Unit Tests"
    
    if [ ! -d "$UNIT_DIR" ]; then
        print_error "Unit test directory not found: $UNIT_DIR"
        return 1
    fi
    
    go test -v $UNIT_DIR/... -timeout=5m
    
    if [ $? -eq 0 ]; then
        print_success "Unit tests passed"
    else
        print_error "Unit tests failed"
        return 1
    fi
}

# Run integration tests
run_integration_tests() {
    print_header "Running Integration Tests"
    
    if [ ! -d "$INTEGRATION_DIR" ]; then
        print_error "Integration test directory not found: $INTEGRATION_DIR"
        return 1
    fi
    
    # Check if required services are running
    if ! check_clickhouse; then
        print_error "Integration tests require ClickHouse to be running"
        print_warning "Start ClickHouse with: docker run -d --name clickhouse-test -p 9000:9000 clickhouse/clickhouse-server"
        return 1
    fi
    
    go test -v $INTEGRATION_DIR/... -timeout=10m
    
    if [ $? -eq 0 ]; then
        print_success "Integration tests passed"
    else
        print_error "Integration tests failed"
        return 1
    fi
}

# Run e2e tests
run_e2e_tests() {
    print_header "Running End-to-End Tests"
    
    if [ ! -f "$E2E_DIR/test_e2e.sh" ]; then
        print_error "E2E test script not found: $E2E_DIR/test_e2e.sh"
        return 1
    fi
    
    chmod +x $E2E_DIR/test_e2e.sh
    $E2E_DIR/test_e2e.sh
    
    if [ $? -eq 0 ]; then
        print_success "E2E tests passed"
    else
        print_error "E2E tests failed"
        return 1
    fi
}

# Generate test coverage
generate_coverage() {
    print_header "Generating Test Coverage"
    
    mkdir -p coverage
    
    # Run tests with coverage
    go test -v $UNIT_DIR/... -coverprofile=coverage/unit.out -timeout=5m
    go test -v $INTEGRATION_DIR/... -coverprofile=coverage/integration.out -timeout=10m
    
    # Merge coverage files
    echo "mode: set" > coverage/total.out
    tail -n +2 coverage/unit.out >> coverage/total.out
    tail -n +2 coverage/integration.out >> coverage/total.out
    
    # Generate HTML report
    go tool cover -html=coverage/total.out -o coverage/coverage.html
    
    # Display coverage summary
    go tool cover -func=coverage/total.out | tail -1
    
    print_success "Coverage report generated at coverage/coverage.html"
}

# Clean test artifacts
clean_test_artifacts() {
    print_header "Cleaning Test Artifacts"
    
    rm -rf coverage/
    rm -rf test/tmp/
    rm -rf test/testdata/temp_*
    
    print_success "Test artifacts cleaned"
}

# Help function
show_help() {
    echo "Test Runner for Storage Control Plane"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  unit          Run unit tests only"
    echo "  integration   Run integration tests only"
    echo "  e2e           Run end-to-end tests only"
    echo "  all           Run all tests"
    echo "  coverage      Generate test coverage report"
    echo "  clean         Clean test artifacts"
    echo "  help          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 unit                    # Run unit tests"
    echo "  $0 integration            # Run integration tests"
    echo "  $0 all                    # Run all tests"
    echo "  $0 coverage               # Generate coverage report"
}

# Main execution
main() {
    case "${1:-all}" in
        "unit")
            load_test_env
            run_unit_tests
            ;;
        "integration")
            load_test_env
            run_integration_tests
            ;;
        "e2e")
            load_test_env
            run_e2e_tests
            ;;
        "all")
            load_test_env
            run_unit_tests && run_integration_tests && run_e2e_tests
            ;;
        "coverage")
            load_test_env
            generate_coverage
            ;;
        "clean")
            clean_test_artifacts
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            print_error "Unknown command: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# Execute main function
main "$@"
