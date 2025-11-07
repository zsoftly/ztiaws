# CI/CD Authentication Guide

## Overview

This guide explains how to use `ztictl` in CI/CD pipelines where interactive authentication is not possible.

**Key Point:** AWS SSO authentication (`ztictl auth login`) **cannot be used in CI/CD** because it requires browser interaction. Instead, you must use IAM-based authentication methods.

## Table of Contents

- [Why SSO Doesn't Work in CI/CD](#why-sso-doesnt-work-in-cicd)
- [Authentication Methods](#authentication-methods)
- [Recommended Approach by Platform](#recommended-approach-by-platform)
- [Configuration](#configuration)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)

---

## Why SSO Doesn't Work in CI/CD

AWS SSO uses the **OAuth2 Device Authorization Grant** flow (RFC 8628), which requires:

1. User opens a browser
2. User navigates to AWS SSO portal
3. User enters a code displayed in the terminal
4. User authenticates with username/password or MFA
5. Browser redirects and provides credentials

**CI/CD pipelines cannot:**
- Open web browsers
- Display interactive prompts
- Wait for user input
- Complete MFA challenges

**Therefore:** `ztictl auth login` will always fail in automated environments.

---

## Authentication Methods

### Method 1: OIDC Federation (Recommended)

**Best for:** GitHub Actions, GitLab CI, modern CI/CD platforms

**Advantages:**
- No long-lived credentials stored
- Automatic credential rotation
- Fine-grained permission control
- Audit trail via CloudTrail

**How it works:**
1. CI/CD platform provides OIDC token
2. AWS STS exchanges token for temporary credentials
3. `ztictl` uses temporary credentials via AWS SDK

**Platforms with OIDC support:**
- GitHub Actions
- GitLab CI/CD
- CircleCI
- Bitbucket Pipelines
- Azure DevOps

### Method 2: EC2 Instance Profile

**Best for:** Self-hosted runners on EC2 instances

**Advantages:**
- No credential management needed
- Automatic credential rotation
- Seamless integration with AWS services

**How it works:**
1. Attach IAM role to EC2 instance
2. Instance metadata service provides credentials
3. `ztictl` automatically uses instance credentials

### Method 3: IAM Access Keys

**Best for:** Quick testing, legacy systems

**Disadvantages:**
- Long-lived credentials (security risk)
- Manual rotation required
- Stored in secrets manager (additional exposure)

**Use only when:**
- OIDC not available
- Not running on AWS infrastructure
- Temporary solution during migration

**How it works:**
1. Create IAM user with access keys
2. Store keys in CI/CD secrets
3. Export as environment variables
4. `ztictl` uses credentials via AWS SDK

### Method 4: ECS Task Role

**Best for:** Containerized CI/CD on ECS/Fargate

**Advantages:**
- No credential management
- Automatic credential rotation
- Container-level isolation

**How it works:**
1. Define task role in ECS task definition
2. ECS provides credentials to container
3. `ztictl` uses credentials automatically

---

## Recommended Approach by Platform

### GitHub Actions

**Recommended:** OIDC Federation

**Required IAM Role Trust Policy:**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::123456789012:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:your-org/your-repo:*"
        }
      }
    }
  ]
}
```

**Workflow Example:**
```yaml
name: Deploy with ztictl
on: [push]

permissions:
  id-token: write  # Required for OIDC
  contents: read

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@v2
        with:
          role-to-assume: arn:aws:iam::123456789012:role/GitHubActionsRole
          aws-region: ca-central-1

      - name: Install ztictl
        run: |
          curl -L https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-amd64 -o ztictl
          chmod +x ztictl
          sudo mv ztictl /usr/local/bin/

      - name: Initialize ztictl config
        run: ztictl config init --non-interactive
        env:
          ZTICTL_DEFAULT_REGION: ca-central-1

      - name: List instances
        run: ztictl ssm list --table
```

See: [docs/examples/github-actions-oidc.yml](examples/github-actions-oidc.yml)

### GitLab CI

**Recommended:** OIDC Federation

**Required IAM Role Trust Policy:**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::123456789012:oidc-provider/gitlab.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "gitlab.com:aud": "https://gitlab.com"
        },
        "StringLike": {
          "gitlab.com:sub": "project_path:your-group/your-project:ref_type:branch:ref:main"
        }
      }
    }
  ]
}
```

