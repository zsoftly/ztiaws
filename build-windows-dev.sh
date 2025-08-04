#!/bin/bash
# Quick Windows build script for development

# Variables
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$SCRIPT_DIR/ztictl"
WINDOWS_TOOLS_DIR="/mnt/c/Tools"
BUILD_NAME="ztictl.exe"

# Source logging utilities
source "$SCRIPT_DIR/src/00_utils.sh"

# Initialize logging for this script
init_logging "build-windows-dev.sh"

# Get dynamic version from git
cd "$PROJECT_DIR"
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "dev")
VERSION="2.1.$GIT_COMMIT"

log_info "Building $BUILD_NAME for Windows development..."
log_info "Project directory: $PROJECT_DIR"
log_info "Target directory: $WINDOWS_TOOLS_DIR"
log_info "Version: $VERSION"
echo ""

cd "$PROJECT_DIR"

# Build with flags that might reduce antivirus detection
log_info "Building with antivirus-friendly flags..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build \
    -a \
    -installsuffix cgo \
    -ldflags "-X main.version=$VERSION -s -w -buildid=" \
    -trimpath \
    -o "builds/$BUILD_NAME" \
    ./cmd/ztictl

if [ $? -eq 0 ]; then
    log_info "Build successful"
    
    # Copy to Windows Tools directory
    cp "builds/$BUILD_NAME" "$WINDOWS_TOOLS_DIR/$BUILD_NAME"
    log_info "Copied to $WINDOWS_TOOLS_DIR\\$BUILD_NAME"
    
    # Show file info
    ls -la "$WINDOWS_TOOLS_DIR/$BUILD_NAME"
    echo ""
    log_warn "To avoid antivirus issues:"
    echo "1. Add C:\\Tools to Windows Defender exclusions"
    echo "2. Or run: Unblock-File -Path 'C:\\Tools\\$BUILD_NAME'"
    echo ""
    log_info "Ready to test on Windows!"
    log_info "Run: cd C:\\Tools && .\\$BUILD_NAME --help"
else
    log_error "Build failed"
    log_completion
    exit 1
fi

# Log completion
log_completion