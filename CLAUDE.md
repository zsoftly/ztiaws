# ZTiAWS - Claude Memory File

## Repository Overview
**ZTiAWS** (ZSoftly Tools for AWS) - A collection of AWS management CLI tools by ZSoftly that simplify AWS SSO authentication and EC2 Systems Manager operations.

## Architecture
- **Legacy Production**: Bash scripts (`authaws`, `ssm`) - currently in production
- **Next-Gen Go Tool**: `ztictl` - modern replacement with enhanced features
- **Migration Strategy**: Both tools coexist during transition period

## Key Directories

### Root Level
- `src/` - Bash script modules (00_*.sh, 01_*.sh, etc.)
- `ztictl/` - Go implementation (primary focus)
- `docs/` - Documentation (IAM_PERMISSIONS.md, TROUBLESHOOTING.md, etc.)
- `tools/` - Utility scripts for testing and releases

### Go Implementation (`ztictl/`)
```
ztictl/
├── cmd/ztictl/           # CLI commands
│   ├── main.go           # Entry point
│   ├── root.go           # Root command (Version: 2.1.0)
│   ├── auth.go           # AWS SSO authentication
│   ├── config.go         # Configuration management
│   ├── ssm*.go          # SSM operations (connect, exec, transfer, list, etc.)
│   └── table_formatter.go
├── internal/             # Internal packages
│   ├── auth/sso.go       # SSO authentication logic
│   ├── config/config.go  # Configuration management
│   ├── splash/splash.go  # Welcome screen
│   └── ssm/              # SSM managers (IAM, S3 lifecycle)
├── pkg/                 # Public packages  
│   ├── aws/             # AWS clients & utilities
│   ├── colors/          # Terminal colors
│   ├── errors/          # Error handling
│   └── logging/         # Logging infrastructure
└── go.mod               # Go 1.24, AWS SDK v2
```

## Key Features

### ztictl (Go Version)
- **Cross-platform**: Linux, macOS, Windows (AMD64/ARM64)
- **Smart file transfers**: <1MB direct SSM, ≥1MB via S3 with lifecycle management
- **Advanced IAM**: Temporary policies with automatic cleanup and filesystem locking
- **Professional logging**: Thread-safe, timestamped, platform-specific locations
- **Modern CLI**: Cobra/Viper with comprehensive help and validation

### Legacy Bash Tools
- **authaws**: AWS SSO authentication with interactive account/role selection
- **ssm**: EC2 instance management via SSM (connect, exec, file transfer, port forwarding)

## Common Commands

### Development
```bash
# Build Go version
cd ztictl && make build-local

# Run tests  
make test
```

### **Standard Development Quality Checklist**
**IMPORTANT**: Run these commands after ANY development work in the `ztictl/` directory:

```bash
# Navigate to ztictl directory
cd ztictl

# 1. Format code
go fmt ./...

# 2. Check for issues
go vet ./...

# 3. Verify compilation
go build ./...

# 4. Run all tests
go test ./...

# 5. Run linter (golangci-lint must be installed)
export PATH=$PATH:$(go env GOPATH)/bin
golangci-lint run ./...
```

**Installation of golangci-lint** (one-time setup):
```bash
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest
```

**These quality checks are MANDATORY before:**
- Creating commits
- Opening pull requests  
- Considering development work "complete"

### Usage Examples
```bash
# ztictl (recommended)
ztictl config check
ztictl auth login
ztictl ssm list --region us-east-1
ztictl ssm connect i-1234567890abcdef0 --region us-east-1

# Legacy bash tools
authaws --check
ssm cac1  # List instances in Canada Central
ssm i-1234abcd  # Connect to instance
```

## Configuration
- **ztictl**: `~/.ztictl.yaml` (YAML format)
- **Legacy tools**: `.env` file with SSO_START_URL, SSO_REGION, DEFAULT_PROFILE

## Build & Release
- **Build system**: Makefiles for both root and ztictl
- **CI/CD**: GitHub Actions for cross-platform builds
- **Release process**: Git tags trigger automated builds and releases

## Important Files
- `README.md` - Main documentation
- `ztictl/README.md` - Go tool documentation  
- `INSTALLATION.md` - Platform-specific setup instructions
- `CONTRIBUTING.md` - Development guidelines
- `docs/IAM_PERMISSIONS.md` - Required AWS permissions

## Development Notes
- Repository uses MIT License
- Active development focused on ztictl Go implementation
- Bash tools maintained for backward compatibility
- Cross-platform testing important for Go version
- AWS CLI and Session Manager plugin required

## Testing Commands
```bash
# Check system requirements
ztictl config check  # or authaws --check

# Verify installation
ztictl --version
authaws --help
ssm --help
```