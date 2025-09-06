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
- **⚡ Smart file transfers**: Automatic S3 routing for large files with lifecycle management  
- **🔒 Advanced IAM management**: Temporary policies with automatic cleanup
- **🛠️ Modern CLI**: Flag-based interface with comprehensive help and validation
- **📊 Professional logging**: Thread-safe, timestamped logs with debug capabilities
- **🔄 Intelligent operations**: Concurrent-safe with filesystem locking

**Legacy bash tools (deprecated):**
- **ssm**: Connect to EC2 instances, execute commands, tag-based operations
- **authaws**: AWS SSO authentication with interactive account/role selection
- Color-coded output and region shortcodes for faster workflows

## 📋 Prerequisites

- AWS CLI installed ([official installation guide](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html))
- AWS Session Manager plugin (automatically checked by `ztictl config check`)
- AWS credentials configured (`aws configure` or AWS SSO)
- Proper IAM permissions for SSM Session Manager and SSO access

## ⚡ Installation

### **For End Users (Simple Installation):**

**Step 1: Download**
```bash
git clone https://github.com/zsoftly/ztiaws.git
cd ztiaws
```

**Step 2: Install**
```bash
./install.sh
```

**Step 3: Verify**
```bash
authaws --check
ssm --help
```

The installation script automatically:
- ✅ Installs `authaws` and `ssm` commands globally
- ✅ Copies all required modules to `/usr/local/bin/src/`
- ✅ Sets up proper permissions
- ✅ Verifies installation works correctly

**To uninstall:** `./uninstall.sh`

---

### **For Developers (Advanced Setup):**

**Development Environment:**
```bash
git clone https://github.com/zsoftly/ztiaws.git
cd ztiaws
make dev          # Sets up development environment
```

**Development Tools:**
```bash
make test         # Run shellcheck and basic tests
make clean        # Clean up temporary files
make help         # Show all available targets
```

**Development Testing:**
```bash
# Test local development versions (before installation)
./authaws --check
./ssm --help

# After make dev or make install
authaws --check
ssm --help
```

**Note for Developers:** Use `make` targets for development workflow. End users should use `./install.sh` for simpler installation without requiring build tools.

See [INSTALLATION.md](INSTALLATION.md) for comprehensive installation instructions including:
- Platform-specific instructions  
- Windows PATH setup (detailed)
- ztictl Go binary installation
- Troubleshooting guide

## 🔄 Updating ZTiAWS

To update to the latest version, see the update instructions in [INSTALLATION.md](INSTALLATION.md).

**Quick update:**
- **ztictl**: Re-run the installation command from INSTALLATION.md
- **Bash tools**: `git pull origin main` in your cloned directory

## 📘 Usage

### ztictl (Recommended)

#### Quick Start
```bash
# Check system requirements
ztictl config check

# Configure AWS authentication
ztictl auth configure

# List instances in a region
ztictl ssm list --region ca-central-1

# Connect to an instance
ztictl ssm connect i-1234567890abcdef0 --region ca-central-1

# Execute commands remotely
ztictl ssm exec i-1234567890abcdef0 "systemctl status nginx" --region ca-central-1

# Advanced file transfers (with automatic S3 routing for large files)
ztictl ssm transfer upload i-1234567890abcdef0 large-file.zip /opt/data.zip --region ca-central-1
```

#### Configuration Management
```bash
# Show current configuration
ztictl config show

# Validate setup
ztictl config validate

# Get comprehensive help
ztictl --help
ztictl ssm --help
```

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

# List instances and connect
ztictl ssm list --region ca-central-1
ztictl ssm connect i-1234567890abcdef0 --region ca-central-1

# Advanced file transfers with S3 routing
ztictl ssm transfer upload i-1234567890abcdef0 large-file.zip /opt/data.zip
```

**📚 Complete Documentation:** [ztictl/README.md](ztictl/README.md) | [Installation Guide](INSTALLATION.md) | [Release Process](RELEASE.md)

## 🌎 Supported Regions

For a complete list of regions and their status, see [docs/REGIONS.md](docs/REGIONS.md).

For required IAM permissions, see [docs/IAM_PERMISSIONS.md](docs/IAM_PERMISSIONS.md).

For troubleshooting common issues, see [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md).

For CI/CD pipeline architecture and development workflow, see [docs/CI_CD_PIPELINE.md](docs/CI_CD_PIPELINE.md).

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

These tools require AWS credentials and access to your AWS resources. Always:
- Keep your AWS credentials secure
- Use appropriate IAM permissions
- Review security best practices in the [AWS Security Documentation](https://docs.aws.amazon.com/security/)

## ✨ About ZSoftly

ZSoftly is a forward-thinking Managed Service Provider dedicated to empowering businesses with cutting-edge technology solutions. Founded by industry veterans, we combine technical expertise with a client-first approach while maintaining ZTiAWS as a free, open-source project to support the developer community.

[Visit our website](https://zsoftly.com) to learn more about our services.

---

<p align="center">
  <strong>Simplify your AWS workflow with ZTiAWS</strong><br>
  Made with ❤️ by ZSoftly
</p>
