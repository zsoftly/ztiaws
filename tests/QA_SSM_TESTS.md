# SSM Flag-Based Parameters QA Test Suite

## Overview

This document outlines the comprehensive test suite for the new flag-based parameter support in the `ssm` tool. The feature adds support for both positional and flag-based syntax while maintaining full backward compatibility.

**Feature**: Support both positional and flag-based parameters for improved UX and scalability  
**Module**: `src/07_ssm_parameter_parser.sh`  
**Version**: 1.0.0  
**Status**: Ready for Testing

## Command Usage Context

**Development Testing** (used in this QA suite):
```bash
./ssm --help    # Tests local development version
```

**Production Usage** (after `make install`):
```bash
ssm --help      # Uses globally installed version
```

This QA document uses `./ssm` to test the local development version before installation.

## Test Environment Requirements

### Prerequisites
- AWS CLI v2.x installed and configured
- AWS Session Manager plugin installed
- `jq` dependency installed (for some operations)
- Valid AWS credentials configured
- Test EC2 instances with SSM agent enabled
- Proper IAM permissions for SSM Session Manager

### Test Data Setup
```bash
# Ensure test instances are available
# ⚠️  WARNING: Replace placeholder values with your actual AWS resources
# Do not commit real instance IDs to version control
export TEST_REGION="<YOUR_AWS_REGION>"              # e.g., "us-east-1", "ca-central-1"
export TEST_INSTANCE_1="<YOUR_INSTANCE_ID_1>"       # e.g., "i-1234567890abcdef0"
export TEST_INSTANCE_2="<YOUR_INSTANCE_ID_2>"       # e.g., "i-0987654321fedcba0"  
export TEST_TAG_KEY="<YOUR_TAG_KEY>"                # e.g., "Environment"
export TEST_TAG_VALUE="<YOUR_TAG_VALUE>"            # e.g., "test"

# Create test files for upload/download testing
echo "Test content for upload" > test-upload.txt
mkdir -p test-downloads
```

## Test Categories

### 1. Backward Compatibility Tests

#### Test 1.1: Positional Parameter Support
**Objective**: Verify existing positional syntax continues to work

**Test Cases**:
```bash
# Test 1.1.1: No arguments (show help)
./ssm
Expected: Shows help message

# Test 1.1.2: Single region argument (list instances)
./ssm $TEST_REGION
Expected: Lists instances in test region

# Test 1.1.3: Instance only (connect with default region)
./ssm $TEST_INSTANCE_1
Expected: Connects to instance in default region (ca-central-1)

# Test 1.1.4: Region and instance (connect)
./ssm $TEST_REGION $TEST_INSTANCE_1
Expected: Connects to instance in specified region

# Test 1.1.5: System commands
./ssm check
./ssm version
./ssm help
Expected: All commands work as before

# Test 1.1.6: Exec command
./ssm exec $TEST_REGION $TEST_INSTANCE_1 "uptime"
Expected: Executes command on instance

# Test 1.1.7: Exec-tagged command
./ssm exec-tagged $TEST_REGION $TEST_TAG_KEY $TEST_TAG_VALUE "hostname"
Expected: Executes command on tagged instances

# Test 1.1.8: Upload command (if available)
./ssm upload $TEST_REGION $TEST_INSTANCE_1 test-upload.txt /tmp/test-upload.txt
Expected: Uploads file to instance

# Test 1.1.9: Download command (if available)
./ssm download $TEST_REGION $TEST_INSTANCE_1 /tmp/test-upload.txt ./test-downloads/downloaded.txt
Expected: Downloads file from instance
```

**Pass Criteria**: All existing positional commands work exactly as before

#### Test 1.2: Existing Functionality Preservation
**Objective**: Ensure all existing features work with new parser

**Test Cases**:
```bash
# Test 1.2.1: Help variations
./ssm help
./ssm -h
./ssm --help
Expected: All show help message

# Test 1.2.2: Version command
./ssm version
Expected: Shows version information

# Test 1.2.3: Check command
./ssm check
Expected: Validates system requirements

# Test 1.2.4: Install command
./ssm install
Expected: Shows installation instructions
```

