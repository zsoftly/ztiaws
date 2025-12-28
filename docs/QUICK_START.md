# Quick Installation & Release Guide

## 🚀 Quick Install

> **Choose Your Version**: ZTiAWS offers both stable shell scripts and a new unified Go binary

### 🐚 Shell Scripts (Production Stable)

```bash
# Download production-stable shell scripts
curl -L -o ssm https://raw.githubusercontent.com/zsoftly/ztiaws/main/ssm
curl -L -o authaws https://raw.githubusercontent.com/zsoftly/ztiaws/main/authaws
chmod +x ssm authaws
sudo mv ssm authaws /usr/local/bin/

# Download support files
sudo mkdir -p /usr/local/bin/src
for file in 00_utils.sh 01_regions.sh 02_ssm_instance_resolver.sh 03_ssm_command_runner.sh 04_ssm_file_transfer.sh; do
  curl -L -o "/tmp/$file" "https://raw.githubusercontent.com/zsoftly/ztiaws/main/src/$file"
  sudo mv "/tmp/$file" "/usr/local/bin/src/"
done
```

### 🚀 Go Binary (New Unified Tool)

#### Linux/macOS

```bash
# Download and install latest ztictl
platform="linux"  # or "darwin" for macOS
arch="amd64"       # or "arm64"

curl -L -o ztictl.tar.gz "https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-${platform}-${arch}.tar.gz"
tar -xzf ztictl.tar.gz
chmod +x ztictl-${platform}-${arch}
sudo mv ztictl-${platform}-${arch} /usr/local/bin/ztictl
rm ztictl.tar.gz
```

#### Windows (PowerShell)

```powershell
# Download and install latest ztictl
Invoke-WebRequest -Uri "https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-windows-amd64.zip" -OutFile "ztictl.zip"
Expand-Archive -Path "ztictl.zip" -DestinationPath "." -Force
$installDir = "$env:USERPROFILE\bin"
New-Item -ItemType Directory -Force -Path $installDir
Move-Item ztictl-windows-amd64.exe "$installDir\ztictl.exe"
Remove-Item ztictl.zip
```

### Verify Installation

```bash
# Shell scripts
ssm --version
authaws --version

# Go binary
ztictl --version

# Test functionality
ssm list --region us-east-1        # Shell script
ztictl ssm list --region us-east-1 # Go binary
```

## 📦 Quick Release Process

### 1. Create Tag and Release

```bash
# Update version and create tag
git tag -a v1.1.0 -m "Release v1.1.0"
git push origin v1.1.0
```

### 2. GitHub Actions Automatically:

- ✅ Builds for all platforms (Linux, macOS, Windows)
- ✅ Runs tests and quality checks
- ✅ Creates GitHub release with binaries
- ✅ Makes binaries available for download

### 3. Verify Release

```bash
# Check build status
gh run list --workflow=build.yml

# Test download
curl -I https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-amd64
```

## 📚 Full Documentation

- **Installation Guide**: [INSTALLATION.md](../INSTALLATION.md)
- **Release Process**: [RELEASE.md](development/RELEASE.md)
- **User Guide**: [README.md](../README.md)
