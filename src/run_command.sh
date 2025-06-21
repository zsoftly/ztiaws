#!/usr/bin/env bash

# SSM Run Command functionality
# Repository: https://github.com/ZSoftly/ztiaws

# SCRIPT_DIR dependency: set by calling script for proper sourcing
if [ -n "${SCRIPT_DIR:-}" ] && [ -f "${SCRIPT_DIR}/src/utils.sh" ]; then
    # shellcheck disable=SC1091
    source "${SCRIPT_DIR}/src/utils.sh"
elif [ -f "/usr/local/bin/src/utils.sh" ]; then
    # shellcheck source=/dev/null
    source "/usr/local/bin/src/utils.sh"
else
    echo "[ERROR] src/utils.sh not found. run_command.sh requires utils.sh." >&2
    echo "This script is typically sourced by 'ssm' or 'authaws', which should handle this error." >&2
fi

# Execute a command on a remote EC2 instance using SSM Run Command
run_remote_command() {
    # Trap for debugging unexpected script exits
    trap 'echo -e "\nTRAP: Script exited unexpectedly at line $LINENO in run_remote_command" >&2' ERR
    
    SSM_DEBUG=${SSM_DEBUG:-false}
    debug_log() {
        if [ "$SSM_DEBUG" = true ]; then
            echo -e "\n[DEBUG] $*" >&2
        fi
    }
    
    local instance_id=$1
    local region=$2
    local command=$3
    local comment=${4:-"Command executed via ztiaws"}

    if [[ -z "$instance_id" || -z "$region" || -z "$command" ]]; then
        echo "Error: Missing required parameters."
        echo "Usage: run_remote_command <instance-id> <region> <command> [comment]"
        trap - ERR
        return 1
    fi

    # Escape quotes for JSON array formatting in AWS API
    local escaped_command
    escaped_command=$(printf '%s' "$command" | sed 's/"/\\"/g')
    
    echo "Executing command on instance $instance_id in region $region:"
    echo "$command"
    
    debug_log "Preparing to send command to AWS SSM"
    
    local response
    local aws_error_log
    aws_error_log=$(mktemp)
    local aws_exit_code

    # Temporarily disable exit on error to capture AWS CLI errors
    set +e
    response=$(aws ssm send-command \
        --instance-ids "$instance_id" \
        --document-name "AWS-RunShellScript" \
        --parameters commands="[\"$escaped_command\"]" \
        --comment "$comment" \
        --region "$region" \
        --output json 2> "$aws_error_log")
    aws_exit_code=$?
    set -e
    
    debug_log "AWS SSM send-command exit code: $aws_exit_code"

    if [ $aws_exit_code -ne 0 ]; then
        log_error "AWS CLI command failed (send-command for single instance) with exit code $aws_exit_code."
        if [ -s "$aws_error_log" ]; then
            echo -e "${RED}--- AWS CLI Error Details ---${NC}"
            while IFS= read -r line; do echo -e "${RED}${line}${NC}"; done < "$aws_error_log"
            echo -e "${RED}-----------------------------${NC}"
        else
            log_error "No specific error message captured from AWS CLI, but command failed."
        fi
        rm -f "$aws_error_log"
        trap - ERR
        return 1
    fi
    rm -f "$aws_error_log"
    
    debug_log "Parsing command ID from response"
    
    local command_id
    # Use temporary file to avoid subshell issues with jq parsing
    local tmp_file
    tmp_file=$(mktemp)
    echo "$response" > "$tmp_file"
    command_id=$(jq -r '.Command.CommandId' < "$tmp_file")
    rm -f "$tmp_file"
    
    debug_log "Extracted command ID: $command_id"
    
    if [[ -z "$command_id" || "$command_id" == "null" ]]; then
        log_error "Failed to parse CommandId from AWS response (send-command for single instance)."
        log_info "AWS Response: $response"
        trap - ERR
        return 1
    fi
    
    echo "Command ID: $command_id"
    echo "Waiting for command to complete..."
    
    # Allow command setup time before first status check
    debug_log "Initial wait of 3 seconds before checking command status"
    sleep 3
    
    # Get command result
    local max_retries=30
    local retry_count=0
    local status="Pending"
    
    while true; do
        debug_log "Polling attempt $((retry_count + 1))/$max_retries"
        
        if [[ $retry_count -ge $max_retries ]]; then
            echo ""
            log_error "Command polling timed out after $max_retries retries."
            trap - ERR
            return 1
        fi
        
        debug_log "Executing get-command-invocation"
        
        # Use files to avoid subshell issues
        local tmp_result_file
        tmp_result_file=$(mktemp)
        local tmp_exit_code_file
        tmp_exit_code_file=$(mktemp)
        set +e
        aws ssm get-command-invocation \
            --command-id "$command_id" \
            --instance-id "$instance_id" \
            --region "$region" \
            --output json > "$tmp_result_file" 2>&1
        echo $? > "$tmp_exit_code_file"
        set -e
        
        local aws_cli_exit_code
        aws_cli_exit_code=$(cat "$tmp_exit_code_file")
        local command_result
        command_result=$(cat "$tmp_result_file")
        
        debug_log "get-command-invocation exit code: $aws_cli_exit_code"
        
        if [[ $aws_cli_exit_code -ne 0 ]]; then
            echo ""
            log_error "AWS CLI command failed (get-command-invocation) with exit code $aws_cli_exit_code"
            log_error "Output was: $command_result"
            rm -f "$tmp_result_file" "$tmp_exit_code_file"
            trap - ERR
            return 1
        fi
        
        # Safe parsing of JSON using files
        debug_log "Parsing JSON status from response"
        set +e
        status=$(jq -r '.Status' < "$tmp_result_file")
        local jq_exit_code=$?
        set -e
        
        debug_log "jq exit code: $jq_exit_code, parsed status: $status"
        
        if [[ $jq_exit_code -ne 0 ]]; then
            echo ""
            log_error "Failed to parse command status with jq (exit code $jq_exit_code)"
            log_error "Input was: $(cat "$tmp_result_file")"
            rm -f "$tmp_result_file" "$tmp_exit_code_file"
            trap - ERR
            return 1
        fi
        
        rm -f "$tmp_result_file" "$tmp_exit_code_file"
        
        if [[ -z "$status" || "$status" == "null" ]]; then
            echo ""
            log_error "Command status was empty or null"
            trap - ERR
            return 1
        fi
        
        debug_log "Command status: $status"
        
        # Command is still running
        if [[ "$status" == "Pending" || "$status" == "InProgress" ]]; then
            echo -n "."
            # Protect sleep operation from error trapping
            set +e
            sleep 2
            retry_count=$((retry_count + 1))
            set -e
            continue
        fi
        
        # Command completed (success or failure)
        echo ""
        
        # Get final results
        debug_log "Command completed with status: $status. Getting final results."
        local final_result_file
        final_result_file=$(mktemp)
        
        set +e
        aws ssm get-command-invocation \
            --command-id "$command_id" \
            --instance-id "$instance_id" \
            --region "$region" \
            --output json > "$final_result_file" 2>&1
        local final_aws_exit_code=$?
        set -e
        
        debug_log "Final get-command-invocation exit code: $final_aws_exit_code"
        
        if [[ $final_aws_exit_code -ne 0 ]]; then
            log_error "Failed to get final command result (exit code $final_aws_exit_code)"
            log_error "Output was: $(cat "$final_result_file")"
            rm -f "$final_result_file"
            trap - ERR
            return 1
        fi
        
        # Parse output safely
        local std_out
        local std_err
        local final_status
        
        debug_log "Parsing final output"
        final_status=$(jq -r '.Status' < "$final_result_file")
        std_out=$(jq -r '.StandardOutputContent' < "$final_result_file")
        std_err=$(jq -r '.StandardErrorContent' < "$final_result_file")
        
        # Handle null values
        std_out=$([[ "$std_out" == "null" ]] && echo "" || echo "$std_out")
        std_err=$([[ "$std_err" == "null" ]] && echo "" || echo "$std_err")
        
        debug_log "Final status: $final_status"
        echo "Status: $final_status"
        echo "--------- Command Output ---------"
        echo -e "${CYAN}$std_out${NC}"
            
        if [[ -n "$std_err" ]]; then
            echo "--------- Command Error ---------"
            echo -e "${CYAN}$std_err${NC}"
        fi
            
        rm -f "$final_result_file"
            
        if [[ "$final_status" != "Success" ]]; then
            echo "Command failed with status: $final_status"
            trap - ERR
            return 1
        fi
        
        # Success - exit the loop
        debug_log "Command completed successfully"
        break
    done
    
    trap - ERR
    return 0
}

