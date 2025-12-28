# ztictl Release Process

## Overview

Simple release process - manually write release notes, tag, and push.

## Release Steps

### 1. Ensure Main is Up-to-Date

```bash
git checkout main
git pull origin main
```

### 2. Update Release Notes

Edit `RELEASE_NOTES.md` with:

- What's new (features)
- Bug fixes
- Breaking changes (if any)
- Installation instructions

### 3. Commit Release Notes

```bash
git add RELEASE_NOTES.md
git commit -m "Release X.Y.Z"
git push origin main
```

### 4. Create and Push Tag

**Important:** Use format `X.Y.Z` (no `v` prefix)

```bash
git tag X.Y.Z
git push origin X.Y.Z
```

This triggers the automated pipeline which:

- Builds binaries for 6 platforms (Linux/macOS/Windows x AMD64/ARM64)
- Generates SHA256 checksums
- Creates GitHub Release with all artifacts
- Includes install scripts (install.sh, install.ps1)
- Sends notification to Google Chat

### 5. Verify Release

Check the GitHub Actions workflow and verify:

- All builds completed successfully
- GitHub Release was created
- All binaries and checksums are attached

## Quick Reference

```bash
# Full release flow
git checkout main && git pull origin main
# Edit RELEASE_NOTES.md
git add RELEASE_NOTES.md && git commit -m "Release X.Y.Z"
git push origin main
git tag X.Y.Z && git push origin X.Y.Z
```

## Installation (Post-Release)

Users can install using one-liner commands:

**Linux/macOS:**

```bash
curl -fsSL https://github.com/zsoftly/ztiaws/releases/latest/download/install.sh | bash
```

**Windows (PowerShell):**

```powershell
irm https://github.com/zsoftly/ztiaws/releases/latest/download/install.ps1 | iex
```

## Troubleshooting

### Build Failures

```bash
# Check workflow logs
gh run list
gh run view <run-id> --log

# Test locally
cd ztictl && make build
```

### Version Issues

- Tag format must be `X.Y.Z` (no `v` prefix)
- Version is auto-detected from git tags via `git describe`
