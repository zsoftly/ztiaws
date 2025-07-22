# QA Test Scenarios

**Prerequisites:** AWS CLI configured, EC2 instances in ca-central-1, SSM agent installed on instances.

**File Transfer Prerequisites (NEW):** 
- For large files (≥1MB): Target instances must have AWS CLI installed and proper IAM permissions for S3 access
- For small files (<1MB): No additional requirements beyond SSM
- S3 bucket will be auto-created in the same region as the target instance

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

## File Transfer (NEW)
```bash
# Small file upload (< 1MB - uses direct SSM transfer)
echo "Hello World" > test_small.txt
./ssm upload cac1 i-1234567890abcdef0 test_small.txt /tmp/test_small.txt
./ssm upload cac1 web-server test_small.txt /tmp/test_small_name.txt

# Small file download (< 1MB - uses direct SSM transfer)
./ssm download cac1 i-1234567890abcdef0 /tmp/test_small.txt ./downloaded_small.txt
./ssm download cac1 web-server /tmp/test_small_name.txt ./downloaded_small_name.txt

# Large file upload (≥ 1MB - uses S3 intermediary)
dd if=/dev/zero of=test_large.txt bs=1M count=2  # Creates 2MB file
./ssm upload cac1 i-1234567890abcdef0 test_large.txt /tmp/test_large.txt
./ssm upload cac1 web-server test_large.txt /tmp/test_large_name.txt

# Large file download (≥ 1MB - uses S3 intermediary)
./ssm download cac1 i-1234567890abcdef0 /tmp/test_large.txt ./downloaded_large.txt
./ssm download cac1 web-server /tmp/test_large_name.txt ./downloaded_large_name.txt

# File transfer to nested directories
./ssm upload cac1 i-1234567890abcdef0 test_small.txt /opt/app/config/settings.txt
./ssm download cac1 i-1234567890abcdef0 /var/log/application.log ./logs/app.log

# File transfer error cases
./ssm upload cac1 i-1234567890abcdef0 nonexistent.txt /tmp/test.txt        # File not found
./ssm upload cac1 fake-name test_small.txt /tmp/test.txt                   # Instance not found
./ssm download cac1 i-1234567890abcdef0 /nonexistent/file.txt ./test.txt   # Remote file not found
./ssm download cac1 fake-name /tmp/test.txt ./test.txt                     # Instance not found

# File transfer parameter validation
./ssm upload cac1                                      # Missing parameters
./ssm upload cac1 i-1234567890abcdef0                  # Missing file paths
./ssm download cac1                                    # Missing parameters
./ssm download cac1 i-1234567890abcdef0                # Missing file paths
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
- [ ] Small file upload works (< 1MB, direct SSM) (NEW)
- [ ] Small file download works (< 1MB, direct SSM) (NEW)
- [ ] Large file upload works (≥ 1MB, S3 intermediary) (NEW)
- [ ] Large file download works (≥ 1MB, S3 intermediary) (NEW)
- [ ] File transfer with instance names works (NEW)
- [ ] File transfer to nested directories works (NEW)
- [ ] File transfer error handling works (NEW)
- [ ] S3 bucket auto-creation works (NEW)
- [ ] S3 file auto-cleanup works (24hr lifecycle) (NEW)
- [ ] Error handling is clear
- [ ] No regressions

## EC2 Test Instance Management (NEW)
For comprehensive testing, use the EC2 manager script to create and destroy test instances:

```bash
# Create test instances for QA testing
./tools/01_ec2_test_manager.sh create --count 2 --owner QA-Team --name-prefix test-instance

# List created instances to get IDs and names
./tools/01_ec2_test_manager.sh verify

# Run QA tests with the created instances
# (Replace instance IDs/names in test commands above with actual values)

# Clean up test instances after testing
./tools/01_ec2_test_manager.sh delete

# Create instances with custom settings
./tools/01_ec2_test_manager.sh create --count 1 --name-prefix qa-server --owner QA-Team
```

## Advanced File Transfer Testing (NEW)
Additional comprehensive tests for the file transfer functionality:

```bash
# Test file transfer with various file types and sizes
# Create test files of different sizes
echo "Small test content" > test_tiny.txt                               # Very small file
dd if=/dev/zero of=test_medium.txt bs=512K count=1                      # 512KB file  
dd if=/dev/zero of=test_boundary.txt bs=1M count=1                      # Exactly 1MB (boundary test)
dd if=/dev/zero of=test_large_2mb.txt bs=1M count=2                     # 2MB file
dd if=/dev/zero of=test_large_10mb.txt bs=1M count=10                   # 10MB file

# Test boundary conditions (files around 1MB threshold)
./ssm upload cac1 web-server test_medium.txt /tmp/test_medium.txt       # Should use SSM direct
./ssm upload cac1 web-server test_boundary.txt /tmp/test_boundary.txt   # Should use S3
./ssm upload cac1 web-server test_large_2mb.txt /tmp/test_large_2mb.txt # Should use S3

