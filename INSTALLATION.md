# ztictl Installation Guide

ztictl is a cross-platform command-line tool for AWS Systems Manager operations. This guide covers installation on Windows, macOS, and Linux.

## Quick Install (Recommended)

### Prerequisites
- AWS CLI configured with appropriate credentials
- EC2 instances with SSM agent installed and proper IAM roles

## Installation Methods

### 📦 Method 1: Download Pre-built Binaries (Recommended)

#### For Linux (x86_64/AMD64)
```bash
# Download latest release
curl -L -o ztictl https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-amd64

# Make executable
chmod +x ztictl

# Install system-wide (optional)
sudo mv ztictl /usr/local/bin/

# Verify installation
ztictl --version
```

#### For Linux (ARM64)
```bash
# Download ARM64 version
curl -L -o ztictl https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-arm64

# Make executable and install
chmod +x ztictl
sudo mv ztictl /usr/local/bin/

# Verify installation
ztictl --version
```

#### For macOS (Intel)
```bash
# Download Intel version
curl -L -o ztictl https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-darwin-amd64

# Make executable
chmod +x ztictl

# Install system-wide (optional)
sudo mv ztictl /usr/local/bin/

# Verify installation
ztictl --version
```

#### For macOS (Apple Silicon - M1/M2/M3)
```bash
# Download Apple Silicon version
curl -L -o ztictl https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-darwin-arm64

# Make executable
chmod +x ztictl

# Install system-wide (optional)
sudo mv ztictl /usr/local/bin/

# Verify installation
ztictl --version
```

#### For Windows (x86_64/AMD64)

**Option A: PowerShell (Recommended)**
```powershell
# Download latest release
Invoke-WebRequest -Uri "https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-windows-amd64.exe" -OutFile "ztictl.exe"

# Add to PATH (optional) - creates directory and adds to user PATH
$installDir = "$env:USERPROFILE\bin"
New-Item -ItemType Directory -Force -Path $installDir
Move-Item ztictl.exe "$installDir\ztictl.exe"
$env:PATH += ";$installDir"
[Environment]::SetEnvironmentVariable("PATH", $env:PATH, [EnvironmentVariableTarget]::User)

# Verify installation
ztictl --version
```

**Option B: Manual Download**
1. Go to [Releases](https://github.com/zsoftly/ztiaws/releases/latest)
2. Download `ztictl-windows-amd64.exe`
3. Rename to `ztictl.exe`
4. Place in a directory in your PATH or create a new directory and add it to PATH

#### For Windows (ARM64)
```powershell
# Download ARM64 version for ARM-based Windows systems
Invoke-WebRequest -Uri "https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-windows-arm64.exe" -OutFile "ztictl.exe"

# Follow same installation steps as above
```

### 🛠️ Method 2: Build from Source

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
```bash
# List SSM-enabled instances
ztictl ssm list

# Connect to an instance
ztictl ssm connect i-1234567890abcdef0

# Execute a command
ztictl ssm command i-1234567890abcdef0 "uptime"

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
