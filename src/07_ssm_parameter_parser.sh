#!/usr/bin/env bash

# SSM Parameter Parser Module
# Version: 1.0.0
# Repository: https://github.com/ZSoftly/ztiaws
# 
# This module provides unified parameter parsing for ssm, supporting both
# positional and flag-based syntax while maintaining full backward compatibility.
# 
# Compatibility: bash 3.2+ (macOS compatible)
# Usage: source this file in ssm script, then call parse_ssm_parameters "$@"

# Global variables to store parsed parameters
# Using regular variable assignment for bash 3.2 compatibility
SSM_REGION=""
SSM_INSTANCE=""
SSM_COMMAND=""
SSM_TAG_KEY=""
SSM_TAG_VALUE=""
SSM_LOCAL_FILE=""
SSM_REMOTE_FILE=""
SSM_LOCAL_PATH=""
SSM_REMOTE_PATH=""
SSM_LOCAL_PORT=""
SSM_REMOTE_PORT=""
SSM_OPERATION=""  # connect, list, exec, exec-tagged, upload, download, forward, check, version, help
SSM_HELP="false"
SSM_VERSION="false"
SSM_CHECK="false"
SSM_DEBUG="false"

# Function to detect if arguments contain flags
has_flags() {
    local args=("$@")
    for arg in "${args[@]}"; do
        if [[ "$arg" =~ ^- ]]; then
            return 0  # Found a flag
        fi
    done
    return 1  # No flags found
}

# Function to show parameter parser help
show_parameter_help() {
    cat << 'EOL'
SSM Parameter Parser Help

This module supports both positional and flag-based parameter syntax:

POSITIONAL SYNTAX (current - backward compatible):
  ssm [region]                     # List instances in region
  ssm [region] [instance]          # Connect to instance
  ssm exec [region] [instance] [command]    # Execute command
  ssm upload [region] [instance] [local] [remote]  # Upload file
  ssm download [region] [instance] [remote] [local]  # Download file
  ssm check                        # Check system requirements
  ssm version                      # Show version
  ssm help                         # Show help

FLAG-BASED SYNTAX (new - enterprise friendly):
  ssm --region [region]            # List instances in region
  ssm --region [region] --instance [instance]  # Connect to instance
  ssm --exec --region [region] --instance [instance] --command [command]  # Execute command
  ssm --upload --region [region] --instance [instance] --local-file [local] --remote-path [remote]  # Upload file
  ssm --download --region [region] --instance [instance] --remote-file [remote] --local-path [local]  # Download file
  ssm --check                      # Check system requirements
  ssm --version                    # Show version
  ssm --help                       # Show help

MIXED SYNTAX (also supported):
  ssm cac1 --instance i-1234567890abcd
  ssm --region cac1 i-1234567890abcd

AVAILABLE FLAGS:
  --region, -r <region>            # AWS region or region code
  --instance, -i <instance>        # Instance ID or name
  --command, -c <command>          # Command to execute
  --tag-key <key>                  # Tag key for exec-tagged
  --tag-value <value>              # Tag value for exec-tagged
  --local-file <path>              # Local file path for upload
  --remote-file <path>             # Remote file path for download
  --local-path <path>              # Local path for download
  --remote-path <path>             # Remote path for upload
  --local-port <port>              # Local port for forwarding
  --remote-port <port>             # Remote port for forwarding
  --exec                           # Execute command operation
  --exec-tagged                    # Execute command on tagged instances
  --upload                         # Upload file operation
  --download                       # Download file operation
  --forward                        # Port forwarding operation
  --list                           # List instances operation
  --connect                        # Connect to instance operation
  --check                          # Check system requirements
  --help, -h                       # Show help
  --version, -v                    # Show version
  --debug                          # Enable debug mode

EXAMPLES:
  # Power user workflow (current syntax)
  ssm cac1                         # List instances in Canada Central
  ssm cac1 i-1234                  # Connect to instance
  ssm exec cac1 i-1234 "uptime"    # Execute command

  # Enterprise workflow (new syntax)
  ssm --region cac1 --list         # List instances
  ssm --region cac1 --instance i-1234 --connect  # Connect to instance
  ssm --exec --region cac1 --instance i-1234 --command "uptime"  # Execute command

  # Mixed workflow
  ssm cac1 --instance i-1234       # List region with specific instance
  ssm --region cac1 i-1234         # Use flag for region, positional for instance
EOL
}

