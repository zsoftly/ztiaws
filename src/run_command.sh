#!/usr/bin/env bash

# SSM Run Command functionality
# Repository: https://github.com/ZSoftly/ztiaws

# Get SCRIPT_DIR and source utilities
# This assumes SCRIPT_DIR is set by the calling script (ssm or authaws)
# If run_command.sh is ever called directly in a way that SCRIPT_DIR is not set,
# this sourcing might fail or need adjustment.
if [ -n "${SCRIPT_DIR:-}" ] && [ -f "${SCRIPT_DIR}/src/utils.sh" ]; then
    # shellcheck source=./utils.sh
    source "${SCRIPT_DIR}/src/utils.sh"
elif [ -f "/usr/local/bin/src/utils.sh" ]; then # For system-wide installation
    # shellcheck source=/dev/null
    source "/usr/local/bin/src/utils.sh"
else
    # utils.sh is mandatory for run_command.sh as well, as it uses log_error etc.
    # However, since this script is sourced, echoing directly might be problematic for tests or other consumers.
    # The primary scripts (ssm, authaws) will catch the absence of utils.sh and exit.
    # If this script *were* to be run standalone and utils.sh was missing, it would likely fail when log_error is called.
    # For now, we assume the calling script handles the mandatory nature of utils.sh.
    # To make it truly standalone-safe, it would need its own exit here.
    echo "[ERROR] src/utils.sh not found. run_command.sh requires utils.sh." >&2
    echo "This script is typically sourced by 'ssm' or 'authaws', which should handle this error." >&2
    # Consider adding 'exit 1' here if standalone execution is a primary concern and it shouldn't proceed.
fi

# Execute a command on a remote EC2 instance using SSM Run Command
run_remote_command() {
    local instance_id=$1
    local region=$2
    local command=$3
    local comment=${4:-"Command executed via ztiaws"}

    # Validate input parameters
    if [[ -z "$instance_id" || -z "$region" || -z "$command" ]]; then
        echo "Error: Missing required parameters."
        echo "Usage: run_remote_command <instance-id> <region> <command> [comment]"
        return 1
    fi

    # Format the command for AWS SSM Run Command
    # We need to escape quotes and format the command as a JSON array string
    local escaped_command
    escaped_command=$(printf '%s' "$command" | sed 's/"/\\"/g')
    
    echo "Executing command on instance $instance_id in region $region:"
    echo "$command"
    
    # Execute the command using AWS SSM Send-Command
    local response
    local aws_error_log
    aws_error_log=$(mktemp)
    local aws_exit_code

    set +e # Temporarily disable exit on error to capture AWS CLI errors
    response=$(aws ssm send-command \
        --instance-ids "$instance_id" \
        --document-name "AWS-RunShellScript" \
        --parameters commands="[\"$escaped_command\"]" \
        --comment "$comment" \
        --region "$region" \
        --output json 2> "$aws_error_log")
    aws_exit_code=$?
    set -e # Re-enable exit on error

    if [ $aws_exit_code -ne 0 ]; then
        log_error "AWS CLI command failed (send-command for single instance) with exit code $aws_exit_code."
        if [ -s "$aws_error_log" ]; then # Check if error log is not empty
            echo -e "${RED}--- AWS CLI Error Details ---${NC}"
            while IFS= read -r line; do echo -e "${RED}${line}${NC}"; done < "$aws_error_log"
            echo -e "${RED}-----------------------------${NC}"
        else
            log_error "No specific error message captured from AWS CLI, but command failed."
        fi
        rm -f "$aws_error_log"
        return 1 # Propagate error
    fi
    rm -f "$aws_error_log" # Clean up temp file on success
    
    local command_id
    command_id=$(echo "$response" | jq -r '.Command.CommandId')
    
    if [[ -z "$command_id" || "$command_id" == "null" ]]; then
        log_error "Failed to parse CommandId from AWS response (send-command for single instance)."
        log_info "AWS Response: $response"
        return 1
    fi
    
    echo "Command ID: $command_id"
    echo "Waiting for command to complete..."
    
    # Wait for command to complete
    sleep 2
    
    # Get command result
    local max_retries=30
    local retry_count=0
    local status="Pending"
    
    while [[ "$status" == "Pending" || "$status" == "InProgress" ]] && (( retry_count < max_retries )); do
        local command_result
        command_result=$(aws ssm get-command-invocation \
            --command-id "$command_id" \
            --instance-id "$instance_id" \
            --region "$region" \
            --output json 2>/dev/null)
        
        status=$(echo "$command_result" | jq -r '.Status')
        
        if [[ -z "$status" || "$status" == "null" ]]; then
            echo "Error: Failed to get command status. AWS Response:"
            echo "$command_result"
            return 1
        fi
        
        if [[ "$status" == "Pending" || "$status" == "InProgress" ]]; then
            echo -n "."
            sleep 1
            ((retry_count++))
        else
            echo ""
            # Display command output
            local command_result_std_out command_result_std_err
            command_result_std_out=$(echo "$command_result" | jq -r '.StandardOutputContent')
            command_result_std_err=$(echo "$command_result" | jq -r '.StandardErrorContent')

            local std_out=$([[ "$command_result_std_out" == "null" ]] && echo "" || echo "$command_result_std_out")
            local std_err=$([[ "$command_result_std_err" == "null" ]] && echo "" || echo "$command_result_std_err")
            
            echo "Status: $status"
            echo "--------- Command Output ---------"
            echo -e "${CYAN}$std_out${NC}"
            
            if [[ -n "$std_err" ]]; then
                echo "--------- Command Error ---------"
                echo -e "${CYAN}$std_err${NC}"
            fi
            
            if [[ "$status" != "Success" ]]; then
                echo "Command failed with status: $status"
                return 1
            fi
            
            return 0
        fi
    done
    
    if (( retry_count >= max_retries )); then
        echo ""
        echo "Error: Command timed out after $max_retries retries."
        return 1
    fi
}