**Pipeline Example:**
```yaml
deploy:
  image: alpine:latest
  id_tokens:
    GITLAB_OIDC_TOKEN:
      aud: https://gitlab.com
  before_script:
    - apk add --no-cache aws-cli curl
    - |
      export AWS_WEB_IDENTITY_TOKEN_FILE=/tmp/web-identity-token
      echo $GITLAB_OIDC_TOKEN > $AWS_WEB_IDENTITY_TOKEN_FILE
      export AWS_ROLE_ARN=arn:aws:iam::123456789012:role/GitLabCIRole
      export AWS_REGION=ca-central-1
    - |
      curl -L https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-amd64 -o ztictl
      chmod +x ztictl
      mv ztictl /usr/local/bin/
  script:
    - ztictl config init --non-interactive
    - ztictl ssm list --table
  variables:
    ZTICTL_DEFAULT_REGION: ca-central-1
```

See: [docs/examples/gitlab-ci-oidc.yml](examples/gitlab-ci-oidc.yml)

### Jenkins

**Recommended:** EC2 Instance Profile (if self-hosted on EC2) or IAM Access Keys

**Option A: EC2 Instance Profile**
```groovy
pipeline {
    agent { label 'ec2' }  // EC2-based Jenkins agent

    stages {
        stage('Deploy') {
            steps {
                sh '''
                    # Credentials automatically from instance metadata
                    ztictl config init --non-interactive
                    ztictl ssm list --table
                '''
            }
        }
    }

    environment {
        ZTICTL_DEFAULT_REGION = 'ca-central-1'
        ZTICTL_NON_INTERACTIVE = 'true'
    }
}
```

**Option B: IAM Access Keys (stored in Jenkins credentials)**
```groovy
pipeline {
    agent any

    stages {
        stage('Deploy') {
            steps {
                withCredentials([
                    string(credentialsId: 'aws-access-key-id', variable: 'AWS_ACCESS_KEY_ID'),
                    string(credentialsId: 'aws-secret-access-key', variable: 'AWS_SECRET_ACCESS_KEY')
                ]) {
                    sh '''
                        ztictl config init --non-interactive
                        ztictl ssm list --table
                    '''
                }
            }
        }
    }

    environment {
        AWS_DEFAULT_REGION = 'ca-central-1'
        ZTICTL_DEFAULT_REGION = 'ca-central-1'
        ZTICTL_NON_INTERACTIVE = 'true'
    }
}
```

See: [docs/examples/jenkins-iam-keys.groovy](examples/jenkins-iam-keys.groovy)

### CircleCI

**Recommended:** OIDC Federation or IAM Access Keys

**OIDC Example:**
```yaml
version: 2.1

orbs:
  aws-cli: circleci/aws-cli@3.1

jobs:
  deploy:
    docker:
      - image: cimg/base:stable
    steps:
      - checkout
      - aws-cli/setup:
          role-arn: arn:aws:iam::123456789012:role/CircleCIRole
          aws-region: ca-central-1
      - run:
          name: Install ztictl
          command: |
            curl -L https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-amd64 -o ztictl
            chmod +x ztictl
            sudo mv ztictl /usr/local/bin/
      - run:
          name: Configure ztictl
          command: ztictl config init --non-interactive
          environment:
            ZTICTL_DEFAULT_REGION: ca-central-1
      - run:
          name: List instances
          command: ztictl ssm list --table

workflows:
  deploy:
    jobs:
      - deploy
```

### AWS CodeBuild

**Recommended:** Service Role (automatic)

**buildspec.yml:**
```yaml
version: 0.2

phases:
  install:
    commands:
      - curl -L https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-amd64 -o ztictl
      - chmod +x ztictl
      - mv ztictl /usr/local/bin/

  pre_build:
    commands:
      - export ZTICTL_DEFAULT_REGION=$AWS_DEFAULT_REGION
      - ztictl config init --non-interactive

  build:
    commands:
      - ztictl ssm list --table
      - ztictl ssm exec $AWS_DEFAULT_REGION $INSTANCE_ID "deploy.sh"

env:
  variables:
    ZTICTL_NON_INTERACTIVE: "true"
```

**Note:** CodeBuild automatically provides credentials via the service role attached to the build project.

---

## Configuration

### Required Permissions

The IAM role or user must have permissions for the operations you plan to perform:

