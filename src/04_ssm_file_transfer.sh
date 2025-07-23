#!/usr/bin/env bash

# AWS SSM File Transfer Module
# Handles file upload/download operations via SSM
# This module is sourced by the main ssm script
#
# Security and Configuration Notes:
# - Policy names use unique identifiers (timestamp + hostname + random) instead of PIDs
# - Temporary policy files use secure permissions (600) and proper temp directories
# - IAM propagation delay is configurable via environment variable:
#   - IAM_PROPAGATION_DELAY: Delay for IAM changes to propagate (default: 5 seconds)
# - Emergency cleanup uses robust pattern matching for instance ID extraction
# - Filesystem locking prevents concurrent IAM policy operations on the same instance
# - Locks are automatically cleaned up and include stale lock detection (5 minute timeout)

# File size threshold for choosing transfer method (1MB)
FILE_SIZE_THRESHOLD=$((1024 * 1024))

# S3 bucket for large file transfers (will be created if needed)
S3_BUCKET_PREFIX="ztiaws-ssm-file-transfer"

# Global variable to track current region for cleanup purposes
# This should be set by the main script or via set_file_transfer_region()
# No default value - region must be explicitly provided
CURRENT_REGION=""

# Configuration for IAM propagation wait time
IAM_PROPAGATION_DELAY="${IAM_PROPAGATION_DELAY:-5}"  # Default 5 seconds

# Lock directory for preventing concurrent policy operations on the same instance
LOCK_DIR="/tmp/ztiaws-locks"

# Dedicated directory for ztiaws temporary files to avoid scanning entire /tmp
ZTIAWS_TEMP_DIR="/tmp/ztiaws-temp"

# Registry file to track active policies for efficient cleanup
POLICY_REGISTRY_FILE="${ZTIAWS_TEMP_DIR}/policy-registry"

# Generate a unique identifier for policy names and temp files
# Combines timestamp, hostname, and random number for uniqueness
generate_unique_id() {
    local timestamp
    local hostname
    local random
    timestamp=$(date +%s)
    hostname=$(hostname -s 2>/dev/null || echo "unknown")
    random=$(od -An -N2 -tu2 < /dev/urandom | tr -d ' ')
    echo "${timestamp}-${hostname}-${random}"
}

# Initialize the ztiaws temporary directory and registry
init_temp_directory() {
    # Ensure ztiaws temp directory exists with proper permissions
    if ! mkdir -p "$ZTIAWS_TEMP_DIR" 2>/dev/null; then
        log_warn "Could not create ztiaws temp directory: $ZTIAWS_TEMP_DIR"
        return 1
    fi
    chmod 700 "$ZTIAWS_TEMP_DIR" 2>/dev/null || true
    
    # Ensure lock directory exists
    mkdir -p "$LOCK_DIR" 2>/dev/null || true
    
    # Initialize registry file if it doesn't exist
    if [[ ! -f "$POLICY_REGISTRY_FILE" ]]; then
        touch "$POLICY_REGISTRY_FILE"
        chmod 600 "$POLICY_REGISTRY_FILE" 2>/dev/null || true
    fi
    
    return 0
}

# Add a policy to the registry
# Format: instance_id|region|policy_arn|policy_file|timestamp
add_policy_to_registry() {
    local instance_id="$1"
    local region="$2"
    local policy_arn="$3"
    local policy_file="$4"
    local timestamp
    timestamp=$(date +%s)
    
    init_temp_directory || return 1
    
    # Use a lock file to prevent concurrent registry modifications
    local registry_lock="${ZTIAWS_TEMP_DIR}/.registry.lock"
    local lock_acquired=false
    local attempts=0
    local max_attempts=10
    
    while [ $attempts -lt $max_attempts ]; do
        if mkdir "$registry_lock" 2>/dev/null; then
            lock_acquired=true
            break
        fi
        sleep 0.1
        attempts=$((attempts + 1))
    done
    
    if [[ "$lock_acquired" == "true" ]]; then
        echo "${instance_id}|${region}|${policy_arn}|${policy_file}|${timestamp}" >> "$POLICY_REGISTRY_FILE"
        rmdir "$registry_lock" 2>/dev/null || true
        debug_log "Added policy to registry: $policy_arn for instance $instance_id"
        return 0
    else
        log_warn "Failed to acquire registry lock for adding policy"
        return 1
    fi
}

# Remove policies from registry for a specific instance
remove_policies_from_registry() {
    local instance_id="$1"
    
    init_temp_directory || return 1
    
    # Use a lock file to prevent concurrent registry modifications
    local registry_lock="${ZTIAWS_TEMP_DIR}/.registry.lock"
    local lock_acquired=false
    local attempts=0
    local max_attempts=10
    
    while [ $attempts -lt $max_attempts ]; do
        if mkdir "$registry_lock" 2>/dev/null; then
            lock_acquired=true
            break
        fi
        sleep 0.1
        attempts=$((attempts + 1))
    done
    
    if [[ "$lock_acquired" == "true" ]]; then
        # Create a temporary file with entries not matching the instance_id
        local temp_registry="${ZTIAWS_TEMP_DIR}/.registry.tmp"
        if [[ -f "$POLICY_REGISTRY_FILE" ]]; then
            grep -v "^${instance_id}|" "$POLICY_REGISTRY_FILE" > "$temp_registry" 2>/dev/null || true
            mv "$temp_registry" "$POLICY_REGISTRY_FILE"
        fi
        rmdir "$registry_lock" 2>/dev/null || true
        debug_log "Removed policies from registry for instance: $instance_id"
        return 0
    else
        log_warn "Failed to acquire registry lock for removing policies"
        return 1
    fi
}

# Get policies for a specific instance from registry
get_policies_for_instance() {
    local instance_id="$1"
    
    if [[ -f "$POLICY_REGISTRY_FILE" ]]; then
        grep "^${instance_id}|" "$POLICY_REGISTRY_FILE" 2>/dev/null || true
    fi
}

