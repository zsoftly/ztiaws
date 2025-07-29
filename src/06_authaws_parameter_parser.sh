#!/usr/bin/env bash

# AuthAWS Parameter Parser Module
# Version: 1.0.0
# Repository: https://github.com/ZSoftly/ztiaws
# 
# This module provides unified parameter parsing for authaws, supporting both
# positional and flag-based syntax while maintaining full backward compatibility.
# 
# Compatibility: bash 3.2+ (macOS compatible)
# Usage: source this file in authaws script, then call parse_authaws_parameters "$@"

# Global variables to store parsed parameters
# Using regular variable assignment for bash 3.2 compatibility
AUTH_PROFILE=""
AUTH_COMMAND=""
AUTH_REGION=""
AUTH_SSO_URL=""
AUTH_EXPORT="false"
AUTH_LIST_PROFILES="false"
AUTH_CHECK="false"
AUTH_CREDS="false"
AUTH_HELP="false"
AUTH_VERSION="false"
AUTH_DEBUG="false"

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
AuthAWS Parameter Parser Help

This module supports both positional and flag-based parameter syntax:

POSITIONAL SYNTAX (current - backward compatible):
  authaws [profile-name]           # Use specific profile
  authaws check                    # Check system requirements
  authaws creds [profile]          # Show credentials
  authaws help                     # Show help
  authaws version                  # Show version

FLAG-BASED SYNTAX (new - enterprise friendly):
  authaws --profile <name>         # Use specific profile
  authaws --check                  # Check system requirements
  authaws --creds [--profile <name>] # Show credentials
  authaws --help                   # Show help
  authaws --version                # Show version

MIXED SYNTAX (also supported):
  authaws dev-profile --region us-west-2
  authaws --profile dev-profile --export

AVAILABLE FLAGS:
  --profile, -p <name>             # Profile name to use
  --region, -r <region>            # Override default region
  --sso-url <url>                  # Override SSO start URL
  --export                         # Export credentials to stdout
  --list-profiles                  # List available profiles
  --check                          # Check system requirements
  --creds                          # Show credentials
  --help, -h                       # Show help
  --version, -v                    # Show version
  --debug                          # Enable debug mode

EXAMPLES:
  # Power user workflow (current syntax)
  authaws dev-profile
  authaws creds

  # Enterprise workflow (new syntax)
  authaws --profile dev-profile
  authaws --profile prod --region us-east-1
  authaws --creds --profile dev-profile

  # Future extensibility
  authaws --profile dev --sso-url https://alt.awsapps.com/start
  authaws --profile dev --export
EOL
}

