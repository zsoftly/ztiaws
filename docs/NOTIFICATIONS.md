# ZTiAWS Notification System

## Overview

ZTiAWS uses automated notifications to keep the team informed about important repository events. The notification system is integrated into the CI/CD pipeline and sends alerts to Google Chat for key development milestones.

## Notification Types

### 1. Pull Request Notifications

**Triggered when:** A pull request is opened targeting the `main` branch
**Sent after:** Security scans complete successfully
**Channel:** Google Chat room (same as zsoftly-services)

**Message format:** Google Chat App Card with:
- **Header:** "New Pull Request" with GitHub avatar
- **Key-Value sections:** PR title, author, repository, PR number
- **Action buttons:** "🔍 Review PR" and "📁 View Files"
- **Professional styling** with icons and structured layout

**Purpose:** 
- Keep team aware of new contributions
- Enable quick review and collaboration
- Maintain visibility into main branch changes

### 2. Release Notifications

**Triggered when:** A new version tag (e.g., `v1.0.0`) is pushed
**Sent after:** GitHub release is created successfully
**Channel:** Google Chat room (same as zsoftly-services)

**Message format:** Google Chat App Card with:
- **Header:** "🚀 New Release Available" with GitHub avatar
- **Key-Value sections:** Version number, repository, deployment status
- **Action buttons:** "📋 View Release", "⬇️ Download", "📝 Changelog"
- **Professional styling** with icons and structured layout

**Purpose:**
- Announce new releases to stakeholders
- Provide direct access to release notes and binaries
- Coordinate deployment activities

## Google Chat App Cards

### Why App Cards Over Simple Text

The notification system uses **Google Chat App Cards** instead of simple text messages to provide:

- **🎨 Rich visual formatting** with headers, icons, and structured sections
- **📱 Interactive elements** like clickable buttons for direct actions
- **🏗️ Professional appearance** that matches enterprise-grade tools
- **⚡ Quick access** to related resources (files, releases, changelog)
- **👀 Better readability** with organized key-value pairs
- **🔗 Direct navigation** without needing to copy/paste URLs

### App Card Structure

**PR Notifications:**
- Header with "New Pull Request" title and GitHub avatar
- Structured sections showing PR details (title, author, repository, number)
- Action buttons for reviewing PR and viewing changed files

**Release Notifications:**
- Header with "New Release Available" title and GitHub avatar  
- Structured sections showing release details (version, status)
- Action buttons for viewing release, downloading, and checking changelog

## Technical Implementation

### Shell Script Architecture

Following the zsoftly-services pattern, notifications are implemented using dedicated shell scripts with embedded Google Chat App Card styling:

- **`scripts/send-pr-notification.sh`** - PR notification handler
- **`scripts/send-release-notification.sh`** - Release notification handler
- **Rich Google Chat App Cards** with professional formatting
- **Base64 webhook encoding** for enhanced security
- **Comprehensive logging** and error handling

### Integration with CI/CD Pipeline

The notification system uses **embedded notifications** architecture in the main CI/CD workflow (`.github/workflows/build.yml`):

- **PR notifications** are embedded within each critical job (`test-shell`, `test-go`, `security`)
- **Release notifications** run after the `release` job completes
- **Immediate feedback** - notifications sent as soon as any test fails
- **Final success notification** sent only when all tests pass
- Both use shell scripts with the same Google Chat webhook as zsoftly-services
- Scripts are checked out and executed with proper environment variables

### Shell Script Benefits

**Why Shell Scripts Over Inline Curl:**
- **🎨 Rich Google Chat App Cards** with embedded styling
- **🔧 Maintainable code** separated from workflow logic  
- **🛡️ Enhanced security** with base64 webhook encoding
- **📝 Comprehensive logging** with colored output and debug mode
- **⚡ Reusable components** for different notification types
- **🧪 Testable independently** of GitHub Actions
- **📋 Professional formatting** matching enterprise standards

### Configuration

**Webhook Secret:** `GOOGLE_CHAT_WEBHOOK`
- Stored in GitHub repository secrets
- Can be base64 encoded for enhanced security (following zsoftly-services pattern)
- Shared across ZSoftly repositories for consistency
- Points to team's Google Chat room

**Embedded Notification Conditions:**
```yaml
# Test Failure Notifications (immediate)
if: failure() && matrix.os == 'ubuntu-latest' && github.event_name == 'pull_request'

# Success Notification (final, from security job)
if: success() && github.event_name == 'pull_request'

# Release Notification  
if: startsWith(github.ref, 'refs/tags/v')
```

