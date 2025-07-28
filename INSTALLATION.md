# ZTiAWS Installation Guide

ZTiAWS provides AWS Systems Manager operations through two complementary tools:

> **🏗️ Dual Installation Strategy**:
> - **Shell Scripts** (`ssm` & `authaws`) - **Production Stable** (v1.4.x) - Battle-tested, in active production use
> - **Go Binary** (`ztictl`) - **Preview/Testing** (v2.0.x) - New unified tool with enhanced features
> 
> **Recommendation**: Install both versions. Use shell scripts for production workflows, test `ztictl` for new features.

## Prerequisites
- AWS CLI configured with appropriate credentials
- EC2 instances with SSM agent installed and proper IAM roles

## Installation Methods

### 📜 Shell Scripts (Production Stable - Recommended for Production)

**Status**: ✅ Production stable, actively maintained, battle-tested  
**Version**: v1.4.x series  
**Use Case**: Production workflows, established environments

#### Quick Install
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

#### Usage Examples
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

---

### 🚀 Go Binary (Preview/Testing - New Features)

**Status**: 🧪 Preview/Testing phase, active development  
**Version**: v2.0.x series  
**Use Case**: Testing new features, development environments, future migration

The new Go-based unified tool that combines both `ssm` and `authaws` functionality with enhanced features and better performance.

#### Quick Install Options

**Option A: Direct Binary Download (Recommended - No extraction needed)**
```bash
# Linux AMD64
curl -L -o ztictl https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-amd64
chmod +x ztictl
sudo mv ztictl /usr/local/bin/

# Linux ARM64  
curl -L -o ztictl https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-arm64
chmod +x ztictl
sudo mv ztictl /usr/local/bin/

# macOS Intel
curl -L -o ztictl https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-darwin-amd64
chmod +x ztictl
sudo mv ztictl /usr/local/bin/

# macOS Apple Silicon (M1/M2/M3)
curl -L -o ztictl https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-darwin-arm64
chmod +x ztictl
sudo mv ztictl /usr/local/bin/

# Verify installation
ztictl --version
```

**Option B: Archive Download (If you prefer archives)**
```bash
# Linux AMD64
curl -L -o ztictl.tar.gz https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-amd64.tar.gz
tar -xzf ztictl.tar.gz
chmod +x ztictl-linux-amd64
sudo mv ztictl-linux-amd64 /usr/local/bin/ztictl
rm ztictl.tar.gz

# Linux ARM64
curl -L -o ztictl.tar.gz https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-arm64.tar.gz
tar -xzf ztictl.tar.gz
chmod +x ztictl-linux-arm64
sudo mv ztictl-linux-arm64 /usr/local/bin/ztictl
rm ztictl.tar.gz

# macOS Intel
curl -L -o ztictl.tar.gz https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-darwin-amd64.tar.gz
tar -xzf ztictl.tar.gz
chmod +x ztictl-darwin-amd64
sudo mv ztictl-darwin-amd64 /usr/local/bin/ztictl
rm ztictl.tar.gz

# macOS Apple Silicon
curl -L -o ztictl.tar.gz https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-darwin-arm64.tar.gz
tar -xzf ztictl.tar.gz
chmod +x ztictl-darwin-arm64
sudo mv ztictl-darwin-arm64 /usr/local/bin/ztictl
rm ztictl.tar.gz

# Verify installation
ztictl --version
```

#### Windows Installation

**Option A: Direct Binary Download**
```powershell
# Windows AMD64
Invoke-WebRequest -Uri "https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-windows-amd64.exe" -OutFile "ztictl.exe"

# Windows ARM64
Invoke-WebRequest -Uri "https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-windows-arm64.exe" -OutFile "ztictl.exe"

# Move to a directory in your PATH (optional)
Move-Item ztictl.exe $env:USERPROFILE\bin\ztictl.exe

# Verify installation
ztictl --version
```

