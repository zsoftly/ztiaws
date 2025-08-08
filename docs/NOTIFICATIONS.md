# ZTiAWS Notification System

## Overview

ZTiAWS uses automated notifications to keep the team informed about important repository events. The notification system is integrated into the CI/CD pipeline and sends alerts to Google Chat for key development milestones.

## Notification Types

### 1. Pull Request Notifications

**Triggered when:** A pull request is opened targeting the `main` branch
**Sent after:** Security scans complete successfully
**Channel:** Google Chat room (same as zsoftly-services)

**Message format:**
```
🔔 New PR opened in ztiaws!
📝 *[PR Title]*
👤 Author: [GitHub Username]
🔗 Link: [Direct PR URL]
```

**Purpose:** 
- Keep team aware of new contributions
- Enable quick review and collaboration
- Maintain visibility into main branch changes

### 2. Release Notifications

**Triggered when:** A new version tag (e.g., `v1.0.0`) is pushed
**Sent after:** GitHub release is created successfully
**Channel:** Google Chat room (same as zsoftly-services)

**Message format:**
```
🚀 New ztiaws release available!
📦 *Version: [Version Number]*
🔗 Release: [GitHub Release URL]
🎉 Ready for deployment!
```

**Purpose:**
- Announce new releases to stakeholders
- Provide direct access to release notes and binaries
- Coordinate deployment activities

## Technical Implementation

### Integration with CI/CD Pipeline

The notification system is implemented as conditional jobs in the main CI/CD workflow (`.github/workflows/build.yml`):

- **PR notifications** run after the `security` job completes
- **Release notifications** run after the `release` job completes
- Both use the same Google Chat webhook as the zsoftly-services repository

### Configuration

**Webhook Secret:** `GOOGLE_CHAT_WEBHOOK`
- Stored in GitHub repository secrets
- Shared across ZSoftly repositories for consistency
- Points to team's Google Chat room

**Job Conditions:**
```yaml
# PR Notification
if: github.event_name == 'pull_request' && github.event.action == 'opened' && github.base_ref == 'main'

# Release Notification  
if: startsWith(github.ref, 'refs/tags/v')
```

## Notification Flow

### Pull Request Flow
1. Developer opens PR to `main` branch
2. CI/CD triggers: `test` → `security` → `pr-notification`
3. If all previous jobs succeed, notification is sent
4. Team receives immediate alert in Google Chat

### Release Flow
1. Maintainer pushes version tag (e.g., `git push origin v1.0.0`)
2. CI/CD triggers: `build` → `release` → `release-notification`
3. After successful release creation, notification is sent
4. Team and stakeholders are notified of new version

## Benefits

### For Development Team
- **Immediate awareness** of new PRs requiring review
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