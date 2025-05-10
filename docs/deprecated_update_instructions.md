# Updating from Older Versions (quickssm)

This document provides instructions for updating if you were using a version of this toolset from before March 2025, when the repository was named "quickssm".

## Option 1: Clean Update (Recommended)

```bash
# Navigate to your installation directory
cd /path/to/old/quickssm

# Backup your .env file if you have one
cp .env .env.backup

# Clone the new repository
cd ..
git clone https://github.com/zsoftly/ztiaws.git

# Copy your .env file if needed
cp /path/to/old/quickssm/.env.backup ztiaws/.env

# Update your path in your shell config file
# (Replace the old path with the new one)
```

## Option 2: In-place Migration

```bash
# Navigate to your installation directory
cd /path/to/old/quickssm

# Update your remote URL
git remote set-url origin https://github.com/zsoftly/ztiaws.git

# Pull the latest changes
git pull origin main

# Make the scripts executable
chmod +x ssm authaws
```
