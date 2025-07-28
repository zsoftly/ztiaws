# ztictl Installation Guide

ztictl is a cross-platform command-line tool for AWS Systems Manager operations. This guide covers installation on Windows, macOS, and Linux.

> **📋 Note**: ZTiAWS provides **two versions**:
> - **Shell Scripts** (`ssm` & `authaws`) - Production stable version (v1.4.x)
> - **Go Binary** (`ztictl`) - New unified tool in testing/preview (v2.0.x)
> 
> You can install both or choose the one that fits your needs.

## Quick Install (Recommended)

### Prerequisites
- AWS CLI configured with appropriate credentials
- EC2 instances with SSM agent installed and proper IAM roles

## Installation Methods

### � Option A: Shell Scripts (Production Stable)

The original shell-based tools that are currently in production use.

#### Download Shell Scripts
```bash
# Download both scripts
curl -L -o ssm https://raw.githubusercontent.com/zsoftly/ztiaws/main/ssm
curl -L -o authaws https://raw.githubusercontent.com/zsoftly/ztiaws/main/authaws

# Make executable
chmod +x ssm authaws

# Install system-wide (optional)
sudo mv ssm authaws /usr/local/bin/

# Download supporting files
sudo mkdir -p /usr/local/bin/src
curl -L -o /tmp/utils.sh https://raw.githubusercontent.com/zsoftly/ztiaws/main/src/00_utils.sh
curl -L -o /tmp/regions.sh https://raw.githubusercontent.com/zsoftly/ztiaws/main/src/01_regions.sh
curl -L -o /tmp/instance_resolver.sh https://raw.githubusercontent.com/zsoftly/ztiaws/main/src/02_ssm_instance_resolver.sh
curl -L -o /tmp/command_runner.sh https://raw.githubusercontent.com/zsoftly/ztiaws/main/src/03_ssm_command_runner.sh
curl -L -o /tmp/file_transfer.sh https://raw.githubusercontent.com/zsoftly/ztiaws/main/src/04_ssm_file_transfer.sh
sudo mv /tmp/*.sh /usr/local/bin/src/

# Verify installation
ssm --version
authaws --version
```

#### Shell Scripts Usage
```bash
# List instances
ssm list

# Connect to instance
ssm connect i-1234567890abcdef0

# Run command
ssm command i-1234567890abcdef0 "uptime"

# Configure AWS SSO
authaws configure
```

### 🚀 Option B: Go Binary (New Unified Tool)

The new Go-based unified tool that combines both `ssm` and `authaws` functionality.

#### For Linux (x86_64/AMD64)
```bash
# Download and extract
curl -L -o ztictl.tar.gz https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-amd64.tar.gz
tar -xzf ztictl.tar.gz

# Make executable and install
chmod +x ztictl-linux-amd64
sudo mv ztictl-linux-amd64 /usr/local/bin/ztictl

# Clean up
rm ztictl.tar.gz

# Verify installation
ztictl --version
```

#### For Linux (ARM64)
```bash
# Download ARM64 version
curl -L -o ztictl.tar.gz https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-arm64.tar.gz
tar -xzf ztictl.tar.gz
chmod +x ztictl-linux-arm64
sudo mv ztictl-linux-arm64 /usr/local/bin/ztictl
rm ztictl.tar.gz

# Verify installation
ztictl --version
```

#### For macOS (Intel)
```bash
# Download Intel version
curl -L -o ztictl.tar.gz https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-darwin-amd64.tar.gz
tar -xzf ztictl.tar.gz

# Make executable
chmod +x ztictl-darwin-amd64

# Install system-wide (optional)
sudo mv ztictl-darwin-amd64 /usr/local/bin/ztictl
rm ztictl.tar.gz

# Verify installation
ztictl --version
```

