# Changelog

## [v2.5.0] - 2025-09-07

### Added
- **Enhanced User Interface**: Complete ASCII-only splash screen redesign
  - Beautiful ASCII art banner for universal terminal compatibility
  - Updated feature showcase highlighting latest capabilities
  - Clean, professional appearance without Unicode dependencies

### Security
- **Comprehensive Security Hardening**: Multiple vulnerability fixes and enhancements
  - **G204 (Command Injection)**: Added input validation for all `exec.CommandContext()` calls with regex-based parameter validation
  - **G304 (Directory Traversal)**: Implemented path validation to prevent file inclusion attacks across all file operations
  - **G301 (File Permissions)**: Hardened directory permissions from 0755 to 0750
  - **G104 (Error Handling)**: Added proper error handling for cleanup operations
  - **G401 (Weak Cryptography)**: Migrated from SHA-1 to SHA-256 for all hash operations
  - **G306 (File Permissions)**: Secured configuration files with 0600 permissions
- **Input Validation Framework**: 
  - Instance ID validation (`i-[0-9a-f]{8,17}`)
  - AWS region format validation (`us-east-1`, `eu-west-2`, etc.)
  - Port number range validation (1-65535)
  - Path traversal protection for all file operations
- **Test Coverage**: Added 45+ security-focused test cases covering all validation scenarios

### Enhanced
- **Code Quality Improvements**: String concatenation optimization and code cleanup
- **Error Handling**: Comprehensive error handling improvements across the codebase
- **Documentation**: Updated feature descriptions to highlight security and performance improvements

### Technical
- **Platform Compatibility**: Enhanced cross-platform support with ASCII-only interface elements
- **Performance**: Optimized string operations and reduced complexity in hot paths
- **Maintainability**: Improved code structure with centralized validation functions

## [v2.4.0] - 2025-09-07

### Added
- **EC2 Power Management Commands**: Complete suite of instance power control operations
  - `ztictl ssm start [instance-id]` - Start stopped EC2 instances
  - `ztictl ssm stop [instance-id]` - Stop running EC2 instances  
  - `ztictl ssm reboot [instance-id]` - Reboot running EC2 instances
  - `ztictl ssm start-tagged --tags <tags>` - Start multiple instances by tag
  - `ztictl ssm stop-tagged --tags <tags>` - Stop multiple instances by tag
  - `ztictl ssm reboot-tagged --tags <tags>` - Reboot multiple instances by tag
- **Multi-Instance Support**: All power commands support `--instances` flag for comma-separated instance IDs
- **Parallel Execution**: Configurable parallel processing with `--parallel` flag (defaults to CPU count)
- **Instance Name Resolution**: Support for both instance IDs (`i-1234...`) and instance names
- **Comprehensive Validation**: Mutual exclusion validation prevents conflicting flag combinations
- **Extensive Test Coverage**: 35+ test scenarios covering all power management functionality

### Enhanced
- **README Documentation**: Updated with power management examples and feature descriptions
- **Help System**: All new commands integrated into ztictl help system
- **Error Handling**: Clear, user-friendly error messages for all validation scenarios

### Examples
```bash
# Start/stop single instances
ztictl ssm start i-1234567890abcdef0 --region cac1
ztictl ssm stop web-server-1 --region use1

# Bulk operations using tags
ztictl ssm start-tagged --tags Environment=Production --region cac1
ztictl ssm stop-tagged --tags ManagedBy=ec2-manager --parallel 5 --region use1

# Multiple specific instances
ztictl ssm reboot --instances i-123,i-456,i-789 --parallel 3 --region cac1
```

## [v2.3.0] - 2025-09-06

### Added
- **Parallel execution for exec-tagged command** - All commands now run in parallel by default for massive performance improvements at scale
- **Instance ID filtering** - New `--instances` flag to explicitly target specific instance IDs (comma-separated)
- **Configurable parallelism** - New `--parallel <N>` flag to control maximum concurrent executions (default: CPU cores)
- **Enhanced execution summaries** - Detailed per-instance timing, success/failure counts, and performance metrics
- **Mutual exclusion validation** - Prevent conflicting usage of `--tags` and `--instances` flags

### Changed
- **Breaking: Sequential execution removed** - All exec-tagged operations now run in parallel for better performance
- **Improved scalability** - Worker pool pattern handles large instance sets efficiently with configurable limits
- **Enhanced error handling** - Better validation and user-friendly error messages for invalid parallel values

