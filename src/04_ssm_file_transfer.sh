#!/usr/bin/env bash

# AWS SSM File Transfer Module
# Handles file upload/download operations via SSM

# File size threshold for choosing transfer method (1MB)
FILE_SIZE_THRESHOLD=$((1024 * 1024))

# S3 bucket for large file transfers (will be created if needed)
S3_BUCKET_PREFIX="ssm-file-transfer"

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
    
    # Try Linux format first, then macOS format as fallback
    stat -c%s "$file_path" 2>/dev/null || stat -f%z "$file_path" 2>/dev/null
}

# Generate cross-platform command to get remote file size
get_remote_file_size_command() {
    local escaped_remote_path="$1"
    
    # Try Linux stat format first, then macOS format as fallback
    echo "if [ -f $escaped_remote_path ]; then stat -c%s $escaped_remote_path 2>/dev/null || stat -f%z $escaped_remote_path 2>/dev/null; else echo 'FILE_NOT_FOUND'; fi"
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
    
    cat > "$config_file" << 'EOF'
{
    "Rules": [
        {
            "ID": "SSMFileTransferCleanup",
            "Status": "Enabled",
            "Expiration": {
                "Days": 1
            }
        }
    ]
}
EOF
}

# Create S3 bucket encryption configuration
create_encryption_config() {
    local config_file="$1"
    
    cat > "$config_file" << 'EOF'
{
    "Rules": [
        {
            "ApplyServerSideEncryptionByDefault": {
                "SSEAlgorithm": "AES256"
            },
            "BucketKeyEnabled": true
        }
    ]
}
EOF
}

