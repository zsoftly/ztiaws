# AuthAWS Flag-Based Parameters QA Test Suite

## Overview

This document outlines the comprehensive test suite for the new flag-based parameter support in the `authaws` tool. The feature adds support for both positional and flag-based syntax while maintaining full backward compatibility.

**Feature**: Support both positional and flag-based parameters for improved UX and enterprise adoption  
**Module**: `src/06_authaws_parameter_parser.sh`  
**Version**: 1.0.0  
**Status**: Ready for Testing

## Test Environment Requirements

### Prerequisites
- AWS CLI v2.x installed and configured
- `jq` and `fzf` dependencies installed
- Valid AWS SSO configuration in `.env` file
- Test AWS profiles available
- Access to AWS SSO portal

### Test Data Setup
```bash
# Create test .env file
# ⚠️  WARNING: Replace with your actual SSO configuration values
# Do not commit real credentials to version control
cat > .env << 'EOL'
SSO_START_URL="https://your-organization.awsapps.com/start"
SSO_REGION="us-east-1"
DEFAULT_PROFILE="test-default-profile"
EOL

# Create test profiles
aws configure set profile.test-profile-1.sso_start_url "https://d-xxxxxxxxxx.awsapps.com/start"
aws configure set profile.test-profile-1.sso_region "us-east-1"
aws configure set profile.test-profile-2.sso_start_url "https://d-xxxxxxxxxx.awsapps.com/start"
aws configure set profile.test-profile-2.sso_region "us-west-2"
```

## Test Categories

### 1. Backward Compatibility Tests

#### Test 1.1: Positional Parameter Support
**Objective**: Verify existing positional syntax continues to work

**Test Cases**:
```bash
# Test 1.1.1: No arguments (default profile)
authaws
Expected: Uses default profile from .env

# Test 1.1.2: Profile name as first argument
authaws test-profile-1
Expected: Uses test-profile-1

# Test 1.1.3: Check command
authaws check
Expected: Runs system requirements check

# Test 1.1.4: Creds command with profile
authaws creds test-profile-1
Expected: Shows credentials for test-profile-1

# Test 1.1.5: Creds command without profile
authaws creds
Expected: Uses current AWS_PROFILE or default
```

**Pass Criteria**: All existing positional commands work exactly as before

#### Test 1.2: Existing Functionality Preservation
**Objective**: Ensure all existing features work with new parser

**Test Cases**:
```bash
# Test 1.2.1: Help command
authaws help
Expected: Shows help with both syntaxes

# Test 1.2.2: Version command
authaws version
Expected: Shows version information

# Test 1.2.3: Check command
authaws check
Expected: Validates system requirements
```

### 2. Flag-Based Parameter Tests

#### Test 2.1: Basic Flag Support
**Objective**: Verify new flag-based syntax works correctly

**Test Cases**:
```bash
# Test 2.1.1: Profile flag
authaws --profile test-profile-1
Expected: Uses test-profile-1

# Test 2.1.2: Short profile flag
authaws -p test-profile-1
Expected: Uses test-profile-1

# Test 2.1.3: Help flag
authaws --help
Expected: Shows help

# Test 2.1.4: Short help flag
authaws -h
Expected: Shows help

# Test 2.1.5: Version flag
authaws --version
Expected: Shows version

# Test 2.1.6: Short version flag
authaws -v
Expected: Shows version
```

#### Test 2.2: Command Flags
**Objective**: Test command-specific flags

**Test Cases**:
```bash
# Test 2.2.1: Check flag
authaws --check
Expected: Runs system requirements check

# Test 2.2.2: Creds flag
authaws --creds
Expected: Shows credentials for current/default profile

# Test 2.2.3: Creds with profile flag
authaws --creds --profile test-profile-1
Expected: Shows credentials for test-profile-1

# Test 2.2.4: List profiles flag
authaws --list-profiles
Expected: Lists available AWS profiles
```

