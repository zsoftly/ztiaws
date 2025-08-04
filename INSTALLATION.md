# ZTiAWS Installation Guide

**ztictl** is the modern, cross-platform AWS Systems Manager CLI tool that simplifies AWS instance management.

> **📦 Primary Tool:** We recommend using the **Go binary (ztictl)** for new installations. The bash tools are maintained for legacy users but are being phased out.

## Prerequisites
- AWS CLI configured with appropriate credentials
- EC2 instances with SSM agent installed and proper IAM roles

## Quick Install (Recommended)

**Linux/macOS - One-liner with automatic platform detection:**
```bash
curl -L -o /tmp/ztictl "https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/; s/aarch64/arm64/')" && chmod +x /tmp/ztictl && sudo mv /tmp/ztictl /usr/local/bin/ztictl && ztictl --version
```

**Windows PowerShell - Full setup:**
```powershell
# Download and setup ztictl
Invoke-WebRequest -Uri "https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-windows-amd64.exe" -OutFile "ztictl.exe"
$toolsDir = "$env:USERPROFILE\Tools"
if (-not (Test-Path $toolsDir)) { New-Item -ItemType Directory -Path $toolsDir }
Move-Item "ztictl.exe" "$toolsDir\ztictl.exe"

# Add to PATH permanently
$currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($currentPath -notlike "*$toolsDir*") {
    $newPath = $currentPath + ";$toolsDir"
    [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
    $env:PATH = $newPath
}

# Verify installation
ztictl --version
```

## Platform-Specific Installation

### Linux

**AMD64 (Intel/AMD):**
```bash
curl -L -o /tmp/ztictl https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-amd64
chmod +x /tmp/ztictl && sudo mv /tmp/ztictl /usr/local/bin/ztictl
```

**ARM64 (ARM processors):**
```bash
curl -L -o /tmp/ztictl https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-arm64
chmod +x /tmp/ztictl && sudo mv /tmp/ztictl /usr/local/bin/ztictl
```

### macOS

**Intel Macs:**
```bash
curl -L -o /tmp/ztictl https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-darwin-amd64
chmod +x /tmp/ztictl && sudo mv /tmp/ztictl /usr/local/bin/ztictl
```

**Apple Silicon (M1/M2/M3):**
```bash
curl -L -o /tmp/ztictl https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-darwin-arm64
chmod +x /tmp/ztictl && sudo mv /tmp/ztictl /usr/local/bin/ztictl
```

### Windows

**Option 1: Manual Download**
1. Download: https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-windows-amd64.exe
2. Rename to `ztictl.exe`
3. Follow the PATH setup instructions below

**Option 2: PowerShell (Recommended)**
```powershell
# Download
Invoke-WebRequest -Uri "https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-windows-amd64.exe" -OutFile "ztictl.exe"

# Setup Tools directory
$toolsDir = "$env:USERPROFILE\Tools"
if (-not (Test-Path $toolsDir)) { New-Item -ItemType Directory -Path $toolsDir }
Move-Item "ztictl.exe" "$toolsDir\ztictl.exe"

# Add to PATH (see Windows PATH setup below)
```

## Windows PATH Setup

**Method 1: PowerShell (Recommended)**
```powershell
$toolsDir = "$env:USERPROFILE\Tools"
$currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($currentPath -notlike "*$toolsDir*") {
    $newPath = $currentPath + ";$toolsDir"
    [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
    $env:PATH = $newPath  # Update current session
}
```

**Method 2: GUI**
1. Press `Win + R`, type `sysdm.cpl`, press Enter
2. Click "Environment Variables"
3. Under "User variables", select "Path" and click "Edit"
4. Click "New" and add your tools directory path (e.g., `C:\Users\YourName\Tools`)
5. Click "OK" on all dialogs
6. Restart PowerShell/Command Prompt

## Usage

### Quick Start
```bash
# Check system requirements
ztictl config check

# Configure AWS authentication
ztictl auth configure

# List instances in a region
ztictl ssm list --region ca-central-1

# Connect to an instance
ztictl ssm connect i-1234567890abcdef0 --region ca-central-1

# Execute remote commands
ztictl ssm exec i-1234567890abcdef0 "uptime" --region ca-central-1

# Advanced file transfers (automatic S3 routing for large files)
ztictl ssm transfer upload i-1234567890abcdef0 large-file.zip /opt/data.zip --region ca-central-1
```

### Configuration Management
```bash
# Show current configuration
ztictl config show

# Validate setup
ztictl config validate

# Get help
ztictl --help
ztictl ssm --help
```

## Updating ZTiAWS

### Updating ztictl

**Simple Update (Recommended):**
```bash
# Download to a temporary location to avoid conflicts
curl -L -o /tmp/ztictl "https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/; s/aarch64/arm64/')" && chmod +x /tmp/ztictl && sudo mv /tmp/ztictl /usr/local/bin/ztictl && ztictl --version
```

**Step-by-step Update:**
```bash
# 1. Download latest version to temporary location
curl -L -o /tmp/ztictl "https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/; s/aarch64/arm64/')"

# 2. Make it executable
chmod +x /tmp/ztictl

# 3. Replace the old version
sudo mv /tmp/ztictl /usr/local/bin/ztictl

# 4. Verify the update
ztictl --version
```