# Clean up stale registry entries (older than 24 hours)
cleanup_stale_registry_entries() {
    init_temp_directory || return 1
    
    local registry_lock="${ZTIAWS_TEMP_DIR}/.registry.lock"
    local lock_acquired=false
    local attempts=0
    local max_attempts=10
    
    while [ $attempts -lt $max_attempts ]; do
        if mkdir "$registry_lock" 2>/dev/null; then
            lock_acquired=true
            break
        fi
        sleep 0.1
        attempts=$((attempts + 1))
    done
    
    if [[ "$lock_acquired" == "true" ]]; then
        if [[ -f "$POLICY_REGISTRY_FILE" ]]; then
            local current_time
            current_time=$(date +%s)
            local temp_registry="${ZTIAWS_TEMP_DIR}/.registry.tmp"
            local cleaned_entries=0
            
            while IFS='|' read -r instance_id region policy_arn policy_file timestamp; do
                # Skip empty lines
                [[ -z "$instance_id" ]] && continue
                
                # Check if entry is older than 24 hours (86400 seconds)
                local age=$((current_time - timestamp))
                if [ $age -gt 86400 ]; then
                    debug_log "Removing stale registry entry for instance $instance_id (age: ${age}s)"
                    # Clean up associated policy file if it exists
                    [[ -f "$policy_file" ]] && rm -f "$policy_file"
                    ((cleaned_entries++))
                else
                    # Keep this entry
                    echo "${instance_id}|${region}|${policy_arn}|${policy_file}|${timestamp}" >> "$temp_registry"
                fi
            done < "$POLICY_REGISTRY_FILE"
            
            # Replace registry with cleaned version
            if [[ -f "$temp_registry" ]]; then
                mv "$temp_registry" "$POLICY_REGISTRY_FILE"
            else
                # No entries left, create empty registry
                : > "$POLICY_REGISTRY_FILE"
            fi
            
            if [ $cleaned_entries -gt 0 ]; then
                debug_log "Cleaned up $cleaned_entries stale registry entries"
            fi
        fi
        rmdir "$registry_lock" 2>/dev/null || true
        return 0
    else
        log_warn "Failed to acquire registry lock for cleanup"
        return 1
    fi
}

# Create a secure temporary file for storing policy ARNs
# Returns the path to the temporary file with restricted permissions
create_secure_temp_file() {
    local prefix="$1"
    local temp_file
    
    # Ensure ztiaws temp directory exists
    mkdir -p "$ZTIAWS_TEMP_DIR" 2>/dev/null || true
    chmod 700 "$ZTIAWS_TEMP_DIR" 2>/dev/null || true
    
    # Create a temporary file with restricted permissions (600) in our dedicated directory
    temp_file=$(mktemp -p "$ZTIAWS_TEMP_DIR" "${prefix}.XXXXXX")
    chmod 600 "$temp_file"
    echo "$temp_file"
}

# Acquire a lock for IAM operations on a specific instance
# Uses filesystem locking to prevent concurrent policy modifications
acquire_instance_lock() {
    local instance_id="$1"
    local lock_file="${LOCK_DIR}/iam-${instance_id}.lock"
    local max_wait=30  # Maximum wait time for lock acquisition
    local wait_interval=1
    local elapsed=0
    
    # Ensure lock directory exists
    mkdir -p "$LOCK_DIR" 2>/dev/null || true
    
    debug_log "Attempting to acquire IAM lock for instance: $instance_id"
    
    # Try to acquire lock with timeout
    while [ $elapsed -lt $max_wait ]; do
        # Use mkdir as an atomic operation for locking
        if mkdir "$lock_file" 2>/dev/null; then
            debug_log "Acquired IAM lock for instance: $instance_id"
            echo "$lock_file"
            return 0
        fi
        
        sleep $wait_interval
        elapsed=$((elapsed + wait_interval))
        
        # Check if lock is stale (older than 5 minutes)
        if [ -d "$lock_file" ]; then
            local lock_age
            if lock_age=$(stat -c %Y "$lock_file" 2>/dev/null); then
                local current_time
                current_time=$(date +%s)
                local age_seconds=$((current_time - lock_age))
                
                if [ $age_seconds -gt 300 ]; then  # 5 minutes
                    debug_log "Removing stale lock for instance: $instance_id (age: ${age_seconds}s)"
                    rmdir "$lock_file" 2>/dev/null || true
                fi
            fi
        fi
    done
    
    log_warn "Failed to acquire IAM lock for instance $instance_id within ${max_wait} seconds"
    return 1
}

# Release a lock for IAM operations on a specific instance
release_instance_lock() {
    local lock_file="$1"
    
    if [ -n "$lock_file" ] && [ -d "$lock_file" ]; then
        if rmdir "$lock_file" 2>/dev/null; then
            debug_log "Released IAM lock: $(basename "$lock_file")"
        else
            log_warn "Failed to release IAM lock: $lock_file"
        fi
    fi
}

# Wait for IAM changes to propagate
# Currently just waits for the configured delay time
# TODO: Implement actual IAM propagation validation in future versions
wait_for_iam_propagation() {
    local delay="${IAM_PROPAGATION_DELAY}"
    
    debug_log "Waiting for IAM changes to propagate (${delay}s)"
    sleep "$delay"
}

# Set up cleanup trap for emergency situations
cleanup_on_exit() {
    local exit_code=$?
    if [ $exit_code -ne 0 ]; then
        debug_log "Script exiting with error code $exit_code, performing emergency cleanup"
        
        # Use CURRENT_REGION if set, otherwise try to get from AWS_DEFAULT_REGION or AWS CLI config
        local cleanup_region="$CURRENT_REGION"
        if [[ -z "$cleanup_region" ]]; then
            cleanup_region="${AWS_DEFAULT_REGION:-}"
        fi
        if [[ -z "$cleanup_region" ]]; then
            cleanup_region=$(aws configure get region 2>/dev/null || echo "")
        fi
        
        if [[ -n "$cleanup_region" ]]; then
            debug_log "Using region for emergency cleanup: $cleanup_region"
            emergency_cleanup_s3_permissions "$cleanup_region" >/dev/null 2>&1 || true
        else
            log_warn "No region available for emergency cleanup - temporary policies may remain"
            debug_log "Consider setting AWS_DEFAULT_REGION environment variable or AWS CLI default region"
            # Still try to clean up files even without region
            emergency_cleanup_s3_permissions "" >/dev/null 2>&1 || true
        fi
    fi
}

# Set trap for cleanup on script exit
trap cleanup_on_exit EXIT INT TERM

# Initialize temp directory and registry on script load
init_temp_directory >/dev/null 2>&1 || true