#### Test 2.3: Advanced Flags
**Objective**: Test new advanced functionality

**Test Cases**:
```bash
# Test 2.3.1: Region override
authaws --profile test-profile-1 --region us-west-2
Expected: Uses us-west-2 region instead of default

# Test 2.3.2: SSO URL override
authaws --profile test-profile-1 --sso-url https://alt.awsapps.com/start
Expected: Uses custom SSO URL

# Test 2.3.3: Export flag
authaws --creds --profile test-profile-1 --export
Expected: Outputs credentials in export format only

# Test 2.3.4: Debug flag
authaws --profile test-profile-1 --debug
Expected: Shows debug information during execution
```

### 3. Mixed Syntax Tests

#### Test 3.1: Positional + Flag Combinations
**Objective**: Test mixed syntax scenarios

**Test Cases**:
```bash
# Test 3.1.1: Positional profile + flag
authaws test-profile-1 --region us-west-2
Expected: Uses test-profile-1 with us-west-2 region

# Test 3.1.2: Positional profile + multiple flags
authaws test-profile-1 --region us-west-2 --export
Expected: Uses test-profile-1 with region override and export mode

# Test 3.1.3: Flag profile + positional (should fail)
authaws --profile test-profile-1 test-profile-2
Expected: Error - unexpected positional argument
```

### 4. Error Handling Tests

#### Test 4.1: Invalid Flags
**Objective**: Test error handling for invalid parameters

**Test Cases**:
```bash
# Test 4.1.1: Unknown flag
authaws --unknown-flag
Expected: Error message about unknown flag

# Test 4.1.2: Missing flag value
authaws --profile
Expected: Error message about missing value

# Test 4.1.3: Invalid region
authaws --profile test-profile-1 --region invalid-region
Expected: Error message about invalid region

# Test 4.1.4: Invalid SSO URL
authaws --profile test-profile-1 --sso-url invalid-url
Expected: Error message about invalid SSO URL
```

#### Test 4.2: Conflicting Commands
**Objective**: Test validation of mutually exclusive commands

**Test Cases**:
```bash
# Test 4.2.1: Multiple commands
authaws --check --creds
Expected: Error message about multiple commands

# Test 4.2.2: Command + positional
authaws check test-profile-1
Expected: Error message about unexpected arguments
```

### 5. Edge Case Tests

#### Test 5.1: Boundary Conditions
**Objective**: Test edge cases and boundary conditions

**Test Cases**:
```bash
# Test 5.1.1: Empty profile name
authaws --profile ""
Expected: Error or uses default profile

# Test 5.1.2: Very long profile name
authaws --profile "$(printf 'a%.0s' {1..1000})"
Expected: Handles long profile names gracefully

# Test 5.1.3: Special characters in profile name
authaws --profile "test-profile-with-special-chars_123"
Expected: Handles special characters correctly

# Test 5.1.4: Unicode characters
authaws --profile "test-プロファイル"
Expected: Handles Unicode characters correctly
```

#### Test 5.2: Environment Variables
**Objective**: Test interaction with environment variables

**Test Cases**:
```bash
# Test 5.2.1: AWS_PROFILE environment variable
export AWS_PROFILE=env-profile
authaws --creds
Expected: Uses env-profile if no profile specified

# Test 5.2.2: AWS_PROFILE override
export AWS_PROFILE=env-profile
authaws --profile flag-profile --creds
Expected: Uses flag-profile, ignores AWS_PROFILE
```

### 6. Integration Tests

#### Test 6.1: Full Authentication Flow
**Objective**: Test complete authentication workflow with new syntax

**Test Cases**:
```bash
# Test 6.1.1: Full login with flags
authaws --profile test-profile-1 --region us-west-2
Expected: Complete SSO authentication flow

# Test 6.1.2: Full login with mixed syntax
authaws test-profile-1 --region us-west-2
Expected: Complete SSO authentication flow

# Test 6.1.3: Credential export workflow
authaws --profile test-profile-1 --export
Expected: Authenticates and exports credentials
```