**SSM Operations:**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeInstances",
        "ec2:DescribeTags",
        "ssm:DescribeInstanceInformation",
        "ssm:StartSession",
        "ssm:SendCommand",
        "ssm:GetCommandInvocation",
        "ssm:TerminateSession"
      ],
      "Resource": "*"
    }
  ]
}
```

For complete permission requirements, see: [IAM_PERMISSIONS.md](IAM_PERMISSIONS.md)

### Environment Variables

**Required for ztictl:**
```bash
# AWS credentials (provided automatically by most CI/CD platforms)
AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE        # (or via OIDC/instance profile)
AWS_SECRET_ACCESS_KEY=wJalrXUt...             # (or via OIDC/instance profile)
AWS_DEFAULT_REGION=ca-central-1               # AWS SDK default region

# ztictl configuration
ZTICTL_DEFAULT_REGION=ca-central-1            # Default region for operations
ZTICTL_NON_INTERACTIVE=true                   # Disable all interactive prompts
```

**Optional:**
```bash
ZTICTL_LOG_ENABLED=false                      # Disable file logging in CI
ZTICTL_LOG_LEVEL=info                         # Log verbosity
ZTICTL_REGIONS=us-east-1,ca-central-1         # Regions to operate in
ZTICTL_INSTANCE_ID=i-1234567890abcdef0        # Default instance for SSM commands
```

### Non-Interactive Mode

`ztictl` automatically detects CI/CD environments and enables non-interactive mode when:

1. `CI` environment variable is set (most CI/CD platforms set this)
2. `ZTICTL_NON_INTERACTIVE=true` is set explicitly
3. `--non-interactive` flag is used

**In non-interactive mode:**
- Splash screen is suppressed
- All interactive prompts are skipped
- Commands requiring input will fail with clear error messages
- Operations use environment variables or fail fast

**Example:**
```bash
# Explicit non-interactive mode
ztictl --non-interactive ssm list --table

# Via environment variable
export ZTICTL_NON_INTERACTIVE=true
ztictl ssm list --table

# Auto-detected (CI=true set by platform)
ztictl ssm list --table  # Works automatically in CI
```

---

## Examples

### Complete GitHub Actions Workflow

```yaml
name: Deploy Application
on:
  push:
    branches: [main]

permissions:
  id-token: write
  contents: read

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Configure AWS Credentials (OIDC)
        uses: aws-actions/configure-aws-credentials@v2
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
          aws-region: ca-central-1

      - name: Install ztictl
        run: |
          curl -L https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-amd64 -o ztictl
          chmod +x ztictl
          sudo mv ztictl /usr/local/bin/
          ztictl --version

      - name: Initialize ztictl
        run: ztictl config init --non-interactive
        env:
          ZTICTL_DEFAULT_REGION: ca-central-1

      - name: List all instances
        run: ztictl ssm list --region ca-central-1 --table

      - name: Deploy to production instances via tags
        run: |
          ztictl ssm exec-tagged ca-central-1 \
            --tags Name=web-server-prod \
            "cd /app && git pull && systemctl restart app"

      - name: Verify deployment
        run: |
          ztictl ssm exec-tagged ca-central-1 \
            --tags Name=web-server-prod \
            "systemctl status app"
```

### Multi-Region Deployment

```yaml
- name: Deploy to all regions
  run: |
    for region in us-east-1 ca-central-1 eu-west-1; do
      echo "Deploying to $region..."
      ztictl ssm exec-tagged $region \
        --tags Environment=production,Service=web \
        "/app/deploy.sh"
    done
```

### Instance Power Management

```yaml
- name: Start all test instances
  run: |
    ztictl ssm start-tagged \
      --region ca-central-1 \
      --tags Environment=test \
      --tags AutoStart=true

- name: Wait for instances to be ready
  run: sleep 60

- name: Run integration tests
  run: |
    ztictl ssm exec-tagged ca-central-1 \
      --tags Environment=test \
      "cd /app && npm test"

- name: Stop test instances
  if: always()
  run: |
    ztictl ssm stop-tagged \
      --region ca-central-1 \
      --tags Environment=test
```

### File Transfer

```yaml
- name: Upload configuration files
  run: |
    ztictl ssm transfer upload \
      i-1234567890abcdef0 \
      ./config/production.json \
      /etc/app/config.json \
      --region ca-central-1

- name: Download logs
  run: |
    ztictl ssm transfer download \
      i-1234567890abcdef0 \
      /var/log/app/error.log \
      ./artifacts/error.log \
      --region ca-central-1

