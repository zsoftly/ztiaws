#!/usr/bin/env bash

resolve_instance_identifier() {
    local identifier="$1"
    local region="$2"
    
    # Direct instance ID validation saves API calls for well-formed IDs
    if [[ "$identifier" =~ ^i-[a-zA-Z0-9]{8,}$ ]]; then
        local instance_check
        instance_check=$(aws ec2 describe-instances \
            --region "$region" \
            --instance-ids "$identifier" \
            --query 'Reservations[0].Instances[0].InstanceId' \
            --output text 2>/dev/null)
        
        if [[ "$instance_check" == "$identifier" ]]; then
            echo "$identifier"
            return 0
        else
            log_error "Instance ID '$identifier' not found in region '$region'" >&2
            return 1
        fi
    fi
    
    log_info "Searching for instance named: $identifier" >&2
    
    # Filter by both name tag and instance state to avoid terminated instances
    local instances
    instances=$(aws ec2 describe-instances \
        --region "$region" \
        --filters "Name=tag:Name,Values=$identifier" "Name=instance-state-name,Values=running,stopped" \
        --query "Reservations[].Instances[].[InstanceId,Tags[?Key==\`Name\`].Value|[0]]" \
        --output json)
    
    local count
    count=$(echo "$instances" | jq length)
    
    if [[ "$count" -eq 0 ]]; then
        log_error "No instance found with name '$identifier'" >&2
        return 1
    elif [[ "$count" -eq 1 ]]; then
        local instance_id
        instance_id=$(echo "$instances" | jq -r '.[0][0]')
        log_info "Found: $identifier → $instance_id" >&2
        echo "$instance_id"
        return 0
    else
        log_error "Multiple instances found with name '$identifier'. Please be more specific." >&2
        return 1
    fi
}

is_potential_instance_name() {
    local identifier="$1"
    
    # Exclude instance IDs from name validation
    if [[ "$identifier" =~ ^i-[a-zA-Z0-9]{8,}$ ]]; then
        return 1
    fi
    
    # Allow standard naming conventions for AWS resources
    if [[ "$identifier" =~ ^[a-zA-Z0-9._-]+$ ]]; then
        return 0
    fi
    
    return 1
}