# Function to set the current region (can be called by main script)
# This is optional - the region will be set automatically when upload_file/download_file are called
set_file_transfer_region() {
    local region_code="$1"
    
    if [[ -z "$region_code" ]]; then
        log_warn "set_file_transfer_region called with empty region"
        return 1
    fi
    
    # Validate region code if the validation function is available
    if command -v validate_region_code >/dev/null 2>&1; then
        local validated_region
        if ! validate_region_code "$region_code" validated_region; then
            log_error "Invalid region code: $region_code"
            return 1
        fi
        CURRENT_REGION="$validated_region"
        debug_log "File transfer module region set to: $validated_region (from code: $region_code)"
    else
        # Fallback: assume it's already a valid AWS region name
        CURRENT_REGION="$region_code"
        debug_log "File transfer module region set to: $region_code (validation not available)"
    fi
    
    return 0
}

# Validate file exists and is readable
validate_local_file() {
    local file_path="$1"

    if [ ! -f "$file_path" ]; then
        log_error "File not found: $file_path"
        return 1
    fi

    if [ ! -r "$file_path" ]; then
        log_error "File not readable: $file_path"
        return 1
    fi

    return 0
}

# Get file size in bytes
get_file_size() {
    local file_path="$1"

    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        stat -f%z "$file_path" 2>/dev/null
    else
        # Linux
        stat -c%s "$file_path" 2>/dev/null
    fi
}

# Generate unique S3 bucket name for this AWS account
get_s3_bucket_name() {
    local region="$1"
    local account_id

    if ! account_id=$(aws sts get-caller-identity --query Account --output text 2>/dev/null); then
        log_error "Failed to get AWS account ID"
        return 1
    fi

    echo "${S3_BUCKET_PREFIX}-${account_id}-${region}"
}

# Create lifecycle configuration for S3 bucket auto-cleanup
create_lifecycle_config() {
    local config_file="$1"

    if [[ -z "$config_file" ]]; then
        log_error "create_lifecycle_config: config file path is required"
        return 1
    fi

    # Create comprehensive lifecycle configuration
    cat > "$config_file" << 'EOF'
{
    "Rules": [
        {
            "ID": "SSMFileTransferCleanup",
            "Status": "Enabled",
            "Filter": {
                "Prefix": ""
            },
            "Expiration": {
                "Days": 1
            },
            "AbortIncompleteMultipartUpload": {
                "DaysAfterInitiation": 1
            }
        }
    ]
}
EOF

    # Validate that the file was created and contains valid JSON
    if [[ ! -f "$config_file" ]]; then
        log_error "Failed to create lifecycle configuration file: $config_file"
        return 1
    fi

    # Basic JSON validation using python if available, or jq if available
    if command -v python3 >/dev/null 2>&1; then
        if ! python3 -m json.tool "$config_file" >/dev/null 2>&1; then
            log_error "Invalid JSON in lifecycle configuration file"
            rm -f "$config_file"
            return 1
        fi
    elif command -v jq >/dev/null 2>&1; then
        if ! jq . "$config_file" >/dev/null 2>&1; then
            log_error "Invalid JSON in lifecycle configuration file"
            rm -f "$config_file"
            return 1
        fi
    fi

    debug_log "Lifecycle configuration file created successfully: $config_file"
    return 0
}

# Apply lifecycle configuration to ensure auto-cleanup
apply_lifecycle_config() {
    local bucket_name="$1"
    local region="$2"
    
    debug_log "Applying lifecycle configuration to bucket: $bucket_name"
    
    # Create lifecycle configuration file
    local lifecycle_config_file
    lifecycle_config_file=$(mktemp)
    
    create_lifecycle_config "$lifecycle_config_file"
    
    # Apply lifecycle configuration with proper error handling
    if aws s3api put-bucket-lifecycle-configuration \
        --bucket "$bucket_name" \
        --region "$region" \
        --lifecycle-configuration "file://$lifecycle_config_file" 2>/dev/null; then
        debug_log "Lifecycle configuration applied successfully to bucket: $bucket_name"
        rm -f "$lifecycle_config_file"
        return 0
    else
        log_warn "Failed to apply lifecycle configuration to bucket: $bucket_name"
        rm -f "$lifecycle_config_file"
        return 1
    fi
}

# Verify lifecycle configuration is active
verify_lifecycle_config() {
    local bucket_name="$1"
    local region="$2"
    
    debug_log "Verifying lifecycle configuration for bucket: $bucket_name"
    
    # Check if lifecycle configuration exists and is enabled
    local lifecycle_status
    # shellcheck disable=SC2016
    if lifecycle_status=$(aws s3api get-bucket-lifecycle-configuration \
        --bucket "$bucket_name" \
        --region "$region" \
        --query 'Rules[?ID==`SSMFileTransferCleanup`].Status' \
        --output text 2>/dev/null); then
        
        if [[ "$lifecycle_status" == "Enabled" ]]; then
            debug_log "Lifecycle configuration verified as enabled for bucket: $bucket_name"
            return 0
        else
            log_warn "Lifecycle configuration exists but is not enabled for bucket: $bucket_name"
            return 1
        fi
    else
        debug_log "No lifecycle configuration found for bucket: $bucket_name"
        return 1
    fi
}

# Create S3 bucket if it doesn't exist
ensure_s3_bucket() {
    local bucket_name="$1"
    local region="$2"

    log_info "Checking S3 bucket: $bucket_name"

    local bucket_created=false

    # Check if bucket exists
    if aws s3api head-bucket --bucket "$bucket_name" --region "$region" >/dev/null 2>&1; then
        log_info "S3 bucket already exists: $bucket_name"
    else
        log_info "Creating S3 bucket: $bucket_name"

        # Create bucket with appropriate configuration
        if [ "$region" = "us-east-1" ]; then
            if aws s3api create-bucket \
                --bucket "$bucket_name" \
                --region "$region" >/dev/null 2>&1; then
                bucket_created=true
            else
                log_error "Failed to create S3 bucket: $bucket_name"
                return 1
            fi
        else
            if aws s3api create-bucket \
                --bucket "$bucket_name" \
                --region "$region" \
                --create-bucket-configuration LocationConstraint="$region" >/dev/null 2>&1; then
                bucket_created=true
            else
                log_error "Failed to create S3 bucket: $bucket_name"
                return 1
            fi
        fi
    fi

    # Ensure lifecycle configuration is applied (for both existing and new buckets)
    if ! verify_lifecycle_config "$bucket_name" "$region"; then
        log_info "Applying lifecycle configuration for auto-cleanup..."
        if ! apply_lifecycle_config "$bucket_name" "$region"; then
            if [[ "$bucket_created" == "true" ]]; then
                log_error "Failed to apply lifecycle configuration to newly created bucket"
                # Clean up the bucket we just created since it's not properly configured
                aws s3api delete-bucket --bucket "$bucket_name" --region "$region" >/dev/null 2>&1
                return 1
            else
                log_warn "Failed to apply lifecycle configuration to existing bucket (continuing anyway)"
                # For existing buckets, we'll continue even if lifecycle config fails
                # as the bucket may have other lifecycle rules or permissions issues
            fi
        fi
    else
        debug_log "Lifecycle configuration already properly configured for bucket: $bucket_name"
    fi

    if [[ "$bucket_created" == "true" ]]; then
        log_info "S3 bucket created successfully with lifecycle configuration: $bucket_name"
    else
        log_info "S3 bucket verified and configured: $bucket_name"
    fi
    
    return 0
}

