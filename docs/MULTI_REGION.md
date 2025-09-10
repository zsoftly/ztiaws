# Multi-Region Operations Guide

This guide covers multi-region command execution capabilities in `ztictl` v2.6+.

## Overview

Multi-region operations allow you to execute commands across multiple AWS regions simultaneously, significantly reducing operational overhead for distributed infrastructure management.

## Configuration

### Region Configuration File

Multi-region operations are configured in `~/.ztictl.yaml`:

```yaml
# ~/.ztictl.yaml
regions:
  # List of enabled regions (shortcodes or full names)
  enabled:
    - cac1          # ca-central-1
    - use1          # us-east-1
    - use2          # us-east-2
    - euw1          # eu-west-1
    - apse1         # ap-southeast-1
  
  # Region groups for logical organization
  groups:
    production:
      - use1
      - euw1
      - apse1
    development:
      - use2
      - cac1
    disaster-recovery:
      - usw2
      - euc1
```

### Interactive Setup

Initialize or update region configuration:

```bash
# Interactive region setup
ztictl config init --interactive

# The setup will prompt you to:
# 1. Enter regions (comma-separated)
# 2. Create region groups (optional)
# 3. Set default behaviors
```

## Command Usage

### Basic Multi-Region Execution

```bash
# Positional regions argument
ztictl ssm exec-multi cac1,use1,euw1 --tags "Environment=prod" "uptime"

# Using --regions flag
ztictl ssm exec-multi --regions cac1,use1,euw1 --tags "App=web" "df -h"
```

### Using All Configured Regions

```bash
# Execute across all enabled regions
ztictl ssm exec-multi --all-regions --tags "Component=database" "systemctl status postgresql"
```

### Using Region Groups

```bash
# Use predefined region group
ztictl ssm exec-multi --region-group production --tags "App=api" "health-check"

# Combine with instance filtering
ztictl ssm exec-multi --region-group development --instances "i-1234,i-5678" "deploy.sh"
```

## Execution Control

### Parallelism

Control how many regions are processed simultaneously:

```bash
# Process 3 regions at a time (default: 5)
ztictl ssm exec-multi --all-regions --tags "Type=web" "nginx -t" --parallel 3

# Sequential execution (one region at a time)
ztictl ssm exec-multi --all-regions --tags "Critical=true" "backup.sh" --parallel 1
```

### Output Formatting

Multi-region execution provides structured output:

```
🌍 Executing across regions: cac1, use1, euw1
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📍 Region: ca-central-1 (cac1)
├─ 📊 Instances: 3 found, 3 targeted
├─ ⏱️  Started: 2024-01-15 10:30:45 EST
│
├─ Instance: i-1234567890abcdef0 (prod-web-01)
│  ✅ Success
│  Output: 10:30:45 up 45 days, 2:15, 0 users, load average: 0.05, 0.03, 0.00
│
├─ Instance: i-0987654321fedcba0 (prod-web-02)
│  ✅ Success
│  Output: 10:30:45 up 45 days, 2:15, 0 users, load average: 0.08, 0.05, 0.01
│
└─ ✅ Region Complete (2.3s)

📍 Region: us-east-1 (use1)
├─ 📊 Instances: 5 found, 5 targeted
...
```

## Filtering and Targeting

### Combining Filters

Multi-region commands support all standard filtering options:

```bash
# By tags across regions
ztictl ssm exec-multi --all-regions --tags "Environment=prod,App=web" "status.sh"

# By instance IDs (must exist in targeted regions)
ztictl ssm exec-multi cac1,use1 --instances "i-1234,i-5678" "uptime"

# Mixed filtering
ztictl ssm exec-multi --region-group production \
  --tags "Type=application" \
  --running \
  "deployment-check.sh"
```

### Region-Specific Behavior

The command automatically handles region-specific aspects:
- Validates region names and converts shortcodes
- Establishes separate AWS clients per region
- Handles region-specific errors gracefully
- Aggregates results across all regions

## Error Handling

### Partial Failures

Multi-region execution continues even if some regions fail:

