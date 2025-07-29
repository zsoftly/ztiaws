# ZTiAWS Installation and Testing Guide for Mac Intel (Sonoma)

## Overview

ZTiAWS (Zero Trust in AWS) is a powerful DevOps tool that simplifies AWS operations through secure SSM connections. As a DevOps cloud engineer, this tool will help you:

- **Securely connect to EC2 instances** without SSH keys
- **Execute remote commands** across multiple instances  
- **Transfer files** securely through AWS infrastructure
- **Manage multi-region operations** efficiently
- **Integrate with AWS SSO** for enterprise security

## Two Versions Available

The project provides both stable production tools and new enhanced features:

### 1. Shell Scripts (Production Stable - v1.4.x) ✅ Recommended for Production
- **Status**: Battle-tested, actively maintained
- **Tools**: `ssm` and `authaws` scripts
- **Use Case**: Production workflows, established environments

### 2. Go Binary (Preview - v2.0.x) 🧪 Testing Phase  
- **Status**: Preview/Testing, enhanced features
- **Tool**: `ztictl` unified binary
- **Use Case**: Testing new features, development environments

## Prerequisites

Before installation, ensure you have:

1. **macOS Sonoma** (your current setup)
2. **AWS CLI configured** with appropriate credentials
3. **Homebrew installed** (recommended)
4. **Terminal access** with basic command-line knowledge

## Installation Steps

### Step 1: Install Shell Scripts (Production Stable)

Open Terminal and run these commands:

```bash
# Create a directory for ZTiAWS tools
mkdir -p ~/ztiaws
cd ~/ztiaws

# Download the shell scripts
curl -L -o ssm https://raw.githubusercontent.com/zsoftly/ztiaws/main/ssm
curl -L -o authaws https://raw.githubusercontent.com/zsoftly/ztiaws/main/authaws

# Make them executable
chmod +x ssm authaws

# Download supporting files
mkdir -p src
curl -L -o src/00_utils.sh https://raw.githubusercontent.com/zsoftly/ztiaws/main/src/00_utils.sh
curl -L -o src/01_regions.sh https://raw.githubusercontent.com/zsoftly/ztiaws/main/src/01_regions.sh
curl -L -o src/02_ssm_instance_resolver.sh https://raw.githubusercontent.com/zsoftly/ztiaws/main/src/02_ssm_instance_resolver.sh
curl -L -o src/03_ssm_command_runner.sh https://raw.githubusercontent.com/zsoftly/ztiaws/main/src/03_ssm_command_runner.sh
curl -L -o src/04_ssm_file_transfer.sh https://raw.githubusercontent.com/zsoftly/ztiaws/main/src/04_ssm_file_transfer.sh

# Install to system PATH (optional but recommended)
sudo cp ssm authaws /usr/local/bin/
sudo cp -r src /usr/local/bin/
```

### Step 2: Install Go Binary (Preview Version)

```bash
# Download ztictl for Mac Intel
curl -L -o ztictl https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-darwin-amd64

# Make executable
chmod +x ztictl

# Install to system PATH (optional)
sudo mv ztictl /usr/local/bin/
```

### Step 3: Verify Installation

```bash
# Test shell scripts
ssm --version
authaws --version

# Test Go binary
ztictl --version
```

## Initial Configuration

### Configure AWS CLI (if not already done)

```bash
# Configure AWS credentials
aws configure

# Or configure a specific profile
aws configure --profile your-profile-name
```

### Configure AWS SSO Authentication

```bash
# Using shell scripts
authaws configure

# Using Go binary (alternative)
ztictl auth configure
```

## Testing the Installation

### Test 1: Basic Functionality

```bash
# Test shell scripts
./ssm --help
./authaws --help

# Test Go binary
ztictl --help
```

### Test 2: AWS Authentication

```bash
# Shell script version
authaws login

# Go binary version
ztictl auth login
```

### Test 3: List SSM Instances

```bash
# Shell script version
ssm list

# Go binary version  
ztictl ssm list
```

### Test 4: Region Operations

```bash
# Shell script version - list instances in specific region
ssm list --region us-east-1

# Go binary version
ztictl ssm list --region us-east-1

# Test region shortcuts (Go binary)
ztictl ssm list --region use1  # shortcut for us-east-1
```

## Usage Examples

### Connecting to Instances

```bash
# Shell scripts
ssm connect i-1234567890abcdef0

# Go binary
ztictl ssm connect i-1234567890abcdef0
```

### Running Remote Commands

```bash
# Shell scripts - single instance
ssm exec us-east-1 i-1234567890abcdef0 "uptime"

# Go binary - single instance
ztictl ssm exec --region us-east-1 --instance-id i-1234567890abcdef0 --command "uptime"

# Go binary - multiple instances by tag
ztictl ssm exec-tagged --region us-east-1 --tag-key Environment --tag-value production --command "df -h"
```

### File Transfer

```bash
# Shell scripts
ssm transfer local-file.txt i-1234567890abcdef0:/tmp/

# Go binary
ztictl ssm transfer local-file.txt i-1234567890abcdef0:/tmp/
```

## Troubleshooting

### Common Issues

1. **"Command not found"**
   - Ensure scripts are in your PATH or use full path: `./ssm` instead of `ssm`
   - Check file permissions: `chmod +x ssm authaws ztictl`

2. **"Permission denied"**
   - Make files executable: `chmod +x ssm authaws ztictl`
   - For system installation: Use `sudo`

3. **AWS Authentication Issues**
   - Verify AWS CLI configuration: `aws sts get-caller-identity`
   - Check SSO configuration: `authaws configure` or `ztictl auth configure`

4. **SSM Connection Issues**
   - Ensure EC2 instances have SSM agent installed
   - Verify IAM roles have SSM permissions
   - Check security groups allow outbound HTTPS (443)

### Getting Help

```bash
# Shell scripts
ssm --help
authaws --help

# Go binary
ztictl --help
ztictl ssm --help
ztictl auth --help
```

## What to Test and Document

As a DevOps engineer evaluating these tools, focus on:

### 1. Installation Experience
- Which version was easier to install?
- Any missing dependencies?
- Installation time and complexity

### 2. Performance Comparison
- Speed of commands between shell scripts vs Go binary
- Resource usage
- Startup time

### 3. Feature Differences
- Compare available commands
- User interface quality (colors, progress bars, etc.)
- Error message clarity

### 4. Production Readiness
- Stability during testing
- Error handling
- Documentation quality

### 5. Integration Testing
- AWS SSO authentication flow
- Multi-region operations
- File transfer capabilities
- Remote command execution

## Expected Outcomes

After completing this installation and testing:

1. **You'll have both versions installed** and understand their differences
2. **You'll know which version** works better for your workflow
3. **You'll understand the tool's capabilities** for DevOps operations
4. **You'll be able to recommend** which version to use in production
5. **You'll have hands-on experience** with modern AWS DevOps tooling

## Next Steps

1. **Install both versions** following this guide
2. **Test basic functionality** with your AWS environment
3. **Compare performance** and usability
4. **Document your findings** for your team
5. **Choose the version** that best fits your production needs

## Support Resources

- **GitHub Repository**: https://github.com/zsoftly/ztiaws
- **Issues**: https://github.com/zsoftly/ztiaws/issues
- **Documentation**: Check the repo's README and wiki

Good luck with your installation and testing! This is an excellent opportunity to evaluate cutting-edge DevOps tooling for AWS operations.