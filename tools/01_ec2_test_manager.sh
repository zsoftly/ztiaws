#!/bin/bash

# Efficient EC2 Test Instance Manager
# Usage: ./tools/01_ec2_test_manager.sh [create|verify|delete] [options]
# 
# Enhanced Features:
# - Auto-detects latest Linux (Amazon Linux 2023) and Windows (Server 2022) AMIs
# - Supports creating both Linux and Windows instances by default
# - Automatically upgrades Windows instances to t3.small for better performance
# - OS type selection: linux, windows, both (default: both)
# - Improved instance naming with OS type identification
# - Enhanced verification display with OS type column

# Get script directory and source utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source utilities from the parent src directory
if [ -f "${SCRIPT_DIR}/../src/00_utils.sh" ]; then
    # shellcheck source=../src/00_utils.sh disable=SC1091
    source "${SCRIPT_DIR}/../src/00_utils.sh"
else
    echo "[ERROR] src/00_utils.sh not found. This script requires 00_utils.sh for logging functions." >&2
    echo "Please ensure src/00_utils.sh is present in the ../src/ directory." >&2
    exit 1
fi

# Initialize logging (console-only by default, can be enabled with ENABLE_EC2_MANAGER_LOGGING=true env var)
init_logging "ec2-manager" "${ENABLE_EC2_MANAGER_LOGGING:-false}"

# Configuration
# AMI ID for the EC2 instances. Leave empty for auto-detection (recommended).
# Override by setting the AMI_ID environment variable to force a specific AMI for all instances.
# When AMI_ID is set, the same AMI will be used for both Linux and Windows instances.
# For mixed OS deployments, leave AMI_ID unset to enable per-OS auto-detection.
AMI_ID="${AMI_ID:-}"

INSTANCE_TYPE="t2.micro"

# Subnet ID for EC2 instances. Default is for ca-central-1 region.
# REQUIRED: Override by setting the SUBNET_ID environment variable for your environment.
# To find available subnets in your region, use:
# aws ec2 describe-subnets --query 'Subnets[*].[SubnetId,VpcId,AvailabilityZone,CidrBlock]' --output table
# The subnet has access to the internet via a NAT gateway
# If they are attached to a public route table with an internet gateway, use --associate-public-ip-address
SUBNET_ID="${SUBNET_ID:-""}"

# Security Group ID for EC2 instances. Default is for ca-central-1 region.
# REQUIRED: Override by setting the SECURITY_GROUP environment variable for your environment.
# To find available security groups in your VPC, use:
# aws ec2 describe-security-groups --query 'SecurityGroups[*].[GroupId,GroupName,Description,VpcId]' --output table
# Has Outbound: HTTPS (443) to 0.0.0.0/0
# Have SSM VPC endpoint or allow outbound internet access for SSM to work
SECURITY_GROUP="${SECURITY_GROUP:-""}"

IAM_ROLE_NAME="EC2-SSM-Role"
IAM_INSTANCE_PROFILE="EC2-SSM-InstanceProfile"
INSTANCE_FILE="ec2-instances.txt"

# Default values
COUNT=1
OWNER=""
NAME_PREFIX="web-server"
OS_TYPE="both"  # Options: linux, windows, both
ASSIGN_PUBLIC_IP=false  # Default: no public IP (use NAT/VPC endpoints for SSM)

# Auto-detect latest AMI IDs for Linux and Windows
get_latest_linux_ami() {
    local ami_id
    # Use Amazon Linux 2023 FULL AMI (not minimal) which has SSM agent pre-installed
    # The standard AL2023 AMI pattern excludes minimal versions
    ami_id=$(aws ec2 describe-images \
        --owners amazon \
        --filters "Name=name,Values=al2023-ami-2*kernel*" "Name=architecture,Values=x86_64" "Name=state,Values=available" \
        --query 'Images | sort_by(@, &CreationDate) | [-1].ImageId' \
        --output text 2>/dev/null)
    
    if [[ "$ami_id" != "None" && -n "$ami_id" ]]; then
        log_info "Found latest Amazon Linux 2023 AMI with SSM: $ami_id"
        echo "$ami_id"
        return 0
    else
        log_warn "Failed to detect latest Linux AMI. Using fallback AMI for ca-central-1."
        echo "ami-08379337a6fc559cd"  # Fallback Amazon Linux 2023 for ca-central-1
        return 0
    fi
}