# Test download of various sizes
./ssm download cac1 web-server /tmp/test_medium.txt ./downloaded_medium.txt
./ssm download cac1 web-server /tmp/test_boundary.txt ./downloaded_boundary.txt
./ssm download cac1 web-server /tmp/test_large_2mb.txt ./downloaded_large_2mb.txt

# Test concurrent file transfers (if needed)
./ssm upload cac1 test-instance-1 test_small.txt /tmp/concurrent1.txt &
./ssm upload cac1 test-instance-2 test_small.txt /tmp/concurrent2.txt &
wait

# Test file transfers with special characters in paths
./ssm upload cac1 web-server test_small.txt "/tmp/path with spaces/test file.txt"
./ssm download cac1 web-server "/tmp/path with spaces/test file.txt" "./downloaded with spaces.txt"

# Test permission edge cases
./ssm exec cac1 web-server 'sudo mkdir -p /opt/restricted && sudo chmod 755 /opt/restricted'
./ssm upload cac1 web-server test_small.txt /opt/restricted/test.txt    # Should work
./ssm exec cac1 web-server 'sudo chmod 000 /opt/restricted'
./ssm upload cac1 web-server test_small.txt /opt/restricted/fail.txt    # Should fail gracefully

# Test S3 error scenarios (simulate network issues)
# Note: These tests require manual intervention to simulate network issues
# ./ssm upload cac1 web-server test_large_10mb.txt /tmp/test_large.txt   # Disconnect network during transfer
```

## Performance and Reliability Testing (NEW)
```bash
# Test large file transfer reliability
./ssm upload cac1 web-server test_large_10mb.txt /tmp/perf_test.txt
./ssm exec cac1 web-server 'md5sum /tmp/perf_test.txt'                  # Verify integrity
md5sum test_large_10mb.txt                                              # Compare checksums

# Test file transfer with very long paths
./ssm exec cac1 web-server 'mkdir -p /tmp/very/deep/nested/directory/structure/for/testing/purposes'
./ssm upload cac1 web-server test_small.txt /tmp/very/deep/nested/directory/structure/for/testing/purposes/deep_file.txt

# Test multiple rapid transfers
for i in {1..5}; do
    echo "Test file $i" > "test_rapid_$i.txt"
    ./ssm upload cac1 web-server "test_rapid_$i.txt" "/tmp/rapid_$i.txt"
done

# Test S3 bucket lifecycle and cleanup (wait 25+ hours to verify automatic cleanup)
# Note: This is a long-running test - document S3 objects created during testing
./ssm upload cac1 web-server test_large_2mb.txt /tmp/lifecycle_test.txt
# Check S3 console after 25 hours to verify automatic cleanup
```

## New Features (Developers add here)
When adding new features, developers must:
1. Add test commands below showing the new functionality
2. Test that existing core features (connect, exec, exec-tagged) still work with the new feature
3. Run full end-to-end testing to ensure no regressions

```bash
# File Transfer Feature - Added in feature/ssm-file-transfer branch
# Test both small and large file transfers to ensure no regressions
./ssm upload cac1 web-server test_small.txt /tmp/regression_test.txt
./ssm exec cac1 web-server 'ls -la /tmp/regression_test.txt'
./ssm download cac1 web-server /tmp/regression_test.txt ./regression_downloaded.txt

# Add new test commands when implementing features
# Example: ./ssm new-command cac1 web-server 'test'
```

## Test Cleanup (NEW)
After running file transfer tests, clean up test files:

```bash
# Remove local test files
rm -f test_small.txt test_large.txt downloaded_*.txt regression_downloaded.txt
rm -f test_tiny.txt test_medium.txt test_boundary.txt test_large_2mb.txt test_large_10mb.txt
rm -f test_rapid_*.txt "downloaded with spaces.txt"

# Clean up remote test files (optional - they don't consume much space)
./ssm exec cac1 i-1234567890abcdef0 'rm -f /tmp/test_*.txt /tmp/regression_test.txt /tmp/rapid_*.txt /tmp/perf_test.txt /tmp/lifecycle_test.txt'
./ssm exec cac1 web-server 'rm -f /tmp/test_*.txt /tmp/regression_test.txt /tmp/rapid_*.txt /tmp/perf_test.txt /tmp/lifecycle_test.txt'
./ssm exec cac1 web-server 'rm -rf "/tmp/path with spaces" /tmp/very/deep/nested/directory /opt/restricted'

# Clean up test instances (if using ec2 test manager script)
./tools/01_ec2_test_manager.sh delete

# Note: S3 files are automatically cleaned up after 24 hours via lifecycle policy
# S3 bucket remains for future use to avoid recreation costs
```
