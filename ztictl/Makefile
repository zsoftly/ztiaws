# ztictl Makefile
# Cross-platform build automation

.PHONY: all build clean test help install

# Default target
all: build

# Version - can be overridden via environment
BASE_VERSION ?= 2.5.2
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION := $(BASE_VERSION)-$(GIT_COMMIT)

# Build directories
BUILD_DIR := builds
CMD_DIR := ./cmd/ztictl

# Go build flags
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -s -w"

# Cross-platform commands
ifeq ($(OS),Windows_NT)
	RM = del /Q /F
	RMDIR = rmdir /S /Q
	MKDIR = mkdir
	LS = dir
	COPY = copy
	SEP = \\
	EXE_EXT = .exe
else
	RM = rm -f
	RMDIR = rm -rf
	MKDIR = mkdir -p
	LS = ls -lh
	COPY = cp
	SEP = /
	EXE_EXT =
endif

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
ifeq ($(OS),Windows_NT)
	@if not exist $(BUILD_DIR) $(MKDIR) $(BUILD_DIR)
else
	@$(MKDIR) $(BUILD_DIR)
endif
	@$(foreach platform,$(PLATFORMS),$(call build_platform,$(platform)))
	@echo ""
	@echo "Build Summary:"
	@echo "=============="
ifeq ($(OS),Windows_NT)
	@$(LS) $(BUILD_DIR)
else
	@$(LS) $(BUILD_DIR)/
endif
	@echo ""
	@echo "✓ All builds completed successfully!"

# Build for current platform only
build-local:
	@echo "Building ztictl for current platform..."
	@go build $(LDFLAGS) -o ztictl$(EXE_EXT) $(CMD_DIR)
	@echo "✓ Local build completed: ./ztictl$(EXE_EXT)"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
ifeq ($(OS),Windows_NT)
	@if exist $(BUILD_DIR) $(RMDIR) $(BUILD_DIR)
	@if exist ztictl.exe $(RM) ztictl.exe
	@if exist ztictl $(RM) ztictl
else
	@$(RMDIR) $(BUILD_DIR)/
	@$(RM) ztictl ztictl.exe
endif
	@echo "✓ Clean completed"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...
	@echo "✓ Tests completed"

# Install locally (for development)
install: build-local
	@echo "Installing ztictl to local system..."
ifeq ($(OS),Windows_NT)
	@echo "Manual installation required on Windows:"
	@echo "Copy ztictl.exe to a directory in your PATH"
else
	@sudo $(COPY) ztictl /usr/local/bin/
	@echo "✓ ztictl installed to /usr/local/bin/"
endif

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
	@echo "  VERSION=2.1.0 make      # Build with custom version"

# Function to build for a specific platform
define build_platform
	$(eval GOOS := $(word 1,$(subst /, ,$(1))))
	$(eval GOARCH := $(word 2,$(subst /, ,$(1))))
	$(eval OUTPUT := $(BUILD_DIR)$(SEP)ztictl-$(GOOS)-$(GOARCH)$(if $(filter windows,$(GOOS)),.exe))
	@echo "Building for $(GOOS)/$(GOARCH)..."
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(OUTPUT) $(CMD_DIR)
	@echo "✓ Built: $(OUTPUT)"
endef
