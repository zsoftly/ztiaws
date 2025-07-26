# ztictl Release Process

This document outlines the complete process for releasing new versions of ztictl, including versioning, building, testing, and distribution.

## Release Overview

ztictl uses semantic versioning (SemVer) and automated GitHub Actions for building and releasing cross-platform binaries.

### Release Types
- **Major** (1.0.0 → 2.0.0): Breaking changes
- **Minor** (1.0.0 → 1.1.0): New features, backward compatible
- **Patch** (1.0.0 → 1.0.1): Bug fixes, backward compatible

## Prerequisites

### Required Access
- Write access to the GitHub repository
- Ability to create tags and releases
- Understanding of semantic versioning

### Tools Needed
- Git
- Go 1.24+
- GitHub CLI (optional, but recommended)

## Release Process

### 1. Pre-Release Preparation

#### Update Version Documentation
```bash
# Update version references in documentation
grep -r "version.*1\.0\.0" docs/ README.md INSTALLATION.md
# Update any hardcoded version references
```

#### Run Full Test Suite
```bash
cd ztictl

# Run all tests
make test

# Build all platforms locally to verify
make build

# Test key functionality
./ztictl --version
./ztictl ssm list --region us-east-1  # Replace with valid region
```

#### Update CHANGELOG.md
Create or update `CHANGELOG.md` with:
```markdown
## [1.1.0] - 2025-07-26

### Added
- New feature X
- Support for Y

### Changed
- Improved Z performance

### Fixed
- Bug fix for issue #123

### Security
- Updated dependencies
```

### 2. Version Tagging

#### Create and Push Tag
```bash
# Ensure you're on main branch and up to date
git checkout main
git pull origin main

# Create annotated tag
git tag -a v1.1.0 -m "Release version 1.1.0

### Added
- New feature X
- Support for Y

### Changed
- Improved Z performance

### Fixed
- Bug fix for issue #123"

# Push tag to trigger release workflow
git push origin v1.1.0
```

#### Alternative: Using GitHub CLI
```bash
# Create release with GitHub CLI
gh release create v1.1.0 \
  --title "ztictl v1.1.0" \
  --notes "### Added
- New feature X
- Support for Y

### Changed
- Improved Z performance

### Fixed
- Bug fix for issue #123" \
  --draft  # Remove --draft when ready to publish
```

### 3. Automated Build Process

The GitHub Actions workflow (`.github/workflows/build.yml`) automatically:

1. **Triggers on tag push** matching `v*` pattern
2. **Builds for all platforms**:
   - Linux (AMD64, ARM64)
   - macOS (Intel, Apple Silicon)
   - Windows (AMD64, ARM64)
3. **Runs tests** and quality checks
4. **Creates release artifacts**
5. **Publishes GitHub release** with binaries

#### Monitor Build Progress
```bash
# Using GitHub CLI
gh run list --workflow=build.yml

# Check specific run
gh run view <run-id>

# View in browser
gh workflow view build.yml --web
```

### 4. Manual Release Process (if needed)

#### Build All Platforms
```bash
cd ztictl

# Clean previous builds
make clean

# Build all platforms
make build

# Verify builds
ls -la builds/
```

#### Create Release Archives
```bash
cd builds

# Create archives for distribution
for file in ztictl-*; do
  if [[ "$file" == *.exe ]]; then
    # Windows - create ZIP
    zip "${file%.exe}.zip" "$file"
  else
    # Unix systems - create tar.gz
    tar -czf "${file}.tar.gz" "$file"
  fi
done
```

#### Upload to GitHub Release
```bash
# Create release
gh release create v1.1.0 \
  --title "ztictl v1.1.0" \
  --notes-file CHANGELOG.md

# Upload binaries
gh release upload v1.1.0 builds/ztictl-*
```

### 5. Post-Release Tasks

#### Update Installation Documentation
Verify that installation links work:
```bash
# Test download URLs
curl -I https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-amd64
curl -I https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-darwin-amd64
curl -I https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-windows-amd64.exe
```

#### Announce Release
- Update README.md with new features
- Post release notes
- Notify users of breaking changes (if major version)

#### Create Next Development Version
```bash
# Start work on next version
git checkout -b feature/next-version
# Update version references for development
```

## Release Workflow Details

### GitHub Actions Workflow

The automated workflow (`.github/workflows/build.yml`) performs:

```yaml
# Triggered by tags matching v*
on:
  push:
    tags: ['v*']

# Build matrix for all platforms
strategy:
  matrix:
    include:
      - goos: linux, goarch: amd64
      - goos: linux, goarch: arm64
      - goos: darwin, goarch: amd64
      - goos: darwin, goarch: arm64
      - goos: windows, goarch: amd64
      - goos: windows, goarch: arm64

# Automatic version extraction from tag
VERSION=${GITHUB_REF#refs/tags/v}

# Build with optimization flags
go build -ldflags "-X main.version=${VERSION} -s -w"
```

### Version Embedding

The build process embeds the version using Go's `-ldflags`:
```bash
# Version is set at build time
go build -ldflags "-X main.version=1.1.0" ./cmd/ztictl
```

In the code (`main.go`):
```go
var version = "dev" // Default, overridden at build time

func init() {
    rootCmd.Version = version
}
```

## Hotfix Release Process

### For Critical Bug Fixes

1. **Create hotfix branch from tag**:
   ```bash
   git checkout v1.0.0
   git checkout -b hotfix/v1.0.1
   ```

2. **Apply minimal fix**:
   ```bash
   # Make necessary changes
   git add .
   git commit -m "Fix critical issue X"
   ```

3. **Tag and release**:
   ```bash
   git tag v1.0.1
   git push origin v1.0.1
   ```

4. **Merge back to main**:
   ```bash
   git checkout main
   git merge hotfix/v1.0.1
   ```

## Release Checklist

### Pre-Release
- [ ] All tests pass locally
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version references updated
- [ ] No breaking changes in minor/patch releases

### Release
- [ ] Tag created with proper version
- [ ] Tag pushed to GitHub
- [ ] GitHub Actions build successful
- [ ] All platform binaries generated
- [ ] Release notes published

### Post-Release
- [ ] Installation links verified
- [ ] Download tests successful
- [ ] Documentation reflects new version
- [ ] Users notified of release

## Troubleshooting Releases

### Build Failures
```bash
# Check workflow logs
gh run view --log

# Local debugging
make build
./builds/ztictl-linux-amd64 --version
```

### Missing Binaries
- Check GitHub Actions workflow completion
- Verify all platforms built successfully
- Check artifact upload step

### Version Issues
- Ensure tag format matches `v*` pattern
- Verify version embedding in binary
- Check go.mod version compatibility

## Security Considerations

### Binary Signing (Future Enhancement)
Consider implementing:
- Code signing for Windows binaries
- Notarization for macOS binaries
- Checksums and signatures for all releases

### Release Verification
Users should verify downloads:
```bash
# Generate checksums (manual process)
sha256sum builds/* > checksums.txt

# Users can verify
sha256sum -c checksums.txt
```

## Continuous Improvement

### Metrics to Track
- Download counts per platform
- Issue reports by version
- Performance regressions
- User feedback

### Process Improvements
- Automated testing on real AWS infrastructure
- Performance benchmarking
- Dependency vulnerability scanning
- Automated changelog generation