#### For macOS (Apple Silicon - M1/M2/M3)
```bash
# Download Apple Silicon version
curl -L -o ztictl.tar.gz https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-darwin-arm64.tar.gz
tar -xzf ztictl.tar.gz

# Make executable
chmod +x ztictl-darwin-arm64

# Install system-wide (optional)
sudo mv ztictl-darwin-arm64 /usr/local/bin/ztictl
rm ztictl.tar.gz

# Verify installation
ztictl --version
```

#### For Windows (x86_64/AMD64)

**Option A: PowerShell (Recommended)**
```powershell
# Download latest release
Invoke-WebRequest -Uri "https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-windows-amd64.zip" -OutFile "ztictl.zip"

# Extract
Expand-Archive -Path "ztictl.zip" -DestinationPath "." -Force

# Add to PATH (optional) - creates directory and adds to user PATH
$installDir = "$env:USERPROFILE\bin"
New-Item -ItemType Directory -Force -Path $installDir
Move-Item ztictl-windows-amd64.exe "$installDir\ztictl.exe"
$env:PATH += ";$installDir"
[Environment]::SetEnvironmentVariable("PATH", $env:PATH, [EnvironmentVariableTarget]::User)

# Clean up
Remove-Item ztictl.zip

# Verify installation
ztictl --version
```

