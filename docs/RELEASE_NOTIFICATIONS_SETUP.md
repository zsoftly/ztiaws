# Release Notifications Setup Guide

This guide explains how to configure automated release notifications for the ZTiAWS repository to send updates to your Google Chat room.

## 🎯 Overview

When a new release is published in the ztiaws repository, an automated notification will be sent to your team's Google Chat room containing:

- ✅ **Version number** and release type (stable/pre-release)
- ✅ **Release notes** (truncated if too long)
- ✅ **Direct links** to view the release, download files, and access documentation
- ✅ **Professional formatting** with emojis and clear structure

## 🔧 Setup Instructions

### Step 1: Create Google Chat Webhook

1. **Open Google Chat** and navigate to the space where you want to receive notifications
2. **Click the space name** at the top to open space settings
3. **Select "Apps & integrations"**
4. **Click "Manage webhooks"**
5. **Create a new webhook:**
   - Name: `ZTiAWS Release Notifications`
   - Avatar URL: (optional) `https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png`
6. **Copy the webhook URL** - it will look like:
   ```
   https://chat.googleapis.com/v1/spaces/SPACE_ID/messages?key=KEY&token=TOKEN
   ```

### Step 2: Configure Repository Secret

1. **Go to your ztiaws repository** on GitHub
2. **Navigate to Settings → Secrets and variables → Actions**
3. **Click "New repository secret"**
4. **Create the secret:**
   - **Name:** `GOOGLE_CHAT_WEBHOOK`
   - **Value:** The webhook URL you copied from Step 1
5. **Click "Add secret"**

### Step 3: Verify Workflow File

The release notification workflow is located at `.github/workflows/release-notifications.yml`. This file:

- ✅ Triggers automatically when a release is published
- ✅ Uses the secure `GOOGLE_CHAT_WEBHOOK` secret
- ✅ Includes comprehensive error handling and fallback notifications
- ✅ Follows GitHub Actions security best practices

## 🧪 Testing the Setup

### Option 1: Create a Test Release (Recommended)

1. **Create a test tag:**
   ```bash
   git tag -a v0.0.1-test -m "Test release for notifications"
   git push origin v0.0.1-test
   ```

2. **Create a release from the tag:**
   - Go to GitHub → Releases → "Draft a new release"
   - Select the `v0.0.1-test` tag
   - Add some test release notes
   - **Important:** Check "Set as a pre-release" to avoid confusion
   - Click "Publish release"

3. **Check your Google Chat room** for the notification

4. **Clean up:** Delete the test release and tag when done

### Option 2: Manual Workflow Test

1. **Go to Actions tab** in your repository
2. **Find the "Release Notifications" workflow**
3. **Note:** This workflow only triggers on actual releases, not manual dispatch

## 🔒 Security Best Practices

This setup follows security best practices:

### ✅ **Secure Secret Storage**
- Webhook URL is stored as an encrypted repository secret
- Never exposed in logs or workflow files
- Only accessible by the workflow during execution

### ✅ **Minimal Permissions**
- Workflow only requests `contents: read` permission
- No write access to repository
- Restricted to notification functionality only

### ✅ **Input Validation**
- Release notes are truncated to prevent oversized messages
- Special characters are properly escaped
- Draft releases are ignored (only published releases trigger notifications)

### ✅ **Error Handling**
- Fallback notification if primary notification fails
- Comprehensive error logging for debugging
- Graceful degradation without breaking the release process

## 📋 Notification Format

Your team will receive notifications like this:

```
🚀 New Release: ZTiAWS v2.1.0

📋 Release Type: Stable Release

📝 Release Notes:
- Added flag-based parameter support to authaws
- Fixed shellcheck warnings
- Enhanced cross-platform compatibility

🔗 Links:
• 📖 View Release
• ⬇️ Download (tar.gz)
• 📦 Download (zip)
• 📚 Documentation
• 🔧 Installation Guide

📊 Repository: zsoftly/ztiaws
```

## 🔧 Troubleshooting

### Notifications Not Appearing

1. **Check webhook URL:**
   - Ensure `GOOGLE_CHAT_WEBHOOK` secret is set correctly
   - Verify the webhook URL is active in Google Chat

2. **Check workflow execution:**
   - Go to Actions tab and verify the workflow ran
   - Check logs for any error messages

3. **Verify trigger conditions:**
   - Workflow only triggers on published releases (not drafts)
   - Pre-releases will still trigger notifications

### Permission Issues

If you see permission errors:

1. **Check repository access:**
   - Ensure you have admin access to configure secrets
   - Verify Actions are enabled for the repository

2. **Check Google Chat permissions:**
   - Ensure you have permission to manage webhooks in the space
   - Verify the webhook is still active

### Message Formatting Issues

1. **Long release notes:** Notes are automatically truncated at 1000 characters
2. **Special characters:** Automatically escaped for JSON compatibility
3. **Links not working:** Verify the release was published successfully

## 🎯 Advanced Configuration

### Custom Message Format

To customize the notification format, edit the `text:` section in `.github/workflows/release-notifications.yml`.

### Additional Recipients

To send notifications to multiple Google Chat rooms:

1. Create additional webhooks for each room
2. Add them as separate secrets (`GOOGLE_CHAT_WEBHOOK_TEAM2`, etc.)
3. Add additional notification steps in the workflow

### Integration with Other Services

The workflow can be extended to send notifications to other services:
- Slack (using slack-webhook actions)
- Microsoft Teams (using teams-webhook actions)
- Email (using email action providers)

## 📚 References

- [Google Chat Webhooks Documentation](https://developers.google.com/chat/how-tos/webhooks)
- [GitHub Actions Security](https://docs.github.com/en/actions/security-guides)
- [GitHub Release Events](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#release)

---

**Need Help?** 
- Check the [workflow logs](../../actions) for detailed error information
- Review this setup guide for missed steps
- Consult the team lead for webhook access if needed