#!/usr/bin/env bash

# SSM Run Command functionality
# Repository: https://github.com/ZSoftly/ztiaws

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
    response=$(aws ssm send-command \
        --instance-ids "$instance_id" \
        --document-name "AWS-RunShellScript" \
        --parameters commands="[\"$escaped_command\"]" \
        --comment "$comment" \
        --region "$region" \
        --output json)
    
    local command_id
    command_id=$(echo "$response" | jq -r '.Command.CommandId')
    
    if [[ -z "$command_id" || "$command_id" == "null" ]]; then
        echo "Error: Failed to execute command. AWS Response:"
        echo "$response"
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
            echo "$std_out"
            
            if [[ -n "$std_err" ]]; then
                echo "--------- Command Error ---------"
                echo "$std_err"
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
    response=$(aws ssm send-command \
        --targets "Key=tag:$tag_key,Values=$tag_value" \
        --document-name "AWS-RunShellScript" \
        --parameters commands="[\"$escaped_command\"]" \
        --comment "$comment" \
        --region "$region" \
        --output json)
    
    local command_id
    command_id=$(echo "$response" | jq -r '.Command.CommandId')
    
    if [[ -z "$command_id" || "$command_id" == "null" ]]; then
        echo "Error: Failed to execute command. AWS Response:"
        echo "$response"
        return 1
    fi
    
    echo "Command ID: $command_id"
    echo "Waiting for command to complete on all instances..."

    local max_retries=60 # Increased retries for multiple instances
    local retry_count=0
    local all_done=false
    local overall_status=0

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
        echo "$std_out"

        if [[ -n "$std_err" ]]; then
            echo "--------- Command Error ---------"
            echo "$std_err"
        fi

        if [[ "$status_from_get" != "Success" ]]; then
            overall_status=1 # Mark overall failure if any instance fails
        fi
    done
    echo "----------------------------------------"

    if [[ $overall_status -ne 0 ]]; then
        echo "One or more commands failed or timed out."
    fi
    
    return $overall_status
}