# Function to parse positional parameters (current behavior)
parse_positional_parameters() {
    local args=("$@")
    
    if [[ ${#args[@]} -eq 0 ]]; then
        # No arguments - show help
        SSM_HELP="true"
        SSM_OPERATION="help"
        return 0
    fi
    
    local first_arg="${args[0]}"
    
    # Check for commands first
    case "$first_arg" in
        "check")
            SSM_OPERATION="check"
            SSM_CHECK="true"
            return 0
            ;;
        "version")
            SSM_OPERATION="version"
            SSM_VERSION="true"
            return 0
            ;;
        "help"|"-h"|"--help")
            SSM_OPERATION="help"
            SSM_HELP="true"
            return 0
            ;;
        "install")
            SSM_OPERATION="install"
            return 0
            ;;
        "exec")
            if [[ ${#args[@]} -lt 4 ]]; then
                log_error "Missing required parameters for exec command"
                log_error "Usage: $(basename "$0") exec <region> <instance-identifier> \"<command>\""
                return 1
            fi
            SSM_OPERATION="exec"
            SSM_REGION="${args[1]}"
            SSM_INSTANCE="${args[2]}"
            SSM_COMMAND="${args[3]}"
            return 0
            ;;
        "exec-tagged")
            if [[ ${#args[@]} -lt 5 ]]; then
                log_error "Missing required parameters for exec-tagged command"
                log_error "Usage: $(basename "$0") exec-tagged <region> <tag-key> <tag-value> \"<command>\""
                return 1
            fi
            SSM_OPERATION="exec-tagged"
            SSM_REGION="${args[1]}"
            SSM_TAG_KEY="${args[2]}"
            SSM_TAG_VALUE="${args[3]}"
            SSM_COMMAND="${args[4]}"
            return 0
            ;;
        "upload")
            if [[ ${#args[@]} -lt 5 ]]; then
                log_error "Missing required parameters for upload command"
                log_error "Usage: $(basename "$0") upload <region> <instance-identifier> <local-file> <remote-path>"
                return 1
            fi
            SSM_OPERATION="upload"
            SSM_REGION="${args[1]}"
            SSM_INSTANCE="${args[2]}"
            SSM_LOCAL_FILE="${args[3]}"
            SSM_REMOTE_PATH="${args[4]}"
            return 0
            ;;
        "download")
            if [[ ${#args[@]} -lt 5 ]]; then
                log_error "Missing required parameters for download command"
                log_error "Usage: $(basename "$0") download <region> <instance-identifier> <remote-file> <local-path>"
                return 1
            fi
            SSM_OPERATION="download"
            SSM_REGION="${args[1]}"
            SSM_INSTANCE="${args[2]}"
            SSM_REMOTE_FILE="${args[3]}"
            SSM_LOCAL_PATH="${args[4]}"
            return 0
            ;;
        *)
            # Check if it's a region code
            local test_region
            if validate_region_code "$first_arg" test_region; then
                SSM_REGION="$test_region"
                if [[ ${#args[@]} -eq 1 ]]; then
                    # Just region - list instances
                    SSM_OPERATION="list"
                elif [[ ${#args[@]} -eq 2 ]]; then
                    # Region and instance - connect
                    SSM_OPERATION="connect"
                    SSM_INSTANCE="${args[1]}"
                else
                    log_error "Too many arguments for region/instance syntax"
                    return 1
                fi
                return 0
            else
                # Assume it's an instance identifier for default region
                SSM_OPERATION="connect"
                SSM_REGION="ca-central-1"  # Default region
                SSM_INSTANCE="$first_arg"
                return 0
            fi
            ;;
    esac
}

# Function to parse flag-based parameters
parse_flag_parameters() {
    local args=("$@")
    local i=0
    
    while [[ $i -lt ${#args[@]} ]]; do
        local arg="${args[$i]}"
        
        case "$arg" in
            # Region flags
            "--region"|"-r")
                if [[ $((i+1)) -lt ${#args[@]} ]]; then
                    SSM_REGION="${args[$((i+1))]}"
                    i=$((i+2))
                else
                    log_error "Error: --region requires a value"
                    return 1
                fi
                ;;
            
            # Instance flags
            "--instance"|"-i")
                if [[ $((i+1)) -lt ${#args[@]} ]]; then
                    SSM_INSTANCE="${args[$((i+1))]}"
                    i=$((i+2))
                else
                    log_error "Error: --instance requires a value"
                    return 1
                fi
                ;;
            
            # Command flag
            "--command"|"-c")
                if [[ $((i+1)) -lt ${#args[@]} ]]; then
                    SSM_COMMAND="${args[$((i+1))]}"
                    i=$((i+2))
                else
                    log_error "Error: --command requires a value"
                    return 1
                fi
                ;;
            
            # Tag flags
            "--tag-key")
                if [[ $((i+1)) -lt ${#args[@]} ]]; then
                    SSM_TAG_KEY="${args[$((i+1))]}"
                    i=$((i+2))
                else
                    log_error "Error: --tag-key requires a value"
                    return 1
                fi
                ;;
            
            "--tag-value")
                if [[ $((i+1)) -lt ${#args[@]} ]]; then
                    SSM_TAG_VALUE="${args[$((i+1))]}"
                    i=$((i+2))
                else
                    log_error "Error: --tag-value requires a value"
                    return 1
                fi
                ;;
            
            # File path flags
            "--local-file")
                if [[ $((i+1)) -lt ${#args[@]} ]]; then
                    SSM_LOCAL_FILE="${args[$((i+1))]}"
                    i=$((i+2))
                else
                    log_error "Error: --local-file requires a value"
                    return 1
                fi
                ;;
            
            "--remote-file")
                if [[ $((i+1)) -lt ${#args[@]} ]]; then
                    SSM_REMOTE_FILE="${args[$((i+1))]}"
                    i=$((i+2))
                else
                    log_error "Error: --remote-file requires a value"
                    return 1
                fi
                ;;
            
            "--local-path")
                if [[ $((i+1)) -lt ${#args[@]} ]]; then
                    SSM_LOCAL_PATH="${args[$((i+1))]}"
                    i=$((i+2))
                else
                    log_error "Error: --local-path requires a value"
                    return 1
                fi
                ;;
            
            "--remote-path")
                if [[ $((i+1)) -lt ${#args[@]} ]]; then
                    SSM_REMOTE_PATH="${args[$((i+1))]}"
                    i=$((i+2))
                else
                    log_error "Error: --remote-path requires a value"
                    return 1
                fi
                ;;
            
            # Port flags
            "--local-port")
                if [[ $((i+1)) -lt ${#args[@]} ]]; then
                    SSM_LOCAL_PORT="${args[$((i+1))]}"
                    i=$((i+2))
                else
                    log_error "Error: --local-port requires a value"
                    return 1
                fi
                ;;
            
            "--remote-port")
                if [[ $((i+1)) -lt ${#args[@]} ]]; then
                    SSM_REMOTE_PORT="${args[$((i+1))]}"
                    i=$((i+2))
                else
                    log_error "Error: --remote-port requires a value"
                    return 1
                fi
                ;;
            
            # Operation flags
            "--exec")
                SSM_OPERATION="exec"
                i=$((i+1))
                ;;
            
            "--exec-tagged")
                SSM_OPERATION="exec-tagged"
                i=$((i+1))
                ;;
            
            "--upload")
                SSM_OPERATION="upload"
                i=$((i+1))
                ;;
            
            "--download")
                SSM_OPERATION="download"
                i=$((i+1))
                ;;
            
            "--forward")
                SSM_OPERATION="forward"
                i=$((i+1))
                ;;
            
            "--list")
                SSM_OPERATION="list"
                i=$((i+1))
                ;;
            
            "--connect")
                SSM_OPERATION="connect"
                i=$((i+1))
                ;;
            
            # System flags
            "--check")
                SSM_CHECK="true"
                SSM_OPERATION="check"
                i=$((i+1))
                ;;
            
            "--help"|"-h")
                SSM_HELP="true"
                SSM_OPERATION="help"
                i=$((i+1))
                ;;
            
            "--version"|"-v")
                SSM_VERSION="true"
                SSM_OPERATION="version"
                i=$((i+1))
                ;;
            
            "--debug")
                SSM_DEBUG="true"
                i=$((i+1))
                ;;
            
            # Unknown flag
            -*)
                log_error "Error: Unknown flag '$arg'"
                log_error "Use 'ssm --help' for usage information"
                return 1
                ;;
            
            # Positional argument (fallback for mixed syntax)
            *)
                # If we haven't set a region yet and this looks like a region, use it
                local test_region
                if [[ -z "$SSM_REGION" ]] && validate_region_code "$arg" test_region; then
                    SSM_REGION="$test_region"
                # If we haven't set an instance yet, treat this as an instance
                elif [[ -z "$SSM_INSTANCE" ]]; then
                    SSM_INSTANCE="$arg"
                # If we haven't set a command yet and this doesn't start with -, treat as command
                elif [[ -z "$SSM_COMMAND" && ! "$arg" =~ ^- ]]; then
                    SSM_COMMAND="$arg"
                else
                    log_error "Error: Unexpected positional argument '$arg'"
                    log_error "Use 'ssm --help' for usage information"
                    return 1
                fi
                i=$((i+1))
                ;;
        esac
    done
    
    return 0
}

# Main parameter parsing function
parse_ssm_parameters() {
    local args=("$@")
    
    # Reset global variables
    SSM_REGION=""
    SSM_INSTANCE=""
    SSM_COMMAND=""
    SSM_TAG_KEY=""
    SSM_TAG_VALUE=""
    SSM_LOCAL_FILE=""
    SSM_REMOTE_FILE=""
    SSM_LOCAL_PATH=""
    SSM_REMOTE_PATH=""
    SSM_LOCAL_PORT=""
    SSM_REMOTE_PORT=""
    SSM_OPERATION=""
    SSM_HELP="false"
    SSM_VERSION="false"
    SSM_CHECK="false"
    SSM_DEBUG="false"
    
    # Check if arguments contain flags
    if has_flags "${args[@]}"; then
        # Parse as flag-based parameters
        if ! parse_flag_parameters "${args[@]}"; then
            return 1
        fi
    else
        # Parse as positional parameters (backward compatibility)
        if ! parse_positional_parameters "${args[@]}"; then
            return 1
        fi
    fi
    
    # Validate parameter combinations and infer operation if not set
    if ! validate_and_infer_operation; then
        return 1
    fi
    
    # Debug output if debug mode is enabled
    if [[ "$SSM_DEBUG" == "true" ]]; then
        debug_log "Parsed parameters:"
        debug_log "  Region: '$SSM_REGION'"
        debug_log "  Instance: '$SSM_INSTANCE'"
        debug_log "  Command: '$SSM_COMMAND'"
        debug_log "  Tag Key: '$SSM_TAG_KEY'"
        debug_log "  Tag Value: '$SSM_TAG_VALUE'"
        debug_log "  Local File: '$SSM_LOCAL_FILE'"
        debug_log "  Remote File: '$SSM_REMOTE_FILE'"
        debug_log "  Local Path: '$SSM_LOCAL_PATH'"
        debug_log "  Remote Path: '$SSM_REMOTE_PATH'"
        debug_log "  Local Port: '$SSM_LOCAL_PORT'"
        debug_log "  Remote Port: '$SSM_REMOTE_PORT'"
        debug_log "  Operation: '$SSM_OPERATION'"
        debug_log "  Help: '$SSM_HELP'"
        debug_log "  Version: '$SSM_VERSION'"
        debug_log "  Check: '$SSM_CHECK'"
        debug_log "  Debug: '$SSM_DEBUG'"
    fi
    
    return 0
}

# Function to validate parameter combinations and infer operation
validate_and_infer_operation() {
    # If no operation is set, try to infer it
    if [[ -z "$SSM_OPERATION" ]]; then
        # Infer operation based on parameters
        if [[ -n "$SSM_COMMAND" && -n "$SSM_TAG_KEY" && -n "$SSM_TAG_VALUE" ]]; then
            SSM_OPERATION="exec-tagged"
        elif [[ -n "$SSM_COMMAND" ]]; then
            SSM_OPERATION="exec"
        elif [[ -n "$SSM_LOCAL_FILE" && -n "$SSM_REMOTE_PATH" ]]; then
            SSM_OPERATION="upload"
        elif [[ -n "$SSM_REMOTE_FILE" && -n "$SSM_LOCAL_PATH" ]]; then
            SSM_OPERATION="download"
        elif [[ -n "$SSM_LOCAL_PORT" && -n "$SSM_REMOTE_PORT" ]]; then
            SSM_OPERATION="forward"
        elif [[ -n "$SSM_INSTANCE" ]]; then
            SSM_OPERATION="connect"
        elif [[ -n "$SSM_REGION" ]]; then
            SSM_OPERATION="list"
        fi
    fi
    
    # Validate region if specified
    if [[ -n "$SSM_REGION" ]]; then
        # Check if it's already a full region name (e.g., "us-east-1")
        if [[ "$SSM_REGION" =~ ^[a-z]+-[a-z]+-[0-9]+$ ]]; then
            # Already a full region name, no need to validate
            :
        else
            # It's a region code, validate and convert
            local validated_region
            if ! validate_region_code "$SSM_REGION" validated_region; then
                log_error "Error: Invalid region '$SSM_REGION'"
                return 1
            fi
            # Update SSM_REGION with the validated full region name
            SSM_REGION="$validated_region"
        fi
    fi
    
    # Set default region if not specified and operation requires it
    if [[ -z "$SSM_REGION" && "$SSM_OPERATION" =~ ^(connect|exec|exec-tagged|upload|download|forward|list)$ ]]; then
        SSM_REGION="ca-central-1"  # Default region
    fi
    
    # Validate required parameters for each operation
    case "$SSM_OPERATION" in
        "exec")
            if [[ -z "$SSM_INSTANCE" ]]; then
                log_error "Error: --instance is required for exec operation"
                return 1
            fi
            if [[ -z "$SSM_COMMAND" ]]; then
                log_error "Error: --command is required for exec operation"
                return 1
            fi
            ;;
        "exec-tagged")
            if [[ -z "$SSM_TAG_KEY" ]]; then
                log_error "Error: --tag-key is required for exec-tagged operation"
                return 1
            fi
            if [[ -z "$SSM_TAG_VALUE" ]]; then
                log_error "Error: --tag-value is required for exec-tagged operation"
                return 1
            fi
            if [[ -z "$SSM_COMMAND" ]]; then
                log_error "Error: --command is required for exec-tagged operation"
                return 1
            fi
            ;;
        "upload")
            if [[ -z "$SSM_INSTANCE" ]]; then
                log_error "Error: --instance is required for upload operation"
                return 1
            fi
            if [[ -z "$SSM_LOCAL_FILE" ]]; then
                log_error "Error: --local-file is required for upload operation"
                return 1
            fi
            if [[ -z "$SSM_REMOTE_PATH" ]]; then
                log_error "Error: --remote-path is required for upload operation"
                return 1
            fi
            ;;
        "download")
            if [[ -z "$SSM_INSTANCE" ]]; then
                log_error "Error: --instance is required for download operation"
                return 1
            fi
            if [[ -z "$SSM_REMOTE_FILE" ]]; then
                log_error "Error: --remote-file is required for download operation"
                return 1
            fi
            if [[ -z "$SSM_LOCAL_PATH" ]]; then
                log_error "Error: --local-path is required for download operation"
                return 1
            fi
            ;;
        "forward")
            if [[ -z "$SSM_INSTANCE" ]]; then
                log_error "Error: --instance is required for forward operation"
                return 1
            fi
            if [[ -z "$SSM_LOCAL_PORT" ]]; then
                log_error "Error: --local-port is required for forward operation"
                return 1
            fi
            if [[ -z "$SSM_REMOTE_PORT" ]]; then
                log_error "Error: --remote-port is required for forward operation"
                return 1
            fi
            ;;
        "connect")
            if [[ -z "$SSM_INSTANCE" ]]; then
                log_error "Error: --instance is required for connect operation"
                return 1
            fi
            ;;
    esac
    
    return 0
}

# Function to get parsed parameters (for use in main script)
get_ssm_region() {
    echo "$SSM_REGION"
}

get_ssm_instance() {
    echo "$SSM_INSTANCE"
}

get_ssm_command() {
    echo "$SSM_COMMAND"
}

get_ssm_tag_key() {
    echo "$SSM_TAG_KEY"
}

get_ssm_tag_value() {
    echo "$SSM_TAG_VALUE"
}

get_ssm_local_file() {
    echo "$SSM_LOCAL_FILE"
}

get_ssm_remote_file() {
    echo "$SSM_REMOTE_FILE"
}

get_ssm_local_path() {
    echo "$SSM_LOCAL_PATH"
}

get_ssm_remote_path() {
    echo "$SSM_REMOTE_PATH"
}

get_ssm_local_port() {
    echo "$SSM_LOCAL_PORT"
}

get_ssm_remote_port() {
    echo "$SSM_REMOTE_PORT"
}

get_ssm_operation() {
    echo "$SSM_OPERATION"
}

get_ssm_help() {
    echo "$SSM_HELP"
}

get_ssm_version() {
    echo "$SSM_VERSION"
}

get_ssm_check() {
    echo "$SSM_CHECK"
}

get_ssm_debug() {
    echo "$SSM_DEBUG"
}

# Function to check if a specific flag is set
is_ssm_flag_set() {
    local flag="$1"
    case "$flag" in
        "help") [[ "$SSM_HELP" == "true" ]] && return 0 ;;
        "version") [[ "$SSM_VERSION" == "true" ]] && return 0 ;;
        "check") [[ "$SSM_CHECK" == "true" ]] && return 0 ;;
        "debug") [[ "$SSM_DEBUG" == "true" ]] && return 0 ;;
        *) return 1 ;;
    esac
    return 1
}
