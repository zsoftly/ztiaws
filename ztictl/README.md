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

- **🖥️ Multi-OS instance support**: Full Linux (bash) and Windows Server (PowerShell) command execution
- **🤖 Automatic platform detection**: Detects instance OS via SSM/EC2 APIs and adapts commands
- **🛡️ Advanced security**: PowerShell injection protection, UNC path validation, path traversal prevention
- **Advanced file transfers**: Intelligent routing (direct <1MB, S3 ≥1MB) with automatic cleanup
- **Comprehensive IAM management**: Temporary policies with lifecycle tracking and emergency cleanup
- **S3 lifecycle integration**: Automatic bucket management with expiration policies
- **Robust error handling**: Detailed logging and graceful recovery procedures

### 🌍 **Cross-Platform Support**

**Client Platforms** (where ztictl runs):

- **Linux**: AMD64 and ARM64 (Intel/AMD and ARM processors)
- **macOS**: Intel and Apple Silicon (M1/M2/M3)
- **Windows**: AMD64 and ARM64 architectures

**Target Instance Support** (what instances you can manage):

- **✅ Linux instances**: Amazon Linux, Ubuntu, RHEL, CentOS, SUSE (bash commands)
- **✅ Windows instances**: Windows Server 2016/2019/2022 (PowerShell commands)
- **🤖 Auto-detection**: Automatically detects instance OS and uses appropriate command syntax

### ⚡ **Performance Benefits**

- **Native binaries**: No runtime dependencies or script interpretation
- **Optimized transfers**: Efficient handling of large files via S3 intermediary
- **Concurrent operations**: Safe multi-instance operations with filesystem locking
- **Centralized logging**: Thread-safe timestamped logs with platform-specific locations

## Quick Start

### Installation

See [INSTALLATION.md](../INSTALLATION.md) for detailed installation instructions.

**Quick Install (Linux/macOS):**

```bash
curl -L -o /tmp/ztictl "https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/; s/aarch64/arm64/')"
chmod +x /tmp/ztictl
sudo mv /tmp/ztictl /usr/local/bin/ztictl
```

### Configuration

```bash
# Initialize configuration interactively (recommended)
ztictl config init --interactive

# Check system requirements
ztictl config check --fix

# Verify installation
ztictl --version

# Customize account/role selector display (optional)
export ZTICTL_SELECTOR_HEIGHT=10  # Show 10 items instead of default 5
```

See [Configuration Guide](../docs/CONFIGURATION.md) for detailed configuration options.

### Interactive Instance Selection

Most SSM commands support **interactive fuzzy finder** for instance selection. Simply omit the instance ID to launch the interactive selector:

```bash
# Traditional way (still supported)
ztictl ssm connect i-1234567890abcdef0 --region cac1

# New interactive way - launches fuzzy finder
ztictl ssm connect --region cac1
# → Type to search instances by name or ID
# → Arrow keys or j/k to navigate
# → Enter to select

# Works for all SSM operations
ztictl ssm exec --region cac1 "uptime"
ztictl ssm transfer upload --region cac1 ./config.yml /etc/app/config.yml
ztictl ssm start --region cac1
ztictl ssm stop --region cac1
```

**Features:**

- 🔍 **Fuzzy search** - Type to filter instances by name, ID, or tags
- ⌨️ **Keyboard shortcuts** - Vim-style navigation (j/k), arrow keys, Page Up/Down
- 🖱️ **Mouse support** - Click to select, scroll to navigate
- ✅ **Smart validation** - Automatically checks instance state and SSM agent status
- 💡 **Helpful errors** - Clear feedback with suggested commands when operations cannot proceed

See [Fuzzy Finder Features](docs/FUZZY_FINDER_FEATURES.md) for complete keyboard shortcuts, mouse controls, and validation details.

## Documentation

📚 **Complete documentation is available in the docs directory:**

- **[Command Reference](../docs/COMMANDS.md)** - All commands with detailed examples
- **[Configuration Guide](../docs/CONFIGURATION.md)** - Setup and configuration options
- **[Multi-Region Operations](../docs/MULTI_REGION.md)** - Cross-region execution guide
- **[CI/CD Authentication](../docs/CI_CD_AUTHENTICATION.md)** - Using ztictl in CI/CD pipelines
- **[Troubleshooting](../docs/TROUBLESHOOTING.md)** - Common issues and solutions

## Authentication Methods

`ztictl` supports multiple authentication methods depending on your environment:

### Interactive Development (AWS SSO)

For local development and manual operations:

```bash
# AWS SSO authentication (requires browser)
ztictl auth login

# Check current credentials
ztictl auth whoami
```

**Best for:**

- Local development
- Manual operations
- Multi-account access with role switching

**Requirements:**

- AWS SSO configured
- Browser access for authentication
- Interactive terminal

### CI/CD Pipelines (IAM-based)

For automated pipelines, AWS SSO **cannot be used** (requires browser interaction). Use IAM-based authentication instead:

