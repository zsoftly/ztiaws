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
```
