# ZTiAWS Installation Guide

**ztictl** - Unified AWS Systems Manager CLI tool

## Prerequisites
- AWS CLI configured with appropriate credentials
- EC2 instances with SSM agent installed and proper IAM roles

## Installation

### Quick Install (Recommended)
```bash
# Linux AMD64
curl -L -o ztictl https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-amd64
chmod +x ztictl && sudo mv ztictl /usr/local/bin/

# Linux ARM64  
curl -L -o ztictl https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-arm64
chmod +x ztictl && sudo mv ztictl /usr/local/bin/

# macOS Intel
curl -L -o ztictl https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-darwin-amd64
chmod +x ztictl && sudo mv ztictl /usr/local/bin/

# macOS Apple Silicon
curl -L -o ztictl https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-darwin-arm64
chmod +x ztictl && sudo mv ztictl /usr/local/bin/

# Windows AMD64
# Download: https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-windows-amd64.exe

# Verify installation
ztictl --version
```

## Usage

### Authentication
```bash
# Configure AWS SSO
ztictl auth configure
```

### Instance Management
```bash
# List instances
ztictl ssm list

# Connect to instance
ztictl ssm connect i-1234567890abcdef0

# Execute commands
ztictl ssm exec i-1234567890abcdef0 "uptime"

# Transfer files
ztictl ssm transfer local-file.txt i-1234567890abcdef0:/tmp/
```

### Configuration
```bash
# Show current configuration
ztictl config show

# Validate setup
ztictl config validate

# Advanced SSM management
ztictl ssm manage

# Cleanup resources
ztictl cleanup
```

## Troubleshooting

### Binary Not Found
```bash
# Check if downloaded
ls -la ztictl*

# Make executable
chmod +x ztictl

# Check PATH
which ztictl
```

### Permission Issues
```bash
# Ensure executable
chmod +x ztictl

# System-wide install needs sudo
sudo mv ztictl /usr/local/bin/
```

### Architecture Mismatch
```bash
# Check system architecture
uname -m
# x86_64 = amd64, aarch64 = arm64

# Download correct binary for your system
```