**Option B: Archive Download**
```powershell
# Windows AMD64
Invoke-WebRequest -Uri "https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-windows-amd64.zip" -OutFile "ztictl.zip"
Expand-Archive ztictl.zip -DestinationPath .
Move-Item .\ztictl-windows-amd64.exe $env:USERPROFILE\bin\ztictl.exe
Remove-Item ztictl.zip

# Verify installation
ztictl --version
```

#### ztictl Usage Examples
```bash
# Configure (replaces authaws configure)
ztictl auth configure

# List instances (replaces ssm list)
ztictl ssm list

# Connect to instance (replaces ssm connect)
ztictl ssm connect i-1234567890abcdef0

# Run command (replaces ssm command)
ztictl ssm exec i-1234567890abcdef0 "uptime"

# Transfer files (replaces ssm transfer)
ztictl ssm transfer local-file.txt i-1234567890abcdef0:/tmp/

# Enhanced features (new in ztictl)
ztictl config show          # Show current configuration
ztictl config validate      # Validate setup
ztictl ssm manage           # Advanced SSM management
ztictl cleanup             # Cleanup resources
```

---

## Migration Guide

### For Shell Script Users

If you're currently using the shell scripts and want to test `ztictl`:

1. **Keep your current shell scripts** - they remain fully supported
2. **Install `ztictl` alongside** - both can coexist
3. **Test `ztictl` commands** - use the command mapping above
4. **Gradually migrate** - move workflows when comfortable

### Command Mapping

| Shell Script Command | ztictl Equivalent | Notes |
|---------------------|-------------------|-------|
| `authaws configure` | `ztictl auth configure` | Enhanced configuration |
| `ssm list` | `ztictl ssm list` | Same functionality |
| `ssm connect <id>` | `ztictl ssm connect <id>` | Same functionality |
| `ssm command <id> <cmd>` | `ztictl ssm exec <id> <cmd>` | Enhanced output |
| `ssm transfer <src> <dst>` | `ztictl ssm transfer <src> <dst>` | Improved progress |
| N/A | `ztictl config show` | New feature |
| N/A | `ztictl config validate` | New feature |
| N/A | `ztictl ssm manage` | New feature |
| N/A | `ztictl cleanup` | New feature |

---

## Troubleshooting

### Common Issues

#### Binary Not Found After Download
```bash
# Check if binary was downloaded
ls -la ztictl*

# Ensure it's executable
chmod +x ztictl

# Check if it's in PATH
echo $PATH
which ztictl
```

#### Permission Denied
```bash
# Make sure binary is executable
chmod +x ztictl

# If installing system-wide, use sudo
sudo mv ztictl /usr/local/bin/
```

#### Wrong Architecture
```bash
# Check your system architecture
uname -m
# x86_64 = amd64
# aarch64 = arm64

# Download the correct binary for your architecture
```

#### Version Check
```bash
# For shell scripts
ssm --version
authaws --version

# For Go binary
ztictl --version
```

---

## Development Installation

### From Source (Go Binary)
```bash
# Clone repository
git clone https://github.com/zsoftly/ztiaws.git
cd ztiaws/ztictl

# Build from source
go mod download
go build -o ztictl ./cmd/ztictl

# Install
sudo mv ztictl /usr/local/bin/
```

### Shell Scripts Development
```bash
# Clone repository
git clone https://github.com/zsoftly/ztiaws.git
cd ztiaws

# Make scripts executable
chmod +x ssm authaws

# Symlink for development
sudo ln -sf "$(pwd)/ssm" /usr/local/bin/ssm
sudo ln -sf "$(pwd)/authaws" /usr/local/bin/authaws
```

---

## Support

- 📝 **Issues**: [GitHub Issues](https://github.com/zsoftly/ztiaws/issues)
- 📖 **Documentation**: [GitHub Wiki](https://github.com/zsoftly/ztiaws/wiki)  
- 🔄 **Updates**: Watch the repository for releases

For production use, stick with the shell scripts until `ztictl` reaches stable status.