# Function to parse positional parameters (current behavior)
parse_positional_parameters() {
    local args=("$@")
    
    if [[ ${#args[@]} -eq 0 ]]; then
        # No arguments - use default profile
        AUTH_PROFILE=""
        AUTH_COMMAND=""
        return 0
    fi
    
    local first_arg="${args[0]}"
    
    # Check for commands first
    case "$first_arg" in
        "check")
            AUTH_COMMAND="check"
            AUTH_CHECK="true"
            return 0
            ;;
        "creds")
            AUTH_COMMAND="creds"
            AUTH_CREDS="true"
            # Check if second argument is a profile
            if [[ ${#args[@]} -gt 1 ]]; then
                AUTH_PROFILE="${args[1]}"
            fi
            return 0
            ;;
        "help")
            AUTH_COMMAND="help"
            AUTH_HELP="true"
            return 0
            ;;
        "version")
            AUTH_COMMAND="version"
            AUTH_VERSION="true"
            return 0
            ;;
        *)
            # Assume it's a profile name
            AUTH_PROFILE="$first_arg"
            AUTH_COMMAND="login"
            return 0
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
            # Profile flags
            "--profile"|"-p")
                if [[ $((i+1)) -lt ${#args[@]} ]]; then
                    AUTH_PROFILE="${args[$((i+1))]}"
                    i=$((i+2))
                else
                    log_error "Error: --profile requires a value"
                    return 1
                fi
                ;;
            
            # Region flag
            "--region"|"-r")
                if [[ $((i+1)) -lt ${#args[@]} ]]; then
                    AUTH_REGION="${args[$((i+1))]}"
                    i=$((i+2))
                else
                    log_error "Error: --region requires a value"
                    return 1
                fi
                ;;
            
            # SSO URL flag
            "--sso-url")
                if [[ $((i+1)) -lt ${#args[@]} ]]; then
                    AUTH_SSO_URL="${args[$((i+1))]}"
                    i=$((i+2))
                else
                    log_error "Error: --sso-url requires a value"
                    return 1
                fi
                ;;
            
            # Boolean flags
            "--export")
                AUTH_EXPORT="true"
                i=$((i+1))
                ;;
            
            "--list-profiles")
                AUTH_LIST_PROFILES="true"
                AUTH_COMMAND="list-profiles"
                i=$((i+1))
                ;;
            
            "--check")
                AUTH_CHECK="true"
                AUTH_COMMAND="check"
                i=$((i+1))
                ;;
            
            "--creds")
                AUTH_CREDS="true"
                AUTH_COMMAND="creds"
                i=$((i+1))
                ;;
            
            "--help"|"-h")
                AUTH_HELP="true"
                AUTH_COMMAND="help"
                i=$((i+1))
                ;;
            
            "--version"|"-v")
                AUTH_VERSION="true"
                AUTH_COMMAND="version"
                i=$((i+1))
                ;;
            
            "--debug")
                AUTH_DEBUG="true"
                i=$((i+1))
                ;;
            
            # Unknown flag
            -*)
                log_error "Error: Unknown flag '$arg'"
                log_error "Use 'authaws --help' for usage information"
                return 1
                ;;
            
            # Positional argument (fallback for mixed syntax)
            *)
                # If we haven't set a profile yet, treat this as a profile name
                if [[ -z "$AUTH_PROFILE" && -z "$AUTH_COMMAND" ]]; then
                    AUTH_PROFILE="$arg"
                    AUTH_COMMAND="login"
                else
                    log_error "Error: Unexpected positional argument '$arg'"
                    log_error "Use 'authaws --help' for usage information"
                    return 1
                fi
                i=$((i+1))
                ;;
        esac
    done
    
    return 0
}

# Main parameter parsing function
parse_authaws_parameters() {
    local args=("$@")
    
    # Reset global variables
    AUTH_PROFILE=""
    AUTH_COMMAND=""
    AUTH_REGION=""
    AUTH_SSO_URL=""
    AUTH_EXPORT="false"
    AUTH_LIST_PROFILES="false"
    AUTH_CHECK="false"
    AUTH_CREDS="false"
    AUTH_HELP="false"
    AUTH_VERSION="false"
    AUTH_DEBUG="false"
    
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
    
    # Validate parameter combinations
    if ! validate_parameter_combinations; then
        return 1
    fi
    
    # Debug output if debug mode is enabled
    if [[ "$AUTH_DEBUG" == "true" ]]; then
        debug_log "Parsed parameters:"
        debug_log "  Profile: '$AUTH_PROFILE'"
        debug_log "  Command: '$AUTH_COMMAND'"
        debug_log "  Region: '$AUTH_REGION'"
        debug_log "  SSO URL: '$AUTH_SSO_URL'"
        debug_log "  Export: '$AUTH_EXPORT'"
        debug_log "  List Profiles: '$AUTH_LIST_PROFILES'"
        debug_log "  Check: '$AUTH_CHECK'"
        debug_log "  Creds: '$AUTH_CREDS'"
        debug_log "  Help: '$AUTH_HELP'"
        debug_log "  Version: '$AUTH_VERSION'"
        debug_log "  Debug: '$AUTH_DEBUG'"
    fi
    
    return 0
}

