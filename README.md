# ZTiAWS

![Ubuntu](https://github.com/zsoftly/ztiaws/actions/workflows/test.yml/badge.svg)
![macOS](https://github.com/zsoftly/ztiaws/actions/workflows/test.yml/badge.svg)
[![MIT License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

<p align="center">
  <img src="assets/no_padding.png" alt="ZSoftly Logo" width="800"/>
  <br>
  <em>Simplify your AWS workflow</em>
</p>

**ZTiAWS** (ZSoftly Tools for AWS) is a collection of streamlined CLI tools that make AWS management effortless. Developed by [ZSoftly](https://zsoftly.com), these open-source utilities help you connect to EC2 instances and authenticate with AWS SSO without the typical complexity.

> **"Life's too short for long AWS commands"** - ZSoftly Team

## 🚀 Key Features

- **ssm**:
    - Connect to EC2 instances via AWS Systems Manager Session Manager using intuitive short region codes.
    - Execute commands remotely on a single EC2 instance (`ssm exec`).
    - Execute commands remotely on multiple EC2 instances based on AWS tags (`ssm exec-tagged`).
- **authaws**: Streamlined AWS SSO authentication with interactive account/role selection.
- Smart interactive listing of available instances (for `ssm <region>`) and accounts/roles (for `authaws`).
- Automatic validation of AWS CLI and required plugins.
- Enhanced error reporting: Clear feedback for AWS CLI issues and specific handling for scenarios like no instances matching tags during command execution.
- Support for multiple AWS regions with simple shortcodes.
- Color-coded output for enhanced readability
- Time-saving workflows designed by AWS practitioners for real-world use

## 📋 Prerequisites

- AWS CLI installed ([official installation guide](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html))
- AWS Session Manager plugin (can be installed interactively via `ssm check`)
- AWS credentials configured (`aws configure`)
- Bash, Zsh, or PowerShell
- Proper IAM permissions for SSM Session Manager and SSO access
- Additional utilities: `jq` and `fzf` (required for `authaws`)

## ⚡ Quick Start

One-liner to download, install, and start using both tools:

**Bash users:**
```bash
git clone https://github.com/zsoftly/ztiaws.git && cd ztiaws && chmod +x ssm authaws && ./ssm check && echo -e "\n# Add ZTiAWS to PATH\nexport PATH=\"\$PATH:$(pwd)\"" >> ~/.bashrc && source ~/.bashrc
```

**Zsh users:**
```bash
git clone https://github.com/zsoftly/ztiaws.git && cd ztiaws && chmod +x ssm authaws && ./ssm check && echo -e "\n# Add ZTiAWS to PATH\nexport PATH=\"\$PATH:$(pwd)\"" >> ~/.zshrc && source ~/.zshrc
```

**PowerShell users:**
```powershell
git clone https://github.com/zsoftly/ztiaws.git
cd ztiaws
# Follow the PowerShell setup in the detailed installation section
```

After running the appropriate command for your shell, you can use the tools by simply typing `ssm` or `authaws` from anywhere.

For detailed installation instructions, see [docs/INSTALLATION.md](docs/INSTALLATION.md).

## 🔄 Updating ZTiAWS

To update ZTiAWS to the latest version, navigate to your cloned repository directory and run:
```bash
git pull origin main
# Ensure scripts remain executable (if needed)
chmod +x ssm authaws
```
If you are updating from a version prior to March 2025 (when the repository was named "quickssm"), please see [docs/deprecated_update_instructions.md](docs/deprecated_update_instructions.md) for specific instructions.

## 📘 Usage

### SSM Session Manager Tool

#### Check System Requirements
```bash
ssm check
```

#### List Instances in a Region
```bash
ssm cac1  # Lists instances in Canada Central
```

#### Connect to an Instance
```bash
ssm i-1234abcd              # Connect to instance in default region (Canada Central)
ssm use1 i-1234abcd         # Connect to instance in US East
```

#### Execute Commands Remotely
Execute commands on a single instance:
```bash
ssm exec cac1 i-1234 "systemctl status nginx"
```

Execute commands on instances matching specific tags:
```bash
ssm exec-tagged use1 Role web "df -h"
```
This will run `df -h` on all instances in the `us-east-1` region that have a tag with `Key=Role` and `Value=web`. The script provides clear feedback if no instances match the specified tags.

#### Show Help
```bash
ssm help
```

### AWS SSO Authentication Tool

#### First-time Setup
```bash
authaws check       # Check dependencies
authaws help        # Show help information
```

Before using `authaws`, set up a `.env` file in the same directory with the following content:
```
SSO_START_URL="https://your-sso-url.awsapps.com/start"
SSO_REGION="your-region"
DEFAULT_PROFILE="your-default-profile"
```

You can create a template file by running `authaws` without a valid .env file.

#### Log in to AWS SSO
```bash
authaws             # Use default profile from .env
authaws myprofile   # Use a specific profile name
```

The tool will:
1. Check for valid cached credentials
2. Initiate AWS SSO login if needed
3. Show an interactive list of accounts
4. Show an interactive list of roles for the selected account
5. Configure your AWS profile with the selected account and role

#### View AWS Credentials
```bash
authaws creds           # Show credentials for current profile
authaws creds myprofile # Show credentials for a specific profile
```

This will display your AWS access key, secret key, and session token for the specified profile.

## 🚀 Next Generation: ztictl

**ztictl** is the modern Go implementation of AWS SSM operations, designed to eventually replace the current bash-based tools with enhanced features and cross-platform support.

> **⚠️ Current Status:** The bash tools (`ssm` and `authaws`) above remain the **production tools**. ztictl is under active development and testing.

### Key Advantages of ztictl:
- **🌍 Cross-platform**: Native binaries for Linux, macOS, and Windows (AMD64/ARM64)
- **⚡ Enhanced performance**: Intelligent file transfer routing and S3 integration
- **🔒 Advanced security**: Comprehensive IAM lifecycle management and automatic cleanup
- **🛠️ Professional tooling**: Built-in logging, debugging, and resource management

### Quick Example:
```bash
# Install (Linux/macOS)
curl -L -o ztictl "https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/')"
chmod +x ztictl && sudo mv ztictl /usr/local/bin/

# Use (enhanced capabilities)
ztictl ssm list --region ca-central-1
ztictl ssm connect i-1234567890abcdef0 --region ca-central-1
ztictl ssm transfer upload i-1234567890abcdef0 large-file.zip /opt/data.zip  # Advanced file transfers
```

**📚 Complete Documentation:** [ztictl/README.md](ztictl/README.md) | [Installation Guide](INSTALLATION.md) | [Release Process](RELEASE.md)

## 🌎 Supported Regions (for SSM tool)

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
