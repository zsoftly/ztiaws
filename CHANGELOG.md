# Changelog

## [v2.9.0] - 2025-10-09

### Added
- feat: add configurable display item to improve UX
- feat(auth): improve account and role selection UI

### Fixed
- fix: add notes for pointer usage in fuzzy finder functions


## [v2.8.2] - 2025-10-07

### Added
- AWS SSO pagination support: Fetch all available accounts and roles instead of only first page (5-10 items)
- Users can now search and select from complete account/role inventory using fuzzy finder
- Comprehensive pagination tests covering single/multiple pages, empty results, and error scenarios
- Modern Go error handling patterns using `errors.As()` throughout codebase

### Fixed
- Session Manager Plugin detection: Correctly validate exit code 255 instead of accepting any exit error
- Authentication error handling: Return `(false, nil)` on config errors to maintain function contract
- Import ordering: Standardized to stdlib, external, internal per Go conventions
- Code quality: Removed empty else blocks, improved error wrapping with `%w` format verb


## [v2.8.1] - 2025-10-03

### Added
- feat: add automatic version update check (queries GitHub for latest release)
- feat: implement 24-hour version check caching to reduce API calls

### Fixed
- fix: correct semantic version comparison by numerically comparing major.minor.patch (e.g., 2.10.0 > 2.2.0)
- fix: clean shell completion output by skipping splash/log output during completion requests
- fix: change "Using config file" log message from Info to Debug level to prevent unwanted output


## [v2.8.0] - 2025-09-15

### Added
- feat: enhance multi-OS support and automatic platform detection in ztictl
- feat: enhance Windows command validation and update breaking changes in platform builders
- Implement Linux and Windows command builders for platform-specific operations
- feat: add cross-platform test writing guidelines and improve error handling in sample config tests

### Fixed
- fix: update AMI ID handling and improve logging for EC2 instance creation
- fix: ensure Unix-style path handling in LinuxBuilder commands
- fix: Resolve Windows test isolation issues in TestAuthLoginCmd
- fix: Resolve formatting issues in ztictl
- fixed pipeline failure issure
- fix: disable Go module caching in CI to resolve go.sum path issue
- fix: Resolve config validation issues in ztictl

### Changed
- refactor: remove outdated breaking change comments and improve code clarity
- Refactor SSM client management and enhance platform builders
- Refactor config tests for improved clarity and coverage
- Fix Windows CI: Add timeout and environment isolation to TestExecCommandSeparationOfConcerns
- Fix Windows CI: Add environment isolation to TestInitializeConfigFile
- Fix Windows CI: Complete environment isolation for config tests
- Fix Windows CI: Add timeout and environment isolation to TestConnectionSeparationOfConcerns
- Fix Windows CI: Add getUserHomeDir() to root.go for complete environment isolation
- Fix Windows CI: Add missing test isolation to auth tests
- Fix Windows CI: Handle config loading gracefully in CI environments


## [v2.7.0] - 2025-09-10

### Added
- feat: auto-update version in Makefile and root.go during release
- Enhance bash completion installation logic: add support for system path installation with sudo, improve path validation, and update test cases for mocked sudo behavior.
- Add AWS credential handling to tests: disable EC2 IMDS to prevent timeouts and CI/CD failures
- Enhance test setup and maintenance guidelines
- Enhance AWS SSO configuration: simplify domain ID input, set default region to ca-central-1, and improve URL construction logic. Add tests for domain ID extraction and configuration validation.
- Add comprehensive documentation for ztictl commands, configuration, and multi-region operations
- Implement comprehensive AWS region validation and refactor related tests
- Add multi-region command execution tests and region configuration setup

### Changed
- Refactor test credentials: replace hardcoded AWS credentials with mock values from testutil for improved test reliability and maintainability
- Refactor test functions: rename TestPathValidation to TestTransferPathValidation for clarity and consistency
- Refactor AWS credential handling in tests: centralize mock credentials, improve test environment setup, and remove deprecated code
- Refactor logging initialization and improve test coverage
- Refactor tilde path expansion logic in TestExpandPathTildeExpansion and remove unused mockStdin function
- Refactor AWS SSO configuration and validation
- chore: remove deprecated v2.6.1 section from CHANGELOG.md


## [v2.6.1] - 2025-09-08

