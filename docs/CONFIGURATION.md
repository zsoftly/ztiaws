# Configuration Reference

Complete reference for `ztictl` configuration file (`~/.ztictl.yaml`).

## Configuration File Location

| Platform | Default Location |
|----------|-----------------|
| Linux/macOS | `~/.ztictl.yaml` |
| Windows | `%USERPROFILE%\.ztictl.yaml` |

Override with environment variable: `ZTICTL_CONFIG=/path/to/config.yaml`

## File Structure

```yaml
# ~/.ztictl.yaml - Complete configuration example

# AWS SSO Configuration
sso:
  start_url: "https://d-1234567890.awsapps.com/start"  # Required for SSO
  region: "ca-central-1"                                # SSO service region (default)

# Default AWS region for operations
default_region: "ca-central-1"                         # Default region for all operations

# Multi-region configuration
regions:
  # List of enabled regions (shortcodes or full names)
  enabled:
    - cac1          # ca-central-1
    - use1          # us-east-1
    - use2          # us-east-2
    - euw1          # eu-west-1
    - euw2          # eu-west-2
    - apse1         # ap-southeast-1
    - apse2         # ap-southeast-2
  
  # Region groups for logical organization
  groups:
    all:            # Special group containing all enabled regions
      - cac1
      - use1
      - use2
      - euw1
      - euw2
      - apse1
      - apse2
    production:
      - use1
      - euw1
      - apse1
    development:
      - use2
      - cac1
    europe:
      - euw1
      - euw2
    americas:
      - use1
      - use2
      - cac1
    asia-pacific:
      - apse1
      - apse2

# Logging configuration
logging:
  directory: "~/logs"         # Log file directory
  file_logging: true          # Enable file logging
  level: "info"              # Log level: debug, info, warn, error
  max_size: 100              # Max log file size in MB
  max_backups: 5             # Number of old log files to keep
  max_age: 30                # Days to keep old log files

# System configuration
system:
  iam_propagation_delay: 5   # Seconds to wait for IAM changes
  file_size_threshold: 1048576  # Bytes (1MB) - files larger use S3
  s3_bucket_prefix: "ztictl-ssm-file-transfer"
  temp_directory: "/tmp"     # Temporary file directory
  parallel_operations: 5     # Default parallelism for multi-operations
  command_timeout: 30        # Default command timeout in seconds
```

## Configuration Sections

### SSO Configuration

Controls AWS SSO authentication settings.

```yaml
sso:
  start_url: "https://d-1234567890.awsapps.com/start"  # Your AWS SSO portal URL
  region: "ca-central-1"                                # Region where SSO is configured (default)
```

**Required for**: 
- `ztictl auth login`
- Any operation requiring AWS credentials

**How to find your SSO domain ID**:
1. Log into AWS SSO portal
2. Look at the URL: `https://YOUR-DOMAIN-ID.awsapps.com/start`
3. The domain ID is the part between `https://` and `.awsapps.com`
4. During setup, you only need to enter the domain ID (e.g., `d-1234567890` or `zsoftly`)

### Default Region

Sets the default AWS region for operations when not specified via `--region` flag.

```yaml
default_region: "ca-central-1"  # Can use shortcode or full name
```

**Precedence order**:
1. Command flag: `--region`
2. Environment variable: `AWS_REGION`
3. Config file: `default_region`
4. Fallback: `ca-central-1`

### Regions Configuration

Defines regions for multi-region operations.

```yaml
regions:
  enabled:           # Regions available for --all-regions
    - cac1          # Shortcodes supported
    - us-east-1     # Full names supported
  
  groups:           # Logical groupings for --region-group
    production:
      - use1
      - euw1
```

**Special groups**:
- `all`: Automatically created, contains all enabled regions

### Logging Configuration

Controls logging behavior and output.

```yaml
logging:
  directory: "~/logs"        # Where to store log files
  file_logging: true         # Enable/disable file logging
  level: "info"             # Verbosity level
  max_size: 100             # MB per log file
  max_backups: 5            # Number of old files to keep
  max_age: 30               # Days to retain logs
```

**Log levels** (least to most verbose):
- `error`: Only errors
- `warn`: Warnings and errors
- `info`: Informational messages (default)
- `debug`: Detailed debugging information

**Log file locations**:
- Linux/macOS: `~/logs/ztictl-YYYYMMDD.log`
- Windows: `%USERPROFILE%\logs\ztictl-YYYYMMDD.log`

### System Configuration

Advanced system behavior settings.

```yaml
system:
  iam_propagation_delay: 5      # Seconds to wait after IAM changes
  file_size_threshold: 1048576  # Bytes - threshold for S3 transfer
  s3_bucket_prefix: "ztictl"    # Prefix for temporary S3 buckets
  temp_directory: "/tmp"         # Temporary file storage
  parallel_operations: 5         # Default parallelism
  command_timeout: 30           # Default timeout in seconds
```

## Initial Setup

### Interactive Configuration

The easiest way to create your configuration:

```bash
# First-time setup
ztictl config init --interactive

# This will prompt for:
# 1. SSO domain ID (e.g., d-1234567890 or zsoftly)
# 2. SSO region (default: ca-central-1)
# 3. Default region (default: ca-central-1)
# 4. Multi-region setup
# 5. Logging preferences
```

### Manual Configuration

Create `~/.ztictl.yaml` manually:

