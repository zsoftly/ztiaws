# ZTiAWS Notification Scripts

This directory contains shell scripts for sending rich Google Chat App Card notifications, following the zsoftly-services notification pattern.

## Scripts Overview

### 📝 `send-pr-notification.sh`
Sends Google Chat App Card notifications for Pull Request events.

**Features:**
- Rich visual formatting with GitHub avatar
- Interactive buttons for PR review and file viewing
- Structured key-value sections for PR details
- Professional styling with icons

### 🚀 `send-release-notification.sh`
Sends Google Chat App Card notifications for Release events.

**Features:**
- Release announcement with version information
- Download and changelog action buttons
- Repository information and deployment status
- Professional styling with icons

## Design Philosophy

These scripts follow the zsoftly-services pattern:

1. **Rich Google Chat App Cards** instead of simple text
2. **Base64 encoding** for webhook URLs (security)
3. **Comprehensive logging** with colored output
4. **Error handling** and dependency validation
5. **Flexible parameter handling** (command line + environment)

## Usage Examples

### PR Notification
```bash
# Using environment variable (recommended)
export GOOGLE_CHAT_WEBHOOK=$(echo -n "https://chat.googleapis.com/..." | base64)
./scripts/send-pr-notification.sh \
  --pr-title "Add new feature" \
  --pr-number "123" \
  --pr-url "https://github.com/org/repo/pull/123" \
  --author "developer" \
  --repository "org/repo"

# Using webhook URL directly
./scripts/send-pr-notification.sh \
  --webhook-url "https://chat.googleapis.com/v1/spaces/..." \
  --pr-title "Fix bug" \
  --pr-number "124" \
  --pr-url "https://github.com/org/repo/pull/124" \
  --author "dev" \
  --repository "org/repo"
```

### Release Notification
```bash
# Using environment variable (recommended)
export GOOGLE_CHAT_WEBHOOK=$(echo -n "https://chat.googleapis.com/..." | base64)
./scripts/send-release-notification.sh \
  --version "v1.2.0" \
  --release-url "https://github.com/org/repo/releases/tag/v1.2.0" \
  --repository "org/repo"

# With custom changelog URL
./scripts/send-release-notification.sh \
  --webhook-url "https://chat.googleapis.com/v1/spaces/..." \
  --version "v1.2.0" \
  --release-url "https://github.com/org/repo/releases/tag/v1.2.0" \
  --repository "org/repo" \
  --changelog-url "https://github.com/org/repo/blob/main/CHANGELOG.md"
```

## Debug Mode

Enable debug output for troubleshooting:

```bash
# Environment variable
export DEBUG=true
./scripts/send-pr-notification.sh --pr-title "..." --pr-number "..." ...

# Command line flag
./scripts/send-pr-notification.sh --debug --pr-title "..." --pr-number "..." ...
```

## Dependencies

- `curl` - for HTTP requests
- `base64` - for webhook URL decoding

## Integration with GitHub Actions

These scripts are integrated into the ztiaws CI/CD pipeline in `.github/workflows/build.yml`:

- **PR notifications** trigger on PRs opened to main branch
- **Release notifications** trigger on version tag pushes
- Both use the `GOOGLE_CHAT_WEBHOOK` repository secret

## Security Considerations

- Webhook URLs are base64 encoded for security
- Sensitive data is handled through environment variables
- No webhook URLs are logged in debug mode (truncated display)
- JSON injection protection through proper escaping

## Google Chat App Card Structure

### PR Notifications
- **Header:** "🔔 New Pull Request" with GitHub avatar
- **Sections:** PR title, author, repository, PR number
- **Buttons:** "🔍 Review PR", "📁 View Files"

### Release Notifications  
- **Header:** "🚀 New Release Available" with GitHub avatar
- **Sections:** Version, repository, deployment status
- **Buttons:** "📋 View Release", "⬇️ Download", "📝 Changelog"

## Benefits Over Simple Text

✅ **Rich visual formatting** with headers and icons  
✅ **Interactive elements** with clickable buttons  
✅ **Professional appearance** for enterprise use  
✅ **Better readability** with structured sections  
✅ **Direct navigation** without copy/paste URLs  
✅ **Consistent branding** across notifications  

## Troubleshooting

1. **Check webhook URL encoding:**
   ```bash
   echo "YOUR_WEBHOOK_URL" | base64
   ```

2. **Test webhook manually:**
   ```bash
   curl -X POST "YOUR_WEBHOOK_URL" \
     -H "Content-Type: application/json" \
     -d '{"text": "Test message"}'
   ```

3. **Enable debug mode:**
   ```bash
   DEBUG=true ./scripts/send-pr-notification.sh --help
   ```

4. **Verify script permissions:**
   ```bash
   ls -la scripts/
   chmod +x scripts/*.sh
   ```

For more information, see the main project documentation at [docs/NOTIFICATIONS.md](../docs/NOTIFICATIONS.md).