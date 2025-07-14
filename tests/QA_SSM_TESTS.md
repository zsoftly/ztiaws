# QA Test Scenarios

**Prerequisites:** AWS CLI configured, EC2 instances in ca-central-1, SSM agent installed on instances.

**Setup:** Navigate to your local ztiaws directory, switch to the feature branch, use `./ssm` for all commands below.

Replace `i-1234567890abcdef0` with real instance ID and `web-server` with real instance name.

**Testing other regions:** Replace `cac1` with `use1`, `usw2`, etc. **Testing multiple instances:** Use different instance names/IDs in same region.

## Basic Commands
```bash
./ssm help                     # Shows help
./ssm version                  # Shows version
./ssm check                    # Verifies requirements
```

## Instance Listing
```bash
./ssm cac1                     # Lists instances
./ssm xyz1                     # Invalid region error
```

## Connect to Instance
```bash
# Using Instance ID
./ssm cac1 i-1234567890abcdef0 # Connects via SSM
./ssm cac1 i-123               # Invalid ID error

# Using Instance Name (NEW)
./ssm cac1 web-server          # Resolves name and connects
./ssm cac1 fake-name           # Instance not found error
```

## Execute Commands
```bash
# Using Instance ID
./ssm exec cac1 i-1234567890abcdef0 'hostname'  # Runs command
./ssm exec cac1 i-123 'ls'                      # Invalid ID error

# Using Instance Name (NEW)
./ssm exec cac1 web-server 'hostname'           # Resolves and runs command
./ssm exec cac1 fake-name 'ls'                  # Instance not found
```

## Tagged Execution
```bash
./ssm exec-tagged cac1 Environment prod 'hostname'  # Runs on tagged instances
./ssm exec-tagged cac1 Role                         # Missing parameters
```

## Error Cases
```bash
./ssm invalid-command                    # Shows help
./ssm exec cac1                          # Missing parameters
./ssm exec cac1 web-server               # Missing command
```

## QA Checklist
- [ ] Basic commands work
- [ ] Instance listing works
- [ ] Connect with Instance ID works
- [ ] Connect with instance name works (NEW)
- [ ] Exec with Instance ID works  
- [ ] Exec with instance name works (NEW)
- [ ] Tagged execution works
- [ ] Error handling is clear
- [ ] No regressions

## New Features (Developers add here)
When adding new features, developers must:
1. Add test commands below showing the new functionality
2. Test that existing core features (connect, exec, exec-tagged) still work with the new feature
3. Run full end-to-end testing to ensure no regressions

```bash
# Add new test commands when implementing features
# Example: ./ssm new-command cac1 web-server 'test'
```
