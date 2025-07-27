# Contributing to ZTiAWS

Thank you for your interest in contributing to ZTiAWS! We especially welcome contributions for new AWS regions.

## Adding a New Region

1. Fork the repository
2. Add your region to `src/regions.sh`
3. Follow this format:
```bash
case "$1" in
    "SHORTCODE") echo "aws-region-name" ;;  # Location/City
```
Example:
```bash
    "euw2") echo "eu-west-2" ;;  # London
```

### Region Code Guidelines
- Use 4 characters: area (2) + region number (2)
- Examples:
  - `use1` - US East 1
  - `euw2` - EU West 2
  - `aps1` - Asia Pacific Singapore

### Required Information
When submitting a new region, include:
1. AWS Region name (e.g., `eu-west-2`)
2. Location/City (e.g., "London")
3. AWS documentation reference
4. Region availability confirmation

## Development Process

1. Fork the repository
2. Create a feature branch
```bash
git checkout -b feature/add-region-euw2
```
3. Make your changes
4. Run tests locally
```bash
# For Go code (ztictl)
cd ztictl && make test

# For shell scripts
./tests/test_ssm.sh
```
5. Submit a Pull Request

> **📚 CI/CD Information:** See [docs/CI_CD_PIPELINE.md](docs/CI_CD_PIPELINE.md) for details on our automated testing and build process.

## Pull Request Process

1. Update REGIONS.md with new region details (if adding regions)
2. Update tests to cover new functionality
3. Ensure all CI checks pass:
   - **Quick tests** run automatically on all PRs
   - **Security scans** run on PRs to main branch
   - **Builds** are triggered only for releases
4. Update documentation if needed
3. Ensure all tests pass
4. Update documentation if needed

## Code Style

- Follow existing bash scripting style
- Use shellcheck for linting
- Add comments for non-obvious code
- Keep functions focused and small

## Testing

Test your changes:
```bash
./tests/test_ssm.sh
```

## Commit Messages

Format:
```
type(scope): description

[optional body]
[optional footer]
```

Types:
- feat: New feature
- fix: Bug fix
- docs: Documentation
- test: Tests
- chore: Maintenance

Example:
```
feat(regions): add EU West 2 London region

Added support for AWS EU West 2 (London) region
with shortcode 'euw2'

Closes #123
```

## Questions?

- Open an issue for discussion
- Tag with 'question' label
- Provide context and examples

Thank you for contributing!

## 🚀 Releasing a New Version

ZTiAWS uses an automated CI/CD pipeline for releases. See [docs/CI_CD_PIPELINE.md](docs/CI_CD_PIPELINE.md) for detailed architecture.

### Quick Release Process:

1. **Prepare release** on main branch:
   ```bash
   git checkout main && git pull origin main
   ```

2. **Update version** in relevant files:
   - `ssm`/`authaws` scripts (VERSION variable)
   - `CHANGELOG.md` (add new entry)
   - `RELEASE_NOTES.txt` (release description)

3. **Create and push tag**:
   ```bash
   git add . && git commit -m "Bump version to vX.Y.Z"
   git tag -a vX.Y.Z -m "Version X.Y.Z: Brief description"
   git push origin main vX.Y.Z
   ```

4. **Automated pipeline** handles:
   - Cross-platform builds (6 platforms)
   - GitHub release creation
   - Binary distribution
   - Release notes from RELEASE_NOTES.txt

> **💡 Pro tip:** Use semantic versioning (major.minor.patch) and watch the GitHub Actions pipeline for build status.