# ZTiAWS

[![CI/CD](https://github.com/zsoftly/ztiaws/actions/workflows/build.yml/badge.svg)](https://github.com/zsoftly/ztiaws/actions/workflows/build.yml)
[![Go Coverage](https://img.shields.io/codecov/c/github/zsoftly/ztiaws?label=Go%20Coverage&logo=codecov)](https://codecov.io/github/zsoftly/ztiaws)
[![MIT License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

<p align="center">
  <img src="assets/no_padding.png" alt="ZSoftly Logo" width="800"/>
  <br>
  <em>Simplify your AWS workflow</em>
</p>

**ZTiAWS** (ZSoftly Tools for AWS) is a collection of streamlined CLI tools that make AWS management effortless. Developed by [ZSoftly](https://zsoftly.com), these open-source utilities help you connect to EC2 instances and authenticate with AWS SSO without the typical complexity.

> **"Life's too short for long AWS commands"** - ZSoftly Team

## 🚀 Key Features

**ztictl (Primary Tool):**
- **🌍 Cross-platform**: Native binaries for Linux, macOS, and Windows
- **🖥️ Multi-OS support**: Full Linux (bash) and Windows Server (PowerShell) command execution
- **🤖 Smart platform detection**: Automatically detects instance OS and adapts commands accordingly
- **⚡ Smart file transfers**: Automatic S3 routing for large files with lifecycle management
- **🔒 Advanced IAM management**: Temporary policies with automatic cleanup
- **🛡️ Enhanced security**: PowerShell injection protection, path traversal prevention, UNC validation
- **🔋 Power management**: Start, stop, and reboot EC2 instances individually or in bulk via tags
- **🛠️ Modern CLI**: Flag-based interface with comprehensive help and validation
- **📊 Professional logging**: Thread-safe, timestamped logs with debug capabilities
- **🔄 Intelligent operations**: Concurrent-safe with filesystem locking and parallel execution
- **🎨 Clean UI**: Customizable fuzzy finder with pagination support for AWS SSO account/role selection

**Legacy bash tools (deprecated):**
- **ssm**: Connect to EC2 instances, execute commands, power management, tag-based operations
- **authaws**: AWS SSO authentication with interactive account/role selection
- Color-coded output and region shortcodes for faster workflows

## 📋 Prerequisites

- AWS CLI installed ([official installation guide](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html))
- AWS Session Manager plugin (automatically checked by `ztictl config check`)
- AWS credentials configured (`aws configure` or AWS SSO)
- Proper IAM permissions for SSM Session Manager and SSO access

## ⚡ Installation

### Quick Install - ztictl (Recommended)

**Linux/macOS:**
```bash
curl -L -o /tmp/ztictl "https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/; s/aarch64/arm64/')" && chmod +x /tmp/ztictl && sudo mv /tmp/ztictl /usr/local/bin/ztictl && ztictl --version
```

**Windows PowerShell:**
```powershell
Invoke-WebRequest -Uri "https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-windows-amd64.exe" -OutFile "$env:TEMP\ztictl.exe"; New-Item -ItemType Directory -Force "$env:USERPROFILE\Tools" | Out-Null; Move-Item "$env:TEMP\ztictl.exe" "$env:USERPROFILE\Tools\ztictl.exe"; [Environment]::SetEnvironmentVariable("PATH", "$env:PATH;$env:USERPROFILE\Tools", "User"); $env:PATH += ";$env:USERPROFILE\Tools"; ztictl --version
```

### Other Installation Options

See [INSTALLATION.md](INSTALLATION.md) for:
- **Platform-specific binaries** (Linux, macOS, Windows - AMD64/ARM64)
- **Building from source** (requires Go 1.24+)
- **Legacy bash tools** (for existing users)
- **Developer setup** (contributing to the project)
- **Update instructions**
- **Troubleshooting guide**

## 🔄 Updating ZTiAWS

To update to the latest version, see the update instructions in [INSTALLATION.md](INSTALLATION.md).

**Quick update:**
- **ztictl**: Re-run the installation command from INSTALLATION.md
- **Bash tools**: `git pull origin main` in your cloned directory

## 📘 Usage

### ztictl (Recommended)

> **📚 Complete Documentation:**
> - [Command Reference](docs/COMMANDS.md) - All commands with examples
> - [Configuration Guide](docs/CONFIGURATION.md) - Setup and configuration
> - [Multi-Region Operations](docs/MULTI_REGION.md) - Cross-region execution

#### Quick Start
```bash
# Initialize configuration interactively (simplified setup)
ztictl config init --interactive
# Only asks for: SSO domain ID (not full URL), uses ca-central-1 defaults

# Check system requirements
ztictl config check --fix

# Authenticate with AWS SSO
ztictl auth login

# List instances in a region (shortcode or full name)
ztictl ssm list --region cac1  # or ca-central-1

# Connect to an instance
ztictl ssm connect i-1234567890abcdef0 --region use1

# Execute commands on tagged instances
ztictl ssm exec --tags "Environment=prod" "uptime" --region euw1
```

#### New Features (v2.4+)

**🔋 Power Management:**
```bash
# Start/stop instances
ztictl ssm start i-1234567890abcdef0 --region cac1
ztictl ssm stop --instances "i-1234,i-5678" --region use1

# Bulk operations by tags
ztictl ssm start-tagged --tags "AutoStart=true" --region euw1
ztictl ssm stop-tagged --tags "Environment=dev" --force --region cac1
```

**🌍 Multi-Region Operations (v2.6+):**
```bash
# Execute across multiple regions
ztictl ssm exec-multi cac1,use1,euw1 --tags "App=web" "health-check"

# Use all configured regions
ztictl ssm exec-multi --all-regions --tags "Type=api" "status"

# Use region groups from config
ztictl ssm exec-multi --region-group production --tags "Critical=true" "backup.sh"
```

See [docs/COMMANDS.md](docs/COMMANDS.md) for complete command reference.

#### Configuration Management
```bash
# Interactive setup (recommended for first-time users)
ztictl config init --interactive
# Simplified: Enter domain ID only (e.g., 'd-1234567890' or 'zsoftly')

# Repair invalid configuration
ztictl config repair

# Show current configuration
ztictl config show

# Get help
ztictl --help
ztictl ssm --help
```

See [docs/CONFIGURATION.md](docs/CONFIGURATION.md) for detailed configuration options.

### Legacy Bash Tools (Deprecated)

> **⚠️ Deprecation Notice:** The bash tools are being phased out. New users should use `ztictl` above.

#### SSM Session Manager Tool

**Available flags**: `--region`, `--instance`, `--command`, `--tag-key`, `--tag-value`, `--local-file`, `--remote-file`, `--local-path`, `--remote-path`, `--local-port`, `--remote-port`, `--exec`, `--exec-tagged`, `--upload`, `--download`, `--forward`, `--list`, `--connect`, `--check`, `--help`, `--version`, `--debug`

##### Check System Requirements
```bash
# Traditional syntax (backward compatible)
ssm check

# Flag-based syntax (new)
ssm --check
```

##### List Instances in a Region
```bash
# Traditional syntax (backward compatible)
ssm cac1  # Lists instances in Canada Central

# Flag-based syntax (new)
ssm --region cac1 --list
ssm --region cac1  # Equivalent to --list
```

##### Connect to an Instance
```bash
# Traditional syntax (backward compatible)
ssm i-1234abcd              # Connect to instance in default region (Canada Central)
ssm use1 i-1234abcd         # Connect to instance in US East

# Flag-based syntax (new)
ssm --region use1 --instance i-1234abcd --connect
ssm --region use1 --instance i-1234abcd  # Equivalent to --connect

# Mixed syntax (also supported)
ssm cac1 --instance i-1234abcd
ssm --region use1 i-1234abcd
```

##### Execute Commands Remotely
Execute commands on a single instance:
```bash
# Traditional syntax (backward compatible)
ssm exec cac1 i-1234 "systemctl status nginx"

# Flag-based syntax (new)
ssm --exec --region cac1 --instance i-1234 --command "systemctl status nginx"

# Mixed syntax (also supported)
ssm exec cac1 --instance i-1234 --command "systemctl status nginx"
```

Execute commands on instances matching specific tags:
```bash
# Traditional syntax (backward compatible)
ssm exec-tagged use1 Role web "df -h"

# Flag-based syntax (new)
ssm --exec-tagged --region use1 --tag-key Role --tag-value web --command "df -h"

# Mixed syntax (also supported)
ssm exec-tagged use1 --tag-key Role --tag-value web --command "df -h"
```

**🆕 ztictl Multi-Tag Filtering (Enhanced)**
```bash
# Single tag filtering
ztictl ssm exec-tagged use1 --tags Environment=production "df -h"
# Multiple tag filtering (AND logic)
ztictl ssm exec-tagged use1 --tags Environment=prod,Role=web,Team=backend "df -h"  
# Short flag syntax
ztictl ssm exec-tagged use1 -t "Owner=james,Environment=dev" "systemctl status nginx"
```

This will run `df -h` on all instances that match **ALL** specified tags. The script provides clear feedback if no instances match the specified tags.

##### File Transfer Operations
Upload and download files with automatic size-based routing:
```bash
# Traditional syntax (backward compatible)
ssm upload cac1 i-1234 ./config.txt /etc/app/config.txt
ssm download cac1 i-1234 /var/log/app.log ./app.log

# Flag-based syntax (new)
ssm --upload --region cac1 --instance i-1234 --local-file ./config.txt --remote-path /etc/app/config.txt
ssm --download --region cac1 --instance i-1234 --remote-file /var/log/app.log --local-path ./app.log

# Mixed syntax (also supported)
ssm upload cac1 --instance i-1234 --local-file ./config.txt --remote-path /etc/app/config.txt
```

##### Show Help
```bash
# Traditional syntax (backward compatible)
ssm help

# Flag-based syntax (new)
ssm --help
ssm -h
```

#### AWS SSO Authentication Tool

#### First-time Setup
```bash
authaws check       # Check dependencies
authaws help        # Show help information
```

**Available flags**: `--profile`, `--region`, `--sso-url`, `--export`, `--list-profiles`, `--debug`, `--help`, `--version`, `--check`, `--creds`

Before using `authaws`, set up a `.env` file in the same directory with the following content:
```
SSO_START_URL="https://your-sso-url.awsapps.com/start"
SSO_REGION="your-region"
DEFAULT_PROFILE="your-default-profile"
```

You can create a template file by running `authaws` without a valid .env file.

#### Log in to AWS SSO
```bash
# Traditional syntax (backward compatible)
authaws             # Use default profile from .env
authaws myprofile   # Use a specific profile name

# Flag-based syntax (new)
authaws --profile myprofile                    # Use specific profile
authaws --profile prod --region us-east-1      # Override region
authaws --profile dev --sso-url https://alt.awsapps.com/start  # Override SSO URL
```

The tool will:
1. Check for valid cached credentials
2. Initiate AWS SSO login if needed
3. Show an interactive list of accounts
4. Show an interactive list of roles for the selected account
5. Configure your AWS profile with the selected account and role

#### View AWS Credentials
```bash
# Traditional syntax
authaws creds           # Show credentials for current profile
authaws creds myprofile # Show credentials for a specific profile

# Flag-based syntax
authaws --creds                           # Show credentials for current profile
authaws --creds --profile myprofile       # Show credentials for specific profile
authaws --creds --profile myprofile --export  # Export format for shell evaluation
```

This will display your AWS access key, secret key, and session token for the specified profile.

## 🚀 Production Tool: ztictl

**ztictl** is our **recommended production tool** - a modern Go implementation of AWS SSM operations with enhanced features and full cross-platform support.

> **✅ Production Ready:** ztictl is now the primary tool. The bash tools are maintained for legacy compatibility but new features are only added to ztictl.

### Why Choose ztictl:
- **🌍 Cross-platform**: Native binaries for Linux, macOS, and Windows (AMD64/ARM64)
- **⚡ Enhanced performance**: Intelligent file transfer routing and S3 integration
- **🔒 Advanced security**: Comprehensive IAM lifecycle management and automatic cleanup
- **🛠️ Professional tooling**: Built-in logging, debugging, and resource management
- **🏗️ Modern CLI**: Flag-based interface with comprehensive help and validation

### Get Started:

**Installation:** See [INSTALLATION.md](INSTALLATION.md) for complete setup instructions.

**Usage Examples:**
```bash
# Check system requirements
ztictl config check

# List instances (shows OS type detection)
ztictl ssm list --region ca-central-1
# Output shows: Linux/UNIX, Windows Server 2022, etc.

# Connect to any instance (Linux or Windows)
ztictl ssm connect i-1234567890abcdef0 --region ca-central-1

# Execute commands (automatically adapts to OS)
# Linux instance - uses bash
ztictl ssm exec ca-central-1 i-linux123 "echo 'Hello Linux'; uname -a"

# Windows instance - uses PowerShell  
ztictl ssm exec ca-central-1 i-windows456 "Write-Output 'Hello Windows'; Get-ComputerInfo"

# Cross-platform file transfers
ztictl ssm transfer upload i-linux123 file.txt /tmp/file.txt
ztictl ssm transfer upload i-windows456 file.txt C:\temp\file.txt
```

**📚 Complete Documentation:** [ztictl/README.md](ztictl/README.md) | [Installation Guide](INSTALLATION.md) | [Release Process](RELEASE.md)

## 🌎 Supported Regions

For a complete list of regions and their status, see [docs/REGIONS.md](docs/REGIONS.md).

For required IAM permissions, see [docs/IAM_PERMISSIONS.md](docs/IAM_PERMISSIONS.md).

For troubleshooting common issues, see [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md).

For CI/CD pipeline architecture and development workflow, see [docs/CI_CD_PIPELINE.md](docs/CI_CD_PIPELINE.md).

## 📚 Documentation

### Core Documentation
- **[Command Reference](docs/COMMANDS.md)** - Complete list of all commands with examples
- **[Configuration Guide](docs/CONFIGURATION.md)** - Detailed configuration file reference  
- **[Multi-Region Operations](docs/MULTI_REGION.md)** - Guide for cross-region command execution
- **[Installation Guide](INSTALLATION.md)** - Platform-specific installation instructions
- **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Common issues and solutions
- **[IAM Permissions](docs/IAM_PERMISSIONS.md)** - Required AWS permissions

### Additional Resources
- **[CI/CD Pipeline](docs/CI_CD_PIPELINE.md)** - Automated build and release process
- **[Release Notifications](docs/NOTIFICATIONS.md)** - Google Chat integration
- **[QA Test Guide](tests/QA_SSM_TESTS.md)** - Testing procedures

## 👥 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

Key areas for contribution:
- Adding support for new regions
- Improving documentation
- Adding new features
- Bug fixes

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

**ZTiAWS is completely free and open source** for anyone to use, modify, and distribute. While developed by ZSoftly, we believe in giving back to the community and making AWS management easier for everyone.

## 🚀 Releasing a New Version

For maintainers who want to create a new release, please see [CONTRIBUTING.md](CONTRIBUTING.md).

## 🔐 Security

### Built-in Security Features
- **Directory Traversal Protection**: Cross-platform path validation prevents unauthorized file access
- **Input Validation**: Comprehensive validation for AWS resource identifiers and parameters
- **IAM Lifecycle Management**: Automatic cleanup of temporary policies and permissions
- **Secure File Handling**: Protected file operations with permission validation

### Best Practices
These tools require AWS credentials and access to your AWS resources. Always:
- Keep your AWS credentials secure
- Use appropriate IAM permissions with least privilege
- Review security best practices in the [AWS Security Documentation](https://docs.aws.amazon.com/security/)
- Ensure your AWS CLI and Session Manager Plugin are up to date

## ✨ About ZSoftly

ZSoftly is a forward-thinking Managed Service Provider dedicated to empowering businesses with cutting-edge technology solutions. Founded by industry veterans, we combine technical expertise with a client-first approach while maintaining ZTiAWS as a free, open-source project to support the developer community.

[Visit our website](https://zsoftly.com) to learn more about our services.

---

<p align="center">
  <strong>Simplify your AWS workflow with ZTiAWS</strong><br>
  Made with ❤️ by ZSoftly
</p>