### 2. Flag-Based Parameter Tests

#### Test 2.1: Basic Flag Support
**Objective**: Verify new flag-based syntax works correctly

**Test Cases**:
```bash
# Test 2.1.1: List with region flag
./ssm --region $TEST_REGION --list
Expected: Lists instances in test region

# Test 2.1.2: Connect with flags
./ssm --region $TEST_REGION --instance $TEST_INSTANCE_1 --connect
Expected: Connects to instance

# Test 2.1.3: Short flags
./ssm -r $TEST_REGION --list
./ssm --region $TEST_REGION -i $TEST_INSTANCE_1
Expected: Short flags work correctly

# Test 2.1.4: Help flags
./ssm --help
./ssm -h
Expected: Shows help

# Test 2.1.5: Version flags
./ssm --version
./ssm -v
Expected: Shows version

# Test 2.1.6: Check flag
./ssm --check
Expected: Runs system requirements check
```

#### Test 2.2: Operation Flags
**Objective**: Test operation-specific flags

**Test Cases**:
```bash
# Test 2.2.1: Exec operation
./ssm --exec --region $TEST_REGION --instance $TEST_INSTANCE_1 --command "uptime"
Expected: Executes command on instance

# Test 2.2.2: Exec-tagged operation
./ssm --exec-tagged --region $TEST_REGION --tag-key $TEST_TAG_KEY --tag-value $TEST_TAG_VALUE --command "hostname"
Expected: Executes command on tagged instances

# Test 2.2.3: Upload operation (if available)
./ssm --upload --region $TEST_REGION --instance $TEST_INSTANCE_1 --local-file test-upload.txt --remote-path /tmp/test-flag-upload.txt
Expected: Uploads file to instance

# Test 2.2.4: Download operation (if available)
./ssm --download --region $TEST_REGION --instance $TEST_INSTANCE_1 --remote-file /tmp/test-flag-upload.txt --local-path ./test-downloads/flag-downloaded.txt
Expected: Downloads file from instance

# Test 2.2.5: Debug flag
./ssm --debug --region $TEST_REGION --list
Expected: Shows debug information during execution
```

#### Test 2.3: Advanced Operations
**Objective**: Test advanced flag functionality

**Test Cases**:
```bash
# Test 2.3.1: Port forwarding (if supported)
./ssm --forward --region $TEST_REGION --instance $TEST_INSTANCE_1 --local-port 8080 --remote-port 80
Expected: Sets up port forwarding (cancel after verification)

# Test 2.3.2: Region code validation
./ssm --region invalid-region --list
Expected: Error message about invalid region

# Test 2.3.3: Complex command with multiple flags
./ssm --exec --region $TEST_REGION --instance $TEST_INSTANCE_1 --command "ps aux | grep ssh" --debug
Expected: Executes complex command with debug output
```

### 3. Mixed Syntax Tests

#### Test 3.1: Positional + Flag Combinations
**Objective**: Test mixed syntax scenarios

**Test Cases**:
```bash
# Test 3.1.1: Positional region + flag instance
./ssm $TEST_REGION --instance $TEST_INSTANCE_1
Expected: Connects to instance using mixed syntax

# Test 3.1.2: Flag region + positional instance
./ssm --region $TEST_REGION $TEST_INSTANCE_1
Expected: Connects to instance using mixed syntax

# Test 3.1.3: Exec with mixed syntax
./ssm exec $TEST_REGION --instance $TEST_INSTANCE_1 --command "date"
Expected: Executes command using mixed syntax

# Test 3.1.4: Upload with mixed syntax
./ssm upload $TEST_REGION --instance $TEST_INSTANCE_1 --local-file test-upload.txt --remote-path /tmp/mixed-upload.txt
Expected: Uploads file using mixed syntax

# Test 3.1.5: Invalid mixed syntax
./ssm --region $TEST_REGION $TEST_INSTANCE_1 $TEST_INSTANCE_2
Expected: Error - too many positional arguments
```

### 4. Error Handling Tests

#### Test 4.1: Invalid Flags
**Objective**: Test error handling for invalid parameters