# Execute a command on multiple EC2 instances using tags
run_remote_command_tagged() {
    local tag_key=$1
    local tag_value=$2
    local region=$3
    local command=$4
    local comment=${5:-"Command executed via ztiaws"}
    
    # Validate input parameters
    if [[ -z "$tag_key" || -z "$tag_value" || -z "$region" || -z "$command" ]]; then
        echo "Error: Missing required parameters."
        echo "Usage: run_remote_command_tagged <tag-key> <tag-value> <region> <command> [comment]"
        return 1
    fi
    
    # Format the command for AWS SSM Run Command
    local escaped_command
    escaped_command=$(printf '%s' "$command" | sed 's/"/\\"/g')
    
    echo "Executing command on instances with tag $tag_key=$tag_value in region $region:"
    echo "$command"
    
    # Execute the command using AWS SSM Send-Command with tag targeting
    local response
    local aws_error_log
    aws_error_log=$(mktemp)
    local aws_exit_code

    set +e # Temporarily disable exit on error to capture AWS CLI errors
    response=$(aws ssm send-command \
        --targets "Key=tag:$tag_key,Values=$tag_value" \
        --document-name "AWS-RunShellScript" \
        --parameters commands="[\"$escaped_command\"]" \
        --comment "$comment" \
        --region "$region" \
        --output json 2> "$aws_error_log")
    aws_exit_code=$?
    set -e # Re-enable exit on error

    if [ $aws_exit_code -ne 0 ]; then
        log_error "AWS CLI command failed (send-command for tags) with exit code $aws_exit_code."
        if [ -s "$aws_error_log" ]; then # Check if error log is not empty
            echo -e "${RED}--- AWS CLI Error Details ---${NC}"
            # Read line by line to ensure color is applied per line
            while IFS= read -r line; do echo -e "${RED}${line}${NC}"; done < "$aws_error_log"
            echo -e "${RED}-----------------------------${NC}"
        else
            log_error "No specific error message captured from AWS CLI, but command failed."
        fi
        rm -f "$aws_error_log"
        return 1 # Propagate error
    fi
    rm -f "$aws_error_log" # Clean up temp file on success
    
    local command_id
    command_id=$(echo "$response" | jq -r '.Command.CommandId')
    
    if [[ -z "$command_id" || "$command_id" == "null" ]]; then
        log_error "Failed to parse CommandId from AWS response (send-command for tags)."
        log_info "AWS Response: $response"
        return 1
    fi

    # Check if any instances were targeted by list-command-invocations
    # This is a proxy for whether the send-command found any targets.
    # The send-command API itself doesn't directly tell you if no targets were found for a tag.
    local initial_invocations_check
    initial_invocations_check=$(aws ssm list-command-invocations \
        --command-id "$command_id" \
        --region "$region" \
        --output json 2>/dev/null)

    local instance_count
    instance_count=$(echo "$initial_invocations_check" | jq -r '.CommandInvocations | length')

    if [[ "$instance_count" -eq 0 ]]; then
        log_error "No instances found matching the specified tags (Tag: $tag_key=$tag_value) in region $region for command $command_id."
        log_error "Command was sent to AWS SSM, but no targets were identified. Halting execution."
        return 1 # Exit with an error status because no instances were targeted
    fi

    local overall_status=0 # Initialize overall_status here, after we know there are instances.

    echo "Waiting for command to complete on all instances..."

    local max_retries=60 # Increased retries for multiple instances
    local retry_count=0
    local all_done=false

    while ! $all_done && (( retry_count < max_retries )); do
        local invocations_response
        invocations_response=$(aws ssm list-command-invocations \
            --command-id "$command_id" \
            --details \
            --region "$region" \
            --output json 2>/dev/null)

        if [[ -z "$invocations_response" ]]; then
            echo "Error: Failed to get command invocations list."
            return 1
        fi

        local pending_or_in_progress
        pending_or_in_progress=$(echo "$invocations_response" | jq -r '.CommandInvocations[] | select(.Status == "Pending" or .Status == "InProgress") | .InstanceId')

        if [[ -z "$pending_or_in_progress" ]]; then
            all_done=true
        else
            echo -n "."
            sleep 2 # Sleep a bit longer when polling for multiple instances
            ((retry_count++))
        fi
    done
    echo ""

    if ! $all_done; then
        echo "Error: Command timed out or failed to get status for all instances after $max_retries retries."
        overall_status=1
    fi

    # Get final results and display
    # Instead of relying on StandardOutputContent from list-command-invocations,
    # iterate and call get-command-invocation for each instance for more reliable output.
    local invocations_summary_response
    invocations_summary_response=$(aws ssm list-command-invocations \
        --command-id "$command_id" \
        --region "$region" \
        --output json 2>/dev/null)

    if [[ -z "$invocations_summary_response" ]]; then
        echo "Error: Failed to get final command invocations list."
        return 1
    fi

    echo "$invocations_summary_response" | jq -c '.CommandInvocations[]' | while IFS= read -r invocation_summary_json; do
        local instance_id status_from_list
        instance_id=$(echo "$invocation_summary_json" | jq -r '.InstanceId')
        status_from_list=$(echo "$invocation_summary_json" | jq -r '.Status') # Status from the list summary

        # Fetch detailed invocation results for this specific instance
        local detailed_command_result
        detailed_command_result=$(aws ssm get-command-invocation \
            --command-id "$command_id" \
            --instance-id "$instance_id" \
            --region "$region" \
            --output json 2>/dev/null)

        if [[ -z "$detailed_command_result" ]]; then
            echo "----------------------------------------"
            echo "Instance ID: $instance_id"
            echo "Status: $status_from_list (Error fetching detailed output)"
            echo "----------------------------------------"
            overall_status=1
            continue
        fi

        local status_from_get command_result_std_out command_result_std_err
        status_from_get=$(echo "$detailed_command_result" | jq -r '.Status')
        command_result_std_out=$(echo "$detailed_command_result" | jq -r '.StandardOutputContent')
        command_result_std_err=$(echo "$detailed_command_result" | jq -r '.StandardErrorContent')

        local std_out=$([[ "$command_result_std_out" == "null" ]] && echo "" || echo "$command_result_std_out")
        local std_err=$([[ "$command_result_std_err" == "null" ]] && echo "" || echo "$command_result_std_err")

        echo "----------------------------------------"
        echo "Instance ID: $instance_id"
        echo "Status: $status_from_get"
        echo "--------- Command Output ---------"
        echo -e "${CYAN}$std_out${NC}"

        if [[ -n "$std_err" ]]; then
            echo "--------- Command Error ---------"
            echo -e "${CYAN}$std_err${NC}"
        fi

        if [[ "$status_from_get" != "Success" ]]; then
            overall_status=1 # Mark overall failure if any instance fails
        fi
    done
    echo "----------------------------------------"

    if [[ $overall_status -ne 0 ]]; then
        # The instance_count check for the specific "no instances targeted" message is now handled earlier.
        log_error "One or more commands failed or timed out on the targeted instances."
    elif [[ "$instance_count" -gt 0 ]]; then # Should always be true if we reached here
        log_info "All commands completed successfully on targeted instances."
    fi # No specific message if instance_count was 0, as we exit earlier.
    
    return $overall_status
}