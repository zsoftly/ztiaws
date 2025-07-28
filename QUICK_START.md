# Quick Installation & Release Guide

## 🚀 Quick Install

### Linux/macOS
```bash
# Download and install latest version
curl -L -o ztictl "https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m | sed 's/x86_64/amd64/')"
chmod +x ztictl
sudo mv ztictl /usr/local/bin/
```

### Windows (PowerShell)
```powershell
Invoke-WebRequest -Uri "https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-windows-amd64.exe" -OutFile "ztictl.exe"
```

### Verify Installation
```bash
ztictl --version
ztictl ssm list --region us-east-1
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
- **Installation Guide**: [INSTALLATION.md](INSTALLATION.md)
- **Release Process**: [RELEASE.md](RELEASE.md)
- **User Guide**: [README.md](README.md)