**Test Cases**:
```bash
# Test 4.1.1: Unknown flag
./ssm --unknown-flag
Expected: Error message about unknown flag

# Test 4.1.2: Missing flag value
./ssm --region
Expected: Error message about missing value

# Test 4.1.3: Invalid region
./ssm --region invalid-region-code
Expected: Error message about invalid region

# Test 4.1.4: Missing required parameters for exec
./ssm --exec --region $TEST_REGION
Expected: Error message about missing instance

# Test 4.1.5: Missing command for exec
./ssm --exec --region $TEST_REGION --instance $TEST_INSTANCE_1
Expected: Error message about missing command
```

#### Test 4.2: Conflicting Operations
**Objective**: Test validation of mutually exclusive operations

**Test Cases**:
```bash
# Test 4.2.1: Multiple operations
./ssm --exec --upload --region $TEST_REGION --instance $TEST_INSTANCE_1
Expected: Should infer operation or handle gracefully

# Test 4.2.2: Incomplete parameter sets
./ssm --upload --region $TEST_REGION --instance $TEST_INSTANCE_1
Expected: Error about missing local-file and remote-path

# Test 4.2.3: Invalid instance identifier
./ssm --region $TEST_REGION --instance invalid-instance-id
Expected: Error message about invalid instance format
```

### 5. Edge Case Tests

#### Test 5.1: Boundary Conditions
**Objective**: Test edge cases and boundary conditions

**Test Cases**:
```bash
# Test 5.1.1: Empty values
./ssm --region "" --list
Expected: Error or uses default region

# Test 5.1.2: Very long command
./ssm --exec --region $TEST_REGION --instance $TEST_INSTANCE_1 --command "$(printf 'echo %.0s' {1..1000})"
Expected: Handles long commands gracefully

# Test 5.1.3: Special characters in command
./ssm --exec --region $TEST_REGION --instance $TEST_INSTANCE_1 --command "echo 'test with spaces and \"quotes\"'"
Expected: Handles special characters correctly

# Test 5.1.4: Unicode in file paths (if supported)
./ssm --upload --region $TEST_REGION --instance $TEST_INSTANCE_1 --local-file test-upload.txt --remote-path "/tmp/test-файл.txt"
Expected: Handles Unicode paths correctly
```

#### Test 5.2: Environment Variables
**Objective**: Test interaction with environment variables

**Test Cases**:
```bash
# Test 5.2.1: SSM_DEBUG environment variable
export SSM_DEBUG=true
./ssm --region $TEST_REGION --list
Expected: Shows debug information

# Test 5.2.2: AWS environment variables
export AWS_DEFAULT_REGION=$TEST_REGION
./ssm --instance $TEST_INSTANCE_1
Expected: Uses environment region if available

# Test 5.2.3: Clean up environment
unset SSM_DEBUG AWS_DEFAULT_REGION
```

### 6. Integration Tests

#### Test 6.1: Full Workflow Tests
**Objective**: Test complete workflows with new syntax

**Test Cases**:
```bash
# Test 6.1.1: Complete file transfer workflow
./ssm --upload --region $TEST_REGION --instance $TEST_INSTANCE_1 --local-file test-upload.txt --remote-path /tmp/integration-test.txt
./ssm --exec --region $TEST_REGION --instance $TEST_INSTANCE_1 --command "cat /tmp/integration-test.txt"
./ssm --download --region $TEST_REGION --instance $TEST_INSTANCE_1 --remote-file /tmp/integration-test.txt --local-path ./test-downloads/integration-downloaded.txt
Expected: Complete workflow works end-to-end

# Test 6.1.2: Command execution workflow
./ssm --exec --region $TEST_REGION --instance $TEST_INSTANCE_1 --command "echo 'test' > /tmp/command-test.txt"
./ssm --exec --region $TEST_REGION --instance $TEST_INSTANCE_1 --command "ls -la /tmp/command-test.txt"
Expected: Command sequence works correctly

# Test 6.1.3: Mixed syntax workflow
./ssm $TEST_REGION --instance $TEST_INSTANCE_1  # Connect and verify
./ssm exec $TEST_REGION --instance $TEST_INSTANCE_1 --command "uptime"  # Execute command
Expected: Mixed syntax works in workflow
```