# Function to validate parameter combinations
validate_parameter_combinations() {
    # Check for mutually exclusive commands
    local command_count=0
    [[ "$AUTH_CHECK" == "true" ]] && command_count=$((command_count + 1))
    [[ "$AUTH_CREDS" == "true" ]] && command_count=$((command_count + 1))
    [[ "$AUTH_HELP" == "true" ]] && command_count=$((command_count + 1))
    [[ "$AUTH_VERSION" == "true" ]] && command_count=$((command_count + 1))
    [[ "$AUTH_LIST_PROFILES" == "true" ]] && command_count=$((command_count + 1))
    
    if [[ $command_count -gt 1 ]]; then
        log_error "Error: Multiple commands specified. Only one command allowed."
        log_error "Commands: check, creds, help, version, list-profiles"
        return 1
    fi
    
    # Validate region if specified
    if [[ -n "$AUTH_REGION" ]]; then
        # Check if validate_region_code function exists (from regions.sh)
        if type validate_region_code >/dev/null 2>&1; then
            if ! validate_region_code "$AUTH_REGION" region_var; then
                log_error "Error: Invalid region '$AUTH_REGION'"
                return 1
            fi
        else
            # Fallback validation for common AWS regions
            if ! validate_aws_region "$AUTH_REGION"; then
                log_error "Error: Invalid region '$AUTH_REGION'"
                return 1
            fi
        fi
    fi
    
    # Validate SSO URL if specified
    if [[ -n "$AUTH_SSO_URL" ]]; then
        if ! validate_sso_url "$AUTH_SSO_URL"; then
            log_error "Error: Invalid SSO URL '$AUTH_SSO_URL'"
            return 1
        fi
    fi
    
    return 0
}

# Function to validate SSO URL format
validate_sso_url() {
    local url="$1"
    
    # Basic URL validation for AWS SSO URLs
    if [[ "$url" =~ ^https://[a-zA-Z0-9.-]+\.awsapps\.com/start$ ]]; then
        return 0
    fi
    
    return 1
}

# Fallback function to validate AWS regions (when regions.sh is not available)
validate_aws_region() {
    local region="$1"
    
    # List of common AWS regions for validation
    local valid_regions=(
        "us-east-1" "us-east-2" "us-west-1" "us-west-2"
        "ca-central-1" "ca-west-1"
        "eu-west-1" "eu-west-2" "eu-west-3" "eu-central-1"
        "ap-south-1" "ap-southeast-1" "ap-southeast-2"
        "ap-northeast-1" "ap-northeast-2" "ap-northeast-3"
        "sa-east-1"
    )
    
    for valid_region in "${valid_regions[@]}"; do
        if [[ "$region" == "$valid_region" ]]; then
            return 0
        fi
    done
    
    return 1
}

# Function to get parsed parameters (for use in main script)
get_auth_profile() {
    echo "$AUTH_PROFILE"
}

get_auth_command() {
    echo "$AUTH_COMMAND"
}

get_auth_region() {
    echo "$AUTH_REGION"
}

get_auth_sso_url() {
    echo "$AUTH_SSO_URL"
}

get_auth_export() {
    echo "$AUTH_EXPORT"
}

get_auth_list_profiles() {
    echo "$AUTH_LIST_PROFILES"
}

get_auth_check() {
    echo "$AUTH_CHECK"
}

get_auth_creds() {
    echo "$AUTH_CREDS"
}

get_auth_help() {
    echo "$AUTH_HELP"
}

get_auth_version() {
    echo "$AUTH_VERSION"
}

get_auth_debug() {
    echo "$AUTH_DEBUG"
}

# Function to check if a specific flag is set
is_flag_set() {
    local flag="$1"
    case "$flag" in
        "export") [[ "$AUTH_EXPORT" == "true" ]] && return 0 ;;
        "list-profiles") [[ "$AUTH_LIST_PROFILES" == "true" ]] && return 0 ;;
        "check") [[ "$AUTH_CHECK" == "true" ]] && return 0 ;;
        "creds") [[ "$AUTH_CREDS" == "true" ]] && return 0 ;;
        "help") [[ "$AUTH_HELP" == "true" ]] && return 0 ;;
        "version") [[ "$AUTH_VERSION" == "true" ]] && return 0 ;;
        "debug") [[ "$AUTH_DEBUG" == "true" ]] && return 0 ;;
        *) return 1 ;;
    esac
    return 1
} 