- name: Upload artifacts
  uses: actions/upload-artifact@v3
  with:
    name: logs
    path: artifacts/
```

---

## Troubleshooting

### "Unable to locate credentials"

**Problem:** AWS SDK cannot find credentials

**Solutions:**
1. Verify OIDC configuration is correct
2. Check that `id-token: write` permission is set (GitHub Actions)
3. Verify IAM role trust policy allows your CI/CD platform
4. Check environment variables are set correctly

**Debug:**
```bash
# Check which credentials are being used
aws sts get-caller-identity

# Verify credentials work
aws ec2 describe-instances --region ca-central-1 --max-items 1
```

### "Access Denied" errors

**Problem:** IAM role/user lacks required permissions

**Solutions:**
1. Review [IAM_PERMISSIONS.md](IAM_PERMISSIONS.md) for required permissions
2. Attach appropriate IAM policies to role/user
3. Verify resource-based policies allow access

**Debug:**
```bash
# Check current identity
aws sts get-caller-identity

# Test specific permission
aws ssm describe-instance-information --region ca-central-1 --max-items 1
```

### "non-interactive mode requires instance ID"

**Problem:** SSM command needs explicit instance identifier in CI

**Solutions:**
1. Provide instance ID as argument:
   ```bash
   ztictl ssm connect i-1234567890abcdef0 --region ca-central-1
   ```

2. Use instance name (auto-resolved):
   ```bash
   ztictl ssm connect web-server-prod --region ca-central-1
   ```

3. Use tag-based commands instead:
   ```bash
   ztictl ssm exec-tagged ca-central-1 --tags Name=web-server-prod "uptime"
   ```

### "AWS SSO authentication requires browser"

**Problem:** Trying to use `ztictl auth login` in CI/CD

**Solution:** Don't use SSO in CI/CD. Use IAM-based authentication instead (see above).

This error appears when:
- Running `ztictl auth login` in a CI environment
- CI environment is detected (CI=true) but no AWS credentials exist

**Fix:** Configure IAM credentials via OIDC, instance profile, or access keys.

### Session Manager plugin not found

**Problem:** `session-manager-plugin` not installed

**Solutions:**

**Ubuntu/Debian:**
```bash
curl "https://s3.amazonaws.com/session-manager-downloads/plugin/latest/ubuntu_64bit/session-manager-plugin.deb" -o "session-manager-plugin.deb"
sudo dpkg -i session-manager-plugin.deb
```

**Amazon Linux/RHEL:**
```bash
curl "https://s3.amazonaws.com/session-manager-downloads/plugin/latest/linux_64bit/session-manager-plugin.rpm" -o "session-manager-plugin.rpm"
sudo yum install -y session-manager-plugin.rpm
```

**Docker (add to Dockerfile):**
```dockerfile
RUN curl "https://s3.amazonaws.com/session-manager-downloads/plugin/latest/ubuntu_64bit/session-manager-plugin.deb" -o "/tmp/session-manager-plugin.deb" && \
    dpkg -i /tmp/session-manager-plugin.deb && \
    rm /tmp/session-manager-plugin.deb
```

### Splash screen blocks execution

**Problem:** First-run splash screen waits for input

**Solution:** Should auto-detect CI and skip splash. If not working:

```bash
# Force non-interactive mode
export ZTICTL_NON_INTERACTIVE=true
ztictl ssm list
```

If issue persists, verify CI environment variable is set:
```bash
export CI=true
```

---

## Best Practices

### 1. Use OIDC When Possible
- No credential management
- Automatic rotation
- Better security posture
- Audit trail

### 2. Minimize Permission Scope
- Use least-privilege IAM policies
- Scope permissions to specific resources when possible
- Separate roles for different environments

### 3. Use Tag-Based Operations
- More resilient than instance IDs
- Works with auto-scaling groups
- Easier to manage

```bash
# Use tag-based commands for multiple instances
ztictl ssm exec-tagged ca-central-1 --tags Environment=prod "deploy.sh"

# Instead of operating on a single instance
ztictl ssm exec ca-central-1 i-1234567890abcdef0 "deploy.sh"
```

### 4. Set Explicit Regions
- Don't rely on defaults
- Specify region in commands or environment variables

```yaml
env:
  AWS_DEFAULT_REGION: ca-central-1
  ZTICTL_DEFAULT_REGION: ca-central-1
