# quickssm

![Ubuntu](https://github.com/ZSoftly/quickssm/actions/workflows/test.yml/badge.svg)
![macOS](https://github.com/ZSoftly/quickssm/actions/workflows/test.yml/badge.svg)

A set of streamlined CLI tools for AWS management, including SSM Session Manager and SSO authentication, making it easier to interact with AWS services across regions.

## Features

- **quickssm**: Connect to EC2 instances using short region codes 
- **auth_aws**: Streamlined AWS SSO authentication with account and role selection
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
- Additional utilities: `jq` and `fzf` (required for `aws_auth`)

## Quick Start

One-liner to download, install, and start using both tools (for bash users):
```bash
git clone https://github.com/ZSoftly/quickssm.git && cd quickssm && chmod +x ssm auth_aws && ./ssm check && echo "export PATH=\"\$PATH:$(pwd)\"" >> ~/.bashrc && source ~/.bashrc
```

For zsh users:
```bash
git clone https://github.com/ZSoftly/quickssm.git && cd quickssm && chmod +x ssm auth_aws && ./ssm check && echo "export PATH=\"\$PATH:$(pwd)\"" >> ~/.zshrc && source ~/.zshrc
```

After running the appropriate command for your shell, you can use the tools by simply typing `ssm` or `auth_aws` from anywhere.

## Installation Options

### Option 1: Local User Installation (Recommended)

For bash users:
```bash
git clone https://github.com/ZSoftly/quickssm.git
cd quickssm
chmod +x ssm auth_aws
./ssm check
./auth_aws check
echo "export PATH=\"\$PATH:$(pwd)\"" >> ~/.bashrc
source ~/.bashrc
```

For zsh users:
```bash
git clone https://github.com/ZSoftly/quickssm.git
cd quickssm
chmod +x ssm auth_aws
./ssm check
./auth_aws check
echo "export PATH=\"\$PATH:$(pwd)\"" >> ~/.zshrc
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
chmod +x ssm auth_aws
./ssm check
./auth_aws check
INSTALL_DIR="$(pwd)"
sudo ln -s "$INSTALL_DIR/ssm" /usr/local/bin/ssm
sudo ln -s "$INSTALL_DIR/auth_aws" /usr/local/bin/auth_aws
sudo ln -s "$INSTALL_DIR/src" /usr/local/bin/src
```

Not recommended because:
- Any user on the system could run the tool and potentially access AWS resources
- Doesn't align well with per-user AWS credential management
- Requires sudo privileges for updates and modifications
- Can lead to security and audit tracking complications
- Makes it harder to manage different AWS configurations for different users

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
auth_aws check       # Check dependencies
auth_aws help        # Show help information
```

Before using `auth_aws`, set up a `.env` file in the same directory with the following content:
```
SSO_START_URL="https://your-sso-url.awsapps.com/start"
SSO_REGION="your-region"
DEFAULT_PROFILE="your-default-profile"
```

You can create a template file by running `auth_aws` without a valid .env file.

#### Log in to AWS SSO
```bash
auth_aws             # Use default profile from .env
auth_aws myprofile   # Use a specific profile name
```

The tool will:
1. Check for valid cached credentials
2. Initiate AWS SSO login if needed
3. Show an interactive list of accounts
4. Show an interactive list of roles for the selected account
5. Configure your AWS profile with the selected account and role

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

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

Key areas for contribution:
- Adding support for new regions
- Improving documentation
- Adding new features
- Bug fixes

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Security

These tools require AWS credentials and access to your AWS resources. Always:
- Keep your AWS credentials secure
- Use appropriate IAM permissions
- Review security best practices in the [AWS Security Documentation](https://docs.aws.amazon.com/security/)