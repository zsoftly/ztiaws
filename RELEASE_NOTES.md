# ztictl v2.11.0 Release Notes

**Installation:** [Installation Guide](https://github.com/zsoftly/ztiaws/blob/release/v2.11.0/INSTALLATION.md)

**Release Date:** November 13, 2025

## Overview

ztictl is a unified AWS SSM management tool that provides both Go binary and bash script implementations. The Go version (`ztictl`) is the primary implementation with enhanced features, while the bash scripts (`authaws`, `ssm`) are maintained for backward compatibility only.

**Note:** The bash scripts are no longer receiving new features or updates. All development efforts are focused on the Go implementation.

## New Features

- Enhance AWS SSO and CI/CD Integration
- Enhance AWS SSO and CI/CD Integration
- Enhance power operation handling and validation
- Add finalized ZTiAWS demo images and update Markdown paths
- Add files via upload
- Create .gitkeep
- Create .gitkeep
- Add ZTiAWS Demo Documentation (Installation, Authentication, and Use Cases)
- feat: Update EC2 Test Manager script with default subnet and security group IDs, add dry run option, and improve resource discovery logging

## Bug Fixes

- fix: Update build scripts and enhance error handling in OIDC examples
- fix: Update script paths and improve error handling in notification scripts
- fix: Increase HTTP client timeout for EC2 instance detection and add warning for proxy parsing errors

## Other Changes

- Wrap commands in proper markdown code blocks
- docs: finalize ZTiAWS demo documentation with consistent numbering and formatting
- Docs: add proper Markdown headers and syntax highlighting for config section
- docs(README): removed emojis
- docs(ztiaws-demo): finalize demo walkthrough with screenshots and stakeholder-focused updates
- Rename 04-confirm ssm-ec2.png to 04-confirm-ssm-ec2.png
- Rename 12-uploading- loca-to-ec2-file.png to 12-uploading-local-to-ec2-file.png
- Rename 11-creating- folder-file.png to 11-creating-folder-file.png
- docs: reordered demo flow, updated region codes, and refined next steps per review feedback
- refactor: Remove unused file modification time check in IAM cache validation
- Refs #130: Refactor EC2 Manager Script to Dynamically Fetch VPC Resources

**Full Changelog**: https://github.com/zsoftly/ztiaws/compare/v2.10.0...v2.11.0
