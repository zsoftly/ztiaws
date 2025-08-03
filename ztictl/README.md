# ztictl - Next Generation AWS SSM Tool

![Build Status](https://github.com/zsoftly/ztiaws/actions/workflows/build.yml/badge.svg)
[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org)
[![Cross Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey.svg)](#installation)

> **🚀 Modern, cross-platform Go implementation of AWS SSM operations**  
> Part of the [ZTiAWS](../README.md) toolkit by [ZSoftly](https://zsoftly.com)

## Overview

`ztictl` is the next-generation Go implementation of AWS Systems Manager operations, designed to eventually replace the current bash-based `ssm` tool. It provides enhanced performance, cross-platform support, and advanced features while maintaining familiar workflows.

> **⚠️ Current Status:** The original bash-based tools ([`ssm`](../README.md#ssm-session-manager-tool) and [`authaws`](../README.md#aws-sso-authentication-tool)) remain in production. `ztictl` is under active development and testing.

## Why ztictl?

### 🎯 **Enhanced Features**
- **Advanced file transfers**: Intelligent routing (direct <1MB, S3 ≥1MB) with automatic cleanup
- **Comprehensive IAM management**: Temporary policies with lifecycle tracking and emergency cleanup
- **S3 lifecycle integration**: Automatic bucket management with expiration policies
- **Robust error handling**: Detailed logging and graceful recovery procedures

### 🌍 **Cross-Platform Support**
- **Linux**: AMD64 and ARM64 (Intel/AMD and ARM processors)
- **macOS**: Intel and Apple Silicon (M1/M2/M3)
- **Windows**: AMD64 and ARM64 architectures

### ⚡ **Performance Benefits**
- **Native binaries**: No runtime dependencies or script interpretation
- **Optimized transfers**: Efficient handling of large files via S3 intermediary
- **Concurrent operations**: Safe multi-instance operations with filesystem locking
- **Centralized logging**: Thread-safe timestamped logs with platform-specific locations

## Quick Start

### Installation

Choose your installation method:

**📦 Pre-built Binaries (Recommended)**
```bash
# Linux/macOS - automatic platform detection
curl -L -o ztictl "https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/')"
chmod +x ztictl
sudo mv ztictl /usr/local/bin/
```

**🪟 Windows (PowerShell)**
```powershell
Invoke-WebRequest -Uri "https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-windows-amd64.exe" -OutFile "ztictl.exe"
```

**🛠️ Build from Source**
```bash
git clone https://github.com/zsoftly/ztiaws.git
cd ztiaws/ztictl
make build-local
sudo cp ztictl /usr/local/bin/  # Linux/macOS
```

> **📚 Detailed Instructions:** See [INSTALLATION.md](../INSTALLATION.md) for platform-specific setup, prerequisites, and troubleshooting.

### Verify Installation
```bash
ztictl --version
ztictl auth whoami
ztictl ssm list --region us-east-1
```

## Core Operations

### 🔐 **Authentication**
```bash
# Check current identity
ztictl auth whoami

# List available regions  
ztictl auth regions

# AWS SSO authentication (planned)
ztictl auth sso --profile myprofile
```

### 🖥️ **Instance Management**
```bash
# List SSM-enabled instances
ztictl ssm list --region us-east-1

# Filter by tags or status
ztictl ssm list --tag "Environment=prod" --status running

# Check SSM agent status
ztictl ssm status i-1234567890abcdef0
```

### 🔗 **Instance Connection**
```bash
# Connect via Session Manager
ztictl ssm connect i-1234567890abcdef0 --region us-east-1

# Execute commands remotely
ztictl ssm command i-1234567890abcdef0 "systemctl status nginx"

# Port forwarding
ztictl ssm forward i-1234567890abcdef0 8080:80
```

### 📁 **File Transfer Operations**

**🚀 Intelligent Transfer Routing**
- **Small files (<1MB)**: Direct SSM transfer for speed
- **Large files (≥1MB)**: S3 intermediary for reliability

```bash
# Upload files (any size)
ztictl ssm transfer upload i-1234567890abcdef0 local-file.txt /tmp/remote-file.txt

# Download files (any size)  
ztictl ssm transfer download i-1234567890abcdef0 /var/log/app.log ./downloaded-log.txt

# Advanced: Large file handling with debugging
ztictl ssm transfer upload i-1234567890abcdef0 large-file.zip /opt/app/data.zip --debug
```

### 🧹 **Resource Management**
```bash
# Clean up temporary resources
ztictl ssm cleanup --region us-east-1

# Emergency cleanup (aggressive)
ztictl ssm emergency-cleanup --region us-east-1
```

## Advanced Features

### 🔒 **Security & IAM**
- **Temporary IAM policies**: Automatically created and cleaned up
- **Filesystem locking**: Prevents concurrent policy conflicts  
- **Registry tracking**: Complete audit trail of temporary resources
- **Emergency procedures**: Comprehensive cleanup capabilities

### 🗃️ **S3 Lifecycle Management**
- **Auto-bucket creation**: Region-specific buckets with lifecycle policies
- **Object expiration**: 1-day automatic cleanup of transfer files
- **Multipart upload handling**: Automatic cleanup of incomplete uploads

### 📊 **Logging & Debugging**
```bash
# Enable debug logging
ztictl ssm list --debug --region us-east-1

# Cross-platform log locations:
# Linux:   ~/.local/share/ztictl/logs/ztictl-YYYY-MM-DD.log
# macOS:   ~/Library/Logs/ztictl/ztictl-YYYY-MM-DD.log  
# Windows: %LOCALAPPDATA%\ztictl\logs\ztictl-YYYY-MM-DD.log

# Custom log directory (all platforms)
export ZTICTL_LOG_DIR="/custom/path"
ztictl ssm list --region us-east-1

# View timestamped logs
tail -f ~/.local/share/ztictl/logs/ztictl-$(date +%Y-%m-%d).log
```

## Migration Path

### 🔄 **From Bash SSM Tool**

The bash `ssm` tool and `ztictl` can coexist during the transition:

**Current (Production)**
```bash
ssm cac1                                    # List instances  
ssm i-1234567890abcdef0                     # Connect
ssm exec cac1 i-1234567890abcdef0 "uptime"  # Execute command
```

**ztictl (Next Generation)**
```bash
ztictl ssm list --region ca-central-1                              # List instances
ztictl ssm connect i-1234567890abcdef0 --region ca-central-1        # Connect  
ztictl ssm command i-1234567890abcdef0 "uptime" --region ca-central-1  # Execute command
```

**🆕 Enhanced Capabilities in ztictl**
```bash
# Advanced file transfers (not available in bash version)
ztictl ssm transfer upload i-1234567890abcdef0 large-file.zip /opt/data.zip

# Comprehensive resource management  
ztictl ssm cleanup --region ca-central-1

# Cross-platform support
# Works natively on Windows, macOS, and Linux
```

### 📈 **Migration Timeline**
1. **Current**: Bash tools (`ssm`, `authaws`) remain in production
2. **Testing Phase**: `ztictl` available for evaluation and testing  
3. **Transition Phase**: Gradual migration to `ztictl` with feature parity
4. **Future**: Complete replacement with enhanced capabilities

## Documentation

Following DRY principles, comprehensive documentation is centralized:

| Topic | Location | Description |
|-------|----------|-------------|
| **Installation** | [INSTALLATION.md](../INSTALLATION.md) | Platform-specific setup and troubleshooting |
| **Release Process** | [RELEASE.md](../RELEASE.md) | Version management and deployment |
| **CI/CD Pipeline** | [docs/CI_CD_PIPELINE.md](../docs/CI_CD_PIPELINE.md) | Build automation and workflow architecture |
| **Build Artifacts** | [BUILD_ARTIFACTS.md](../BUILD_ARTIFACTS.md) | Git workflow and artifact management |
| **Quick Reference** | [QUICK_START.md](../QUICK_START.md) | Essential commands and examples |
| **Root Project** | [../README.md](../README.md) | Current production tools and overview |

## Development

### 🔨 **Build System**
```bash
# Build for current platform
make build-local

# Cross-platform builds
make build

# Run tests
make test

# Clean artifacts
make clean
```

### 🚀 **Release Process**
```bash
# Create and push tag (triggers automated builds)
git tag v1.1.0
git push origin v1.1.0

# GitHub Actions automatically:
# ✅ Builds for all platforms
# ✅ Runs comprehensive tests  
# ✅ Creates GitHub release
# ✅ Uploads cross-platform binaries
```

> **📋 Details:** See [RELEASE.md](../RELEASE.md) for complete release procedures.

## Contributing

We welcome contributions to ztictl! Areas of focus:

- **🐛 Bug fixes** and stability improvements
- **✨ New features** and AWS service integrations
- **📚 Documentation** and examples
- **🧪 Testing** on different platforms and AWS environments
- **🔧 Performance** optimizations

See the main [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

## Support & Compatibility

### 🔧 **Requirements**
- **AWS CLI** configured with appropriate credentials
- **Session Manager plugin** for SSM connections
- **EC2 instances** with SSM agent and proper IAM roles

### 🆘 **Getting Help**
```bash
# Built-in help
ztictl --help
ztictl ssm --help
ztictl ssm transfer --help

# Debug mode for troubleshooting
ztictl ssm list --debug --region us-east-1
```

## License & About

**ztictl** is part of the open-source [ZTiAWS](../README.md) project, licensed under the MIT License.

**Developed by [ZSoftly](https://zsoftly.com)** - Making AWS management effortless for developers worldwide.

---

<p align="center">
  <strong>🚀 Next-generation AWS SSM operations</strong><br>
  <em>Enhanced performance • Cross-platform • Advanced features</em>
</p>