# Upload small file via base64 encoding
upload_file_small() {
    local local_file="$1"
    local remote_path="$2"
    local instance_id="$3"
    local region="$4"

    log_info "Uploading small file via base64 encoding..."

    # Encode file to base64
    local base64_content
    if ! base64_content=$(base64 < "$local_file" | tr -d '\n'); then
        log_error "Failed to encode file to base64"
        return 1
    fi

    # Create the upload command with proper directory creation
    local remote_dir
    remote_dir=$(dirname "$remote_path")
    local upload_command="mkdir -p '$remote_dir' && echo '$base64_content' | base64 -d > '$remote_path'"
    
    # Execute via SSM
    local command_id
    command_id=$(aws ssm send-command \
        --region "$region" \
        --instance-ids "$instance_id" \
        --document-name "AWS-RunShellScript" \
        --parameters "commands=[\"$upload_command\"]" \
        --comment "SSM file upload: $(basename "$local_file")" \
        --query "Command.CommandId" \
        --output text)

    if [ -z "$command_id" ]; then
        log_error "Failed to initiate file upload command"
        return 1
    fi

    # Wait for command completion and check status
    wait_for_command_completion "$command_id" "$instance_id" "$region"
    return $?
}

# Download small file via base64 encoding
download_file_small() {
    local remote_path="$1"
    local local_file="$2"
    local instance_id="$3"
    local region="$4"

    log_info "Downloading small file via base64 encoding..."

    # Create the download command
    local download_command="if [ -f '$remote_path' ]; then base64 '$remote_path'; else echo 'FILE_NOT_FOUND'; fi"

    # Execute via SSM
    local command_id
    command_id=$(aws ssm send-command \
        --region "$region" \
        --instance-ids "$instance_id" \
        --document-name "AWS-RunShellScript" \
        --parameters "commands=[\"$download_command\"]" \
        --comment "SSM file download: $(basename "$remote_path")" \
        --query "Command.CommandId" \
        --output text)

    if [ -z "$command_id" ]; then
        log_error "Failed to initiate file download command"
        return 1
    fi

    # Wait for command completion
    if ! wait_for_command_completion "$command_id" "$instance_id" "$region"; then
        return 1
    fi

    # Get command output
    local output
    if ! output=$(aws ssm get-command-invocation \
        --region "$region" \
        --command-id "$command_id" \
        --instance-id "$instance_id" \
        --query "StandardOutputContent" \
        --output text); then
        log_error "Failed to get command output"
        return 1
    fi

    # Check if file was found
    if echo "$output" | grep -q "FILE_NOT_FOUND"; then
        log_error "Remote file not found: $remote_path"
        return 1
    fi

    # Decode and save the file
    if echo "$output" | base64 -d > "$local_file"; then
        log_info "File downloaded successfully: $local_file"
        return 0
    else
        log_error "Failed to decode downloaded file"
        return 1
    fi
}

# Get instance profile role name for an EC2 instance
get_instance_profile_role() {
    local instance_id="$1"
    local region="$2"

    debug_log "Getting instance profile role for instance: $instance_id"
    
    # Get instance profile name
    local instance_profile_name
    instance_profile_name=$(aws ec2 describe-instances \
        --region "$region" \
        --instance-ids "$instance_id" \
        --query 'Reservations[0].Instances[0].IamInstanceProfile.Arn' \
        --output text 2>/dev/null | awk -F'/' '{print $2}')

    if [[ -z "$instance_profile_name" || "$instance_profile_name" == "None" ]]; then
        log_error "No IAM instance profile found for instance $instance_id"
        return 1
    fi

    # Get role name from instance profile
    local role_name
    role_name=$(aws iam get-instance-profile \
        --instance-profile-name "$instance_profile_name" \
        --query 'InstanceProfile.Roles[0].RoleName' \
        --output text 2>/dev/null)

    if [[ -z "$role_name" || "$role_name" == "None" ]]; then
        log_error "No role found in instance profile $instance_profile_name"
        return 1
    fi

    echo "$role_name"
    return 0
}

# Create S3 policy for the specific bucket
create_s3_policy_document() {
    local bucket_name="$1"
    local policy_file="$2"

    cat > "$policy_file" << EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:GetObject",
                "s3:PutObject",
                "s3:DeleteObject"
            ],
            "Resource": "arn:aws:s3:::${bucket_name}/*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "s3:ListBucket"
            ],
            "Resource": "arn:aws:s3:::${bucket_name}"
        }
    ]
}
EOF
}