**Option B: Manual Download**
1. Go to [Releases](https://github.com/zsoftly/ztiaws/releases/latest)
2. Download `ztictl-windows-amd64.zip`
3. Extract the ZIP file
4. Rename the binary to `ztictl.exe`
5. Place in a directory in your PATH or create a new directory and add it to PATH

#### For Windows (ARM64)
```powershell
# Download ARM64 version for ARM-based Windows systems
Invoke-WebRequest -Uri "https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-windows-arm64.zip" -OutFile "ztictl.zip"

# Follow same extraction and installation steps as above
```

### 🛠️ Option C: Build from Source

#### Prerequisites
- Go 1.24+ installed
- Git

#### Steps
```bash
# Clone repository
git clone https://github.com/zsoftly/ztiaws.git
cd ztiaws/ztictl

# Build for your platform
make build-local

# Install (Linux/macOS)
sudo cp ztictl /usr/local/bin/

# Or on Windows, copy ztictl.exe to a directory in PATH
```

#### Cross-compile for other platforms
```bash
# Build all platforms
make build

# Individual platform builds
GOOS=windows GOARCH=amd64 go build -o ztictl-windows.exe ./cmd/ztictl
GOOS=darwin GOARCH=arm64 go build -o ztictl-macos-arm64 ./cmd/ztictl
```

## Version Comparison

| Feature | Shell Scripts (`ssm`/`authaws`) | Go Binary (`ztictl`) |
|---------|--------------------------------|---------------------|
| **Status** | ✅ Production stable | 🧪 Testing/Preview |
| **Version** | v1.4.x | v2.0.x |
| **Installation** | Individual scripts | Single binary |
| **Dependencies** | Bash, AWS CLI | None (self-contained) |
| **Commands** | `ssm list`, `authaws configure` | `ztictl ssm list`, `ztictl auth configure` |
| **Platforms** | Linux/macOS (bash) | Linux/macOS/Windows |
| **Maintenance** | Separate tools | Unified tool |

## Post-Installation Setup

### 1. AWS Configuration
Ensure AWS credentials are configured:
```bash
# Using AWS CLI
aws configure

# Or using environment variables
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_DEFAULT_REGION="us-east-1"

# Or using AWS profiles
export AWS_PROFILE="your-profile"
```

### 2. Test Installation
```bash
# Check version
ztictl --version

# List available regions
ztictl auth regions

# Test SSM connectivity (replace with your region)
ztictl ssm list --region us-east-1
```

### 3. Basic Usage

#### Shell Scripts
```bash
# List SSM-enabled instances
ssm list

# Connect to an instance
ssm connect i-1234567890abcdef0

# Execute a command
ssm command i-1234567890abcdef0 "uptime"

# Configure AWS SSO
authaws configure

# Upload a file
ssm upload i-1234567890abcdef0 local-file.txt /tmp/remote-file.txt

# Download a file  
ssm download i-1234567890abcdef0 /tmp/remote-file.txt downloaded-file.txt
```

#### Go Binary (ztictl)
```bash
# List SSM-enabled instances
ztictl ssm list

# Connect to an instance
ztictl ssm connect i-1234567890abcdef0

# Execute a command
ztictl ssm command i-1234567890abcdef0 "uptime"

# Configure AWS SSO
ztictl auth configure

# Upload a file
ztictl ssm transfer upload i-1234567890abcdef0 local-file.txt /tmp/remote-file.txt

# Download a file
ztictl ssm transfer download i-1234567890abcdef0 /tmp/remote-file.txt downloaded-file.txt
```

## Platform-Specific Notes

### Linux
- Requires `session-manager-plugin` for SSM sessions:
  ```bash
  # Ubuntu/Debian
  curl "https://s3.amazonaws.com/session-manager-downloads/plugin/latest/ubuntu_64bit/session-manager-plugin.deb" -o "session-manager-plugin.deb"
  sudo dpkg -i session-manager-plugin.deb
  
  # RHEL/CentOS/Amazon Linux
  curl "https://s3.amazonaws.com/session-manager-downloads/plugin/latest/linux_64bit/session-manager-plugin.rpm" -o "session-manager-plugin.rpm"
  sudo yum install -y session-manager-plugin.rpm
  ```

### macOS
- Install Session Manager plugin:
  ```bash
  # Using Homebrew (recommended)
  brew install --cask session-manager-plugin
  
  # Or manual installation
  curl "https://s3.amazonaws.com/session-manager-downloads/plugin/latest/mac/sessionmanager-bundle.zip" -o "sessionmanager-bundle.zip"
  unzip sessionmanager-bundle.zip
  sudo ./sessionmanager-bundle/install -i /usr/local/sessionmanagerplugin -b /usr/local/bin/session-manager-plugin
  ```

### Windows
- Install Session Manager plugin:
  1. Download from [AWS Documentation](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html#install-plugin-windows)
  2. Run the installer
  3. Add to PATH if needed

## Troubleshooting

### Common Issues

#### "Command not found"
- Ensure the binary is in your PATH
- On Linux/macOS: `echo $PATH`
- On Windows: `echo $env:PATH`

#### "Permission denied"
- On Linux/macOS: Ensure binary is executable (`chmod +x ztictl`)
- Check file permissions and ownership

#### AWS Authentication Issues
```bash
# Verify AWS credentials
ztictl auth whoami

# Check region configuration
ztictl auth regions
```

#### SSM Connection Issues
- Verify EC2 instances have SSM agent installed
- Check IAM roles and policies
- Ensure security groups allow outbound HTTPS (443)

### Getting Help
```bash
# General help
ztictl --help

# Command-specific help
ztictl ssm --help
ztictl ssm transfer --help

# Enable debug logging
ztictl ssm list --debug --region us-east-1
```

## Uninstallation

### Linux/macOS
```bash
# Remove binary
sudo rm /usr/local/bin/ztictl

# Remove configuration (optional)
rm -rf ~/.ztictl.yaml
rm -rf ~/logs/ztictl-*.log
```

### Windows
```powershell
# Remove binary from PATH location
Remove-Item "$env:USERPROFILE\bin\ztictl.exe"

# Remove configuration (optional)
Remove-Item "$env:USERPROFILE\.ztictl.yaml"
Remove-Item "$env:USERPROFILE\logs\ztictl-*.log"
```

## Next Steps

After installation, see:
- [User Guide](README.md) - Complete usage documentation
- [Examples](docs/examples.md) - Common use cases and examples
- [Troubleshooting](docs/TROUBLESHOOTING.md) - Detailed troubleshooting guide