get_latest_windows_ami() {
    local ami_id
    # Use Windows Server 2022 which has SSM agent pre-installed
    ami_id=$(aws ec2 describe-images \
        --owners amazon \
        --filters "Name=name,Values=Windows_Server-2022-English-Full-Base-*" "Name=architecture,Values=x86_64" "Name=state,Values=available" \
        --query 'Images | sort_by(@, &CreationDate) | [-1].ImageId' \
        --output text 2>/dev/null)
    
    if [[ "$ami_id" != "None" && -n "$ami_id" ]]; then
        log_info "Found latest Windows Server 2022 AMI with SSM: $ami_id"
        echo "$ami_id"
        return 0
    else
        log_error "Failed to detect Windows Server 2022 AMI in current region."
        log_info "Windows AMIs may not be available in all regions. To find available Windows AMIs:"
        log_info "aws ec2 describe-images --owners amazon --filters \"Name=platform,Values=windows\" --query 'Images[*].{Name:Name,ImageId:ImageId}' --output table"
        return 1
    fi
}

# Get AMI ID based on OS type
get_ami_for_os() {
    local os_type="$1"
    case "$os_type" in
        linux)
            get_latest_linux_ami
            ;;
        windows)
            get_latest_windows_ami
            ;;
        *)
            log_error "Invalid OS type: $os_type. Use 'linux' or 'windows'"
            return 1
            ;;
    esac
}

show_help() {
    cat << 'EOF'
EC2 Test Instance Manager - Cross-Platform Version

Usage: $0 <command> [options]

Commands:
  create    Create EC2 instances
  verify    Verify and show instance status
  delete    Delete all tracked instances

Options:
  -c, --count NUMBER       Number of instances to create (default: 1)
  -o, --owner NAME         Set owner tag (required for create, used as name prefix)
  -n, --name-prefix NAME   Instance name suffix (default: web-server)
  -t, --os-type TYPE       OS type: linux, windows, both (default: both)
  -a, --ami-id AMI_ID      AMI ID to use (overrides auto-detection)
  -s, --subnet-id SUBNET   Override auto-discovered Subnet ID
  -g, --security-group SG  Override auto-discovered Security Group ID
  -p, --public-ip          Assign public IP addresses (default: private IPs)
  -h, --help              Show this help

Environment Variables:
  AMI_ID           Override default AMI ID
  SUBNET_ID        Override auto-discovered subnet ID
  SECURITY_GROUP   Override auto-discovered security group ID

AWS Resource Auto-Discovery:
  The script automatically discovers the following resources using tags:
  - VPC:           tag:Name=ztiaws-poc-vpc
  - Subnet:        tag:Name=ztiaws-poc-vpc-public-subnet-1
  - Security Group: tag:Name=ztiaws-poc-sg

Examples:
  $0 create --owner John               # Uses auto-discovered network resources
  $0 create --owner Bob --subnet-id subnet-123 --security-group sg-456
EOF
}

