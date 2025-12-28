# ZTiAWS Command Reference

This document provides comprehensive documentation for all ZTiAWS commands. For installation instructions, see [INSTALLATION.md](../INSTALLATION.md).

## Table of Contents

- [ztictl Commands](#ztictl-commands)
  - [Authentication](#authentication-commands)
  - [Configuration](#configuration-commands)
  - [SSM Operations](#ssm-operations)
  - [Power Management](#power-management)
  - [Multi-Region Operations](#multi-region-operations)
- [Legacy Bash Commands](#legacy-bash-commands)

---

## ztictl Commands

### Authentication Commands

#### `ztictl auth login`

Authenticate with AWS SSO and select account/role interactively.

```bash
# Basic login
ztictl auth login

# Login with specific profile
ztictl auth login --profile production
```

#### `ztictl auth whoami`

Display current AWS identity and credentials status.

```bash
ztictl auth whoami
```

#### `ztictl auth logout`

Clear AWS SSO cached credentials.

```bash
ztictl auth logout
```

### Configuration Commands

#### `ztictl config init`

Initialize ztictl configuration interactively.

```bash
# Interactive setup
ztictl config init --interactive

# Force overwrite existing config
ztictl config init --force
```

#### `ztictl config check`

Verify system requirements and configuration.

```bash
# Check requirements
ztictl config check

# Auto-fix issues where possible
ztictl config check --fix
```

#### `ztictl config show`

Display current configuration settings.

```bash
ztictl config show
```

#### `ztictl config validate`

Validate configuration file syntax and required fields.

```bash
ztictl config validate
```

#### `ztictl config repair`

**New in v2.5+** - Interactively fix configuration issues.

```bash
# Detects and helps fix invalid values
ztictl config repair
```

### SSM Operations

> **🔍 Interactive Fuzzy Finder**: Many SSM commands now feature an interactive fuzzy finder for enhanced user experience. The fuzzy finder provides real-time search, instance details preview, and keyboard navigation.

#### `ztictl ssm list`

**🔍 Interactive Fuzzy Finder** - Browse and search EC2 instances with real-time filtering.

The fuzzy finder provides:

- **Real-time search**: Type to filter instances by name or ID
- **Instance details**: Shows state, platform, IP addresses, and SSM status
- **Browse mode**: Navigate through instances without connecting
- **Keyboard shortcuts**: Arrow keys to navigate, Enter to select, Escape to cancel

```bash
# Launch interactive instance browser for a region (default)
ztictl ssm list --region cac1

# Pre-filter by tags, then browse with fuzzy finder
ztictl ssm list --region use1 --tags "Environment=prod,App=web"

# Browse instances with status filtering
ztictl ssm list --region euw1 --status running

# Use traditional table format instead of fuzzy finder
ztictl ssm list --region cac1 --table
```

**Flags:**

- `--table` - Display instances in traditional table format (for scripts/automation)

#### `ztictl ssm connect`

**🔍 Interactive Connection** - Connect to instances via Session Manager with fuzzy finder support.

**Interactive Mode** (no instance specified):

- Launches interactive fuzzy finder to search and select instances
- Real-time filtering by typing instance names or IDs
- Shows instance details including SSM status and platform
- Press Enter to connect, Escape to cancel

**Direct Mode** (instance specified):

- Connects directly to the specified instance by ID or name

```bash
# 🔍 Interactive mode - Launch fuzzy finder to search and connect
ztictl ssm connect --region euw1

# 🎯 Direct mode - Connect using instance ID
ztictl ssm connect i-1234567890abcdef0 --region use1

# 🎯 Direct mode - Connect using instance name
ztictl ssm connect prod-web-01 --region cac1
```

#### `ztictl ssm exec`

Execute commands on instances.

```bash
# Single instance
ztictl ssm exec i-1234567890abcdef0 "uptime" --region use1

# Multiple instances by ID
ztictl ssm exec --instances "i-1234,i-5678" "df -h" --region cac1

# By tags
ztictl ssm exec --tags "Environment=prod" "systemctl status nginx" --region euw1
```

#### `ztictl ssm transfer`

Transfer files to/from instances.

```bash
# Upload file (automatic S3 for files ≥1MB)
ztictl ssm transfer /local/file.txt i-1234567890abcdef0:/remote/path/ --region use1

# Download file
ztictl ssm transfer i-1234567890abcdef0:/remote/file.txt /local/path/ --region cac1
```

### Power Management

**New in v2.4+** - EC2 instance power management commands.

#### `ztictl ssm start`

Start stopped EC2 instances.

```bash
# Single instance
ztictl ssm start i-1234567890abcdef0 --region cac1

# Multiple instances
ztictl ssm start --instances "i-1234,i-5678" --region use1

# With parallelism control
ztictl ssm start --instances "i-1234,i-5678,i-9012" --parallel 2 --region euw1
```

#### `ztictl ssm stop`

Stop running EC2 instances.

```bash
# Single instance
ztictl ssm stop i-1234567890abcdef0 --region cac1

# Multiple instances
ztictl ssm stop --instances "i-1234,i-5678" --region use1
```

#### `ztictl ssm start-tagged`

Start instances by tags (parallel execution).

```bash
# Start all instances with specific tags
ztictl ssm start-tagged --tags "Environment=dev" --region cac1

# Multiple tag filters
ztictl ssm start-tagged --tags "Environment=staging,App=web" --region use1

# Control parallelism
ztictl ssm start-tagged --tags "AutoStart=true" --parallel 5 --region euw1
```

#### `ztictl ssm stop-tagged`

Stop instances by tags (parallel execution).

```bash
# Stop all instances with specific tags
ztictl ssm stop-tagged --tags "Environment=dev" --region cac1

# With confirmation bypass
ztictl ssm stop-tagged --tags "AutoStop=true" --force --region use1
```

### Multi-Region Operations

**New in v2.6+** - Execute commands across multiple AWS regions simultaneously.

#### `ztictl ssm exec-multi`

Execute commands across multiple regions. See [MULTI_REGION.md](MULTI_REGION.md) for detailed configuration.

```bash
# Using region list (shortcodes or full names)
ztictl ssm exec-multi cac1,use1,euw1 --tags "Environment=prod" "uptime"

# Using --regions flag
ztictl ssm exec-multi --regions cac1,us-east-1,eu-west-1 --tags "App=api" "health-check"

# All configured regions (from ~/.ztictl.yaml)
ztictl ssm exec-multi --all-regions --tags "Component=web" "systemctl status nginx"

# Using region groups (configured in ~/.ztictl.yaml)
ztictl ssm exec-multi --region-group production --tags "App=api" "curl localhost:8080/health"

# Control parallelism
ztictl ssm exec-multi --all-regions --tags "Type=cache" "redis-cli ping" --parallel 3
```

---

## Interactive Fuzzy Finder

**New in v2.9+** - Enhanced interactive instance selection with real-time search capabilities.

### Features

- **🔍 Real-time Search**: Type to instantly filter instances by name or ID
- **📋 Instance Details**: Preview instance information before connecting
- **⌨️ Keyboard Navigation**:
  - Arrow keys or `j/k` to navigate
  - Enter to select instance
  - Escape or `q` to cancel
  - `/` to focus search input
- **🎨 Color-coded Status**: Visual indicators for SSM status and instance state
- **🏷️ Tag Support**: Pre-filter instances by tags before launching fuzzy finder

### When Fuzzy Finder Launches

The interactive fuzzy finder automatically launches when:

- `ztictl ssm list` - Always launches for browsing instances
- `ztictl ssm connect` - Launches when no instance identifier is provided
- Any command where instance selection is needed but not specified

### Example Workflow

```bash
# 1. Launch fuzzy finder for region
ztictl ssm connect --region ca-central-1

# 2. Type to search (e.g., "web" to find web servers)
#    Results filter in real-time as you type

# 3. Use arrow keys to navigate through matches

# 4. Press Enter to connect to selected instance
#    Or Escape to cancel
```

### Instance Information Displayed

- **Instance ID**: e.g., `i-1234567890abcdef0`
- **Name**: Instance name from Name tag
- **State**: running, stopped, pending, etc.
- **Platform**: Linux, Windows
- **IP Addresses**: Private and public IPs
- **SSM Status**: ✓ Online, ⚠ Lost, ✗ No Agent

---

## Legacy Bash Commands

> ⚠️ **Note:** These bash-based commands remain in production use. Consider migrating to `ztictl` for enhanced features and cross-platform support.

### authaws

AWS SSO authentication tool.

```bash
# Interactive login
authaws

# Check configuration
authaws --check

# Show version
authaws --version
```

### ssm

Session Manager operations tool.

```bash
# List instances
ssm cac1                    # List in Canada Central
ssm use1                    # List in US East 1

# Connect to instance
ssm i-1234567890abcdef0    # Connect by ID
ssm prod-web-01            # Connect by name

# Execute command
ssm i-1234567890abcdef0 -c "uptime"

# Transfer files
ssm i-1234567890abcdef0 -u /local/file.txt:/remote/path/
ssm i-1234567890abcdef0 -d /remote/file.txt:/local/path/
```

---

## Region Shortcodes

Both `ztictl` and legacy tools support region shortcodes:

| Shortcode | Region Name    | Location                 |
| --------- | -------------- | ------------------------ |
| `cac1`    | ca-central-1   | Canada (Montreal)        |
| `use1`    | us-east-1      | US East (N. Virginia)    |
| `use2`    | us-east-2      | US East (Ohio)           |
| `usw1`    | us-west-1      | US West (N. California)  |
| `usw2`    | us-west-2      | US West (Oregon)         |
| `euw1`    | eu-west-1      | EU (Ireland)             |
| `euw2`    | eu-west-2      | EU (London)              |
| `euc1`    | eu-central-1   | EU (Frankfurt)           |
| `apse1`   | ap-southeast-1 | Asia Pacific (Singapore) |
| `apse2`   | ap-southeast-2 | Asia Pacific (Sydney)    |
| `apne1`   | ap-northeast-1 | Asia Pacific (Tokyo)     |

See full list in the source code: `ztictl/pkg/aws/regions.go`

---

## Exit Codes

All commands follow standard exit code conventions:

- `0`: Success
- `1`: General error
- `2`: Misuse of command (invalid arguments)
- `130`: Interrupted (Ctrl+C)

---

## Environment Variables

| Variable           | Description        | Default        |
| ------------------ | ------------------ | -------------- |
| `AWS_PROFILE`      | AWS profile to use | default        |
| `AWS_REGION`       | Default AWS region | us-east-1      |
| `ZTICTL_CONFIG`    | Config file path   | ~/.ztictl.yaml |
| `ZTICTL_LOG_LEVEL` | Logging level      | info           |

---

## See Also

- [CONFIGURATION.md](CONFIGURATION.md) - Configuration file reference
- [MULTI_REGION.md](MULTI_REGION.md) - Multi-region operations guide
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Common issues and solutions
