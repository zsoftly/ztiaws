# QA AuthAWS Test Scenarios

**Prerequisites:** AWS CLI installed, fzf installed, jq installed, valid AWS SSO setup.

**Setup:** Navigate to your local ztiaws directory, switch to the feature branch, use `./authaws` for all commands below.

Replace `your-profile-name` with actual profile names from your environment.

**Testing different profiles:** Use different profile names. **Testing different accounts:** Select different accounts/roles during interactive setup.

## Basic Commands
```bash
./authaws help                    # Shows help
./authaws version                 # Shows version
./authaws check                   # Verifies requirements
```

## Profile Authentication
```bash
# Default profile (from .env file)
./authaws                         # Uses default profile from .env

# Specific profile
./authaws dev-profile             # Authenticates with specific profile
./authaws prod-profile            # Authenticates with different profile
./authaws my-test-profile         # Creates new profile if needed
```

## Credential Management
```bash
# Show credentials for profiles
./authaws creds                   # Shows creds for current/default profile
./authaws creds dev-profile       # Shows creds for specific profile
./authaws creds nonexistent       # Profile not found error
```

## Interactive Flows
```bash
# These require interactive selection via fzf
./authaws new-profile             # Should show account selection
# Select account -> Select role -> Profile configured

./authaws                         # With expired session
# Should prompt for re-authentication
```

## Configuration Setup
```bash
# First time setup (when no .env exists)
./authaws                         # Should offer to create .env file
# Answer 'y' -> Creates sample .env -> Shows edit instructions
```

## Error Cases
```bash
./authaws invalid-command         # Shows help
./authaws creds                   # Without valid session - should show instructions
```

## Environment Testing
```bash
# Test with missing dependencies (temporarily rename)
# mv $(which fzf) $(which fzf).bak
# ./authaws check                 # Should show missing fzf error
# mv $(which fzf).bak $(which fzf)

# Test with no .env file
# mv .env .env.bak
# ./authaws                       # Should offer to create .env
# mv .env.bak .env
```

## QA Checklist
- [ ] Basic commands work
- [ ] Authentication flows work
- [ ] Credential display works
- [ ] Interactive account/role selection works
- [ ] Profile creation works
- [ ] Configuration setup works
- [ ] Error handling is clear
- [ ] Dependencies are checked
- [ ] No regressions

## New Features (Developers add here)
When adding new features, developers must:
1. Add test commands below showing the new functionality
2. Test that existing core features (auth, creds, profiles) still work with the new feature
3. Run full end-to-end testing to ensure no regressions

```bash
# Add new test commands when implementing features
# Example: ./authaws new-command profile-name

# Test all help variations
./authaws help                    # Original syntax
./authaws --help                  # New flag syntax
./authaws -h                      # Short flag syntax

# Test all version variations
./authaws version                 # Original syntax
./authaws --version               # New flag syntax
./authaws -v                      # Short flag syntax

# Test check variations
./authaws check                   # Original syntax
./authaws --check                 # New flag syntax

# Test profile authentication with both syntaxes
./authaws dev-profile             # Original positional syntax
./authaws --profile dev-profile   # New flag syntax
./authaws -p dev-profile          # Short flag syntax

# Test default profile handling
./authaws                         # Should use default from .env
./authaws --profile              # Should show error (missing profile name)
./authaws -p                     # Should show error (missing profile name)

# Test credential display variations
./authaws creds                   # Original syntax (default profile)
./authaws creds dev-profile       # Original syntax (specific profile)
./authaws --creds --profile dev-profile  # New flag syntax
./authaws --creds                 # Should use default/current profile
./authaws --creds --profile       # Should show error (missing profile)

# Test region override functionality
./authaws --profile dev-profile --region us-west-2    # Override region
./authaws -p dev-profile -r eu-west-1                 # Short flags
./authaws --region us-east-1 --profile test-profile   # Flag order shouldn't matter
./authaws --region                                     # Should show error (missing region)
./authaws -r                                          # Should show error (missing region)

# Test invalid flag usage
./authaws --unknown-flag          # Should show error and suggest help
./authaws --profile               # Missing profile name
./authaws --region                # Missing region name
./authaws --check --profile test  # Profile ignored with warning
./authaws --help --profile test   # Profile ignored with warning
./authaws --version --profile test # Profile ignored with warning

# Test conflicting commands
./authaws check version         # Should show error about multiple commands
./authaws help creds              # Should show error about multiple commands
./authaws check creds         # Should show error about multiple commands

# Test that old and new syntaxes produce identical results
./authaws dev-profile             # Old syntax
./authaws --profile dev-profile   # New syntax
# Both should produce identical profile configuration

./authaws creds dev-profile       # Old syntax
./authaws --creds --profile dev-profile  # New syntax
# Both should show identical credential output

# Test interaction with AWS_PROFILE environment variable
export AWS_PROFILE=test-profile
./authaws creds                   # Should use test-profile
./authaws --creds                 # Should use test-profile
unset AWS_PROFILE

# Test with no AWS_PROFILE and no default in .env
./authaws creds                   # Should show appropriate error
./authaws --creds                 # Should show appropriate error

# Test parameter parsing with various combinations
./authaws --profile test --region us-east-1 --export
./authaws -p test -r us-west-2
./authaws --creds --profile test
./authaws --check
./authaws help
./authaws --help
./authaws version
./authaws -v

# Test that AWS CLI integration still works with flag syntax
./authaws --profile test-profile
export AWS_PROFILE=test-profile
aws sts get-caller-identity       # Should work with the authenticated profile
```
