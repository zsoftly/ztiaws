# ztictl Release Process

## Release Overview

Simple release process for CLI tools where users control their updates.

## Release Process

### 1. Create Release Branch

```bash
# Create and push release branch from latest main
git checkout main
git pull origin main
git checkout -b release/v<version>
git push -u origin release/v<version>
```

### 2. Wait for Auto-Generated Documentation

- GitHub Actions automatically detects new `release/*` branch
- Runs documentation generator script
- Creates unified CHANGELOG.md and RELEASE_NOTES.txt
- Commits changes back to release branch
- **Wait for this to complete before next step!**

### 3. Pull Auto-Generated Changes

```bash
# Pull the auto-generated documentation
git pull origin release/v<version>

# Review the generated files
cat CHANGELOG.md | head -50
cat RELEASE_NOTES.txt
```

### 4. Create Release Tag

```bash
# Create and push tag - this triggers the build and release
git tag v<version>
git push origin v<version>
```

### 5. Merge Back to Main

```bash
# After release is successful, merge back to main
git checkout main
git merge release/v<version>
git push origin main
git branch -d release/v<version>
```

## Why This Process is Better

### For CLI Tools Specifically

- **Users control updates**: No forced upgrades, users choose when to update
- **No hotfixes needed**: Critical issues can wait for next planned release
- **Simpler maintenance**: Less overhead for a tool that doesn't break user workflows

### Automation Benefits

- **Consistent formatting**: Automated changelog generation
- **Reduced errors**: No manual version number management
- **Faster releases**: Less manual work, more frequent releases possible
- **Git-based**: Changelog reflects actual development history

### Process Benefits

- **Clear separation**: Preparation vs. release vs. merge back
- **Review opportunity**: Generated content can be reviewed and edited
- **Rollback friendly**: Release branch can be abandoned if needed
- **Audit trail**: All changes are tracked in git history

## Quick Release Checklist

- [ ] Create release branch from main: `git checkout -b release/v<version>`
- [ ] Push release branch: `git push -u origin release/v<version>`
- [ ] **Wait for auto-generation**: Check GitHub Actions for completion
- [ ] Pull auto-generated docs: `git pull origin release/v<version>`
- [ ] Review CHANGELOG.md and RELEASE_NOTES.txt
- [ ] Create release tag: `git tag v<version>`
- [ ] Push release tag: `git push origin v<version>` (triggers build)
- [ ] Verify automated build completes successfully
- [ ] Merge release branch back to main
- [ ] Clean up release branch

### Emergency Release (Rare)

- [ ] Create hotfix branch from latest release tag
- [ ] Apply minimal fix and test
- [ ] Follow standard release process from hotfix branch
- [ ] Merge hotfix back to main

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
