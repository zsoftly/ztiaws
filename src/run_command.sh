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
    command_id=$(echo "$response" | grep -o '"CommandId": "[^"]*"' | cut -d'"' -f4)
    
    if [[ -z "$command_id" ]]; then
        echo "Error: Failed to execute command."
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
        
        status=$(echo "$command_result" | grep -o '"Status": "[^"]*"' | cut -d'"' -f4)
        
        if [[ -z "$status" ]]; then
            echo "Error: Failed to get command status."
            return 1
        fi
        
        if [[ "$status" == "Pending" || "$status" == "InProgress" ]]; then
            echo -n "."
            sleep 1
            ((retry_count++))
        else
            echo ""
            # Display command output
            local std_out
            std_out=$(echo "$command_result" | grep -o '"StandardOutputContent": "[^"]*"' | cut -d'"' -f4)
            local std_err
            std_err=$(echo "$command_result" | grep -o '"StandardErrorContent": "[^"]*"' | cut -d'"' -f4)
            
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
run_remote_command_by_tag() {
    local tag_key=$1
    local tag_value=$2
    local region=$3
    local command=$4
    local comment=${5:-"Command executed via ztiaws"}
    
    # Validate input parameters
    if [[ -z "$tag_key" || -z "$tag_value" || -z "$region" || -z "$command" ]]; then
        echo "Error: Missing required parameters."
        echo "Usage: run_remote_command_by_tag <tag-key> <tag-value> <region> <command> [comment]"
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
    command_id=$(echo "$response" | grep -o '"CommandId": "[^"]*"' | cut -d'"' -f4)
    
    if [[ -z "$command_id" ]]; then
        echo "Error: Failed to execute command."
        return 1
    fi
    
    echo "Command ID: $command_id"
    echo "Command sent to instances. Use AWS Console or AWS CLI to check results."
    echo "To check results with AWS CLI, use:"
    echo "aws ssm list-command-invocations --command-id $command_id --details --region $region"
    
    return 0
}