#!/bin/bash

# ztictl Cross-Platform Build Script
# Builds ztictl for multiple operating systems and architectures

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Version from go.mod or default
VERSION=${VERSION:-"1.0.0"}

echo -e "${BLUE}Building ztictl v${VERSION} for multiple platforms...${NC}"

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
        -ldflags "-X main.version=${VERSION} -s -w" \
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
echo -e "${BLUE}===============${NC}"
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