| Method                   | When to Use                                 | Security        | Setup Complexity |
| ------------------------ | ------------------------------------------- | --------------- | ---------------- |
| **OIDC Federation**      | GitHub Actions, GitLab CI, modern platforms | ⭐⭐⭐⭐⭐ Best | Medium           |
| **EC2 Instance Profile** | Self-hosted runners on EC2                  | ⭐⭐⭐⭐⭐ Best | Easy             |
| **ECS Task Role**        | Containerized CI/CD on ECS/Fargate          | ⭐⭐⭐⭐⭐ Best | Easy             |
| **IAM Access Keys**      | Legacy systems, quick testing               | ⭐⭐ Poor       | Easy             |

**OIDC Federation Example (Recommended):**

```yaml
# GitHub Actions
- name: Configure AWS Credentials
  uses: aws-actions/configure-aws-credentials@v2
  with:
    role-to-assume: arn:aws:iam::123456789012:role/GitHubActionsRole
    aws-region: ca-central-1

- name: Use ztictl
  run: |
    ztictl config init --non-interactive
    ztictl ssm exec-multi --tag Environment=prod --command "deploy.sh"
```

**EC2 Instance Profile Example:**

```bash
# No credential configuration needed - automatic from instance metadata
ztictl config init --non-interactive
ztictl ssm list --table
```

**IAM Access Keys Example (Not Recommended):**

```bash
# Set environment variables (store in CI/CD secrets)
export AWS_ACCESS_KEY_ID="AKIAIOSFODNN7EXAMPLE"
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export AWS_DEFAULT_REGION="ca-central-1"

ztictl config init --non-interactive
ztictl ssm list --table
```

### Non-Interactive Mode

`ztictl` automatically detects CI/CD environments and enables non-interactive mode when:

- `CI` environment variable is set (most platforms set this automatically)
- `ZTICTL_NON_INTERACTIVE=true` is set
- `--non-interactive` flag is used

**In non-interactive mode:**

- No splash screen or prompts
- Commands requiring instance selection will fail with clear error messages
- Use explicit instance IDs or tag-based commands

```bash
# These commands work in CI/CD (no interactive prompts)
ztictl ssm list --table
ztictl ssm exec i-1234567890abcdef0 --command "uptime"
ztictl ssm exec-multi --tag Environment=prod --command "deploy.sh"
ztictl ssm power start --tag Environment=test

# These require interactive selection (will fail in CI/CD)
ztictl ssm connect  # ❌ No instance ID specified
ztictl ssm exec "uptime"  # ❌ No instance ID specified

# Fix: Provide instance identifier
ztictl ssm connect i-1234567890abcdef0  # ✅ Works
ztictl ssm connect web-server-prod  # ✅ Works (name lookup)
```

**Complete Guide:** See [docs/CI_CD_AUTHENTICATION.md](../docs/CI_CD_AUTHENTICATION.md) for:

- Detailed authentication setup for each platform
- IAM permission requirements
- Complete workflow examples
- Troubleshooting

## Core Operations

### Quick Examples

```bash
# Authentication
ztictl auth login
ztictl auth whoami

# Instance operations with auto-platform detection
ztictl ssm list --region cac1
# Output shows: Platform column (Linux/UNIX, Windows Server 2022, etc.)

ztictl ssm connect i-1234567890abcdef0 --region use1
ztictl ssm exec --tags "Environment=prod" "uptime" --region euw1

# Cross-platform commands (auto-adapts to instance OS)
# Linux instances - bash commands
ztictl ssm exec cac1 i-linux123 "ps aux | grep nginx"
ztictl ssm exec cac1 i-linux123 "cat /var/log/app.log | tail -10"

# Windows instances - PowerShell commands
ztictl ssm exec cac1 i-windows456 "Get-Process | Where-Object {$_.Name -like '*iis*'}"
ztictl ssm exec cac1 i-windows456 "Get-EventLog -LogName Application -Newest 10"

# Cross-platform file operations
ztictl ssm transfer upload i-linux123 config.yml /etc/app/config.yml
ztictl ssm transfer upload i-windows456 config.xml C:\inetpub\config.xml

# Power management (v2.4+)
ztictl ssm start i-1234567890abcdef0 --region cac1
ztictl ssm stop-tagged --tags "Environment=dev" --region use1

# Multi-region operations (v2.6+)
ztictl ssm exec-multi cac1,use1,euw1 --tags "App=web" "health-check"
ztictl ssm exec-multi --all-regions --tags "Type=api" "status"
```

For complete command documentation, see [docs/COMMANDS.md](../docs/COMMANDS.md).

````

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

See [Configuration Guide](../docs/CONFIGURATION.md#logging-configuration) for detailed logging setup and locations.

## Migration from Legacy Tools

The bash `ssm` tool and `ztictl` can coexist during the transition. See [Command Reference](../docs/COMMANDS.md#legacy-bash-commands) for comparison and migration guide.


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
````

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

> **📋 Details:** See [RELEASE.md](../docs/development/RELEASE.md) for complete release procedures.

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
