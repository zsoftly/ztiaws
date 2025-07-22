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
# AMI ID for the EC2 instances. Default is Amazon Linux 2023 in ca-central-1 (ami-08379337a6fc559cd).
# Override by setting the AMI_ID environment variable.
# WARNING: This AMI ID is region-specific and may become outdated. Verify availability before use.
# To find the latest Amazon Linux 2023 AMI ID for your region, use:
# aws ec2 describe-images --owners amazon --filters "Name=name,Values=al2023-ami-*" --query 'Images | sort_by(@, &CreationDate) | [-1].ImageId' --output text
AMI_ID="${AMI_ID:-ami-08379337a6fc559cd}"

INSTANCE_TYPE="t2.micro"

# Subnet ID for EC2 instances. Default is for ca-central-1 region.
# REQUIRED: Override by setting the SUBNET_ID environment variable for your environment.
# To find available subnets in your region, use:
# aws ec2 describe-subnets --query 'Subnets[*].[SubnetId,VpcId,AvailabilityZone,CidrBlock]' --output table
SUBNET_ID="${SUBNET_ID:-subnet-0248e681d3349e41c}"

# Security Group ID for EC2 instances. Default is for ca-central-1 region.
# REQUIRED: Override by setting the SECURITY_GROUP environment variable for your environment.
# To find available security groups in your VPC, use:
# aws ec2 describe-security-groups --query 'SecurityGroups[*].[GroupId,GroupName,Description,VpcId]' --output table
SECURITY_GROUP="${SECURITY_GROUP:-sg-0a932ab2ea0066022}"

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
  -a, --ami-id AMI_ID      AMI ID to use (default: $AMI_ID)
  -s, --subnet-id SUBNET   Subnet ID to use (default: $SUBNET_ID)
  -g, --security-group SG  Security Group ID to use (default: $SECURITY_GROUP)
  -h, --help              Show this help

Environment Variables:
  AMI_ID           Override default AMI ID
  SUBNET_ID        Override default subnet ID (REQUIRED for different environments)
  SECURITY_GROUP   Override default security group ID (REQUIRED for different environments)

Examples:
  $0 create --count 3 --owner John    # Creates: John-web-server-1, John-web-server-2, John-web-server-3
  $0 create --owner Alice --subnet-id subnet-123abc --security-group sg-456def
  SUBNET_ID=subnet-123abc SECURITY_GROUP=sg-456def $0 create --owner Bob
  $0 verify
  $0 delete

Finding AWS Resources:
  # Find latest Amazon Linux 2023 AMI:
  aws ec2 describe-images --owners amazon --filters "Name=name,Values=al2023-ami-*" \\
    --query 'Images | sort_by(@, &CreationDate) | [-1].ImageId' --output text
  
  # List available subnets:
  aws ec2 describe-subnets --query 'Subnets[*].[SubnetId,VpcId,AvailabilityZone,CidrBlock]' --output table
  
  # List security groups:
  aws ec2 describe-security-groups --query 'SecurityGroups[*].[GroupId,GroupName,Description,VpcId]' --output table
EOF
}

# Check prerequisites once
check_prereq() {
    command -v aws >/dev/null || { log_error "AWS CLI not found"; exit 1; }
    command -v jq >/dev/null || { log_error "jq not found"; exit 1; }
    aws sts get-caller-identity >/dev/null 2>&1 || { log_error "AWS credentials invalid"; exit 1; }
}

# Validate AWS resources exist and are accessible
validate_aws_resources() {
    log_info "Validating AWS resources..."
    
    # Check if AMI exists and is available
    if ! aws ec2 describe-images --image-ids "$AMI_ID" --query 'Images[0].State' --output text 2>/dev/null | grep -q "available"; then
        log_error "AMI $AMI_ID is not available or doesn't exist in this region"
        log_info "To find the latest Amazon Linux 2023 AMI, run:"
        log_info "aws ec2 describe-images --owners amazon --filters \"Name=name,Values=al2023-ami-*\" --query 'Images | sort_by(@, &CreationDate) | [-1].ImageId' --output text"
        exit 1
    fi
    
    # Check if subnet exists
    if ! aws ec2 describe-subnets --subnet-ids "$SUBNET_ID" >/dev/null 2>&1; then
        log_error "Subnet $SUBNET_ID doesn't exist or is not accessible"
        log_info "To list available subnets, run:"
        log_info "aws ec2 describe-subnets --query 'Subnets[*].[SubnetId,VpcId,AvailabilityZone,CidrBlock]' --output table"
        exit 1
    fi
    
    # Check if security group exists
    if ! aws ec2 describe-security-groups --group-ids "$SECURITY_GROUP" >/dev/null 2>&1; then
        log_error "Security group $SECURITY_GROUP doesn't exist or is not accessible"
        log_info "To list available security groups, run:"
        log_info "aws ec2 describe-security-groups --query 'SecurityGroups[*].[GroupId,GroupName,Description,VpcId]' --output table"
        exit 1
    fi
    
    log_info "AWS resources validated successfully"
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
    
    # Validate required AWS resources configuration
    if [[ "$SUBNET_ID" == "subnet-0248e681d3349e41c" ]]; then
        log_warn "Using default subnet ID for ca-central-1. For other regions, set SUBNET_ID environment variable or use --subnet-id option."
    fi
    
    if [[ "$SECURITY_GROUP" == "sg-0a932ab2ea0066022" ]]; then
        log_warn "Using default security group for ca-central-1. For other regions, set SECURITY_GROUP environment variable or use --security-group option."
    fi
    
    if [[ "$AMI_ID" == "ami-08379337a6fc559cd" ]]; then
        log_warn "Using default AMI ID for Amazon Linux 2023 in ca-central-1. This may become outdated. Consider updating or using --ami-id option."
    fi
    
    # Validate AWS resources before proceeding
    validate_aws_resources
    
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
            -a|--ami-id)
                AMI_ID="$2"
                [[ "$AMI_ID" =~ ^ami-[0-9a-f]{8,17}$ ]] || { log_error "Invalid AMI ID format: $AMI_ID"; exit 1; }
                shift 2 ;;
            -s|--subnet-id)
                SUBNET_ID="$2"
                [[ "$SUBNET_ID" =~ ^subnet-[0-9a-f]{8,17}$ ]] || { log_error "Invalid subnet ID format: $SUBNET_ID"; exit 1; }
                shift 2 ;;
            -g|--security-group)
                SECURITY_GROUP="$2"
                [[ "$SECURITY_GROUP" =~ ^sg-[0-9a-f]{8,17}$ ]] || { log_error "Invalid security group ID format: $SECURITY_GROUP"; exit 1; }
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