## Embedded Notifications Architecture

### **Why Embedded vs Centralized?**

**Previous Design Issues:**
- Centralized `notify` job created dependency chain failures
- No notifications if any prerequisite job was skipped
- Single point of failure blocked all notifications
- Only worked for PRs targeting `main` branch

**New Embedded Design Benefits:**
- **Fail-fast**: Immediate notifications when tests fail
- **Independent**: Each job notifies without dependencies  
- **Reliable**: Cannot be blocked by other job failures
- **Comprehensive**: Works for all PRs, any branch target

### **Implementation Pattern:**
```yaml
- name: Notify on test failure
  if: failure() && matrix.os == 'ubuntu-latest' && github.event_name == 'pull_request'
  env:
    GOOGLE_CHAT_WEBHOOK: ${{ secrets.GOOGLE_CHAT_WEBHOOK }}
  run: |
    ./scripts/send-pr-notification.sh \
      --status "failure" \
      --message "❌ Specific test failed - PR needs attention"
  continue-on-error: true
```

## Notification Flow

### Pull Request Flow
1. Developer opens PR (to any branch)
2. CI/CD triggers: `test-shell` + `test-go` jobs run in parallel
3. **Immediate failure notifications**: If any test fails, notification sent immediately
4. **Security analysis**: If PR targets `main`, security job runs after tests pass
5. **Final success notification**: Only sent when ALL tests pass (from security job)
6. **Fail-fast feedback**: Team gets alerts as soon as issues occur

### Release Flow
1. Maintainer pushes version tag (e.g., `git push origin v1.0.0`)
2. CI/CD triggers: `build` → `release` → `release-notification`
3. After successful release creation, notification is sent
4. Team and stakeholders are notified of new version

## Benefits

### For Development Team
- **Fail-fast notifications** - Immediate alerts when any test fails (no waiting for all jobs)
- **Single success notification** - Clean, consolidated "✅" message when all tests pass
- **No dependency chain failures** - Each job notifies independently
- **Broader coverage** - Works for all PRs, not just ones targeting `main`
- **Automated release announcements** without manual coordination
- **Consistent notification format** across all ZSoftly repositories
- **Integration with existing workflows** (same chat room as zsoftly-services)

### For Project Management
- **Visibility into development velocity** (PR frequency)
- **Release tracking** with direct links to release notes
- **Stakeholder communication** through automated announcements

## Troubleshooting

### Notifications Not Received

1. **Check workflow execution:**
```bash
# View recent workflow runs
gh run list --workflow=build.yml --limit=5

# Check specific run logs
gh run view [RUN_ID] --log
```

2. **Verify webhook configuration:**
- Ensure `GOOGLE_CHAT_WEBHOOK` secret is configured
- Verify webhook URL is active and accessible
- Check Google Chat room permissions

3. **Check job conditions:**
- PR notifications only trigger for PRs to `main` branch
- Release notifications only trigger for version tags starting with `v`
- Jobs must complete successfully in correct order

### Missing Notifications

**Common causes:**
- Previous job failed (notifications won't run)
- Incorrect branch target (feature PRs don't trigger notifications)
- Tag format doesn't match `v*` pattern
- Webhook secret misconfigured

**Debugging steps:**
1. Check GitHub Actions logs for failed jobs
2. Verify PR target branch is `main`
3. Confirm tag format: `v1.0.0` (not `1.0.0` or `release-1.0.0`)
4. Test webhook manually using curl

### Testing Notifications

**Test PR notification:**
1. Create a test branch
2. Open PR to `main` branch
3. Wait for security scans to complete
4. Check Google Chat for notification

**Test release notification:**
1. Create a test tag: `git tag v0.0.0-test`
2. Push tag: `git push origin v0.0.0-test`
3. Wait for release creation
4. Check Google Chat for notification
5. Clean up: Delete test tag and release

## Security Considerations

- **Webhook security:** Uses GitHub secrets for webhook URL
- **Least privilege:** Notification jobs only have necessary permissions
- **No sensitive data:** Messages only contain public repository information
- **Fail-safe design:** Failed notifications don't block development workflows

## Related Documentation

- [CI/CD Pipeline Architecture](CI_CD_PIPELINE.md) - Complete workflow documentation
- [Contributing Guidelines](../CONTRIBUTING.md) - Development process
- [Release Process](../RELEASE.md) - How to create releases
- [zsoftly-services repository](https://github.com/zsoftly/zsoftly-services) - Reference implementation

---

**Note:** This notification system follows the same pattern as the zsoftly-services repository to maintain consistency across ZSoftly projects.