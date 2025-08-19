#!/usr/bin/env bash

# CloudWatch Logs functionality for SSM
# Repository: https://github.com/ZSoftly/ztiaws

# Get SCRIPT_DIR and source utilities
# This assumes SCRIPT_DIR is set by the calling script (ssm)
if [ -n "${SCRIPT_DIR:-}" ] && [ -f "${SCRIPT_DIR}/src/00_utils.sh" ]; then
    # This warning is expected as ShellCheck can't follow dynamically constructed file paths
    # shellcheck disable=SC1091
    source "${SCRIPT_DIR}/src/00_utils.sh"
elif [ -f "/usr/local/bin/src/00_utils.sh" ]; then # For system-wide installation
    # shellcheck source=/dev/null
    source "/usr/local/bin/src/00_utils.sh"
else
    echo "[ERROR] src/00_utils.sh not found. 05_ssm_cloudwatch_logs.sh requires 00_utils.sh." >&2
    echo "This script is typically sourced by 'ssm', which should handle this error." >&2
    exit 1
fi

# Get log groups for an EC2 instance
get_instance_log_groups() {
    local instance_id="$1"
    local region="$2"
    
    debug_log "Getting log groups for instance: $instance_id in region: $region"
    
    # Common EC2 log group patterns
    local log_patterns=(
        "/aws/ec2/${instance_id}"
        "/var/log/messages"
        "/var/log/cloud-init"
        "/var/log/cloud-init-output.log"
        "/aws/amazoncloudwatch-agent/${instance_id}"
    )
    
    local found_groups=()
    
    # Search for log groups matching instance patterns
    for pattern in "${log_patterns[@]}"; do
        local groups
        groups=$(aws logs describe-log-groups \
            --region "$region" \
            --log-group-name-prefix "$pattern" \
            --query 'logGroups[].logGroupName' \
            --output text 2>/dev/null || echo "")
        
        if [ -n "$groups" ] && [ "$groups" != "None" ]; then
            # Convert tab-separated output to array, filtering empty values
            while IFS=$'\t' read -ra group_array; do
                for group in "${group_array[@]}"; do
                    if [ -n "$group" ] && [ "$group" != "None" ]; then
                        found_groups+=("$group")
                    fi
                done
            done <<< "$groups"
        fi
    done
    
    # Also try to find log groups that contain the instance ID
    local instance_groups
    instance_groups=$(aws logs describe-log-groups \
        --region "$region" \
        --query "logGroups[?contains(logGroupName, '$instance_id')].logGroupName" \
        --output text 2>/dev/null || echo "")
    
    if [ -n "$instance_groups" ] && [ "$instance_groups" != "None" ]; then
        while IFS=$'\t' read -ra group_array; do
            for group in "${group_array[@]}"; do
                if [ -n "$group" ] && [ "$group" != "None" ]; then
                    # Avoid duplicates
                    local already_added=false
                    for existing in "${found_groups[@]}"; do
                        if [ "$existing" = "$group" ]; then
                            already_added=true
                            break
                        fi
                    done
                    if [ "$already_added" = false ]; then
                        found_groups+=("$group")
                    fi
                fi
            done
        done <<< "$instance_groups"
    fi
    
    # Return found groups
    printf '%s\n' "${found_groups[@]}"
}

# View recent logs from a log group
view_log_group() {
    local log_group="$1"
    local region="$2"
    local filter_pattern="${3:-}"
    local max_items="${4:-100}"
    
    debug_log "Viewing logs from group: $log_group in region: $region"
    
    # Calculate start time (last 1 hour by default)
    local start_time
    start_time=$(($(date +%s) - 3600))
    start_time=$((start_time * 1000)) # Convert to milliseconds
    
    # Build AWS CLI command
    local aws_cmd=(
        "aws" "logs" "filter-log-events"
        "--region" "$region"
        "--log-group-name" "$log_group"
        "--start-time" "$start_time"
        "--max-items" "$max_items"
    )
    
    # Add filter pattern if provided
    if [ -n "$filter_pattern" ]; then
        aws_cmd+=("--filter-pattern" "$filter_pattern")
    fi
    
    # Add output formatting
    aws_cmd+=(
        "--query" "events[].{Time:timestamp,Message:message}"
        "--output" "table"
    )
    
    debug_log "Executing: ${aws_cmd[*]}"
    
    # Execute the command
    if ! "${aws_cmd[@]}" 2>/dev/null; then
        log_error "Failed to retrieve logs from group: $log_group"
        return 1
    fi
    
    return 0
}