# Attach S3 permissions to instance profile role (with locking)
attach_s3_permissions() {
    local instance_id="$1"
    local region="$2"
    local bucket_name="$3"

    debug_log "Attaching S3 permissions for bucket: $bucket_name"

    # Acquire lock for this instance to prevent concurrent policy modifications
    local lock_file
    if ! lock_file=$(acquire_instance_lock "$instance_id"); then
        log_error "Failed to acquire lock for IAM operations on instance $instance_id"
        return 1
    fi

    # Get the role name
    local role_name
    if ! role_name=$(get_instance_profile_role "$instance_id" "$region"); then
        release_instance_lock "$lock_file"
        return 1
    fi

    debug_log "Found role: $role_name"

    # Create unique policy name using secure identifier
    local unique_id
    unique_id=$(generate_unique_id)
    local policy_name="ZTIaws-SSM-S3-Access-${unique_id}"
    
    # Create policy document with secure permissions
    local policy_file
    policy_file=$(create_secure_temp_file "ztiaws-s3-policy-doc")
    create_s3_policy_document "$bucket_name" "$policy_file"

    # Create and attach the policy
    local policy_arn
    if policy_arn=$(aws iam create-policy \
        --policy-name "$policy_name" \
        --policy-document "file://$policy_file" \
        --description "Temporary S3 access for ztiaws SSM file transfer" \
        --query 'Policy.Arn' \
        --output text 2>/dev/null); then
        
        debug_log "Created policy: $policy_arn"
        
        # Attach policy to role
        if aws iam attach-role-policy \
            --role-name "$role_name" \
            --policy-arn "$policy_arn" >/dev/null 2>&1; then
            
            debug_log "Attached policy to role: $role_name"
            # Store policy ARN in a secure temp file for later cleanup
            local policy_tracking_file
            policy_tracking_file=$(create_secure_temp_file "ztiaws-s3-policy-${instance_id}-${unique_id}")
            echo "$policy_arn" > "$policy_tracking_file"
            
            # Add to registry for efficient cleanup
            add_policy_to_registry "$instance_id" "$region" "$policy_arn" "$policy_tracking_file"
            
            rm -f "$policy_file"
            
            # Release lock before returning
            release_instance_lock "$lock_file"
            return 0
        else
            log_error "Failed to attach policy to role"
            # Clean up policy if attachment failed
            aws iam delete-policy --policy-arn "$policy_arn" >/dev/null 2>&1
            rm -f "$policy_file"
            release_instance_lock "$lock_file"
            return 1
        fi
    else
        log_error "Failed to create S3 policy"
        rm -f "$policy_file"
        release_instance_lock "$lock_file"
        return 1
    fi
}

# Remove S3 permissions from instance profile role (with locking)
remove_s3_permissions() {
    local instance_id="$1"
    local region="$2"

    debug_log "Removing S3 permissions for instance: $instance_id"

    # Acquire lock for this instance to prevent concurrent policy modifications
    local lock_file
    if ! lock_file=$(acquire_instance_lock "$instance_id"); then
        log_warn "Failed to acquire lock for IAM cleanup on instance $instance_id - skipping cleanup"
        return 1
    fi

    # Get policies for this instance from registry
    local policy_entries
    policy_entries=$(get_policies_for_instance "$instance_id")
    
    if [[ -z "$policy_entries" ]]; then
        debug_log "No policy entries found in registry for instance: $instance_id"
        release_instance_lock "$lock_file"
        return 0
    fi

    # Process each policy entry found in registry
    while IFS='|' read -r reg_instance_id reg_region policy_arn policy_file timestamp; do
        # Skip empty lines
        [[ -z "$reg_instance_id" ]] && continue
        
        debug_log "Processing policy from registry: $policy_arn"
        
        # Clean up the policy file
        if [[ -f "$policy_file" ]]; then
            rm -f "$policy_file"
        fi

        if [[ -z "$policy_arn" ]]; then
            debug_log "No policy ARN found in registry entry"
            continue
        fi

        # Get the role name
        local role_name
        if ! role_name=$(get_instance_profile_role "$instance_id" "$region"); then
            debug_log "Could not get role name for cleanup, but continuing with policy deletion"
        else
            # Detach policy from role
            if aws iam detach-role-policy \
                --role-name "$role_name" \
                --policy-arn "$policy_arn" >/dev/null 2>&1; then
                debug_log "Detached policy from role: $role_name"
            else
                log_warn "Failed to detach policy from role (may already be detached)"
            fi
        fi

        # Delete the policy
        if aws iam delete-policy --policy-arn "$policy_arn" >/dev/null 2>&1; then
            debug_log "Deleted policy: $policy_arn"
        else
            log_warn "Failed to delete policy (may already be deleted): $policy_arn"
        fi
    done <<< "$policy_entries"

    # Remove policies from registry for this instance
    remove_policies_from_registry "$instance_id"

    # Release lock after cleanup
    release_instance_lock "$lock_file"
    return 0
}

# Manage S3 permissions with proper locking
manage_s3_permissions() {
    local action="$1"  # "attach" or "remove"
    local instance_id="$2"
    local region="$3"
    local bucket_name="$4"

    case "$action" in
        "attach")
            if attach_s3_permissions "$instance_id" "$region" "$bucket_name"; then
                debug_log "S3 permissions attached successfully"
                return 0
            else
                log_error "Failed to attach S3 permissions"
                return 1
            fi
            ;;
        "remove")
            # Run remove operation synchronously with proper locking
            # Background execution removed to prevent race conditions
            if remove_s3_permissions "$instance_id" "$region"; then
                debug_log "S3 permissions removed successfully"
                return 0
            else
                log_warn "Failed to remove S3 permissions (may already be cleaned up)"
                # Don't return error for removal failures as they're often expected
                return 0
            fi
            ;;
        *)
            log_error "Invalid action for manage_s3_permissions: $action"
            return 1
            ;;
    esac
}