#### Test 6.2: Instance Management
**Objective**: Test instance discovery and management

**Test Cases**:
```bash
# Test 6.2.1: List instances
./ssm --region $TEST_REGION --list
Expected: Shows list of available instances

# Test 6.2.2: Instance name resolution
./ssm --region $TEST_REGION --instance "test-instance-name"
Expected: Resolves instance name to ID (if instance has Name tag)

# Test 6.2.3: Tag-based operations
./ssm --exec-tagged --region $TEST_REGION --tag-key $TEST_TAG_KEY --tag-value $TEST_TAG_VALUE --command "uptime"
Expected: Executes on all instances matching tag
```

## Performance Tests

### Test 7.1: Parser Performance
**Objective**: Ensure parameter parsing doesn't impact performance

**Test Cases**:
```bash
# Test 7.1.1: Parse time measurement
time ./ssm --help
Expected: Parsing completes in <100ms

# Test 7.1.2: Large argument list
./ssm --exec --region $TEST_REGION --instance $TEST_INSTANCE_1 --command "echo test" --debug
Expected: Handles multiple flags efficiently

# Test 7.1.3: Repeated operations
for i in {1..5}; do time ./ssm --region $TEST_REGION --list > /dev/null; done
Expected: Consistent performance across multiple runs
```

## Regression Tests

### Test 8.1: Existing Scripts
**Objective**: Ensure existing automation scripts continue to work

**Test Cases**:
```bash
# Test 8.1.1: CI/CD script compatibility
# Create test script using old syntax
cat > test-old-syntax.sh << 'EOF'
#!/bin/bash
./ssm check
./ssm $TEST_REGION
./ssm exec $TEST_REGION $TEST_INSTANCE_1 "echo 'CI test'"
EOF
chmod +x test-old-syntax.sh
./test-old-syntax.sh
Expected: Script runs without modification

# Test 8.1.2: Existing aliases
alias quick-connect='./ssm $TEST_REGION $TEST_INSTANCE_1'
quick-connect
Expected: Alias works unchanged

# Test 8.1.3: Clean up
rm test-old-syntax.sh
unalias quick-connect
```

## Test Execution Checklist

### Pre-Test Setup
- [ ] AWS CLI configured with test credentials
- [ ] Test instances available and SSM-enabled
- [ ] Test region and instance IDs configured
- [ ] Test files created for upload/download tests
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
# Remove test files
rm -f test-upload.txt
rm -rf test-downloads/

# Clean up remote test files (if needed)
./ssm exec $TEST_REGION $TEST_INSTANCE_1 "rm -f /tmp/test-*.txt /tmp/integration-test.txt /tmp/command-test.txt /tmp/mixed-upload.txt /tmp/flag-upload.txt"

# Clear environment variables
unset TEST_REGION TEST_INSTANCE_1 TEST_INSTANCE_2 TEST_TAG_KEY TEST_TAG_VALUE
```

## Reporting

### Test Report Template
```
SSM Flag-Based Parameters Test Report
=====================================

Test Date: [DATE]
Tester: [NAME]
Environment: [OS/Version]
AWS Region: [REGION]
Test Instances: [INSTANCE_IDS]

Summary:
- Total Tests: [NUMBER]
- Passed: [NUMBER]
- Failed: [NUMBER]
- Skipped: [NUMBER]

Key Findings:
- [List any issues found]
- [Performance observations]
- [User experience notes]

Backward Compatibility:
- [Status of existing functionality]
- [Any breaking changes detected]

New Features:
- [Flag-based syntax validation]
- [Mixed syntax validation]
- [Error handling validation]

Recommendations:
- [Any changes needed]
- [Additional testing required]

Status: [PASS/FAIL/NEEDS_REVISION]
```

## Maintenance

### Ongoing Testing
- Run regression tests after any SSM changes
- Test new flag combinations as they're added
- Monitor performance impact over time
- Update test cases for new features

### Test Case Updates
- Add test cases for new flags
- Update expected results for changed behavior
- Remove obsolete test cases
- Add performance benchmarks for new features