```bash
# Continue execution despite regional failures
ztictl ssm exec-multi --all-regions --tags "App=web" "health-check" --continue-on-error

# The output will show:
# ✅ Successful regions
# ⚠️  Regions with partial failures
# ❌ Failed regions
```

### Timeout Configuration

Set per-region timeout:

```bash
# 30-second timeout per region
ztictl ssm exec-multi --all-regions --tags "Type=batch" "process.sh" --timeout 30
```

## Performance Considerations

### Optimal Parallelism

- **Default**: 5 regions processed simultaneously
- **CPU-bound operations**: Use `--parallel $(nproc)`
- **Network-intensive**: Use `--parallel 3-5`
- **Critical operations**: Use `--parallel 1` for sequential execution

### Large-Scale Deployments

For operations across many regions:

```bash
# Gradual rollout
ztictl ssm exec-multi --region-group canary --tags "App=web" "deploy.sh"
ztictl ssm exec-multi --region-group production --tags "App=web" "deploy.sh"

# With health checks between regions
ztictl ssm exec-multi --all-regions \
  --tags "App=web" \
  "deploy.sh && health-check.sh" \
  --parallel 1 \
  --stop-on-error
```

## Use Cases

### 1. Security Patching
```bash
# Apply security updates across all regions
ztictl ssm exec-multi --all-regions \
  --tags "PatchGroup=critical" \
  "sudo yum update -y --security" \
  --parallel 3
```

### 2. Configuration Auditing
```bash
# Check configuration across regions
ztictl ssm exec-multi --all-regions \
  --tags "Type=webserver" \
  "grep -i 'ssl_protocols' /etc/nginx/nginx.conf"
```

### 3. Service Status Monitoring
```bash
# Check service status globally
ztictl ssm exec-multi --region-group production \
  --tags "Service=api" \
  "systemctl is-active api-service && curl -s localhost:8080/health"
```

### 4. Log Collection
```bash
# Gather logs from all regions
ztictl ssm exec-multi --all-regions \
  --tags "App=frontend" \
  "tail -n 100 /var/log/app/error.log" \
  > global-error-logs.txt
```

## Best Practices

1. **Test in Development First**
   ```bash
   ztictl ssm exec-multi --region-group development --tags "App=test" "your-command"
   ```

2. **Use Appropriate Parallelism**
   - Start with default (5)
   - Adjust based on command complexity
   - Use sequential (1) for critical changes

3. **Implement Health Checks**
   ```bash
   ztictl ssm exec-multi --all-regions \
     --tags "App=web" \
     "deploy.sh && sleep 5 && health-check.sh || rollback.sh"
   ```

4. **Monitor Execution Time**
   - Set reasonable timeouts
   - Consider region-specific latencies
   - Use `--verbose` for detailed timing

5. **Handle Region-Specific Configs**
   ```bash
   # Use region-aware scripts
   ztictl ssm exec-multi --all-regions \
     --tags "Type=database" \
     'REGION=$(curl -s http://169.254.169.254/latest/meta-data/placement/region); configure-db.sh $REGION'
   ```

## Troubleshooting

### Common Issues

1. **Region not found**
   - Verify region name/shortcode
   - Check `~/.ztictl.yaml` configuration
   - Run `ztictl config validate`

2. **No instances found**
   - Verify tags exist in target regions
   - Check instance SSM status: `ztictl ssm list --region <region>`

3. **Timeout errors**
   - Increase timeout: `--timeout 60`
   - Reduce parallelism: `--parallel 2`
   - Check network connectivity

4. **Partial failures**
   - Review per-region error messages
   - Check region-specific IAM permissions
   - Verify SSM agent status in failing regions

### Debug Mode

Enable verbose output for troubleshooting:

```bash
# Verbose output with timing
ztictl ssm exec-multi --all-regions \
  --tags "App=web" "test-command" \
  --verbose \
  --log-level debug
```

## See Also

- [COMMANDS.md](COMMANDS.md) - Complete command reference
- [CONFIGURATION.md](CONFIGURATION.md) - Configuration file details
- [Region Codes](COMMANDS.md#region-shortcodes) - List of all region shortcodes