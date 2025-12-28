# ZTiAWS - Claude Memory File

## Repository Overview

**ZTiAWS** (ZSoftly Tools for AWS) - A collection of AWS management CLI tools by ZSoftly that simplify AWS SSO authentication and EC2 Systems Manager operations.

## Architecture

- **Legacy Production**: Bash scripts (`authaws`, `ssm`) - currently in production
- **Next-Gen Go Tool**: `ztictl` - modern replacement with enhanced features
- **Migration Strategy**: Both tools coexist during transition period

## Key Directories

### Root Level

- `src/` - Bash script modules (00*\*.sh, 01*\*.sh, etc.)
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

### Post-Development Cleanup

After any development session, run:

```bash
# Clean test artifacts
go clean -testcache
rm -rf coverage/ *.out *.test *.html

# Remove any temporary test files created for debugging
find . -name "*_temp_test.go" -delete
find . -name "test_*.sh" -delete
```

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
  - Default region: `ca-central-1` (both for SSO and operations)
  - SSO setup: Only asks for domain ID (e.g., `d-1234567890` or `zsoftly`)
  - Automatically constructs full URL: `https://{domain-id}.awsapps.com/start`
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

## Test Infrastructure

**AWS Credential Handling in Tests**:

- Centralized test utilities in `internal/testutil/aws.go` define mock AWS credentials
- Each test package has an `init_test.go` file that calls `testutil.SetupAWSTestEnvironment()`
- Mock credentials (AWS documentation examples):
  - Access Key: `AKIAIOSFODNN7EXAMPLE`
  - Secret Key: `wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY`
  - Session Token: `test-session-token`
  - Region: `ca-central-1` (matches project default)
- Benefits:
  - Single source of truth for test credentials
  - Easy to update across all tests
  - Prevents real AWS API calls in tests
  - Ensures consistent behavior across all platforms (Linux, macOS, Windows)
- Located in: `cmd/ztictl/init_test.go`, `internal/ssm/init_test.go`, `internal/system/init_test.go`
- Makefile also sets these environment variables for `make test` command

**Cross-Platform Test Writing Guidelines**:
**CRITICAL**: All tests MUST be platform-agnostic and work on Unix (Linux/macOS) AND Windows:

- **File Paths**:
  - ALWAYS use `filepath.Join()` instead of hardcoding paths with `/` or `\`
  - NEVER use absolute Unix paths like `/tmp` or `/var/log` - use `t.TempDir()` or `os.TempDir()`
  - NEVER use Windows-specific paths like `C:\` or `D:\`
- **Path Separators**:
  - Use `filepath.Separator` when you need the OS-specific separator
  - Use `filepath.ToSlash()` to convert paths to forward slashes for URLs/YAML
  - Use `filepath.FromSlash()` to convert from forward slashes to OS format
- **Invalid Path Testing**:
  - Don't rely on Unix-specific invalid paths like `/nonexistent`
  - Create reliably invalid paths by placing files where directories are expected
  - Use permission-based failures cautiously as they behave differently across OS
- **Home Directory**:
  - Use `os.UserHomeDir()` instead of relying on `$HOME` (Unix) or `%USERPROFILE%` (Windows)
  - When testing with environment variables, check both `HOME` and `USERPROFILE`
- **File Permissions**:
  - Be aware that Windows doesn't support Unix permission bits the same way
  - Read-only directories behave differently on Windows
- **Test Examples**:

  ```go
  // BAD - Unix-specific path
  invalidPath := "/nonexistent/directory/config.yaml"

  // GOOD - Platform-agnostic approach
  invalidPath := filepath.Join(t.TempDir(), "subdir", "config.yaml")
  // Then create a file at "subdir" to make the path invalid
  ```

## Test File Management Guidelines

**IMPORTANT**: When creating test files or scripts during development:

### Temporary Files to Clean Up

1. **One-time test files**: Delete after use
   - `*_repair_test.go`, `*_init_test.go` (unless part of permanent suite)
   - `test_coverage.sh`, `run_tests.sh` (temporary scripts)
2. **Coverage artifacts**: Remove after review

   ```bash
   rm -rf coverage/ *.out *.html
   ```

3. **Test binaries and cache**: Clean after testing
   ```bash
   go clean -testcache
   rm -f *.test
   ```

### Permanent Test Files (Keep These)

- `*_test.go` files that test actual functionality
- Test fixtures in `testdata/` directories
- Benchmark tests for performance validation

### Best Practices

- **Before committing**: Clean up all temporary test artifacts
- **After debugging**: Remove one-off test files
- **Duplicate tests**: Update existing tests rather than creating new files
- **Use .gitignore**: Ensure coverage/, _.out, _.test are ignored

## Code Maintenance Guidelines

**CRITICAL**: When removing deprecated, redundant, or obsolete code:

- **NO COMMENTS**: Never leave comments explaining what was removed or why
- **CLEAN REMOVAL**: Completely remove all traces of deprecated functionality
- **NO DEAD CODE**: Remove entire functions, variables, and imports that are no longer needed
- **NO EXPLANATORY COMMENTS**: Never add comments like "// Removed X", "// Deprecated", or "// No longer needed"
- **COMPLETE CLEANUP**: If removing a feature, remove ALL related code including:
  - Function definitions
  - Variable declarations
  - Import statements
  - Test functions
  - Documentation references
  - Configuration options
- **Examples of what NOT to do**:

  ```go
  // BAD - Don't do this:
  // Removed deprecated auth method
  // func oldAuthMethod() { } // Deprecated

  // GOOD - Just remove it completely with no trace
  ```

- **Principal**: When deprecating/removing code, act as a principal engineer - leave the codebase cleaner with no remnants of removed functionality

## Security Validation Guidelines

**IMPORTANT**: Security validations must be comprehensive:

- **Path Traversal Protection**:
  - **Use `ztictl/pkg/security` package**: The project has a comprehensive security package
  - Call `security.ContainsUnsafePath()` for path validation
  - Call `security.ValidateFilePath()` for directory-scoped validation
  - The security package handles all patterns: `../`, `/..`, `..\\`, `\\..`, null bytes, etc.
  - Special handling needed for Windows UNC paths (validate after server/share prefix)
- **Command Injection Prevention**:
  - Validate all user inputs against strict patterns
  - Use parameterized commands where possible
  - Escape shell arguments properly for each platform
- **Here-String Validation**:
  - Check for terminator sequences like `'@` in PowerShell here-strings
  - Validate base64 data before embedding in commands

## Logging Best Practices

**CRITICAL**: Proper logging configuration:

- **Logger Instances**:
  - Use dependency injection for loggers (pass them as parameters)
  - Never hardcode debug levels in production code
  - Provide factory methods like `NewDetectorWithLogger()` for custom loggers
- **Sensitive Data**:
  - Never log full commands at Info level (may contain secrets)
  - Use Debug level for detailed command logging
  - Log command length or hash at Info level instead

## Resource Management

**IMPORTANT**: Prevent resource leaks:

- **AWS Client Pooling**:
  - Reuse AWS clients across operations
  - Implement client pools with proper synchronization
  - Cache clients by region to avoid recreation
- **Thread Safety**:
  - Always use mutexes for shared map access
  - Prefer RWMutex for read-heavy operations
  - Lock before checking and modifying shared state