# Create S3 bucket if it doesn't exist
ensure_s3_bucket() {
    local bucket_name="$1"
    local region="$2"
    
    log_info "Checking S3 bucket: $bucket_name"
    
    # Check if bucket exists
    if aws s3api head-bucket --bucket "$bucket_name" --region "$region" >/dev/null 2>&1; then
        log_info "S3 bucket already exists: $bucket_name"
        
        # Check if encryption is already enabled
        if ! aws s3api get-bucket-encryption --bucket "$bucket_name" --region "$region" >/dev/null 2>&1; then
            log_info "Encryption not enabled on existing bucket, enabling now..."
            
            local encryption_config_file
            encryption_config_file=$(mktemp)
            create_encryption_config "$encryption_config_file"
            
            if aws s3api put-bucket-encryption \
                --bucket "$bucket_name" \
                --server-side-encryption-configuration "file://$encryption_config_file" \
                --region "$region" >/dev/null 2>&1; then
                log_info "Successfully enabled encryption on existing bucket"
            else
                log_warn "Failed to enable encryption on existing bucket"
            fi
            
            rm -f "$encryption_config_file"
        else
            log_info "Encryption already enabled on existing bucket"
        fi
        
        return 0
    fi
    
    log_info "Creating S3 bucket: $bucket_name"
    
    # Create bucket with appropriate configuration
    local create_bucket_success=false
    
    if [ "$region" = "us-east-1" ]; then
        if aws s3api create-bucket \
            --bucket "$bucket_name" \
            --region "$region" >/dev/null 2>&1; then
            create_bucket_success=true
        fi
    else
        if aws s3api create-bucket \
            --bucket "$bucket_name" \
            --region "$region" \
            --create-bucket-configuration LocationConstraint="$region" >/dev/null 2>&1; then
            create_bucket_success=true
        fi
    fi
    
    if [ "$create_bucket_success" = true ]; then
        # Enable server-side encryption
        local encryption_config_file
        encryption_config_file=$(mktemp)
        create_encryption_config "$encryption_config_file"
        
        if ! aws s3api put-bucket-encryption \
            --bucket "$bucket_name" \
            --server-side-encryption-configuration "file://$encryption_config_file" \
            --region "$region" >/dev/null 2>&1; then
            log_warn "Failed to enable bucket encryption, but continuing..."
        else
            log_info "Enabled AES256 encryption on bucket"
        fi
        
        rm -f "$encryption_config_file"
        
        # Set bucket lifecycle to auto-delete files after 1 day
        local lifecycle_config_file
        lifecycle_config_file=$(mktemp)
        create_lifecycle_config "$lifecycle_config_file"
        
        if ! aws s3api put-bucket-lifecycle-configuration \
            --bucket "$bucket_name" \
            --lifecycle-configuration "file://$lifecycle_config_file" \
            --region "$region" >/dev/null 2>&1; then
            log_warn "Failed to set bucket lifecycle, but continuing..."
        else
            log_info "Set bucket lifecycle to auto-delete files after 1 day"
        fi
        
        rm -f "$lifecycle_config_file"
        
        # Block public access
        if ! aws s3api put-public-access-block \
            --bucket "$bucket_name" \
            --public-access-block-configuration \
                "BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true" \
            --region "$region" >/dev/null 2>&1; then
            log_warn "Failed to block public access, but continuing..."
        else
            log_info "Blocked all public access to bucket"
        fi
        
        log_info "S3 bucket created successfully: $bucket_name"
        return 0
    else
        log_error "Failed to create S3 bucket: $bucket_name"
        return 1
    fi
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
    
    # Create the upload command
    local escaped_remote_path
    escaped_remote_path=$(printf '%q' "$remote_path")
    local upload_command="mkdir -p \"\$(dirname $escaped_remote_path)\" && echo '$base64_content' | base64 -d > $escaped_remote_path"
    
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
    local escaped_remote_path
    escaped_remote_path=$(printf '%q' "$remote_path")
    local download_command="if [ -f $escaped_remote_path ]; then base64 $escaped_remote_path; else echo 'FILE_NOT_FOUND'; fi"
    
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

# Upload large file via S3 intermediary
upload_file_large() {
    local local_file="$1"
    local remote_path="$2"
    local instance_id="$3"
    local region="$4"
    
    log_info "Uploading large file via S3 intermediary..."
    
    # Get S3 bucket name
    local bucket_name
    if ! bucket_name=$(get_s3_bucket_name "$region"); then
        return 1
    fi
    
    # Ensure bucket exists
    if ! ensure_s3_bucket "$bucket_name" "$region"; then
        return 1
    fi
    
    # Generate unique S3 key
    local s3_key
    s3_key="uploads/$(date +%s)-$(basename "$local_file")"
    
    log_info "Uploading to S3: s3://$bucket_name/$s3_key"
    
    # Upload to S3 with server-side encryption
    if ! aws s3 cp "$local_file" "s3://$bucket_name/$s3_key" \
        --region "$region" \
        --sse AES256; then
        log_error "Failed to upload file to S3"
        return 1
    fi
    
    # Create download command for instance
    local escaped_remote_path
    escaped_remote_path=$(printf '%q' "$remote_path")
    local download_command="mkdir -p \"\$(dirname $escaped_remote_path)\" && aws s3 cp s3://$bucket_name/$s3_key $escaped_remote_path --region $region && aws s3 rm s3://$bucket_name/$s3_key --region $region"
    
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
        # Clean up S3 object
        aws s3 rm "s3://$bucket_name/$s3_key" --region "$region" >/dev/null 2>&1
        return 1
    fi
    
    # Wait for command completion
    if wait_for_command_completion "$command_id" "$instance_id" "$region"; then
        log_info "Large file uploaded successfully via S3"
        return 0
    else
        # Clean up S3 object if command failed
        aws s3 rm "s3://$bucket_name/$s3_key" --region "$region" >/dev/null 2>&1
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
    
    # Get S3 bucket name
    local bucket_name
    if ! bucket_name=$(get_s3_bucket_name "$region"); then
        return 1
    fi
    
    # Ensure bucket exists
    if ! ensure_s3_bucket "$bucket_name" "$region"; then
        return 1
    fi
    
    # Generate unique S3 key
    local s3_key
    s3_key="downloads/$(date +%s)-$(basename "$remote_path")"
    
    # Create upload command for instance
    local escaped_remote_path
    escaped_remote_path=$(printf '%q' "$remote_path")
    local upload_command="if [ -f $escaped_remote_path ]; then aws s3 cp $escaped_remote_path s3://$bucket_name/$s3_key --region $region --sse AES256; else echo 'FILE_NOT_FOUND'; fi"
    
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
        return 1
    fi
    
    # Wait for command completion
    if ! wait_for_command_completion "$command_id" "$instance_id" "$region"; then
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
        return 1
    fi
    
    # Download from S3
    log_info "Downloading from S3: s3://$bucket_name/$s3_key"
    
    if aws s3 cp "s3://$bucket_name/$s3_key" "$local_file" --region "$region"; then
        # Clean up S3 object
        aws s3 rm "s3://$bucket_name/$s3_key" --region "$region" >/dev/null 2>&1
        log_info "Large file downloaded successfully via S3"
        return 0
    else
        log_error "Failed to download file from S3"
        # Clean up S3 object
        aws s3 rm "s3://$bucket_name/$s3_key" --region "$region" >/dev/null 2>&1
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
    
    local escaped_remote_path
    escaped_remote_path=$(printf '%q' "$remote_path")
    local size_command
    size_command=$(get_remote_file_size_command "$escaped_remote_path")
    
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
    file_size=$(echo "$size_output" | tr -d '\n' | grep -o '[0-9][0-9]*')
    
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