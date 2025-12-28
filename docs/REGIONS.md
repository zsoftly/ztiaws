# AWS Regions Support

This document lists all supported AWS regions in quickssm.

## Currently Supported Regions

| Shortcode | AWS Region   | Location      | Status |
| --------- | ------------ | ------------- | ------ |
| cac1      | ca-central-1 | Montreal      | Active |
| caw1      | ca-west-1    | Calgary       | Active |
| use1      | us-east-1    | N. Virginia   | Active |
| usw1      | us-west-1    | N. California | Active |
| euw1      | eu-west-1    | Ireland       | Active |

## Adding New Regions

Please see [CONTRIBUTING.md](../CONTRIBUTING.md) for instructions on how to add support for new regions.

## Region Status Definitions

- **Active**: Region is fully supported and tested
- **Beta**: Region support is new or under testing
- **Deprecated**: Region support will be removed in future versions

## Region Naming Convention

Shortcodes follow this pattern:

- First 3 characters: Geographic area (e.g., 'use' for US East)
- Last character: Region number

Examples:

- `use1` = US East 1
- `cac1` = Canada Central 1
- `euw1` = EU West 1
