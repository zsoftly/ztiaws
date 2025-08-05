#!/bin/bash

# ztictl Cross-Platform Build Script
# Builds ztictl for multiple operating systems and architectures

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Get version information
BASE_VERSION=${BASE_VERSION:-${VERSION:-"2.1.0"}}
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_VERSION="${BASE_VERSION}-${GIT_COMMIT}"

echo -e "${CYAN}Building ztictl v${BUILD_VERSION} for multiple platforms...${NC}"
echo -e "${YELLOW}Base version: ${BASE_VERSION}${NC}"
echo -e "${YELLOW}Git commit: ${GIT_COMMIT}${NC}"

# Clean previous builds
rm -rf builds/
mkdir -p builds/

# Build matrix: OS/ARCH combinations
declare -a builds=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
)

# Build for each platform
for build in "${builds[@]}"; do
    IFS='/' read -r -a platform <<< "$build"
    os="${platform[0]}"
    arch="${platform[1]}"
    
    output="builds/ztictl-${os}-${arch}"
    if [ "$os" = "windows" ]; then
        output="${output}.exe"
    fi
    
    echo -e "${YELLOW}Building for ${os}/${arch}...${NC}"
    
    # Set environment and build
    GOOS=$os GOARCH=$arch go build \
        -ldflags "-X main.Version=${BUILD_VERSION} -s -w" \
        -o "$output" \
        ./cmd/ztictl
    
    if [ $? -eq 0 ]; then
        size=$(ls -lh "$output" | awk '{print $5}')
        echo -e "${GREEN}✓ Built: $output (${size})${NC}"
    else
        echo -e "${RED}✗ Failed to build for ${os}/${arch}${NC}"
        exit 1
    fi
done

echo ""
echo -e "${GREEN}Build Summary:${NC}"
echo -e "${CYAN}===============${NC}"
ls -lh builds/ | tail -n +2 | while read -r line; do
    echo "  $line"
done

echo ""
echo -e "${GREEN}✓ All builds completed successfully!${NC}"
echo ""
echo -e "${YELLOW}Usage Examples:${NC}"
echo "  Linux (x64):     ./builds/ztictl-linux-amd64"
echo "  Linux (ARM):     ./builds/ztictl-linux-arm64"
echo "  macOS (Intel):   ./builds/ztictl-darwin-amd64"
echo "  macOS (Apple):   ./builds/ztictl-darwin-arm64"
echo "  Windows (x64):   ./builds/ztictl-windows-amd64.exe"
echo "  Windows (ARM):   ./builds/ztictl-windows-arm64.exe"