# Upload large file via S3 intermediary
upload_file_large() {
    local local_file="$1"
    local remote_path="$2"
    local instance_id="$3"
    local region="$4"

    log_info "Uploading large file via S3 intermediary..."

    # Set current region for cleanup purposes
    CURRENT_REGION="$region"

    # Get S3 bucket name
    local bucket_name
    if ! bucket_name=$(get_s3_bucket_name "$region"); then
        return 1
    fi

    # Ensure bucket exists
    if ! ensure_s3_bucket "$bucket_name" "$region"; then
        return 1
    fi

    # Attach S3 permissions to instance role (background operation)
    log_info "Configuring S3 permissions for instance..."
    
    # First validate IAM setup
    if ! validate_instance_iam_setup "$instance_id" "$region"; then
        return 1
    fi
    
    if ! manage_s3_permissions "attach" "$instance_id" "$region" "$bucket_name"; then
        log_error "Failed to configure S3 permissions. The instance may need proper IAM role setup."
        return 1
    fi

    # Wait for IAM changes to propagate with configurable delay
    wait_for_iam_propagation

    # Generate unique S3 key
    local s3_key
    s3_key="uploads/$(date +%s)-$(basename "$local_file")"

    log_info "Uploading to S3: s3://$bucket_name/$s3_key"

    # Upload to S3
    if ! aws s3 cp "$local_file" "s3://$bucket_name/$s3_key" --region "$region"; then
        log_error "Failed to upload file to S3"
        # Clean up permissions in background
        manage_s3_permissions "remove" "$instance_id" "$region" "$bucket_name" >/dev/null 2>&1
        return 1
    fi

    # Create download command for instance
    local remote_dir
    remote_dir=$(dirname "$remote_path")
    local download_command="mkdir -p '$remote_dir' && aws s3 cp s3://$bucket_name/$s3_key '$remote_path' --region $region && aws s3 rm s3://$bucket_name/$s3_key --region $region"

    # Execute download on instance
    local command_id
    command_id=$(aws ssm send-command \
        --region "$region" \
        --instance-ids "$instance_id" \
        --document-name "AWS-RunShellScript" \
        --parameters "commands=[\"$download_command\"]" \
        --comment "SSM large file upload via S3: $(basename "$local_file")" \
        --query "Command.CommandId" \
        --output text)

    if [ -z "$command_id" ]; then
        log_error "Failed to initiate S3 download command on instance"
        # Clean up S3 object and permissions
        aws s3 rm "s3://$bucket_name/$s3_key" --region "$region" >/dev/null 2>&1
        manage_s3_permissions "remove" "$instance_id" "$region" "$bucket_name" >/dev/null 2>&1
        return 1
    fi

    # Wait for command completion
    if wait_for_command_completion "$command_id" "$instance_id" "$region"; then
        log_info "Large file uploaded successfully via S3"
        # Clean up permissions in background after successful transfer
        manage_s3_permissions "remove" "$instance_id" "$region" "$bucket_name" >/dev/null 2>&1
        return 0
    else
        # Clean up S3 object and permissions if command failed
        aws s3 rm "s3://$bucket_name/$s3_key" --region "$region" >/dev/null 2>&1
        manage_s3_permissions "remove" "$instance_id" "$region" "$bucket_name" >/dev/null 2>&1
        return 1
    fi
}

# Download large file via S3 intermediary
download_file_large() {
    local remote_path="$1"
    local local_file="$2"
    local instance_id="$3"
    local region="$4"

    log_info "Downloading large file via S3 intermediary..."

    # Set current region for cleanup purposes
    CURRENT_REGION="$region"

    # Get S3 bucket name
    local bucket_name
    if ! bucket_name=$(get_s3_bucket_name "$region"); then
        return 1
    fi

    # Ensure bucket exists
    if ! ensure_s3_bucket "$bucket_name" "$region"; then
        return 1
    fi

    # Attach S3 permissions to instance role (background operation)
    log_info "Configuring S3 permissions for instance..."
    
    # First validate IAM setup
    if ! validate_instance_iam_setup "$instance_id" "$region"; then
        return 1
    fi
    
    if ! manage_s3_permissions "attach" "$instance_id" "$region" "$bucket_name"; then
        log_error "Failed to configure S3 permissions. The instance may need proper IAM role setup."
        return 1
    fi

    # Wait for IAM changes to propagate with configurable delay
    wait_for_iam_propagation

    # Generate unique S3 key
    local s3_key
    s3_key="downloads/$(date +%s)-$(basename "$remote_path")"

    # Create upload command for instance
    local upload_command="if [ -f '$remote_path' ]; then aws s3 cp '$remote_path' s3://$bucket_name/$s3_key --region $region; else echo 'FILE_NOT_FOUND'; fi"

    # Execute upload on instance
    local command_id
    command_id=$(aws ssm send-command \
        --region "$region" \
        --instance-ids "$instance_id" \
        --document-name "AWS-RunShellScript" \
        --parameters "commands=[\"$upload_command\"]" \
        --comment "SSM large file download via S3: $(basename "$remote_path")" \
        --query "Command.CommandId" \
        --output text)

    if [ -z "$command_id" ]; then
        log_error "Failed to initiate S3 upload command on instance"
        # Clean up permissions
        manage_s3_permissions "remove" "$instance_id" "$region" "$bucket_name" >/dev/null 2>&1
        return 1
    fi

    # Wait for command completion
    if ! wait_for_command_completion "$command_id" "$instance_id" "$region"; then
        # Clean up permissions
        manage_s3_permissions "remove" "$instance_id" "$region" "$bucket_name" >/dev/null 2>&1
        return 1
    fi

    # Check command output for errors
    local output
    output=$(aws ssm get-command-invocation \
        --region "$region" \
        --command-id "$command_id" \
        --instance-id "$instance_id" \
        --query "StandardOutputContent" \
        --output text 2>/dev/null)

    if echo "$output" | grep -q "FILE_NOT_FOUND"; then
        log_error "Remote file not found: $remote_path"
        # Clean up permissions
        manage_s3_permissions "remove" "$instance_id" "$region" "$bucket_name" >/dev/null 2>&1
        return 1
    fi

    # Download from S3
    log_info "Downloading from S3: s3://$bucket_name/$s3_key"

    if aws s3 cp "s3://$bucket_name/$s3_key" "$local_file" --region "$region"; then
        # Clean up S3 object and permissions
        aws s3 rm "s3://$bucket_name/$s3_key" --region "$region" >/dev/null 2>&1
        manage_s3_permissions "remove" "$instance_id" "$region" "$bucket_name" >/dev/null 2>&1
        log_info "Large file downloaded successfully via S3"
        return 0
    else
        log_error "Failed to download file from S3"
        # Clean up S3 object and permissions
        aws s3 rm "s3://$bucket_name/$s3_key" --region "$region" >/dev/null 2>&1
        manage_s3_permissions "remove" "$instance_id" "$region" "$bucket_name" >/dev/null 2>&1
        return 1
    fi
}

