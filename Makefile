# ClippyCLI Makefile

.PHONY: build install clean test fmt vet help

# Default target
all: build

# Build the binary
build:
	@echo "Building clippycli..."
	@go build -o clippycli .

# Install the binary to GOPATH/bin
install:
	@echo "Installing clippycli..."
	@go install .

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f clippycli
	@go clean

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Run all checks
check: fmt vet test

# Show help
help:
	@echo "Available targets:"
	@echo "  build    - Build the clippycli binary"
	@echo "  install  - Install clippycli to GOPATH/bin"
	@echo "  clean    - Remove build artifacts"
	@echo "  test     - Run tests"
	@echo "  fmt      - Format code"
	@echo "  vet      - Run go vet"
	@echo "  check    - Run fmt, vet, and test"
	@echo "  help     - Show this help message" 