### Added
- feat: update Go version to 1.25 and enhance golangci-lint configuration for improved code quality
- feat: add golangci-lint configuration and update SHA1 usage comments for AWS CLI compatibility
- feat: switch from SHA256 to SHA1 for cache filename generation to ensure AWS CLI compatibility
- feat: improve release process documentation with clearer steps and enhanced checklist
- feat: update release notes generation script to use dynamic installation guide link and improve formatting
- feat: update usage instructions and enhance release notes format in documentation generator
- feat: update installation guide link and enhance release notes with overview and deprecation notice
- feat: enhance security by refactoring path validation and adding comprehensive tests for directory traversal detection
- feat: add #nosec annotations to suppress security warnings for various print and environment setup operations
- feat: enhance security by adding #nosec annotations for file read operations
- feat: add #nosec annotations for path validation in file read operations
- feat: enhance security by updating directory permissions to 0750 and adding path validation to prevent directory traversal
- feat: update file permissions to 0600 for enhanced security across configuration and logging files
- feat: enhance security with directory traversal prevention and comprehensive path validation; update changelog and README
- feat: release version 2.5.0 with enhanced security, input validation, and UI improvements; update changelog and README
- feat: update changelog and README for EC2 power management commands; bump version to 2.4.0
- feat: implement the EC2 instance power management commands for the ztictl tool
- feat: Add validation for mutual exclusion of --tags and --instances flags in exec-tagged command
- feat: Update version to 2.3.0 and enhance changelog with new features, changes, and examples for exec-tagged command
- feat: Enhance exec-tagged command with parallel execution and new flags for instance targeting
- feat: Update changelog, README, and version to support multi-tag filtering in exec-tagged command
- feat: Enhance exec-tagged command to support multiple tags and add validation tests
- feat: Update notification messages in CI/CD pipeline to include emojis for better visibility
- feat: Implement embedded notifications for test failures and successes in CI/CD pipeline
- feat: Implement unified PR notification for test results and status summary
- feat: Add embedded notifications for test failures and successes in CI/CD pipeline
- feat: add tests
- feat: Add professional installation system with Makefile
- feat: Add flag-based parameters support to SSM tool
- feat(notifications): implement shell scripts with Google Chat App Cards
- feat(ci): add automated Google Chat notifications for PRs and releases
- Add automated release notifications for Google Chat
- feat: Revise installation instructions and update ztictl usage examples for clarity and consistency
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
- fix: update golangci-lint and gosec installation scripts for accuracy and consistency
- fix: update instance tag filtering to use awssdk.String for better clarity and consistency
- fix: Enhance environment variable handling for cross-platform compatibility in tests
- fix: Skip LoggerMethods test on Windows CI due to output formatting differences
- fix: Enhance PR notification with detailed status reporting and improve logging tests for Windows CI
- fix: Enhance PR notification with status and message, and skip logging test on Windows in CI
- fix: Update shellcheck directives for clarity and safety in scripts
- fix: Enhance logging tests by preserving logger state during failure scenarios
- fix: Improve string quoting for better readability in uninstall script
- fix: Update CI/CD workflows to enhance testing and path filters
- fix: Restore logger state after TestLogFileCreationFailure
- fix: Update icon in Google Chat App Card payload for deployment status
- fix: Tests check platform-appropriate instructions
- fix: Tests check platform-appropriate instructions
- fix: Tests check platform-appropriate instructions
- fix: Tests check platform-appropriate instructions
- fix: Tests check platform-appropriate instructions
- fix: Tests check platform-appropriate instructions
- fix: Tests check platform-appropriate instructions
- fix: Tests check platform-appropriate instructions
- fix: Add validation for unrecognized positional arguments
- fix: Address all Ditah and Copilot feedback
- fix: Implement Copilot review suggestions for SSM tool
- fix: Improve Makefile duplicate PATH prevention
- fix: Resolve region validation inconsistency for exec commands
- fix: Make authaws help text consistent with SSM dynamic command display
- fix: Resolve shellcheck linting issues
- fixed regex errors for google chat webhook
- fix(scripts): handle both plain text and base64 encoded webhook URLs
- fix(ci): update workflow path filters to include notification-related files
- fix: Update GitHub Release step to use RELEASE_NOTES.txt for release notes
- fix: Update help message and improve default behavior for no arguments in authaws
- fix: Address PR review feedback and improve validation
- fix: Resolve shellcheck warnings and bash compatibility issues

### Changed
- chore: remove deprecated v2.6.0 section from CHANGELOG.md
- chore: update version to 2.5.2 in changelog and source files
- test: enhance validation in various test cases for improved coverage and reliability
- Refactor SSM management commands to improve error handling and separation of concerns
- refactor: Consolidate CI/CD workflows and enhance shell script testing in build pipeline
- refactor: Remove legacy CI/CD workflows and update documentation for optimized pipeline structure
- Revert "removed version line from 07 module file"
- Revert "updated release date"
- Revert "updated release date"
- Revert "chore: Prepare for v2.2.0 release"
- chore: Prepare for v2.2.0 release
- updated release date
- updated release date
- removed version line from 07 module file
- Remove the XDG directory setup from install.sh since we simplified the config approach
- refactor: Simplify to address core feedback without overengineering
- docs: Update versioning documentation for v1.6.0 release
- chore: Bump version to 1.6.0 for flag-based parameter feature
- docs: Achieve complete CLI documentation consistency
- make copilot changes
- all logging goes to stderr
- cleaned the code and applied best practices to sergical precision.
- modified condition for pr-notification
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

## [v2.5.2] - 2025-09-07

### Security
- **Critical Directory Traversal Fix**: Enhanced cross-platform path validation security
  - **Directory Traversal Prevention**: Fixed Windows-style directory traversal vulnerability (`..\\` patterns)
  - **Cross-Platform Security**: Added comprehensive Windows (`\`) and Unix (`/`) path separator validation
  - **Enhanced Path Validation**: Now blocks all directory traversal patterns: `../`, `..\\`, `/../`, `\\..\\`, `/..`, `\\..`
  - **Comprehensive Testing**: Added 87+ platform-specific security test cases covering Windows UNC paths, drive letters, and Unix system paths
  - **Runtime Adaptation**: Tests automatically adapt to current OS for platform-specific attack vector validation

### Technical
- **Code Quality**: Improved boolean condition readability in security validation logic
- **Cross-Platform Compatibility**: Replaced string concatenation with `filepath.Join()` for proper path handling
- **Test Coverage**: Enhanced security package with Windows and Unix specific validation scenarios

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