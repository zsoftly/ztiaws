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
4. Run tests
```bash
./tests/test_ssm.sh
```
5. Submit a Pull Request

## Pull Request Process

1. Update REGIONS.md with new region details
2. Update tests to cover new region
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

For maintainers who want to create a new release:

```bash
# Make sure you're on the main branch
git checkout main

# Pull the latest changes (including merged PRs)
git pull origin main

# Ensure all changes are committed and the working directory is clean
git status

# Create an annotated tag
git tag -a v1.x.x -m "Version 1.x.x: Brief description of changes"

# Push the tag to GitHub
git push origin v1.x.x
```

After pushing the tag, go to the GitHub repository and:
1. Click on "Releases"
2. Click "Draft a new release"
3. Select the tag you just pushed
4. Add release notes
5. Publish the release

This process ensures that releases are always created from the stable main branch after code has been properly reviewed and merged.