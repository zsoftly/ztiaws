#!/bin/bash

# Efficient EC2 Test Instance Manager
# Usage: ./tools/01_ec2_test_manager.sh [create|verify|delete] [options]

# Get script directory and source utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source utilities from the parent src directory
if [ -f "${SCRIPT_DIR}/../src/00_utils.sh" ]; then
    # shellcheck source=../src/00_utils.sh
    source "${SCRIPT_DIR}/../src/00_utils.sh"
else
    echo "[ERROR] src/00_utils.sh not found. This script requires 00_utils.sh for logging functions." >&2
    echo "Please ensure src/00_utils.sh is present in the ../src/ directory." >&2
    exit 1
fi

# Initialize logging (console-only by default, can be enabled with ENABLE_EC2_MANAGER_LOGGING=true env var)
init_logging "ec2-manager" "${ENABLE_EC2_MANAGER_LOGGING:-false}"

# Configuration
AMI_ID="ami-08379337a6fc559cd"
INSTANCE_TYPE="t2.micro"
SUBNET_ID="subnet-0248e681d3349e41c"
SECURITY_GROUP="sg-0a932ab2ea0066022"
IAM_ROLE_NAME="EC2-SSM-Role"
IAM_INSTANCE_PROFILE="EC2-SSM-InstanceProfile"
INSTANCE_FILE="ec2-instances.txt"

# Default values
COUNT=1
OWNER=""
NAME_PREFIX="web-server"

show_help() {
    cat << EOF
EC2 Test Instance Manager - Efficient Version

Usage: $0 <command> [options]

Commands:
  create    Create EC2 instances
  verify    Verify and show instance status
  delete    Delete all tracked instances

Options:
  -c, --count NUMBER       Number of instances to create (default: 1)
  -o, --owner NAME         Set owner tag (required for create, used as name prefix)
  -n, --name-prefix NAME   Instance name suffix (default: web-server)
  -h, --help              Show this help

Examples:
  $0 create --count 3 --owner John    # Creates: John-web-server-1, John-web-server-2, John-web-server-3
  $0 verify
  $0 delete
EOF
}

# Check prerequisites once
check_prereq() {
    command -v aws >/dev/null || { log_error "AWS CLI not found"; exit 1; }
    command -v jq >/dev/null || { log_error "jq not found"; exit 1; }
    aws sts get-caller-identity >/dev/null 2>&1 || { log_error "AWS credentials invalid"; exit 1; }
}

# Ensure IAM resources exist (cached check)
ensure_iam() {
    local cache_file=".iam_ready"
    
    # Skip if recently checked
    if [[ -f "$cache_file" && $(($(date +%s) - $(stat -c %Y "$cache_file" 2>/dev/null || echo 0))) -lt 300 ]]; then
        return 0
    fi
    
    # Check and create role if needed
    if ! aws iam get-role --role-name "$IAM_ROLE_NAME" >/dev/null 2>&1; then
        log_info "Creating IAM role..."
        aws iam create-role --role-name "$IAM_ROLE_NAME" \
            --assume-role-policy-document '{
                "Version": "2012-10-17",
                "Statement": [{
                    "Effect": "Allow",
                    "Principal": {"Service": "ec2.amazonaws.com"},
                    "Action": "sts:AssumeRole"
                }]
            }' >/dev/null || { log_error "Failed to create IAM role"; exit 1; }
        
        aws iam attach-role-policy --role-name "$IAM_ROLE_NAME" \
            --policy-arn "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore" >/dev/null
    fi
    
    # Check and create instance profile if needed
    if ! aws iam get-instance-profile --instance-profile-name "$IAM_INSTANCE_PROFILE" >/dev/null 2>&1; then
        log_info "Creating instance profile..."
        aws iam create-instance-profile --instance-profile-name "$IAM_INSTANCE_PROFILE" >/dev/null || { log_error "Failed to create instance profile"; exit 1; }
        aws iam add-role-to-instance-profile --instance-profile-name "$IAM_INSTANCE_PROFILE" --role-name "$IAM_ROLE_NAME" >/dev/null
        sleep 5  # Brief wait for AWS propagation
    fi
    
    touch "$cache_file"
}