### Performance
- **Dramatic speed improvements** - Commands on 20+ instances complete in seconds instead of minutes
- **Resource control** - Configurable parallelism prevents system overload while maximizing throughput
- **Real-time feedback** - Individual execution timing and progress visibility

### Examples
```bash
# Parallel execution with tags (default CPU cores)
ztictl ssm exec-tagged cac1 --tags Environment=production "uptime"

# Custom parallelism for large environments
ztictl ssm exec-tagged use1 --tags Owner=Ditah --parallel 15 "df -h" 

# Direct instance targeting
ztictl ssm exec-tagged cac1 --instances i-123,i-456,i-789 --parallel 5 "systemctl status nginx"
```

## [v2.2.0] - 2025-09-06

### Added
- **Multi-tag filtering for ztictl exec-tagged** - Enhanced `--tags` flag supporting multiple tag filters with AND logic
- **Comprehensive test coverage** - Added unit tests for tag parsing and integration tests for multi-tag functionality
- **Backward compatibility** - Maintained support for legacy single-tag filtering alongside new multi-tag syntax

### Changed
- **Breaking: exec-tagged command syntax** - Changed from positional `<tag-key> <tag-value>` to `--tags key=value,key2=value2` flag format
- **Enhanced documentation** - Updated README files with multi-tag examples and usage patterns
- **Improved error handling** - Better validation and error messages for malformed tag filters

### Examples
```bash
# Single tag filtering
ztictl ssm exec-tagged cac1 --tags Environment=production "df -h"
# Multiple tag filtering (AND logic)  
ztictl ssm exec-tagged use1 --tags Environment=dev,Component=fts,Team=backend "systemctl status nginx"
```

## [v1.6.0] - 2025-08-19

### Added
- **Flag-based parameter support for SSM tool** - New enterprise-friendly syntax with `--region`, `--instance`, `--command` flags
- **Mixed syntax support** - Combination of positional and flag-based parameters (e.g., `ssm cac1 --instance i-1234`)
- **Enhanced user experience** - Self-documenting commands with clear parameter names
- **Professional installation system** - New Makefile with `make install`, `make dev`, `make test` targets
- **Short flag support** - `-h`, `-v`, `-r`, `-i` for common operations
- **Comprehensive parameter parser** - New `src/07_ssm_parameter_parser.sh` module following established patterns
- **Enhanced error messages** - Actionable guidance for missing modules and setup issues

### Changed
- **Backward compatibility maintained** - All existing positional syntax continues to work unchanged
- **Dynamic command naming** - Both `ssm` and `authaws` now use `$(basename "$0")` for consistent help text
- **Improved documentation** - Updated README.md with clear user vs developer installation paths
- **Enhanced testing** - New `tests/QA_SSM_TESTS.md` with comprehensive test scenarios

### Fixed
- **Critical port forwarding bug** - Fixed instance name resolution for port forwarding operations
- **Region validation consistency** - Resolved inconsistent region validation across different SSM commands
- **Duplicate PATH prevention** - Fixed Makefile to prevent duplicate PATH entries in development setup
- **Shellcheck compliance** - Resolved all linting warnings (SC2034, SC2155)

## [v2.1.0] - 2025-08-04

### Added
- feat(ci): add automated Google Chat notifications for PRs and releases with shell scripts
- feat(scripts): add professional notification scripts with embedded Google Chat App Cards
- feat(notifications): implement rich visual formatting following zsoftly-services pattern
- feat: Implement release documentation generator for automated CHANGELOG.md and RELEASE_NOTES.txt creation
- feat: Implement release documentation generator for automated CHANGELOG.md and RELEASE_NOTES.txt creation
- feat: Implement release documentation generator for automated CHANGELOG.md and RELEASE_NOTES.txt creation
- feat: Refactor CHANGELOG.md generation to improve structure and clarity
- feat: Enhance changelog and release notes generation with commit categorization
- feat: Implement auto-generation of release documentation and changelog
- feat: Add automated release preparation workflow
- Enhance logging and color output across SSM commands
- feat: Enhance SSM list command to display all EC2 instances with their SSM status
- feat: Introduce version management module and update version references in authaws and ssm scripts
- feat: Add flag-based parameter support to authaws
- Enhance branch protection: allow pull requests to main from 'release/*' branches
- Enhance CI/CD workflow: include 'release/*' branches in push and pull request triggers

