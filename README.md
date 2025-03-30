# quickssm

![Ubuntu](https://github.com/ZSoftly/quickssm/actions/workflows/test.yml/badge.svg)
![macOS](https://github.com/ZSoftly/quickssm/actions/workflows/test.yml/badge.svg)

A set of streamlined CLI tools for AWS management, including SSM Session Manager and SSO authentication, making it easier to interact with AWS services across regions.

## Features

- **quickssm**: Connect to EC2 instances using short region codes 
- **authaws**: Streamlined AWS SSO authentication with account and role selection
- Interactive listing of available instances and accounts
- Automatic validation of AWS CLI and required plugins
- Support for multiple AWS regions
- Color-coded output for better readability

## Prerequisites

- AWS CLI installed (follow [official installation guide](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html))
- AWS Session Manager plugin (can be installed interactively via `ssm check`)
- AWS credentials configured (`aws configure`)
- Bash or Zsh shell
- Proper IAM permissions for SSM Session Manager and SSO access
- Additional utilities: `jq` and `fzf` (required for `authaws`)

## Quick Start

One-liner to download, install, and start using both tools (for bash users):
```bash
git clone https://github.com/ZSoftly/quickssm.git && cd quickssm && chmod +x ssm authaws && ./ssm check && echo -e "\n# Add quickssm to PATH\nexport PATH=\"\$PATH:$(pwd)\"" >> ~/.bashrc && source ~/.bashrc
```

For zsh users:
```bash
git clone https://github.com/ZSoftly/quickssm.git && cd quickssm && chmod +x ssm authaws && ./ssm check && echo -e "\n# Add quickssm to PATH\nexport PATH=\"\$PATH:$(pwd)\"" >> ~/.zshrc && source ~/.zshrc
```

After running the appropriate command for your shell, you can use the tools by simply typing `ssm` or `authaws` from anywhere.

## Installation Options

### Option 1: Local User Installation (Recommended)

For bash users:
```bash
git clone https://github.com/ZSoftly/quickssm.git
cd quickssm
chmod +x ssm authaws
./ssm check
./authaws check
echo -e "\n# Add quickssm to PATH\nexport PATH=\"\$PATH:$(pwd)\"" >> ~/.bashrc
source ~/.bashrc
```

For zsh users:
```bash
git clone https://github.com/ZSoftly/quickssm.git
cd quickssm
chmod +x ssm authaws
./ssm check
./authaws check
echo -e "\n# Add quickssm to PATH\nexport PATH=\"\$PATH:$(pwd)\"" >> ~/.zshrc
source ~/.zshrc
```

This is the recommended approach because:
- Keeps AWS tooling scoped to your user
- Maintains better security practices
- Makes updates easier without requiring sudo
- Aligns with AWS credentials being stored per-user in ~/.aws/
- Follows principle of least privilege
- Easier to manage different AWS configurations per user

### Option 2: System-wide Installation (Not Recommended)
```bash
git clone https://github.com/ZSoftly/quickssm.git
cd quickssm
chmod +x ssm authaws
./ssm check
./authaws check
INSTALL_DIR="$(pwd)"
sudo ln -s "$INSTALL_DIR/ssm" /usr/local/bin/ssm
sudo ln -s "$INSTALL_DIR/authaws" /usr/local/bin/authaws
sudo ln -s "$INSTALL_DIR/src" /usr/local/bin/src
```

Not recommended because:
- Any user on the system could run the tool and potentially access AWS resources
- Doesn't align well with per-user AWS credential management
- Requires sudo privileges for updates and modifications
- Can lead to security and audit tracking complications
- Makes it harder to manage different AWS configurations for different users

## Updating from Previous Versions

If you're updating from a previous version that used `auth_aws` instead of `authaws`, follow these steps:

### Option 1: Clean Update (Recommended)
```bash
# Navigate to your quickssm directory
cd /path/to/quickssm

# Backup your .env file if you have one
cp .env .env.backup

# Pull the latest changes
git pull

# Make the new scripts executable
chmod +x ssm authaws

# Remove the old symlink if you had one
rm -f /usr/local/bin/auth_aws  # May require sudo

# Update your path if needed or recreate symlinks
```

### Option 2: In-place Migration
```bash
# Navigate to your quickssm directory
cd /path/to/quickssm

# Pull the latest changes
git pull

# Make the new script executable
chmod +x authaws

# Create a symlink from the old name to the new script for compatibility
ln -s "$(pwd)/authaws" "$(pwd)/auth_aws"
```

## Usage

### SSM Session Manager Tool

#### Check System Requirements
```bash
ssm check
```

#### List Instances in a Region
```bash
ssm cac1  # Lists instances in Canada Central
```

#### Connect to an Instance
```bash
ssm i-1234abcd              # Connect to instance in default region (Canada Central)
ssm use1 i-1234abcd         # Connect to instance in US East
```

#### Show Help
```bash
ssm help
```

### AWS SSO Authentication Tool

#### First-time Setup
```bash
authaws check       # Check dependencies
authaws help        # Show help information
```

Before using `authaws`, set up a `.env` file in the same directory with the following content:
```
SSO_START_URL="https://your-sso-url.awsapps.com/start"
SSO_REGION="your-region"
DEFAULT_PROFILE="your-default-profile"
```

You can create a template file by running `authaws` without a valid .env file.

#### Log in to AWS SSO
```bash
authaws             # Use default profile from .env
authaws myprofile   # Use a specific profile name
```

The tool will:
1. Check for valid cached credentials
2. Initiate AWS SSO login if needed
3. Show an interactive list of accounts
4. Show an interactive list of roles for the selected account
5. Configure your AWS profile with the selected account and role

#### View AWS Credentials
```bash
authaws creds           # Show credentials for current profile
authaws creds myprofile # Show credentials for a specific profile
```

This will display your AWS access key, secret key, and session token for the specified profile.

## Supported Regions (for SSM tool)

| Shortcode | AWS Region    | Location     |
|-----------|---------------|--------------|
| cac1      | ca-central-1  | Montreal     |
| caw1      | ca-west-1     | Calgary      |
| use1      | us-east-1     | N. Virginia  |
| usw1      | us-west-1     | N. California|
| euw1      | eu-west-1     | Ireland      |

For a complete list of regions and their status, see [REGIONS.md](docs/REGIONS.md).

## IAM Permissions

### For SSM Session Manager:
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ssm:StartSession",
                "ssm:TerminateSession",
                "ssm:ResumeSession",
                "ec2:DescribeInstances"
            ],
            "Resource": "*"
        }
    ]
}
```

### For AWS SSO Authentication:
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "sso:GetRoleCredentials",
                "sso:ListAccountRoles",
                "sso:ListAccounts"
            ],
            "Resource": "*"
        }
    ]
}
```

