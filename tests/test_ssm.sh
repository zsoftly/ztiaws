#!/usr/bin/env bash

# Exit on error
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_FAILED=0

# Directory setup
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MAIN_SCRIPT="$SCRIPT_DIR/../ssm"
MOCK_DIR="$SCRIPT_DIR/mocks"

# Setup mock AWS CLI
mkdir -p "$MOCK_DIR"
cat > "$MOCK_DIR/aws" << 'EOF'
#!/bin/bash
case "$*" in
    *"ec2 describe-instances"*)
        echo "InstanceID    Name    Platform    State"
        echo "i-12345678   test    Linux       running"
        ;;
    *"ssm start-session"*)
        echo "Starting session with instance $4"
        ;;
    *"ssm send-command"*)
        echo '{"Command":{"CommandId":"11111-22222-33333","Status":"Pending"}}'
        ;;
    *"ssm get-command-invocation"*)
        echo '{"Status":"Success","StandardOutputContent":"Mock command output","StandardErrorContent":""}'
        ;;
    *)
        echo "Mocked AWS CLI called with: $*"
        ;;
esac
EOF
chmod +x "$MOCK_DIR/aws"

# Add mocks to PATH
export PATH="$MOCK_DIR:$PATH"

# Test function
run_test() {
    local test_name="$1"
    local test_cmd="$2"
    local expected_output="$3"
    local expected_exit_code="${4:-0}"

    ((TESTS_RUN++))
    echo
    echo "Running test: $test_name"
    echo "Command: $test_cmd"

    output=$(eval "$test_cmd" 2>&1) || true
    exit_code=$?

    if [ "$exit_code" -ne "$expected_exit_code" ]; then
        echo -e "${RED}FAILED${NC} (wrong exit code)"
        echo "Expected: $expected_exit_code, Got: $exit_code"
        ((TESTS_FAILED++))
        return
    fi

    if [[ "$output" =~ $expected_output ]]; then
        echo -e "${GREEN}PASSED${NC}"
    else
        echo -e "${RED}FAILED${NC} (unexpected output)"
        echo "Expected: $expected_output"
        echo "Got: $output"
        ((TESTS_FAILED++))
    fi
}

# Cleanup mocks
cleanup() {
    rm -rf "$MOCK_DIR"
}
trap cleanup EXIT

# Run tests
echo "Running tests..."

run_test "Show version" "$MAIN_SCRIPT version" "AWS SSM Helper version"
run_test "Show help" "$MAIN_SCRIPT help" "AWS SSM Session Manager Helper"
run_test "List instances in default region" "$MAIN_SCRIPT cac1" "Listing instances in ca-central-1"
run_test "Connect to instance with default region" "$MAIN_SCRIPT i-12345678" "Connecting to instance i-12345678 in ca-central-1"
run_test "Invalid region" "$MAIN_SCRIPT xyz1" "Invalid region" 1
run_test "Invalid instance ID" "$MAIN_SCRIPT cac1 i-123" "Invalid instance ID format" 1
run_test "Connect with valid region and instance" "$MAIN_SCRIPT use1 i-12345678" "Connecting to instance i-12345678 in us-east-1"

# Exec command tests
run_test "Exec command on instance" "$MAIN_SCRIPT exec cac1 i-12345678 'ls -la'" "Executing command on instance i-12345678"
run_test "Exec command with invalid region" "$MAIN_SCRIPT exec xyz1 i-12345678 'ls -la'" "Invalid region code" 1
run_test "Exec command with invalid instance ID" "$MAIN_SCRIPT exec cac1 i-123 'ls -la'" "Invalid instance ID format" 1
run_test "Exec command missing parameters" "$MAIN_SCRIPT exec cac1" "Missing required parameters" 1

# Tag-based exec
run_test "Run command by tag" "$MAIN_SCRIPT exec-tagged cac1 Role web 'ls -la'" "Executing command on instances with tag Role=web"
run_test "Run command by tag missing parameters" "$MAIN_SCRIPT exec-tagged cac1 Role" "Missing required parameters" 1

# 🔥 New region tests
NEW_REGIONS=("usw2" "use2" "euw2" "euw3" "euc1" "euc2" "eun1" "eus1" "eus2" "apse1" "apse2" "apse3" "apse4" "apne1" "apne2" "apne3" "sae1" "mec1" "mees1" "afc1")

for region in "${NEW_REGIONS[@]}"; do
    run_test "List instances in $region" "$MAIN_SCRIPT $region" "Listing instances in"
done

echo
echo "Test Summary:"
echo "Tests run: $TESTS_RUN"
echo "Tests failed: $TESTS_FAILED"

if [ "$TESTS_FAILED" -gt 0 ]; then
    exit 1
fi
exit 0