# Execute a command on multiple EC2 instances using tags
run_remote_command_tagged() {
    local tag_key=$1
    local tag_value=$2
    local region=$3
    local command=$4
    local comment=${5:-"Command executed via ztiaws"}
    
    if [[ -z "$tag_key" || -z "$tag_value" || -z "$region" || -z "$command" ]]; then
        echo "Error: Missing required parameters."
        echo "Usage: run_remote_command_tagged <tag-key> <tag-value> <region> <command> [comment]"
        return 1
    fi
    
    # Escape quotes for JSON array formatting in AWS API
    local escaped_command
    escaped_command=$(printf '%s' "$command" | sed 's/"/\\"/g')
    
    echo "Executing command on instances with tag $tag_key=$tag_value in region $region:"
    echo "$command"
    
    local response
    local aws_error_log
    aws_error_log=$(mktemp)
    local aws_exit_code

    # Temporarily disable exit on error to capture AWS CLI errors
    set +e
    response=$(aws ssm send-command \
        --targets "Key=tag:$tag_key,Values=$tag_value" \
        --document-name "AWS-RunShellScript" \
        --parameters commands="[\"$escaped_command\"]" \
        --comment "$comment" \
        --region "$region" \
        --output json 2> "$aws_error_log")
    aws_exit_code=$?
    set -e

    if [ $aws_exit_code -ne 0 ]; then
        log_error "AWS CLI command failed (send-command for tags) with exit code $aws_exit_code."
        if [ -s "$aws_error_log" ]; then
            echo -e "${RED}--- AWS CLI Error Details ---${NC}"
            # Read line by line to ensure color is applied per line
            while IFS= read -r line; do echo -e "${RED}${line}${NC}"; done < "$aws_error_log"
            echo -e "${RED}-----------------------------${NC}"
        else
            log_error "No specific error message captured from AWS CLI, but command failed."
        fi
        rm -f "$aws_error_log"
        return 1
    fi
    rm -f "$aws_error_log"
    
    local command_id
    command_id=$(echo "$response" | jq -r '.Command.CommandId')
    
    if [[ -z "$command_id" || "$command_id" == "null" ]]; then
        log_error "Failed to parse CommandId from AWS response (send-command for tags)."
        log_info "AWS Response: $response"
        return 1
    fi

    # Verify instances were targeted by checking command invocations
    # AWS send-command doesn't directly indicate if no targets were found
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
        return 1
    fi

    # Create a temp file to store status across subshells
    local status_file
    status_file=$(mktemp)
    echo "0" > "$status_file"

    echo "Waiting for command to complete on all instances..."

    # Increased retries for multiple instances
    local max_retries=60
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
            echo "1" > "$status_file"
            rm -f "$status_file"
            return 1
        fi

        local pending_or_in_progress
        pending_or_in_progress=$(echo "$invocations_response" | jq -r '.CommandInvocations[] | select(.Status == "Pending" or .Status == "InProgress") | .InstanceId')

        if [[ -z "$pending_or_in_progress" ]]; then
            all_done=true
        else
            echo -n "."
            # Sleep longer when polling for multiple instances
            sleep 2
            ((retry_count++))
        fi
    done
    echo ""

    if ! $all_done; then
        echo "Error: Command timed out or failed to get status for all instances after $max_retries retries."
        echo "1" > "$status_file"
    fi

    # Get final results and display
    # Use get-command-invocation for each instance for reliable output
    local invocations_summary_response
    invocations_summary_response=$(aws ssm list-command-invocations \
        --command-id "$command_id" \
        --region "$region" \
        --output json 2>/dev/null)

    if [[ -z "$invocations_summary_response" ]]; then
        echo "Error: Failed to get final command invocations list."
        echo "1" > "$status_file"
        rm -f "$status_file"
        return 1
    fi

    # Process each instance result individually
    local instance_ids
    instance_ids=$(echo "$invocations_summary_response" | jq -r '.CommandInvocations[].InstanceId')
    
    for instance_id in $instance_ids; do
        # Get the status from the summary response
        local status_from_list
        status_from_list=$(echo "$invocations_summary_response" | jq -r --arg id "$instance_id" '.CommandInvocations[] | select(.InstanceId == $id) | .Status')

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
            echo "1" > "$status_file"
            continue
        fi

        local status_from_get command_result_std_out command_result_std_err
        status_from_get=$(echo "$detailed_command_result" | jq -r '.Status')
        command_result_std_out=$(echo "$detailed_command_result" | jq -r '.StandardOutputContent')
        command_result_std_err=$(echo "$detailed_command_result" | jq -r '.StandardErrorContent')

        local std_out
        std_out=$([[ "$command_result_std_out" == "null" ]] && echo "" || echo "$command_result_std_out")
        
        local std_err
        std_err=$([[ "$command_result_std_err" == "null" ]] && echo "" || echo "$command_result_std_err")

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
            echo "1" > "$status_file"
        fi
    done
    echo "----------------------------------------"

    # Read the final status from our temp file
    local overall_status
    overall_status=$(cat "$status_file")
    rm -f "$status_file"

    if [[ $overall_status -ne 0 ]]; then
        log_error "One or more commands failed or timed out on the targeted instances."
    elif [[ "$instance_count" -gt 0 ]]; then
        log_info "All commands completed successfully on targeted instances."
    fi
    
    return "$overall_status"
}