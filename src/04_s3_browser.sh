#!/usr/bin/env bash

# S3 Browser Module for ZTiAWS
# Provides simplified S3 operations

# Check if AWS CLI is configured for S3
check_s3_config() {
    if ! aws sts get-caller-identity >/dev/null 2>&1; then
        log_error "AWS CLI not configured or credentials invalid"
        log_info "Run 'aws configure' to set up your credentials"
        return 1
    fi
    return 0
}

# List all S3 buckets
s3_list_buckets() {
    log_info "Listing all S3 buckets..."
    echo
    
    if ! check_s3_config; then
        return 1
    fi
    
    aws s3 ls | while read -r line; do
        if [[ -n "$line" ]]; then
            bucket_date=$(echo "$line" | awk '{print $1, $2}')
            bucket_name=$(echo "$line" | awk '{print $3}')
            echo "í³¦ $bucket_name (Created: $bucket_date)"
        fi
    done
    
    bucket_count=$(aws s3 ls | wc -l)
    echo
    log_info "Found $bucket_count bucket(s)"
}

# List objects in a specific bucket
s3_list_objects() {
    local bucket_name="$1"
    
    if [[ -z "$bucket_name" ]]; then
        log_error "Usage: ssm s3 ls <bucket-name>"
        return 1
    fi
    
    if ! check_s3_config; then
        return 1
    fi
    
    log_info "Listing contents of bucket: $bucket_name"
    echo
    
    if ! aws s3 ls "s3://$bucket_name" >/dev/null 2>&1; then
        log_error "Bucket '$bucket_name' not found or access denied"
        return 1
    fi
    
    aws s3 ls "s3://$bucket_name" --recursive --human-readable | while read -r line; do
        if [[ -n "$line" ]]; then
            date_time=$(echo "$line" | awk '{print $1, $2}')
            size=$(echo "$line" | awk '{print $3}')
            file_path=$(echo "$line" | awk '{$1=$2=$3=""; print $0}' | sed 's/^ *//')
            echo "í³„ $file_path ($size) - $date_time"
        fi
    done
    
    object_count=$(aws s3 ls "s3://$bucket_name" --recursive | wc -l)
    if [ "$object_count" -eq 0 ]; then
        log_warn "Bucket is empty"
    else
        echo
        log_info "Found $object_count object(s) in bucket"
    fi
}

# Download a file from S3
s3_get_object() {
    local bucket_name="$1"
    local object_key="$2"
    
    if [[ -z "$bucket_name" ]] || [[ -z "$object_key" ]]; then
        log_error "Usage: ssm s3 get <bucket-name> <object-key>"
        return 1
    fi
    
    if ! check_s3_config; then
        return 1
    fi
    
    local local_file
    local_file=$(basename "$object_key")
    
    log_info "Downloading $object_key from bucket $bucket_name..."
    
    if aws s3 cp "s3://$bucket_name/$object_key" "./$local_file"; then
        log_info "âœ… Downloaded $object_key as $local_file"
    else
        log_error "Failed to download $object_key"
        return 1
    fi
}

# Upload a file to S3
s3_put_object() {
    local bucket_name="$1"
    local local_file="$2"
    
    if [[ -z "$bucket_name" ]] || [[ -z "$local_file" ]]; then
        log_error "Usage: ssm s3 put <bucket-name> <local-file>"
        return 1
    fi
    
    if [[ ! -f "$local_file" ]]; then
        log_error "File '$local_file' not found"
        return 1
    fi
    
    if ! check_s3_config; then
        return 1
    fi
    
    log_info "Uploading $local_file to bucket $bucket_name..."
    
    if aws s3 cp "$local_file" "s3://$bucket_name/"; then
        log_info "âœ… Uploaded $local_file successfully"
        log_info "File available at: s3://$bucket_name/$local_file"
    else
        log_error "Failed to upload $local_file"
        return 1
    fi
}

# Handle S3 commands
handle_s3_command() {
    local subcommand="$1"
    shift
    
    case "$subcommand" in
        "list")
            s3_list_buckets
            ;;
        "ls")
            s3_list_objects "$@"
            ;;
        "get")
            s3_get_object "$@"
            ;;
        "put")
            s3_put_object "$@"
            ;;
        "help"|"-h"|"--help")
            echo "SSM S3 Browser Commands:"
            echo ""
            echo "  ssm s3 list                    - List all buckets"
            echo "  ssm s3 ls <bucket>            - List objects in bucket"
            echo "  ssm s3 get <bucket> <file>    - Download a file"
            echo "  ssm s3 put <bucket> <file>    - Upload a file"
            ;;
        *)
            log_error "Unknown S3 command: $subcommand"
            echo "Usage: ssm s3 [list|ls|get|put|help]"
            return 1
            ;;
    esac
}