## Troubleshooting

### AWS CLI Not Found
If AWS CLI is not installed, follow the [official AWS CLI installation guide](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html).

### Session Manager Plugin Missing
Run `ssm check` to install the plugin interactively, or follow the [manual installation instructions](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html).

### Missing jq or fzf
For Ubuntu/Debian: `sudo apt-get install jq fzf`
For macOS: `brew install jq fzf`

### AWS Credentials Not Configured
Run `aws configure` to set up your AWS credentials.

### Permission Errors
Ensure your AWS user/role has the required IAM permissions listed above.

### Shell Configuration
If the commands aren't available after installation, make sure you've added them to your PATH in the correct shell configuration file:
- For Bash users: `~/.bashrc`
- For Zsh users: `~/.zshrc`

You may need to restart your terminal or run `source ~/.bashrc` (or `source ~/.zshrc` for Zsh) for the changes to take effect.

### Script Name Changed
If you're getting "command not found" for `auth_aws`, note that the script has been renamed to `authaws` in v1.4.0+. Update your scripts and aliases accordingly.

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

Key areas for contribution:
- Adding support for new regions
- Improving documentation
- Adding new features
- Bug fixes

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Releasing a New Version

For maintainers who want to create a new release:

```bash
# Make sure you're on the main branch
git checkout main

# Pull the latest changes (including merged PRs)
git pull origin main

# Ensure all changes are committed and the working directory is clean
git status

# Create an annotated tag
git tag -a v1.x.x -m "Version 1.x.x: Brief description of changes"

# Push the tag to GitHub
git push origin v1.x.x
```

After pushing the tag, go to the GitHub repository and:
1. Click on "Releases"
2. Click "Draft a new release"
3. Select the tag you just pushed
4. Add release notes
5. Publish the release

This process ensures that releases are always created from the stable main branch after code has been properly reviewed and merged.

## Security

These tools require AWS credentials and access to your AWS resources. Always:
- Keep your AWS credentials secure
- Use appropriate IAM permissions
- Review security best practices in the [AWS Security Documentation](https://docs.aws.amazon.com/security/)