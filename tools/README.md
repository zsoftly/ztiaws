# Tools Directory

This directory contains utility scripts for development and testing.

## Release Documentation Generator

The `02_release_docs_generator.sh` script automatically generates `CHANGELOG.md` and `RELEASE_NOTES.txt` files for releases. It follows the project's logging standards from `src/00_utils.sh` and supports conventional commit patterns.

### Features

- **Conventional Commits**: Automatically categorizes commits as features, fixes, or other changes
- **Semantic Versioning**: Validates version format (v1.2.3 or 1.2.3)
- **Automatic Git Integration**: Detects latest tags and GitHub repository information
- **Professional Logging**: Uses standardized logging from `src/00_utils.sh`
- **Safe Operations**: Atomic file updates with proper error handling

### Usage

```bash
# Generate docs for a release version
./tools/02_release_docs_generator.sh --version v2.1.0

# Force regeneration of existing files
./tools/02_release_docs_generator.sh --version v2.1.0 --force

# Specify a different base tag for comparison
./tools/02_release_docs_generator.sh --version v2.1.0 --latest-tag v2.0.5
```

### Options

| Option                  | Description                                                       |
| ----------------------- | ----------------------------------------------------------------- |
| `-v, --version VERSION` | Release version (required, format: v1.2.3 or 1.2.3)               |
| `-t, --latest-tag TAG`  | Latest release tag for comparison (auto-detected if not provided) |
| `-f, --force`           | Force regeneration even if files exist                            |
| `-h, --help`            | Show help message                                                 |

## EC2 Test Manager Script

The `01_ec2_test_manager.sh` script was created to streamline EC2 instance management for testing the SSM program. It automatically handles IAM role creation, instance provisioning, and cleanup, making it easy to spin up test environments quickly.

### Why This Script Exists

- **Simplifies testing**: Creates SSM-ready instances with proper IAM roles automatically
- **Prevents configuration drift**: Uses consistent, predefined settings
- **Enables rapid iteration**: Quick create/test/destroy cycles
- **Reduces manual errors**: Automates IAM setup and instance tagging

### Quick Start

```bash
# Create test instances
./tools/01_ec2_test_manager.sh create --count 2 --owner YourName

# Verify they're running
./tools/01_ec2_test_manager.sh verify

# Test SSM functionality (examples)
./ssm cac1 YourName-web-server-1                                    # Connect to instance
./ssm exec cac1 YourName-web-server-1 "hostname"                    # Execute command
./ssm upload cac1 YourName-web-server-1 test.txt /tmp/test.txt      # Upload file
./ssm download cac1 YourName-web-server-2 /var/log/messages ./messages.log  # Download file

# Clean up when done
./tools/01_ec2_test_manager.sh delete
```

### Commands

| Command  | Purpose                                        |
| -------- | ---------------------------------------------- |
| `create` | Create EC2 instances with SSM-enabled IAM role |
| `verify` | Show status of tracked instances               |
| `delete` | Terminate and cleanup all tracked instances    |

### Key Options

- `--count NUMBER`: How many instances to create
- `--owner NAME`: Required - used as instance name prefix
- `--name-prefix TEXT`: Instance suffix (default: "web-server")

### Automatic Features

- **IAM Role Creation**: Creates `EC2-SSM-Role` with `AmazonSSMManagedInstanceCore` policy
- **Instance Tracking**: Stores instance IDs in `ec2-instances.txt`
- **Batch Operations**: Efficient bulk create/verify/delete operations
- **Safety Tags**: All instances tagged with `ManagedBy=ec2-manager`

### Logging

Enable detailed logging with: `ENABLE_EC2_MANAGER_LOGGING=true`

Log files stored in: `~/logs/ec2-YYYY-MM-DD.log`