#---
# Function to discover VPC, Subnet, and Security Group IDs based on tags
#---
discover_network_resources() {
  echo "--> Auto-discovering network resources from AWS..."

  # 1. Discover the VPC ID using its Name tag
  VPC_ID=$(aws ec2 describe-vpcs --filters "Name=tag:Name,Values=ztiaws-poc-vpc" --query "Vpcs[0].VpcId" --output text)
  if [ -z "$VPC_ID" ] || [ "$VPC_ID" == "None" ]; then
    echo "Error: Could not find the VPC with tag Name=ztiaws-poc-vpc." >&2
    exit 1
  fi
  echo "    Found VPC: $VPC_ID"

  # 2. Discover the Subnet ID if not provided by the user
  if [ -z "$SUBNET_ID" ]; then
    SUBNET_ID=$(aws ec2 describe-subnets --filters "Name=vpc-id,Values=$VPC_ID" "Name=tag:Name,Values=ztiaws-poc-vpc-public-subnet-1" --query "Subnets[0].SubnetId" --output text)
    if [ -z "$SUBNET_ID" ] || [ "$SUBNET_ID" == "None" ]; then
      echo "Error: Could not find the Subnet with tag Name=ztiaws-poc-vpc-public-subnet-1 in VPC $VPC_ID." >&2
      exit 1
    fi
    echo "    Auto-detected Subnet ID: $SUBNET_ID"
  fi

  # 3. Discover the Security Group ID if not provided by the user
  if [ -z "$SECURITY_GROUP" ]; then
    SECURITY_GROUP=$(aws ec2 describe-security-groups --filters "Name=vpc-id,Values=$VPC_ID" "Name=tag:Name,Values=ztiaws-poc-sg" --query "SecurityGroups[0].GroupId" --output text)
    if [ -z "$SECURITY_GROUP" ] || [ "$SECURITY_GROUP" == "None" ]; then
      echo "Error: Could not find the Security Group with tag Name=ztiaws-poc-sg in VPC $VPC_ID." >&2
      exit 1
    fi
    echo "    Auto-detected Security Group ID: $SECURITY_GROUP"
  fi
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
    
    # Check if manually specified AMI exists and is available (skip if auto-detecting)
    if [[ -n "$AMI_ID" ]]; then
        if ! aws ec2 describe-images --image-ids "$AMI_ID" --query 'Images[0].State' --output text 2>/dev/null | grep -q "available"; then
            log_error "Manually specified AMI $AMI_ID is not available or doesn't exist in this region"
            log_info "To find the latest Amazon Linux 2023 AMI, run:"
            log_info "aws ec2 describe-images --owners amazon --filters \"Name=name,Values=al2023-ami-*\" --query 'Images | sort_by(@, &CreationDate) | [-1].ImageId' --output text"
            exit 1
        fi
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

# Cross-platform function to read instance IDs from file
read_instance_ids() {
    local instance_ids=()
    local line
    
    if [[ ! -f "$INSTANCE_FILE" || ! -s "$INSTANCE_FILE" ]]; then
        return 1
    fi
    
    # Read file line by line, filtering out empty lines
    while IFS= read -r line; do
        # Skip empty lines and whitespace-only lines
        if [[ -n "${line// }" ]]; then
            instance_ids+=("$line")
        fi
    done < "$INSTANCE_FILE"
    
    # Return the array via a global variable
    INSTANCE_IDS=("${instance_ids[@]}")
    return 0
}
get_file_mtime() {
    local file="$1"
    if [[ ! -f "$file" ]]; then
        echo 0
        return
    fi
    
    # Cross-platform stat command for modification time
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        stat -f %m "$file" 2>/dev/null || echo 0
    else
        # Linux and others
        stat -c %Y "$file" 2>/dev/null || echo 0
    fi
}

# Ensure IAM resources exist (cached check)
ensure_iam() {
    local cache_file=".iam_ready"
    
    # Skip if recently checked (within 5 minutes)
    local current_time
    current_time=$(date +%s)
    local file_mtime
    file_mtime=$(get_file_mtime "$cache_file")
    if [[ -f "$cache_file" && $((current_time - file_mtime)) -lt 300 ]]; then
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
    
    # Auto-discover network resources if not provided
    discover_network_resources
    
    # Validate AWS resources before proceeding
    validate_aws_resources
    ensure_iam
    
    # Determine which OS types to create
    local os_types=()
    case "$OS_TYPE" in
        linux)
            os_types=("linux")
            ;;
        windows)
            os_types=("windows")
            ;;
        both)
            os_types=("linux" "windows")
            ;;
    esac
    
    # Calculate total instances to create
    local total_instances=$((COUNT * ${#os_types[@]}))
    log_info "Creating $total_instances instance(s) ($COUNT per OS type: ${os_types[*]})"
    
    # Prepare batch creation
    local instances=()
    local counter=1
    
    for os_type in "${os_types[@]}"; do
        # Get AMI ID for this OS type (unless manually specified)
        local ami_id="$AMI_ID"
        if [[ -z "$AMI_ID" ]]; then
            log_info "Auto-detecting latest $os_type AMI..."
            ami_id=$(get_ami_for_os "$os_type")
            if [[ $? -ne 0 || -z "$ami_id" ]]; then
                log_error "Failed to get AMI for $os_type"
                continue
            fi
            log_info "Using $os_type AMI: $ami_id"
        fi
        
        # Create instances for this OS type
        for i in $(seq 1 "$COUNT"); do
            local name="$OWNER-$NAME_PREFIX-$os_type"
            if [[ $COUNT -gt 1 ]]; then
                name="$name-$i"
            elif [[ "${#os_types[@]}" -gt 1 ]]; then
                # When creating both OS types with count=1, still append counter for uniqueness
                name="$name-$counter"
            fi
            
            # Set instance type based on OS (Windows typically needs more resources)
            local instance_type="$INSTANCE_TYPE"
            if [[ "$os_type" == "windows" && "$INSTANCE_TYPE" == "t2.micro" ]]; then
                instance_type="t3.small"  # Windows needs more resources than t2.micro
                log_info "Using $instance_type for Windows instance (upgraded from t2.micro)"
            fi
            
            # Create instance with dynamic public IP assignment
            local public_ip_flag
            if [[ "$ASSIGN_PUBLIC_IP" == "true" ]]; then
                public_ip_flag="--associate-public-ip-address"
                log_info "Creating $name with public IP address"
            else
                public_ip_flag="--no-associate-public-ip-address"
                log_info "Creating $name with private IP only (requires NAT/VPC endpoints for SSM)"
            fi
            
            local result
            result=$(aws ec2 run-instances \
                --image-id "$ami_id" \
                --instance-type "$instance_type" \
                --count 1 \
                --subnet-id "$SUBNET_ID" \
                --security-group-ids "$SECURITY_GROUP" \
                $public_ip_flag \
                --iam-instance-profile Name="$IAM_INSTANCE_PROFILE" \
                --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=$name},{Key=Owner,Value=$OWNER},{Key=ManagedBy,Value=ec2-manager},{Key=OSType,Value=$os_type},{Key=Platform,Value=$os_type},{Key=Environment,Value=testing},{Key=SSMEnabled,Value=true}]" \
                --credit-specification CpuCredits=standard \
                --metadata-options HttpTokens=required \
                --output json 2>/dev/null) || { log_error "Failed to create instance $name"; continue; }
            
            local instance_id
            instance_id=$(echo "$result" | jq -r '.Instances[0].InstanceId')
            instances+=("$instance_id")
            echo "$instance_id" >> "$INSTANCE_FILE"
            log_info "$name ($os_type): $instance_id"
            
            ((counter++))
        done
    done
    
    log_info "Created ${#instances[@]} instance(s) successfully"
    if [[ ${#instances[@]} -lt $total_instances ]]; then
        log_warn "Some instances failed to create. Check the logs above."
    fi
    
    verify_instances
}

# Verify instances with efficient batch query
verify_instances() {
    # Read all instance IDs using cross-platform function
    if ! read_instance_ids || [[ ${#INSTANCE_IDS[@]} -eq 0 ]]; then
        log_warn "No instances tracked or no valid instance IDs found"
        return 0
    fi
    
    log_info "Checking ${#INSTANCE_IDS[@]} instance(s)..."
    
    # Batch query all instances
    local result
    # shellcheck disable=SC2016
    # The backticks here are part of AWS CLI's JMESPath syntax, not shell command substitution
    result=$(aws ec2 describe-instances \
        --instance-ids "${INSTANCE_IDS[@]}" \
        --query 'Reservations[].Instances[].[InstanceId,State.Name,Tags[?Key==`Name`].Value|[0],Tags[?Key==`Owner`].Value|[0],Tags[?Key==`OSType`].Value|[0]]' \
        --output json 2>/dev/null)
    
    if [[ $? -eq 0 && "$result" != "null" ]]; then
        printf "%-19s %-12s %-30s %-15s %s\n" "INSTANCE-ID" "STATE" "NAME" "OWNER" "OS-TYPE"
        printf "%-19s %-12s %-30s %-15s %s\n" "-------------------" "------------" "------------------------------" "---------------" "-------"
        echo "$result" | jq -r '.[] | @tsv' | while IFS=$'\t' read -r id state name owner ostype; do
            local status_color=""
            case "$state" in
                running) status_color="$GREEN" ;;
                pending) status_color="$YELLOW" ;;
                terminated|terminating) status_color="$RED" ;;
            esac
            printf "%-19s ${status_color}%-12s${NC} %-30s %-15s %s\n" "$id" "$state" "${name:-N/A}" "${owner:-N/A}" "${ostype:-N/A}"
        done
    else
        log_warn "Failed to get instance details"
        # Fallback: show IDs only
        printf '%s\n' "${INSTANCE_IDS[@]}"
    fi
}

# Delete instances efficiently
delete_instances() {
    # Read all instance IDs using cross-platform function
    if ! read_instance_ids || [[ ${#INSTANCE_IDS[@]} -eq 0 ]]; then
        log_warn "No instances to delete or no valid instance IDs found"
        return 0
    fi
    
    log_info "Terminating ${#INSTANCE_IDS[@]} instance(s)..."
    
    # Batch terminate
    if aws ec2 terminate-instances --instance-ids "${INSTANCE_IDS[@]}" >/dev/null 2>&1; then
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
            exit 0 
            ;;
    esac
    
    local command="$1"
    shift
    
    # Parse remaining arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -c|--count)
                COUNT="$2"
                [[ "$COUNT" =~ ^[1-9][0-9]*$ ]] || { log_error "Count must be positive integer"; exit 1; }
                shift 2 
                ;;
            -o|--owner)
                OWNER="$2"
                shift 2 
                ;;
            -n|--name-prefix)
                NAME_PREFIX="$2"
                shift 2 
                ;;
            -t|--os-type)
                OS_TYPE="$2"
                [[ "$OS_TYPE" =~ ^(linux|windows|both)$ ]] || { log_error "Invalid OS type: $OS_TYPE. Use 'linux', 'windows', or 'both'"; exit 1; }
                shift 2 
                ;;
            -a|--ami-id)
                AMI_ID="$2"
                [[ "$AMI_ID" =~ ^ami-[0-9a-f]{8,17}$ ]] || { log_error "Invalid AMI ID format: $AMI_ID"; exit 1; }
                shift 2 
                ;;
            -s|--subnet-id)
                SUBNET_ID="$2"
                [[ "$SUBNET_ID" =~ ^subnet-[0-9a-f]{8,17}$ ]] || { log_error "Invalid subnet ID format: $SUBNET_ID"; exit 1; }
                shift 2 
                ;;
            -g|--security-group)
                SECURITY_GROUP="$2"
                [[ "$SECURITY_GROUP" =~ ^sg-[0-9a-f]{8,17}$ ]] || { log_error "Invalid security group ID format: $SECURITY_GROUP"; exit 1; }
                shift 2 
                ;;
            -p|--public-ip)
                ASSIGN_PUBLIC_IP=true
                shift 
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1 
                ;;
        esac
    done
    
    case "$command" in
        create) create_instances ;;
        verify) verify_instances ;;
        delete) delete_instances ;;
        *) 
            log_error "Unknown command: $command. Use create, verify, or delete"
            exit 1 
            ;;
    esac
}

# Main execution
check_prereq
parse_args "$@"

# Log completion marker when done
log_completion "ec2-manager"