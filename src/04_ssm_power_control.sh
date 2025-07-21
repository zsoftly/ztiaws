#!/usr/bin/env bash

# 04_ssm_power_control.sh - Instance power state management functions

# Load utilities
if [ -n "${SCRIPT_DIR:-}" ] && [ -f "${SCRIPT_DIR}/src/00_utils.sh" ]; then
    source "${SCRIPT_DIR}/src/00_utils.sh"
elif [ -f "/usr/local/bin/src/00_utils.sh" ]; then
    source "/usr/local/bin/src/00_utils.sh"
else
    echo "[ERROR] 00_utils.sh not found. Exiting." >&2
    exit 1
fi

# Start an EC2 instance
start_instance() {
    local instance_id="$1"
    local region="$2"

    if [[ -z "$instance_id" || -z "$region" ]]; then
        log_error "Missing parameters for start_instance"
        return 1
    fi

    log_info "Starting instance: $instance_id in region: $region"
    aws ec2 start-instances --instance-ids "$instance_id" --region "$region" >/dev/null \
        && log_info "Instance $instance_id started successfully." \
        || log_error "Failed to start instance $instance_id"
}

# Stop an EC2 instance
stop_instance() {
    local instance_id="$1"
    local region="$2"

    if [[ -z "$instance_id" || -z "$region" ]]; then
        log_error "Missing parameters for stop_instance"
        return 1
    fi

    log_info "Stopping instance: $instance_id in region: $region"
    aws ec2 stop-instances --instance-ids "$instance_id" --region "$region" >/dev/null \
        && log_info "Instance $instance_id stopped successfully." \
        || log_error "Failed to stop instance $instance_id"
}

# Reboot an EC2 instance
reboot_instance() {
    local instance_id="$1"
    local region="$2"

    if [[ -z "$instance_id" || -z "$region" ]]; then
        log_error "Missing parameters for reboot_instance"
        return 1
    fi

    log_info "Rebooting instance: $instance_id in region: $region"
    aws ec2 reboot-instances --instance-ids "$instance_id" --region "$region" >/dev/null \
        && log_info "Instance $instance_id rebooted successfully." \
        || log_error "Failed to reboot instance $instance_id"
}

# Tag-based operations - power control multiple instances by tag

manage_instances_by_tag() {
    local action="$1"   # start | stop | reboot
    local region="$2"
    local tag_key="$3"
    local tag_value="$4"

    if [[ -z "$action" || -z "$region" || -z "$tag_key" || -z "$tag_value" ]]; then
        log_error "Missing parameters for manage_instances_by_tag"
        return 1
    fi

    log_info "Fetching instances with tag $tag_key=$tag_value in $region..."
    local instance_ids
    instance_ids=$(aws ec2 describe-instances \
        --filters "Name=tag:$tag_key,Values=$tag_value" "Name=instance-state-name,Values=running,stopped" \
        --region "$region" \
        --query 'Reservations[*].Instances[*].InstanceId' \
        --output text)

    if [[ -z "$instance_ids" ]]; then
        log_error "No instances found with tag $tag_key=$tag_value in $region."
        return 1
    fi

    log_info "Found instances: $instance_ids"

    for instance_id in $instance_ids; do
        case "$action" in
            start)
                start_instance "$instance_id" "$region"
                ;;
            stop)
                stop_instance "$instance_id" "$region"
                ;;
            reboot)
                reboot_instance "$instance_id" "$region"
                ;;
            *)
                log_error "Invalid action: $action"
                return 1
                ;;
        esac
    done
}
