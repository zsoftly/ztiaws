# Scripts Directory

Automation scripts for the ZTiAWS project, focusing on **Google Chat notifications** that provide rich, interactive updates to the development team.

## Why Google Chat App Cards?

These scripts send **professional, interactive notifications** instead of plain text because:

- **Rich visual formatting** with headers, icons, and structured sections
- **Interactive elements** like clickable buttons for direct actions  
- **Enterprise-grade appearance** that matches professional tools
- **Quick access** to related resources without copy/paste

## Notification Scripts

### `send-pr-notification.sh`
**Purpose:** Notify team when PRs are opened to the main branch  
**Triggered:** After security scans complete in CI/CD  
**Benefits:** Enables quick review and collaboration

```bash
# Basic usage
./send-pr-notification.sh \
  --pr-title "Add new feature" \
  --pr-number "123" \
  --pr-url "https://github.com/org/repo/pull/123" \
  --author "developer" \
  --repository "org/repo"
```

### `send-release-notification.sh`  
**Purpose:** Announce new releases to stakeholders  
**Triggered:** After GitHub releases are created  
**Benefits:** Coordinates deployment activities and provides direct access to release assets

```bash
# Basic usage
./send-release-notification.sh \
  --version "v1.0.0" \
  --release-url "https://github.com/org/repo/releases/tag/v1.0.0" \
  --repository "org/repo"
```

## Architecture

- **DRY principle:** Shared utilities in `../src/00_utils.sh`
- **Consistent logging:** Colored output and optional file logging
- **Error handling:** Comprehensive validation and debugging
- **Security:** Base64 webhook encoding support
- **CI/CD integration:** Designed for GitHub Actions workflows

## Configuration

Set `GOOGLE_CHAT_WEBHOOK` in repository secrets (supports plain text or base64 encoding).

For implementation details, see individual script help: `script-name.sh --help`