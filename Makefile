.PHONY: build test clean run help validate

# Build variables
BINARY_NAME=coreos-newsletter
VERSION=0.1.0
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags="-X main.Version=$(VERSION)"

# Directories
BUILD_DIR=dist
CMD_DIR=cmd/coreos-newsletter

help:
	@echo "Fedora CoreOS Newsletter Generator"
	@echo ""
	@echo "Available targets:"
	@echo "  build       - Build the binary"
	@echo "  test        - Run tests"
	@echo "  clean       - Remove build artifacts"
	@echo "  run         - Run with validate command"
	@echo "  install     - Install binary to GOBIN"
	@echo "  dev         - Build for development"
	@echo "  help        - Show this help message"

build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "✓ Built $(BINARY_NAME) to $(BUILD_DIR)/"

dev:
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) ./$(CMD_DIR)
	@echo "✓ Built $(BINARY_NAME)"

test:
	$(GO) test $(GOFLAGS) ./...

clean:
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)
	$(GO) clean
	@echo "✓ Cleaned build artifacts"

run: build
	./$(BUILD_DIR)/$(BINARY_NAME) validate

install: build
	$(GO) install $(LDFLAGS) ./$(CMD_DIR)
	@echo "✓ Installed $(BINARY_NAME)"

fmt:
	$(GO) fmt ./...
	@echo "✓ Formatted code"

vet:
	$(GO) vet ./...
	@echo "✓ Vet passed"

lint: fmt vet
	@echo "✓ Linting passed"