```bash
# Create minimal configuration
cat > ~/.ztictl.yaml << 'EOF'
sso:
  start_url: "https://your-domain-id.awsapps.com/start"
  region: "ca-central-1"
default_region: "ca-central-1"
EOF

# Validate configuration
ztictl config validate
```

## Configuration Management Commands

### Initialize Configuration
```bash
# Interactive setup (recommended)
ztictl config init --interactive

# Create sample configuration
ztictl config init

# Force overwrite existing
ztictl config init --force
```

### Validate Configuration
```bash
# Check for syntax and required fields
ztictl config validate
```

### Show Configuration
```bash
# Display current settings
ztictl config show
```

### Repair Configuration
```bash
# Fix invalid configuration interactively
ztictl config repair
```

### Check Requirements
```bash
# Verify system requirements
ztictl config check

# Auto-fix issues where possible
ztictl config check --fix
```

## Environment Variables

Environment variables can override configuration file settings:

| Variable | Overrides | Example |
|----------|-----------|---------|
| `AWS_PROFILE` | AWS profile selection | `AWS_PROFILE=production` |
| `AWS_REGION` | `default_region` | `AWS_REGION=eu-west-1` |
| `ZTICTL_CONFIG` | Config file location | `ZTICTL_CONFIG=/etc/ztictl.yaml` |
| `ZTICTL_LOG_LEVEL` | `logging.level` | `ZTICTL_LOG_LEVEL=debug` |
| `ZTICTL_LOG_DIR` | `logging.directory` | `ZTICTL_LOG_DIR=/var/log` |
| `ZTICTL_SELECTOR_HEIGHT` | Fuzzy finder display height | `ZTICTL_SELECTOR_HEIGHT=10` |

### UI Customization

#### Fuzzy Finder Display Height

Control the number of items shown in the interactive account/role selector:

```bash
# Show 3 items (compact display)
ZTICTL_SELECTOR_HEIGHT=3 ztictl auth login

# Show 10 items (more visible options)
ZTICTL_SELECTOR_HEIGHT=10 ztictl auth login

# Use default (5 items - matches fzf --height=20%)
ztictl auth login
```

**Settings**:
- **Default**: 5 items (equivalent to fzf `--height=20%`)
- **Range**: 1-20 items
- **Behavior**: Invalid values log a warning and use default

**Features**:
- All accounts/roles remain searchable regardless of display height
- Full keyboard navigation through entire list with arrow keys
- Single bordered rectangle at bottom of terminal
- Preview panel shows selected account/role details

## Configuration Validation

### Automatic Validation

Configuration is validated automatically when:
- Running any `ztictl` command
- Using `ztictl config validate`
- After `ztictl config init`

### Common Validation Errors

1. **Invalid SSO URL**
   ```
   Error: SSO start URL must begin with https://
   Fix: ztictl config repair
   ```

2. **Invalid Region**
   ```
   Error: 'caa-central-1' is not a valid AWS region
   Fix: Use 'ca-central-1' or shortcode 'cac1'
   ```

3. **Missing Required Fields**
   ```
   Error: SSO configuration required
   Fix: ztictl config init --interactive
   ```

## Migration from Legacy Tools

### From .env file (authaws/ssm)

If you have an existing `.env` file:

```bash
# .env file format
SSO_START_URL=https://d-1234567890.awsapps.com/start
SSO_REGION=us-east-1
DEFAULT_PROFILE=default

# ztictl will attempt to read this automatically
# To create proper config:
ztictl config init --interactive
```

### Automatic Migration

On first run, `ztictl` will:
1. Check for `~/.ztictl.yaml`
2. If not found, check for `.env`
3. Offer to migrate settings
4. Create new configuration

## Advanced Configuration

### Per-Profile Configuration (Planned)

Future support for multiple profiles:

```yaml
profiles:
  default:
    sso:
      start_url: "https://personal.awsapps.com/start"
    default_region: "ca-central-1"
  
  work:
    sso:
      start_url: "https://company.awsapps.com/start"
    default_region: "ca-central-1"
```

### Custom Commands (Planned)

Define command aliases:

```yaml
aliases:
  prod-status: "exec-multi --region-group production --tags Environment=prod"
  dev-deploy: "exec-multi --region-group development --tags App=web deploy.sh"
```

## Troubleshooting

### Configuration Not Found
```bash
# Check file location
ls -la ~/.ztictl.yaml

# Create new configuration
ztictl config init
```

### Invalid Configuration
```bash
# Validate and show errors
ztictl config validate

# Interactive repair
ztictl config repair

# Manual edit
nano ~/.ztictl.yaml
```

### Permission Issues
```bash
# Fix file permissions
chmod 600 ~/.ztictl.yaml

# Fix directory permissions
chmod 700 ~/logs
```

## Best Practices

1. **Keep Configuration Secure**
   - Use `chmod 600 ~/.ztictl.yaml`
   - Don't commit to version control
   - Use environment variables for sensitive data

2. **Use Region Groups**
   - Organize regions logically
   - Create environment-specific groups
   - Use meaningful group names

3. **Configure Logging**
   - Enable file logging for audit trails
   - Set appropriate retention periods
   - Use debug level only when troubleshooting

4. **Regular Validation**
   - Run `ztictl config validate` after changes
   - Use `ztictl config check` periodically
   - Keep configuration backed up

## See Also

- [COMMANDS.md](COMMANDS.md) - Command reference
- [MULTI_REGION.md](MULTI_REGION.md) - Multi-region operations
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Common issues