### Fixed
- fix: Update help message and improve default behavior for no arguments in authaws
- fix: Address PR review feedback and improve validation
- fix: Resolve shellcheck warnings and bash compatibility issues

### Changed
- refactor: Improve ANSI code handling and enhance logging error messages
- refactor: Implement dynamic table formatting for EC2 instance display
- refactor: Enhance logging messages for clarity and consistency across IAM and S3 lifecycle management
- refactor: Update test version in ShowSplash test case to 2.2.0-test
- refactor: Update version to 2.1.0 in documentation, build scripts, and tests
- refactor: Update version to 2.1.0 in documentation, build scripts, and tests
- refactor: Simplify NewManager function and enhance logging messages across various components
- refactor: Update logging in Windows build script and enhance EC2 test manager for cross-platform compatibility
- refactor: Improve SSO login command and enhance account/role selection with fuzzy finder
- gofmt -s -w .
- refactor: Migrate logging to centralized package and remove legacy logger
- docs: Enhance authaws section with flag-based parameter support and usage examples


All notable changes to the ZTiAWS project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v2.0.0] - 2025-07-28

### Added
- **Initial release of ztictl** - the Go version of ZTiAWS CLI
- **Cross-platform support** - Native binaries for Linux, macOS, Windows (AMD64 and ARM64)
- **Interactive UI enhancements** - Colorized output, progress bars, and animated splash screens
- **Enhanced configuration management** - YAML-based configuration with validation
- **Comprehensive test suite** - Automated testing with >85% coverage
- **Build automation** - Cross-platform build system with GitHub Actions
- **Improved logging system** - Structured logging with file and console output
- **Interactive profile selection** - Enhanced AWS SSO authentication flow
- **Real-time command feedback** - Live output streaming for remote commands
- **Enhanced error handling** - Detailed error messages with troubleshooting hints

### Changed
- **Complete rewrite from shell scripts to Go** for better performance and maintainability
- **Unified command interface** - Single binary replacing multiple shell scripts
- **Improved user experience** - Interactive menus, real-time feedback, and better error messages
- **Enhanced reliability** - Better error handling, validation, and recovery mechanisms
- **Performance improvements** - Significantly faster execution compared to shell scripts

### Migrated Features
- AWS SSO authentication (enhanced with interactive selection)
- SSM Session Manager connections
- SSM instance listing and management  
- Remote command execution via SSM (`exec` and `exec-tagged` commands)
- File transfer through SSM with S3 support for large files
- Port forwarding through SSM tunnels
- Multi-region support with region shortcuts

## [v1.4.2] - 2025-05-10

### Added
- Remote command execution capabilities
  - New `exec` command to run commands on individual EC2 instances
  - New `exec-tagged` command to run commands across multiple instances with the same tag
  - Real-time status updates during command execution
  - Formatted output display showing both stdout and stderr
- Enhanced error handling for missing instances and command failures

### Fixed
- Improved AWS SSO token management with better cache file detection
- Fixed "base64: invalid input" errors when using exec-tagged command
- Resolved ShellCheck warnings for improved CI pipeline reliability

## [v1.4.1] - 2025-05-10

### Added
- Enhanced error handling for access token retrieval in authaws script

### Fixed
- Improved error messages for SSO configuration issues

## [v1.4.0] - 2025-03-30

### Added
- Renamed auth script from auth_aws to authaws for better usability
- Ensured PATH updates with proper commenting and new line handling

## [v1.3.1] - 2025-03-31

### Changed
- Rebranded repository to ZTiAWS from quickssm
- Updated documentation, README, and issue templates to reflect new branding
- Improved installation and troubleshooting guides

## [v1.3.0] - 2025-05-10

### Added
- Support for running commands on EC2 instances (initially named run_command)
- Improved tests for the SSM script

### Fixed
- Syntax error in detect_os function
- ShellCheck directive placement in cleanup function

## [v1.1.2] - 2025-03-28

### Added
- New auth_aws script for simplified AWS SSO login
- Improved logging functionality
- Better error handling

## [v1.1.0] - 2025-02-06

### Added
- Support for Singapore region (ap-southeast-1)
- Formatted EC2 instance output with clear columns
- PowerShell profile for Windows users

## [v1.0.0] - 2025-01-20 (Initial Release)

### Added
- Core SSM Session Manager functionality
- Support for multiple regions
- Auto-install prompt for AWS SSM plugin
- Basic documentation and README
- Installation support for bash and zsh
- Initial CI/CD setup with badges