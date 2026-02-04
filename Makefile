# lazyas - Lazy Agent Skills package manager
# Build configuration

VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)
DIST_DIR := dist

.PHONY: help build test test-coverage run install clean check lint release deps browse

# Default target - show help
help:
	@echo "lazyas - Lazy Agent Skills package manager"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  help          Show this help message (default)"
	@echo "  build         Build the binary for current platform"
	@echo "  test          Run all tests"
	@echo "  test-coverage Run tests with coverage report"
	@echo "  check         Format code and run vet"
	@echo "  lint          Check formatting without fixing (for CI)"
	@echo "  install       Install to GOPATH/bin"
	@echo "  clean         Remove build artifacts"
	@echo "  release       Build release (infers version from HEAD tag, or use TAG=)"
	@echo ""
	@echo "Examples:"
	@echo "  make build                    # Build dev binary"
	@echo "  make build VERSION=v0.1.0     # Build with specific version"
	@echo "  make release                  # Build release (HEAD must be tagged)"
	@echo "  make release TAG=v0.1.0       # Build release with explicit tag"

# Build the binary
build:
	go build -ldflags="$(LDFLAGS)" -o bin/lazyas ./cmd/lazyas/

# Run all tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build and run
run: build
	./bin/lazyas help

# Install to GOPATH/bin
install:
	go install -ldflags="$(LDFLAGS)" ./cmd/lazyas/

# Clean build artifacts
clean:
	rm -rf bin/ $(DIST_DIR)/ coverage.out coverage.html

# Format and vet (fixes formatting)
check:
	gofmt -s -w .
	go vet ./...

# Lint only (check formatting without fixing, for CI)
lint:
	@test -z "$$(gofmt -s -l .)" || (echo "Formatting issues found:"; gofmt -s -l .; exit 1)
	go vet ./...

# Download and tidy dependencies
deps:
	go mod tidy
	go mod download

# Development helpers
browse:
	go run ./cmd/lazyas browse

# Build release from a git tag
# Usage: make release          (infers version from current HEAD if tagged)
#        make release TAG=v0.1.0  (explicit tag)
# Production build options:
#   CGO_ENABLED=0  - static binary, no C dependencies
#   -trimpath      - remove file paths from binary
#   -s -w          - strip symbol table and debug info
release:
	$(eval TAG ?= $(shell git describe --tags --exact-match 2>/dev/null))
	@if [ -z "$(TAG)" ]; then \
		echo "Error: HEAD is not tagged. Either:"; \
		echo "  1. Checkout a tagged commit: git checkout v0.1.0"; \
		echo "  2. Specify tag explicitly: make release TAG=v0.1.0"; \
		exit 1; \
	fi
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Error: Working tree is dirty. Commit or stash changes first."; \
		git status --short; \
		exit 1; \
	fi
	@echo "Building release $(TAG)..."
	@mkdir -p $(DIST_DIR)
	@echo "Building linux/amd64..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w -X main.version=$(TAG)" -o $(DIST_DIR)/lazyas-$(TAG)-linux-amd64 ./cmd/lazyas/
	@echo "Building linux/arm64..."
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w -X main.version=$(TAG)" -o $(DIST_DIR)/lazyas-$(TAG)-linux-arm64 ./cmd/lazyas/
	@echo ""
	@echo "Release $(TAG) built successfully:"
	@ls -lh $(DIST_DIR)/lazyas-$(TAG)-*
