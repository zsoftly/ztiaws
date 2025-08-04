# ztictl Release Process

## Release Overview

Simple release process for CLI tools where users control their updates.

## Release Process

### 1. Create Release Branch
```bash
# Create release branch from latest main
git checkout main
git pull origin main
git checkout -b release/v<version>
```

### 2. Auto-Generate Documentation
Release branch automatically:
- Compares latest release tag with current repo
- Generates/updates `CHANGELOG.md` from git history
- Generates/updates `RELEASE_NOTES.txt`
- Commits changes directly to release branch

### 3. Create Release
```bash
# Create and push tag from release branch
git tag v<version>
git push origin v<version>
```

### 4. Merge Back
```bash
# Merge release branch to main
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

- [ ] Create release branch from main
- [ ] Auto-generate changelog and release notes
- [ ] Review and edit if needed
- [ ] Create and push release tag from release branch
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
