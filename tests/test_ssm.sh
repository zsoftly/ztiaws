#!/usr/bin/env bash

# Exit on error
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Test counter
TESTS_RUN=0
TESTS_FAILED=0

# Test function
run_test() {
    local test_name="$1"
    local test_cmd="$2"
    local expected_output="$3"
    local expected_exit_code="${4:-0}"
    
    ((TESTS_RUN++))
    
    echo -n "Testing $test_name... "
    
    # Run the command and capture output and exit code
    output=$(eval "$test_cmd" 2>&1) || true
    exit_code=$?
    
    # Check exit code
    if [ "$exit_code" -ne "$expected_exit_code" ]; then
        echo -e "${RED}FAILED${NC}"
        echo "Expected exit code $expected_exit_code, got $exit_code"
        echo "Command output: $output"
        ((TESTS_FAILED++))
        return
    fi
    
    # Check output if expected output is provided
    if [ -n "$expected_output" ]; then
        if [[ "$output" =~ $expected_output ]]; then
            echo -e "${GREEN}PASSED${NC}"
        else
            echo -e "${RED}FAILED${NC}"
            echo "Expected output matching: $expected_output"
            echo "Got output: $output"
            ((TESTS_FAILED++))
            return
        fi
    else
        echo -e "${GREEN}PASSED${NC}"
    fi
}

# Setup
echo "Setting up test environment..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MAIN_SCRIPT="$SCRIPT_DIR/../ssm"

# Make script executable
chmod +x "$MAIN_SCRIPT"

# Test cases
echo "Running tests..."

# Test version command
run_test "version command" \
    "$MAIN_SCRIPT version" \
    "AWS SSM Helper version: [0-9]+\.[0-9]+\.[0-9]+"

# Test help display
run_test "help command" \
    "$MAIN_SCRIPT help" \
    "Usage: ssm \[region\] \[instance-id\]"

# Test invalid region
run_test "invalid region" \
    "$MAIN_SCRIPT invalidregion" \
    "Error: Invalid region" \
    1

# Test valid region format
run_test "valid region format" \
    "$MAIN_SCRIPT cac1" \
    "Available instances"

# Test invalid instance ID format
run_test "invalid instance ID format" \
    "$MAIN_SCRIPT i-123" \
    "Error: Invalid instance ID format" \
    1

# Test valid instance ID format
run_test "valid instance ID format" \
    "$MAIN_SCRIPT i-12345678" \
    "Connecting to instance"

# Report results
echo "Test Summary:"
echo "Tests run: $TESTS_RUN"
echo "Tests failed: $TESTS_FAILED"

if [ "$TESTS_FAILED" -gt 0 ]; then
    exit 1
fi

exit 0