# Build Artifacts - What to Commit vs Ignore

## ❌ **DON'T Commit Build Artifacts**

The following build artifacts should **NEVER** be committed to source control:

### Local Builds
- `ztictl/ztictl` - Local binary for current platform
- `ztictl/ztictl.exe` - Windows binary
- `ztictl/builds/` - Directory containing all cross-platform binaries

### Cross-Platform Builds
- `ztictl-linux-amd64`
- `ztictl-linux-arm64` 
- `ztictl-darwin-amd64`
- `ztictl-darwin-arm64`
- `ztictl-windows-amd64.exe`
- `ztictl-windows-arm64.exe`

## ✅ **DO Commit Build Configuration**

These files should be committed to help others build the project:

### Build Scripts & Configuration
- `ztictl/Makefile` - Build automation
- `ztictl/build.sh` - Cross-platform build script
- `ztictl/go.mod` & `ztictl/go.sum` - Go dependencies
- `.github/workflows/build.yml` - CI/CD build pipeline

## 🛡️ **Protection via .gitignore**

The `.gitignore` file automatically excludes build artifacts:

```gitignore
# Go build artifacts
/ztictl/ztictl
/ztictl/ztictl.exe
/ztictl/builds/
**/ztictl-*
*.exe
*.bin
```

## 🎯 **Why This Matters**

### **Problems with Committing Binaries:**
- **Repository bloat**: Binaries are large (20-30MB each)
- **Platform conflicts**: Different OS/architecture binaries
- **Version confusion**: Outdated binaries vs current code
- **Merge conflicts**: Binary files cause Git issues
- **Security risks**: Executables in source control

### **Benefits of Ignoring:**
- **Clean repository**: Only source code tracked
- **Platform independence**: Each developer builds for their system
- **Always current**: Binaries match current code state
- **CI/CD handles distribution**: Automated builds for releases

## 🔄 **Recommended Workflow**

### During Development
```bash
# Build locally for testing
make build-local
./ztictl --version

# Test your changes
./ztictl ssm list --region us-east-1

# Clean before committing
make clean
git add .
git commit -m "Your changes"
```

### For Releases
```bash
# Tag triggers automated builds
git tag v1.1.0
git push origin v1.1.0

# GitHub Actions automatically:
# 1. Builds all platforms
# 2. Creates release
# 3. Uploads binaries
```

## 🚨 **If You Accidentally Committed Binaries**

```bash
# Remove from staging
git restore --staged ztictl/ztictl ztictl/builds/

# Remove from history (if already committed)
git rm --cached ztictl/ztictl
git rm --cached -r ztictl/builds/
git commit -m "Remove build artifacts from tracking"

# Ensure .gitignore is working
git status --ignored
```

## ✨ **Summary**

- **Source code**: Always commit ✅
- **Build scripts**: Always commit ✅  
- **Binaries**: Never commit ❌
- **Let CI/CD handle distribution** 🚀
