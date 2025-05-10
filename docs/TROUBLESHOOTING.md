# ❓ Troubleshooting

This document provides solutions to common issues encountered while using ZTiAWS.

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
Ensure your AWS user/role has the required IAM permissions. Refer to [IAM_PERMISSIONS.md](IAM_PERMISSIONS.md) for details.

### Shell Configuration
If the commands aren't available after installation, make sure you've added them to your PATH in the correct shell configuration file:
- For Bash users: `~/.bashrc`
- For Zsh users: `~/.zshrc`

You may need to restart your terminal or run `source ~/.bashrc` (or `source ~/.zshrc` for Zsh) for the changes to take effect.

### Script Name Changed
If you're getting "command not found" for `auth_aws`, note that the script has been renamed to `authaws` in v1.4.0+. Update your scripts and aliases accordingly.
