# Changelog

## [v2.1.0] - 2025-08-04

### Added
- feat: Enhance changelog and release notes generation with commit categorization
- feat: Implement auto-generation of release documentation and changelog
- feat: Add automated release preparation workflow
- feat: Enhance SSM list command to display all EC2 instances with their SSM status
- feat: Introduce version management module and update version references in authaws and ssm scripts
- feat: Add flag-based parameter support to authaws

### Fixed
- fix: Update help message and improve default behavior for no arguments in authaws
- fix: Address PR review feedback and improve validation
- fix: Resolve shellcheck warnings and bash compatibility issues

### Changed
- Auto-generate changelog and release notes for v2.1.0
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
- Enhance logging and color output across SSM commands
- docs: Enhance authaws section with flag-based parameter support and usage examples
- Enhance branch protection: allow pull requests to main from 'release/*' branches
- Enhance CI/CD workflow: include 'release/*' branches in push and pull request triggers


## [v2.1.0] - 2025-08-04

- feat: Implement auto-generation of release documentation and changelog
- feat: Add automated release preparation workflow
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
- Enhance logging and color output across SSM commands
- feat: Enhance SSM list command to display all EC2 instances with their SSM status
- feat: Introduce version management module and update version references in authaws and ssm scripts
- docs: Enhance authaws section with flag-based parameter support and usage examples
- fix: Update help message and improve default behavior for no arguments in authaws
- fix: Address PR review feedback and improve validation
- fix: Resolve shellcheck warnings and bash compatibility issues
- feat: Add flag-based parameter support to authaws
- Enhance branch protection: allow pull requests to main from 'release/*' branches
- Enhance CI/CD workflow: include 'release/*' branches in push and pull request triggers