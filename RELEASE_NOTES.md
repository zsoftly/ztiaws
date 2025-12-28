# ztictl 2.12.0 Release Notes

**Release Date:** December 28, 2025

## What's New

### SSM SSH & RDP Commands

- **`ztictl ssm ssh`** - SSH to EC2 instances via SSM tunnel (no open ports required)
- **`ztictl ssm ssh-config`** - Generate `~/.ssh/config` entries for SSM-based SSH
- **`ztictl ssm rdp`** - RDP tunneling for Windows instances with auto-launch support

### RDS Management Commands

- **`ztictl rds list`** - List all RDS instances in a region
- **`ztictl rds start`** - Start a stopped RDS instance (with `--wait` option)
- **`ztictl rds stop`** - Stop a running RDS instance (with `--wait` option)
- **`ztictl rds reboot`** - Reboot an RDS instance (with `--force-failover` for Multi-AZ)

### Modernized CI/CD Pipeline

- Smart change detection using `dorny/paths-filter` - only runs relevant jobs
- Numbered job names for clear pipeline visibility
- SHA256 checksums included in releases
- One-liner install scripts for all platforms

## Bug Fixes

- **Fixed Ctrl+C handling** in WSL2/Windows Terminal - signals now properly pass to SSM subprocess
- **Fixed script injection vulnerability** in GitHub Actions notification steps
- **Fixed signal channel leak** - proper goroutine cleanup for signal handling
- **Fixed RDS polling** - added exponential backoff for transient errors

## Installation

**Linux/macOS:**

```bash
curl -fsSL https://github.com/zsoftly/ztiaws/releases/latest/download/install.sh | bash
```

**Windows (PowerShell):**

```powershell
irm https://github.com/zsoftly/ztiaws/releases/latest/download/install.ps1 | iex
```

## Documentation Updates

- Fixed inconsistent heading syntax in issue templates
- Updated Go version policy documentation (staying on Go 1.24.x intentionally)
- Fixed outdated version references in CI/CD documentation
- Removed duplicate steps in contributing guide

## Breaking Changes

- **Tag format changed**: Use `2.12.0` instead of `v2.12.0` for releases
- **Release notes format**: Changed from `.txt` to `.md`

**Full Changelog**: https://github.com/zsoftly/ztiaws/compare/v2.11.0...2.12.0