#### Test 6.2: Profile Management
**Objective**: Test profile listing and management

**Test Cases**:
```bash
# Test 6.2.1: List profiles
authaws --list-profiles
Expected: Shows all configured AWS profiles

# Test 6.2.2: Profile creation with flags
authaws --profile new-profile --region us-east-1
Expected: Creates new profile with specified settings
```

## Performance Tests

### Test 7.1: Parser Performance
**Objective**: Ensure parameter parsing doesn't impact performance

**Test Cases**:
```bash
# Test 7.1.1: Parse time measurement
time authaws --help
Expected: Parsing completes in <100ms

# Test 7.1.2: Large argument list
authaws --profile test --region us-east-1 --sso-url https://test.awsapps.com/start --export --debug
Expected: Handles multiple flags efficiently
```

## Regression Tests

### Test 8.1: Existing Scripts
**Objective**: Ensure existing automation scripts continue to work

**Test Cases**:
```bash
# Test 8.1.1: CI/CD script compatibility
./ci-script.sh  # Uses authaws with positional parameters
Expected: Script runs without modification

# Test 8.1.2: Existing aliases
alias auth='authaws dev-profile'
auth
Expected: Alias works unchanged
```

## Test Execution Checklist

### Pre-Test Setup
- [ ] AWS CLI configured with test credentials
- [ ] Test .env file created with valid SSO configuration
- [ ] Test profiles created in AWS config
- [ ] Dependencies (jq, fzf) installed
- [ ] Test environment isolated from production

### Test Execution
- [ ] Run all backward compatibility tests
- [ ] Run all flag-based parameter tests
- [ ] Run all mixed syntax tests
- [ ] Run all error handling tests
- [ ] Run all edge case tests
- [ ] Run all integration tests
- [ ] Run performance tests
- [ ] Run regression tests

### Post-Test Validation
- [ ] All tests pass
- [ ] No performance degradation
- [ ] Error messages are clear and helpful
- [ ] Help documentation is accurate
- [ ] Backward compatibility maintained

## Expected Results

### Success Criteria
1. **100% Backward Compatibility**: All existing commands work unchanged
2. **Flag Support**: All new flag-based syntax works correctly
3. **Mixed Syntax**: Positional + flag combinations work as expected
4. **Error Handling**: Clear, helpful error messages for invalid inputs
5. **Performance**: No measurable performance impact
6. **Documentation**: Help text accurately reflects both syntaxes

### Failure Criteria
1. Any existing command fails or behaves differently
2. Flag-based syntax doesn't work as documented
3. Error messages are unclear or unhelpful
4. Performance degradation >10%
5. Help documentation is inaccurate

## Test Data Cleanup

After testing, clean up test data:
```bash
# Remove test profiles
aws configure set profile.test-profile-1.sso_start_url ""
aws configure set profile.test-profile-2.sso_start_url ""

# Remove test .env file
rm -f .env

# Clear any test credentials
aws sso logout --profile test-profile-1
aws sso logout --profile test-profile-2
```

## Reporting

### Test Report Template
```
AuthAWS Flag-Based Parameters Test Report
========================================

Test Date: [DATE]
Tester: [NAME]
Environment: [OS/Version]

Summary:
- Total Tests: [NUMBER]
- Passed: [NUMBER]
- Failed: [NUMBER]
- Skipped: [NUMBER]

Key Findings:
- [List any issues found]
- [Performance observations]
- [User experience notes]

Recommendations:
- [Any changes needed]
- [Additional testing required]

Status: [PASS/FAIL/NEEDS_REVISION]
```

## Maintenance

### Ongoing Testing
- Run regression tests after any authaws changes
- Test new flag combinations as they're added
- Monitor performance impact over time
- Update test cases for new features

### Test Case Updates
- Add test cases for new flags
- Update expected results for changed behavior
- Remove obsolete test cases
- Add performance benchmarks for new features