# Wait for SSM command completion with progress indication
wait_for_command_completion() {
    local command_id="$1"
    local instance_id="$2"
    local region="$3"
    local max_wait=300  # 5 minutes max wait
    local wait_interval=2
    local elapsed=0

    log_info "Waiting for command completion... (Command ID: $command_id)"

    while [ $elapsed -lt $max_wait ]; do
        local status
        status=$(aws ssm get-command-invocation \
            --region "$region" \
            --command-id "$command_id" \
            --instance-id "$instance_id" \
            --query "Status" \
            --output text 2>/dev/null)

        case "$status" in
            "Success")
                echo
                log_info "Command completed successfully"
                return 0
                ;;
            "Failed"|"Cancelled"|"TimedOut")
                echo
                log_error "Command failed with status: $status"
                # Get error details
                local error_output
                error_output=$(aws ssm get-command-invocation \
                    --region "$region" \
                    --command-id "$command_id" \
                    --instance-id "$instance_id" \
                    --query "StandardErrorContent" \
                    --output text 2>/dev/null)
                if [ -n "$error_output" ]; then
                    log_error "Error details: $error_output"
                fi
                return 1
                ;;
            "InProgress"|"Pending"|"Delayed")
                printf "."
                sleep $wait_interval
                elapsed=$((elapsed + wait_interval))
                ;;
            *)
                echo
                log_error "Unknown command status: $status"
                return 1
                ;;
        esac
    done

    echo
    log_error "Command timed out after $max_wait seconds"
    return 1
}

# Main upload function
upload_file() {
    local region="$1"
    local instance_identifier="$2"
    local local_file="$3"
    local remote_path="$4"

    # Basic validation - ensure region is provided
    if [[ -z "$region" ]]; then
        log_error "Region parameter is required for file upload"
        return 1
    fi

    # Set current region for cleanup purposes
    CURRENT_REGION="$region"

    # Validate inputs
    if ! validate_local_file "$local_file"; then
        return 1
    fi

    # Resolve instance ID
    local instance_id
    if ! instance_id=$(resolve_instance_identifier "$instance_identifier" "$region"); then
        return 1
    fi

    # Get file size
    local file_size
    file_size=$(get_file_size "$local_file")
    if [ -z "$file_size" ]; then
        log_error "Failed to get file size"
        return 1
    fi

    log_info "File size: $file_size bytes"

    # Choose transfer method based on file size
    if [ "$file_size" -lt "$FILE_SIZE_THRESHOLD" ]; then
        log_info "Using direct SSM transfer (file < 1MB)"
        upload_file_small "$local_file" "$remote_path" "$instance_id" "$region"
    else
        log_info "Using S3 intermediary transfer (file ≥ 1MB)"
        upload_file_large "$local_file" "$remote_path" "$instance_id" "$region"
    fi
}

# Main download function
download_file() {
    local region="$1"
    local instance_identifier="$2"
    local remote_path="$3"
    local local_file="$4"

    # Basic validation - ensure region is provided
    if [[ -z "$region" ]]; then
        log_error "Region parameter is required for file download"
        return 1
    fi

    # Set current region for cleanup purposes
    CURRENT_REGION="$region"

    # Resolve instance ID
    local instance_id
    if ! instance_id=$(resolve_instance_identifier "$instance_identifier" "$region"); then
        return 1
    fi

    # Create local directory if needed
    local local_dir
    local_dir=$(dirname "$local_file")
    if [ ! -d "$local_dir" ]; then
        if ! mkdir -p "$local_dir"; then
            log_error "Failed to create local directory: $local_dir"
            return 1
        fi
    fi

    # First, try to get remote file size to determine transfer method
    log_info "Checking remote file size..."

    local size_command="if [ -f '$remote_path' ]; then stat -c%s '$remote_path' 2>/dev/null || stat -f%z '$remote_path' 2>/dev/null; else echo 'FILE_NOT_FOUND'; fi"

    local command_id
    command_id=$(aws ssm send-command \
        --region "$region" \
        --instance-ids "$instance_id" \
        --document-name "AWS-RunShellScript" \
        --parameters "commands=[\"$size_command\"]" \
        --comment "SSM file size check: $(basename "$remote_path")" \
        --query "Command.CommandId" \
        --output text)

    if [ -z "$command_id" ]; then
        log_error "Failed to check remote file size"
        return 1
    fi

    # Wait for size check completion
    if ! wait_for_command_completion "$command_id" "$instance_id" "$region"; then
        return 1
    fi

    # Get file size result
    local size_output
    size_output=$(aws ssm get-command-invocation \
        --region "$region" \
        --command-id "$command_id" \
        --instance-id "$instance_id" \
        --query "StandardOutputContent" \
        --output text)

    if echo "$size_output" | grep -q "FILE_NOT_FOUND"; then
        log_error "Remote file not found: $remote_path"
        return 1
    fi

    local file_size
    file_size=$(echo "$size_output" | tr -d '\n' | grep -o '[0-9]*')

    if [ -z "$file_size" ] || [ "$file_size" -eq 0 ]; then
        log_error "Could not determine remote file size or file is empty"
        return 1
    fi

    log_info "Remote file size: $file_size bytes"

    # Choose transfer method based on file size
    if [ "$file_size" -lt "$FILE_SIZE_THRESHOLD" ]; then
        log_info "Using direct SSM transfer (file < 1MB)"
        download_file_small "$remote_path" "$local_file" "$instance_id" "$region"
    else
        log_info "Using S3 intermediary transfer (file ≥ 1MB)"
        download_file_large "$remote_path" "$local_file" "$instance_id" "$region"
    fi
}