```

### 5. Enable Debug Logging for Troubleshooting
```yaml
env:
  ZTICTL_LOG_ENABLED: "true"
  ZTICTL_LOG_LEVEL: "debug"
  ZTICTL_LOG_DIR: "/tmp/ztictl-logs"

# Upload logs as artifacts
- uses: actions/upload-artifact@v3
  if: failure()
  with:
    name: ztictl-logs
    path: /tmp/ztictl-logs/
```

### 6. Use `--table` Output for Parsing
```bash
# Machine-readable table output
ztictl ssm list --table --region ca-central-1 | grep running
```

### 7. Test with `--dry-run` When Available
```bash
ztictl ssm cleanup --region ca-central-1 --dry-run
```

---

## Security Considerations

### 1. Rotate Access Keys Regularly
If using IAM access keys:
- Rotate every 90 days minimum
- Use AWS Secrets Manager or CI/CD secrets
- Never commit keys to source control

### 2. Use Temporary Credentials
- OIDC provides short-lived credentials (1 hour default)
- Instance profiles rotate automatically
- Avoid long-lived access keys

### 3. Restrict OIDC Trust Policies
```json
{
  "Condition": {
    "StringEquals": {
      "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
    },
    "StringLike": {
      "token.actions.githubusercontent.com:sub": "repo:your-org/your-repo:ref:refs/heads/main"
    }
  }
}
```

This restricts role assumption to:
- Specific repository
- Specific branch (main)
- Prevents unauthorized use

### 4. Audit CloudTrail Logs
- Monitor API calls from CI/CD
- Set up alerts for unusual activity
- Review access patterns regularly

### 5. Separate Roles by Environment
```
GitHubActions-Dev-Role   → Limited permissions, dev resources
GitHubActions-Prod-Role  → Full permissions, prod resources
```

Restrict production role to protected branches only.

---

## Additional Resources

- [IAM_PERMISSIONS.md](IAM_PERMISSIONS.md) - Required IAM permissions
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - General troubleshooting
- [GitHub Actions OIDC Guide](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/configuring-openid-connect-in-amazon-web-services)
- [GitLab CI OIDC Guide](https://docs.gitlab.com/ee/ci/cloud_services/aws/)
- [AWS Session Manager Plugin Installation](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html)

---

## Quick Reference

### Authentication Method Decision Tree

```
Are you running on AWS infrastructure?
├─ Yes → Use EC2 Instance Profile or ECS Task Role
└─ No
   └─ Does your CI/CD platform support OIDC?
      ├─ Yes → Use OIDC Federation (recommended)
      └─ No → Use IAM Access Keys (rotate regularly)
```

### Common Commands in CI/CD

```bash
# Initialize configuration
ztictl config init --non-interactive

# List instances (table format for parsing)
ztictl ssm list --region ca-central-1 --table

# Execute on specific instance (with region shortcode)
ztictl ssm exec cac1 i-1234567890abcdef0 "uptime"

# Execute on multiple instances by tag
ztictl ssm exec-tagged cac1 --tags Environment=prod "deploy.sh"

# Power management by tag
ztictl ssm start-tagged --region ca-central-1 --tags Environment=test
ztictl ssm stop-tagged --region ca-central-1 --tags Environment=test
ztictl ssm reboot-tagged --region ca-central-1 --tags Environment=test

# Power management for single instance
ztictl ssm start i-1234567890abcdef0 --region ca-central-1
ztictl ssm stop i-1234567890abcdef0 --region ca-central-1
ztictl ssm reboot i-1234567890abcdef0 --region ca-central-1

# File transfer
ztictl ssm transfer upload i-xxx ./local.txt /remote/path/file.txt --region cac1
ztictl ssm transfer download i-xxx /remote/log.txt ./local.txt --region cac1

# Cleanup old resources
ztictl ssm cleanup --region ca-central-1 --dry-run
```

### Environment Variables Checklist

```bash
# AWS SDK (auto-provided by OIDC/instance profile)
✓ AWS_ACCESS_KEY_ID
✓ AWS_SECRET_ACCESS_KEY
✓ AWS_DEFAULT_REGION

# ztictl required
✓ ZTICTL_DEFAULT_REGION

# ztictl optional
○ ZTICTL_NON_INTERACTIVE=true
○ ZTICTL_INSTANCE_ID=i-xxx
○ ZTICTL_LOG_ENABLED=false
```

---

**Need help?** Open an issue at https://github.com/zsoftly/ztiaws/issues