# List available log groups in a region
list_log_groups() {
    local region="$1"
    local prefix="${2:-}"
    
    debug_log "Listing log groups in region: $region"
    
    local aws_cmd=(
        "aws" "logs" "describe-log-groups"
        "--region" "$region"
        "--query" "logGroups[].{LogGroup:logGroupName,CreationTime:creationTime,SizeBytes:storedBytes}"
        "--output" "table"
    )
    
    # Add prefix filter if provided
    if [ -n "$prefix" ]; then
        aws_cmd+=("--log-group-name-prefix" "$prefix")
    fi
    
    debug_log "Executing: ${aws_cmd[*]}"
    
    # Execute the command
    if ! "${aws_cmd[@]}" 2>/dev/null; then
        log_error "Failed to list log groups in region: $region"
        return 1
    fi
    
    return 0
}

# Handle logs command with various argument combinations
handle_logs_command() {
    local region=""
    local target=""
    local filter_pattern=""
    
    # Parse arguments
    case $# in
        0)
            log_error "logs command requires at least region and target"
            echo "Usage: $(basename "$0") logs <region> <instance-id|log-group> [filter-pattern]"
            return 1
            ;;
        1)
            log_error "logs command requires both region and target"
            echo "Usage: $(basename "$0") logs <region> <instance-id|log-group> [filter-pattern]"
            return 1
            ;;
        2)
            region="$1"
            target="$2"
            ;;
        3)
            region="$1"
            target="$2"
            filter_pattern="$3"
            ;;
        *)
            log_error "Too many arguments for logs command"
            echo "Usage: $(basename "$0") logs <region> <instance-id|log-group> [filter-pattern]"
            return 1
            ;;
    esac
    
    # Validate region
    if ! validate_region_code "$region" region_name; then
        log_error "Invalid region code: $region"
        return 1
    fi
    
    debug_log "Processing logs command: region=$region, target=$target, filter=$filter_pattern"
    
    # Check if target looks like an instance ID (starts with i-)
    if [[ "$target" =~ ^i-[0-9a-f]{8,17}$ ]]; then
        log_info "Getting logs for instance: $target"
        
        # First, verify we can connect to AWS
        if ! aws sts get-caller-identity --region "$region" >/dev/null 2>&1; then
            log_error "Unable to connect to AWS. Please check your credentials and try again."
            log_info "You can configure AWS credentials using:"
            log_info "  aws configure"
            log_info "  aws sso login"
            log_info "  or set AWS_PROFILE environment variable"
            return 1
        fi
        
        # Get log groups for this instance
        local log_groups
        mapfile -t log_groups < <(get_instance_log_groups "$target" "$region")
        
        if [ ${#log_groups[@]} -eq 0 ]; then
            log_warn "No log groups found for instance: $target"
            log_info "You may need to:"
            log_info "  1. Install and configure CloudWatch agent on the instance"
            log_info "  2. Check that the instance has proper IAM permissions for CloudWatch"
            log_info "  3. Verify that logs are being sent to CloudWatch"
            return 1
        fi
        
        log_info "Found ${#log_groups[@]} log group(s) for instance $target:"
        printf "  - %s\n" "${log_groups[@]}"
        echo
        
        # Show logs from each group
        for log_group in "${log_groups[@]}"; do
            log_info "Recent logs from: $log_group"
            echo "----------------------------------------"
            view_log_group "$log_group" "$region" "$filter_pattern"
            echo
        done
        
    else
        # Treat as log group name
        log_info "Viewing logs from group: $target"
        
        # First, verify we can connect to AWS
        if ! aws sts get-caller-identity --region "$region" >/dev/null 2>&1; then
            log_error "Unable to connect to AWS. Please check your credentials and try again."
            log_info "You can configure AWS credentials using:"
            log_info "  aws configure"
            log_info "  aws sso login"
            log_info "  or set AWS_PROFILE environment variable"
            return 1
        fi
        
        # Check if log group exists
        if ! aws logs describe-log-groups \
            --region "$region" \
            --log-group-name-prefix "$target" \
            --query "logGroups[?logGroupName=='$target']" \
            --output text >/dev/null 2>&1; then
            
            log_error "Log group not found: $target"
            log_info "Available log groups in region $region:"
            list_log_groups "$region"
            return 1
        fi
        
        # Show logs from the specified group
        view_log_group "$target" "$region" "$filter_pattern"
    fi
    
    return 0
}