**Windows Update:**
```powershell
# Download to temporary location
$tempFile = "$env:TEMP\ztictl.exe"
Invoke-WebRequest -Uri "https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-windows-amd64.exe" -OutFile $tempFile

# Replace existing binary
$toolsDir = "$env:USERPROFILE\Tools"
Move-Item $tempFile "$toolsDir\ztictl.exe" -Force

# Verify update
ztictl --version
```

**Alternative: Platform-specific updates**

*Linux AMD64:*
```bash
curl -L -o /tmp/ztictl https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-amd64 && chmod +x /tmp/ztictl && sudo mv /tmp/ztictl /usr/local/bin/ztictl
```

*Linux ARM64:*
```bash
curl -L -o /tmp/ztictl https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-arm64 && chmod +x /tmp/ztictl && sudo mv /tmp/ztictl /usr/local/bin/ztictl
```

*macOS Intel:*
```bash
curl -L -o /tmp/ztictl https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-darwin-amd64 && chmod +x /tmp/ztictl && sudo mv /tmp/ztictl /usr/local/bin/ztictl
```

*macOS Apple Silicon:*
```bash
curl -L -o /tmp/ztictl https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-darwin-arm64 && chmod +x /tmp/ztictl && sudo mv /tmp/ztictl /usr/local/bin/ztictl
```

**Troubleshooting Updates:**

*If you get "Is a directory" error:*
```bash
# This happens when there's a directory named 'ztictl' in current folder
# Solution: Always download to /tmp/ as shown above
```

*If you get "cannot overwrite non-directory" error:*
```bash
# Check what ztictl currently is
file /usr/local/bin/ztictl
ls -la /usr/local/bin/ztictl

# If it's somehow a directory, remove it first
sudo rm -rf /usr/local/bin/ztictl
# Then retry the installation
```

*To check current version before updating:*
```bash
ztictl --version
```

*To install a specific version (if needed):*
```bash
# Replace v2.1.0 with desired version
curl -L -o /tmp/ztictl "https://github.com/zsoftly/ztiaws/releases/download/v2.1.0/ztictl-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/; s/aarch64/arm64/')" && chmod +x /tmp/ztictl && sudo mv /tmp/ztictl /usr/local/bin/ztictl
```

### Updating Legacy Bash Tools
Navigate to your cloned repository directory and pull the latest changes:

```bash
cd /path/to/ztiaws
git pull origin main
chmod +x ssm authaws
```

If updating from pre-March 2025 (when repository was named "quickssm"), see [docs/deprecated_update_instructions.md](docs/deprecated_update_instructions.md).

## Legacy Bash Tools (Deprecated)

> **⚠️ Deprecation Notice:** The bash tools (`ssm` and `authaws`) are being phased out in favor of the Go binary. They remain available for existing users but new features will only be added to `ztictl`.

If you need to use the legacy bash tools:

**Installation:**
```bash
# Clone repository (required - bash tools cannot be downloaded as binaries)
git clone https://github.com/zsoftly/ztiaws.git
cd ztiaws

# Make executable
chmod +x ssm authaws

# Add to PATH
echo 'export PATH="$PATH:'$(pwd)'"' >> ~/.bashrc  # bash
echo 'export PATH="$PATH:'$(pwd)'"' >> ~/.zshrc   # zsh
source ~/.bashrc  # or ~/.zshrc

# Verify
ssm check
authaws check
```

**Basic Usage:**
```bash
# List instances
ssm cac1  # Canada Central region

# Connect to instance  
ssm i-1234567890abcdef0

# Authenticate with AWS SSO
authaws
```

## Troubleshooting

### Command Not Found
```bash
# Check installation
which ztictl
ztictl --version

# Check PATH (Linux/macOS)
echo $PATH | grep -o "/usr/local/bin"

# Check PATH (Windows)
echo $env:PATH.Split(';') | Select-String "Tools"
```

### Permission Issues
```bash
# Linux/macOS: Ensure executable
sudo chmod +x /usr/local/bin/ztictl

# Windows: Check execution policy
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Architecture Issues
```bash
# Check your system architecture
uname -m        # Linux/macOS
$env:PROCESSOR_ARCHITECTURE  # Windows

# Common mappings:
# x86_64 / AMD64 → use amd64 binary
# aarch64 / arm64 → use arm64 binary  
```

### AWS Configuration
```bash
# Verify AWS CLI setup
aws --version
aws configure list

# Check SSM plugin (for ztictl)
ztictl config check
```

## Why Choose ztictl?

- **🌍 Cross-platform**: Native binaries for Linux, macOS, and Windows
- **⚡ Enhanced performance**: No runtime dependencies, faster execution
- **🔒 Advanced security**: Comprehensive IAM management and automatic cleanup
- **📁 Smart file transfers**: Automatic routing via S3 for large files
- **🛠️ Modern CLI**: Flag-based interface with comprehensive help
- **📊 Better logging**: Thread-safe, timestamped logs with debug capabilities

**Migrate from bash tools** by simply installing ztictl and using similar commands with modern flag syntax!
```
