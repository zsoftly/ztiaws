# 🔧 Installation Options

This document provides detailed installation instructions for ZTiAWS.

## Option 1: Local User Installation (Recommended)

**Bash users:**
```bash
git clone https://github.com/zsoftly/ztiaws.git
cd ztiaws
chmod +x ssm authaws
./ssm check
./authaws check
echo -e "\n# Add ZTiAWS to PATH\nexport PATH=\"\$PATH:$(pwd)\"" >> ~/.bashrc
source ~/.bashrc
```

**Zsh users:**
```bash
git clone https://github.com/zsoftly/ztiaws.git
cd ztiaws
chmod +x ssm authaws
./ssm check
./authaws check
echo -e "\n# Add ZTiAWS to PATH\nexport PATH=\"\$PATH:$(pwd)\"" >> ~/.zshrc
source ~/.zshrc
```

**PowerShell users:**
```powershell
git clone https://github.com/zsoftly/ztiaws.git
cd ztiaws
# Copy the profile to your PowerShell profile directory
Copy-Item .\Microsoft.PowerShell_profile.ps1 $PROFILE
# Reload your profile
. $PROFILE
```

This is the recommended approach because:
- Keeps AWS tooling scoped to your user
- Maintains better security practices
- Makes updates easier without requiring sudo/admin privileges
- Aligns with AWS credentials being stored per-user
- Follows principle of least privilege
- Easier to manage different AWS configurations per user

## Option 2: System-wide Installation (Not Recommended)
```bash
git clone https://github.com/zsoftly/ztiaws.git
cd ztiaws
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
