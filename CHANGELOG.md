# Changelog

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