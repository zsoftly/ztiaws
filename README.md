# quickssm

![Ubuntu](https://github.com/{your-actual-username}/quickssm/actions/workflows/test.yml/badge.svg)
![macOS](https://github.com/{your-actual-username}/quickssm/actions/workflows/test.yml/badge.svg)

A streamlined CLI tool for AWS SSM Session Manager, making it easier to list and connect to EC2 instances across regions.

## Features

- Quick connection to EC2 instances using short region codes
- Interactive listing of available instances
- Automatic validation of AWS CLI and Session Manager plugin
- Interactive Session Manager plugin installation
- Support for multiple AWS regions
- Color-coded output for better readability

## Prerequisites

- AWS CLI installed (follow [official installation guide](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html))
- AWS Session Manager plugin (can be installed interactively via `ssm check`)
- AWS credentials configured (`aws configure`)
- Bash or Zsh shell
- Proper IAM permissions for SSM Session Manager

## Quick Start

One-liner to download, install, and start using quickssm (for bash users):
```bash
git clone https://github.com/ZSoftly/quickssm.git && cd quickssm && chmod +x ssm && ./ssm check && echo 'export PATH="$PATH:$HOME/quickssm"' >> ~/.bashrc && source ~/.bashrc
```

For zsh users:
```bash
git clone https://github.com/ZSoftly/quickssm.git && cd quickssm && chmod +x ssm && ./ssm check && echo 'export PATH="$PATH:$HOME/quickssm"' >> ~/.zshrc && source ~/.zshrc
```

After running the appropriate command for your shell, you can use the tool by simply typing `ssm` from anywhere.

## Installation Options

### Option 1: Local User Installation (Recommended)

For bash users:
```bash
git clone https://github.com/ZSoftly/quickssm.git
cd quickssm
chmod +x ssm
./ssm check
echo 'export PATH="$PATH:$HOME/quickssm"' >> ~/.bashrc
source ~/.bashrc
```

For zsh users:
```bash
git clone https://github.com/ZSoftly/quickssm.git
cd quickssm
chmod +x ssm
./ssm check
echo 'export PATH="$PATH:$HOME/quickssm"' >> ~/.zshrc
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
chmod +x ssm
./ssm check
sudo ln -s "$(pwd)/ssm" /usr/local/bin/ssm
sudo ln -s "$(pwd)/src" /usr/local/bin/src
```

Not recommended because:
- Any user on the system could run the tool and potentially access AWS resources
- Doesn't align well with per-user AWS credential management
- Requires sudo privileges for updates and modifications
- Can lead to security and audit tracking complications
- Makes it harder to manage different AWS configurations for different users

## Usage

### Check System Requirements
```bash
ssm check
```

### List Instances in a Region
```bash
ssm cac1  # Lists instances in Canada Central
```

### Connect to an Instance
```bash
ssm i-1234abcd              # Connect to instance in default region (Canada Central)
ssm use1 i-1234abcd         # Connect to instance in US East
```

### Show Help
```bash
ssm help
```

## Supported Regions

| Shortcode | AWS Region    | Location     |
|-----------|---------------|--------------|
| cac1      | ca-central-1  | Montreal     |
| caw1      | ca-west-1     | Calgary      |
| use1      | us-east-1     | N. Virginia  |
| usw1      | us-west-1     | N. California|
| euw1      | eu-west-1     | Ireland      |

For a complete list of regions and their status, see [REGIONS.md](docs/REGIONS.md).

## IAM Permissions

Your AWS user/role needs the following permissions:
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

## Troubleshooting

### AWS CLI Not Found
If AWS CLI is not installed, follow the [official AWS CLI installation guide](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html).

### Session Manager Plugin Missing
Run `ssm check` to install the plugin interactively, or follow the [manual installation instructions](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html).

### AWS Credentials Not Configured
Run `aws configure` to set up your AWS credentials.

### Permission Errors
Ensure your AWS user/role has the required IAM permissions listed above.

### Shell Configuration
If the `ssm` command isn't available after installation, make sure you've added it to your PATH in the correct shell configuration file:
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

This tool requires AWS credentials and access to your instances. Always:
- Keep your AWS credentials secure
- Use appropriate IAM permissions
- Review security best practices in the [AWS Security Documentation](https://docs.aws.amazon.com/security/)