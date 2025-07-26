# ztictl Makefile
# Cross-platform build automation

.PHONY: all build clean test help install

# Default target
all: build

# Version - can be overridden via environment
VERSION ?= 1.0.0

# Build directories
BUILD_DIR := builds
CMD_DIR := ./cmd/ztictl

# Go build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION) -s -w"

# Platform targets
PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64 \
	windows/arm64

# Build all platforms
build: clean
	@echo "Building ztictl v$(VERSION) for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@$(foreach platform,$(PLATFORMS),$(call build_platform,$(platform)))
	@echo ""
	@echo "Build Summary:"
	@echo "=============="
	@ls -lh $(BUILD_DIR)/
	@echo ""
	@echo "✓ All builds completed successfully!"

# Build for current platform only
build-local:
	@echo "Building ztictl for current platform..."
	@go build $(LDFLAGS) -o ztictl $(CMD_DIR)
	@echo "✓ Local build completed: ./ztictl"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)/
	@rm -f ztictl ztictl.exe
	@echo "✓ Clean completed"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...
	@echo "✓ Tests completed"

# Install locally (for development)
install: build-local
	@echo "Installing ztictl to local system..."
	@sudo cp ztictl /usr/local/bin/
	@echo "✓ ztictl installed to /usr/local/bin/"

# Show help
help:
	@echo "ztictl Build System"
	@echo "=================="
	@echo ""
	@echo "Available targets:"
	@echo "  all         - Build for all platforms (default)"
	@echo "  build       - Build for all platforms"
	@echo "  build-local - Build for current platform only"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  install     - Install locally (requires sudo)"
	@echo "  help        - Show this help"
	@echo ""
	@echo "Environment variables:"
	@echo "  VERSION     - Set build version (default: $(VERSION))"
	@echo ""
	@echo "Examples:"
	@echo "  make                    # Build all platforms"
	@echo "  make build-local        # Build for current platform"
	@echo "  VERSION=2.0.0 make      # Build with custom version"

# Function to build for a specific platform
define build_platform
	$(eval GOOS := $(word 1,$(subst /, ,$(1))))
	$(eval GOARCH := $(word 2,$(subst /, ,$(1))))
	$(eval OUTPUT := $(BUILD_DIR)/ztictl-$(GOOS)-$(GOARCH)$(if $(filter windows,$(GOOS)),.exe))
	@echo "Building for $(GOOS)/$(GOARCH)..."
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(OUTPUT) $(CMD_DIR)
	@echo "✓ Built: $(OUTPUT)"
endef