# Create instances efficiently
create_instances() {
    [[ -n "$OWNER" ]] || { log_error "Owner required (use -o/--owner)"; exit 1; }
    
    ensure_iam
    log_info "Creating $COUNT instance(s)..."
    
    # Prepare batch creation
    local instances=()
    for i in $(seq 1 "$COUNT"); do
        local name="$OWNER-$NAME_PREFIX"
        [[ $COUNT -gt 1 ]] && name="$name-$i"
        
        # Create instance
        local result
        result=$(aws ec2 run-instances \
            --image-id "$AMI_ID" \
            --instance-type "$INSTANCE_TYPE" \
            --count 1 \
            --subnet-id "$SUBNET_ID" \
            --security-group-ids "$SECURITY_GROUP" \
            --no-associate-public-ip-address \
            --iam-instance-profile Name="$IAM_INSTANCE_PROFILE" \
            --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=$name},{Key=Owner,Value=$OWNER},{Key=ManagedBy,Value=ec2-manager}]" \
            --credit-specification CpuCredits=standard \
            --metadata-options HttpTokens=required \
            --output json 2>/dev/null) || { log_error "Failed to create instance $name"; exit 1; }
        
        local instance_id
        instance_id=$(echo "$result" | jq -r '.Instances[0].InstanceId')
        instances+=("$instance_id")
        echo "$instance_id" >> "$INSTANCE_FILE"
        log_info "$name: $instance_id"
    done
    
    log_info "Created ${#instances[@]} instance(s)"
    verify_instances
}

# Verify instances with efficient batch query
verify_instances() {
    [[ -f "$INSTANCE_FILE" && -s "$INSTANCE_FILE" ]] || { log_warn "No instances tracked"; return 0; }
    
    # Read all instance IDs
    local instance_ids
    mapfile -t instance_ids < <(grep -v '^[[:space:]]*$' "$INSTANCE_FILE")
    [[ ${#instance_ids[@]} -gt 0 ]] || { log_warn "No valid instance IDs found"; return 0; }
    
    log_info "Checking ${#instance_ids[@]} instance(s)..."
    
    # Batch query all instances
    local result
    result=$(aws ec2 describe-instances \
        --instance-ids "${instance_ids[@]}" \
        --query 'Reservations[].Instances[].[InstanceId,State.Name,Tags[?Key==`Name`].Value|[0],Tags[?Key==`Owner`].Value|[0]]' \
        --output json 2>/dev/null)
    
    if [[ $? -eq 0 && "$result" != "null" ]]; then
        echo "$result" | jq -r '.[] | @tsv' | while IFS=$'\t' read -r id state name owner; do
            local status_color=""
            case "$state" in
                running) status_color="$GREEN" ;;
                pending) status_color="$YELLOW" ;;
                terminated|terminating) status_color="$RED" ;;
            esac
            printf "%-19s ${status_color}%-12s${NC} %-20s %s\n" "$id" "$state" "${name:-N/A}" "${owner:-N/A}"
        done
    else
        log_warn "Failed to get instance details"
        # Fallback: show IDs only
        printf '%s\n' "${instance_ids[@]}"
    fi
}

# Delete instances efficiently
delete_instances() {
    [[ -f "$INSTANCE_FILE" && -s "$INSTANCE_FILE" ]] || { log_warn "No instances to delete"; return 0; }
    
    # Read and validate instance IDs
    local instance_ids
    mapfile -t instance_ids < <(grep -v '^[[:space:]]*$' "$INSTANCE_FILE")
    [[ ${#instance_ids[@]} -gt 0 ]] || { log_warn "No valid instances found"; return 0; }
    
    log_info "Terminating ${#instance_ids[@]} instance(s)..."
    
    # Batch terminate
    if aws ec2 terminate-instances --instance-ids "${instance_ids[@]}" >/dev/null 2>&1; then
        rm -f "$INSTANCE_FILE" ".iam_ready"
        log_info "Terminated all instances and cleaned up"
    else
        log_error "Failed to terminate instances"
        exit 1
    fi
}

# Parse arguments
parse_args() {
    [[ $# -eq 0 ]] && { show_help; exit 1; }
    
    # Handle help first
    case "$1" in
        -h|--help|help)
            show_help
            exit 0 ;;
    esac
    
    local command="$1"
    shift
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -c|--count)
                COUNT="$2"
                [[ "$COUNT" =~ ^[1-9][0-9]*$ ]] || { log_error "Count must be positive integer"; exit 1; }
                shift 2 ;;
            -o|--owner)
                OWNER="$2"
                shift 2 ;;
            -n|--name-prefix)
                NAME_PREFIX="$2"
                shift 2 ;;
            -h|--help)
                show_help; exit 0 ;;
            *)
                log_error "Unknown option: $1"
                exit 1 ;;
        esac
    done
    
    case "$command" in
        create) create_instances ;;
        verify) verify_instances ;;
        delete) delete_instances ;;
        *) 
            log_error "Unknown command: $command. Use create, verify, or delete"
            exit 1 ;;
    esac
}

# Main execution
check_prereq
parse_args "$@"

# Log completion marker when done
log_completion "ec2-manager"