# Function to handle multiple instances with S3 permissions (for tagged operations)
manage_multiple_instance_s3_permissions() {
    local action="$1"  # "attach" or "remove"
    local region="$2"
    local bucket_name="$3"
    shift 3
    local instance_ids=("$@")

    local success_count=0
    local failed_instances=()

    for instance_id in "${instance_ids[@]}"; do
        if manage_s3_permissions "$action" "$instance_id" "$region" "$bucket_name"; then
            ((success_count++))
            debug_log "S3 permissions $action successful for instance: $instance_id"
        else
            failed_instances+=("$instance_id")
            log_warn "Failed to $action S3 permissions for instance: $instance_id"
        fi
    done

    if [ ${#failed_instances[@]} -gt 0 ]; then
        log_warn "Failed to $action S3 permissions for ${#failed_instances[@]} instances: ${failed_instances[*]}"
        return 1
    fi

    log_info "Successfully ${action}ed S3 permissions for $success_count instances"
    return 0
}

# Cleanup function for emergency cleanup of all temporary policies
emergency_cleanup_s3_permissions() {
    local region="$1"
    
    debug_log "Performing emergency cleanup of temporary S3 policies..."
    
    # First try to clean up from registry if it exists
    if [[ -f "$POLICY_REGISTRY_FILE" ]]; then
        debug_log "Using registry for emergency cleanup"
        local cleanup_count=0
        
        while IFS='|' read -r instance_id reg_region policy_arn policy_file timestamp; do
            # Skip empty lines
            [[ -z "$instance_id" ]] && continue
            
            debug_log "Emergency cleanup for policy: $policy_arn (instance: $instance_id)"
            
            # Clean up policy file
            if [[ -f "$policy_file" ]]; then
                rm -f "$policy_file"
            fi
            
            # Use provided region or fall back to registry region
            local cleanup_region="${region:-$reg_region}"
            
            if [[ -n "$cleanup_region" && -n "$instance_id" ]]; then
                # Try to get role name and detach policy
                local role_name
                if role_name=$(get_instance_profile_role "$instance_id" "$cleanup_region" 2>/dev/null); then
                    aws iam detach-role-policy \
                        --role-name "$role_name" \
                        --policy-arn "$policy_arn" >/dev/null 2>&1 || true
                fi
            fi
            
            # Delete the policy
            if [[ -n "$policy_arn" ]]; then
                aws iam delete-policy --policy-arn "$policy_arn" >/dev/null 2>&1 || true
                ((cleanup_count++))
            fi
            
        done < "$POLICY_REGISTRY_FILE"
        
        # Clear the registry after cleanup
        : > "$POLICY_REGISTRY_FILE"
        
        if [ $cleanup_count -gt 0 ]; then
            debug_log "Emergency cleanup completed for $cleanup_count policies from registry"
        else
            debug_log "No policies found in registry for cleanup"
        fi
    else
        debug_log "No registry file found, attempting fallback cleanup"
        
        # Fallback: scan our dedicated temp directory only (much more efficient than scanning all of /tmp)
        if [[ -d "$ZTIAWS_TEMP_DIR" ]]; then
            local policy_files
            policy_files=$(find "$ZTIAWS_TEMP_DIR" -name "ztiaws-s3-policy-*" -type f 2>/dev/null || true)
            
            if [[ -n "$policy_files" ]]; then
                local cleanup_count=0
                while IFS= read -r policy_file; do
                    if [[ -f "$policy_file" ]]; then
                        local policy_arn
                        policy_arn=$(cat "$policy_file" 2>/dev/null)
                        
                        if [[ -n "$policy_arn" ]]; then
                            # Extract instance ID from filename using more robust pattern
                            local instance_id
                            if [[ $(basename "$policy_file") =~ ztiaws-s3-policy-(i-[a-f0-9]+) ]]; then
                                instance_id="${BASH_REMATCH[1]}"
                                debug_log "Extracted instance ID: $instance_id from policy file"
                                
                                if [[ -n "$instance_id" && -n "$region" ]]; then
                                    # Try to get role name and detach policy
                                    local role_name
                                    if role_name=$(get_instance_profile_role "$instance_id" "$region" 2>/dev/null); then
                                        aws iam detach-role-policy \
                                            --role-name "$role_name" \
                                            --policy-arn "$policy_arn" >/dev/null 2>&1 || true
                                    fi
                                fi
                            else
                                debug_log "Could not extract instance ID from filename: $(basename "$policy_file")"
                            fi
                            
                            # Delete the policy regardless
                            debug_log "Attempting direct policy cleanup: $policy_arn"
                            aws iam delete-policy --policy-arn "$policy_arn" >/dev/null 2>&1 || true
                            ((cleanup_count++))
                        fi
                        
                        rm -f "$policy_file"
                    fi
                done <<< "$policy_files"
                
                if [ $cleanup_count -gt 0 ]; then
                    debug_log "Fallback cleanup completed for $cleanup_count policy files"
                else
                    debug_log "No policies required cleanup in fallback mode"
                fi
            else
                debug_log "No temporary policy files found in dedicated directory"
            fi
        fi
    fi
    
    # Clean up stale lock files (older than 5 minutes)
    if [[ -d "$LOCK_DIR" ]]; then
        debug_log "Cleaning up stale lock files..."
        local lock_cleanup_count=0
        
        # Store the results of the find command in an array
        local lock_files
        mapfile -t lock_files < <(find "$LOCK_DIR" -name "iam-*.lock" -type d 2>/dev/null)
        
        # Iterate over the array
        for lock_file in "${lock_files[@]}"; do
            if [[ -d "$lock_file" ]]; then
                local lock_age
                if lock_age=$(stat -c %Y "$lock_file" 2>/dev/null); then
                    local current_time
                    current_time=$(date +%s)
                    local age_seconds=$((current_time - lock_age))
                    
                    if [ $age_seconds -gt 300 ]; then  # 5 minutes
                        debug_log "Removing stale lock: $(basename "$lock_file") (age: ${age_seconds}s)"
                        rmdir "$lock_file" 2>/dev/null || true
                        ((lock_cleanup_count++))
                    fi
                fi
            fi
        done
        
        if [ $lock_cleanup_count -gt 0 ]; then
            debug_log "Cleaned up $lock_cleanup_count stale lock files"
        fi
        
        # Try to remove lock directory if empty
        rmdir "$LOCK_DIR" 2>/dev/null || true
    fi
    
    # Clean up stale registry entries
    cleanup_stale_registry_entries >/dev/null 2>&1 || true
    
    return 0
}

# Function to validate that instance has required IAM setup for S3 operations
validate_instance_iam_setup() {
    local instance_id="$1"
    local region="$2"
    
    debug_log "Validating IAM setup for instance: $instance_id"
    
    # Check if instance has IAM instance profile
    local instance_profile_arn
    instance_profile_arn=$(aws ec2 describe-instances \
        --region "$region" \
        --instance-ids "$instance_id" \
        --query 'Reservations[0].Instances[0].IamInstanceProfile.Arn' \
        --output text 2>/dev/null)

    if [[ -z "$instance_profile_arn" || "$instance_profile_arn" == "None" ]]; then
        log_error "Instance $instance_id does not have an IAM instance profile attached"
        log_error "Please attach an IAM instance profile with appropriate permissions to the instance"
        return 1
    fi
    
    debug_log "Instance has IAM instance profile: $instance_profile_arn"
    
    # Get role name and validate it exists
    local role_name
    if ! role_name=$(get_instance_profile_role "$instance_id" "$region"); then
        log_error "Failed to get IAM role for instance $instance_id"
        return 1
    fi
    
    debug_log "IAM validation successful for instance: $instance_id (role: $role_name)"
    return 0
}