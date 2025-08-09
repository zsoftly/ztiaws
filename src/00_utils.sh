#!/usr/bin/env bash

# Common Utility Functions and Definitions
# Repository: https://github.com/ZSoftly/ztiaws

# Color definitions
# Export these variables so they're available to other scripts that source this file
export RED='\033[0;31m'
export GREEN='\033[0;32m'
export YELLOW='\033[1;33m'
export CYAN='\033[0;36m'
export NC='\033[0m' # No Color

# Internal log function for file logging (only if LOG_FILE is set)
_log_to_file() {
    if [[ -n "${LOG_FILE:-}" ]]; then
        local timestamp
        timestamp=$(date '+%Y-%m-%d %H:%M:%S')
        local message="$timestamp - $*"
        if ! echo "$message" >> "$LOG_FILE" 2>/dev/null; then
            # If log file write fails, disable file logging to prevent spam
            unset LOG_FILE
        fi
    fi
}

# Initialize logging for a script
# Usage: init_logging [script_name] [enable_file_logging]
# - script_name: optional, defaults to basename of calling script
# - enable_file_logging: optional, defaults to true
init_logging() {
    local script_name="${1:-}"
    local enable_file_logging="${2:-true}"
    
    # If no script name provided, derive it from the calling script
    if [[ -z "$script_name" ]]; then
        script_name=$(basename "${BASH_SOURCE[1]}")
    fi
    
    if [[ "$enable_file_logging" == "true" ]]; then
        # Set up log directory and file
        export LOG_DIR="${LOG_DIR:-$HOME/logs}"
        if ! mkdir -p "$LOG_DIR" 2>/dev/null; then
            echo "Warning: Failed to create log directory $LOG_DIR, disabling file logging" >&2
            unset LOG_FILE
            return 0
        fi
        LOG_FILE="${LOG_DIR}/${script_name%-*}-$(date +%Y-%m-%d).log"
        export LOG_FILE
        
        # Add spacer and start marker to log file
        if echo "" >> "$LOG_FILE" 2>/dev/null; then
            _log_to_file "========== NEW $(echo "$script_name" | tr '[:lower:]' '[:upper:]') SCRIPT EXECUTION =========="
        else
            # If initial write fails, disable file logging
            unset LOG_FILE
        fi
    else
        # Disable file logging by unsetting LOG_FILE
        unset LOG_FILE
    fi
}

# Log script completion marker
# Usage: log_completion [script_name]
log_completion() {
    local script_name="${1:-}"
    
    # If no script name provided, derive it from the calling script
    if [[ -z "$script_name" ]]; then
        script_name=$(basename "${BASH_SOURCE[1]}")
    fi
    
    if [[ -n "${LOG_FILE:-}" ]]; then
        _log_to_file "========== $(echo "$script_name" | tr '[:lower:]' '[:upper:]') SCRIPT EXECUTION COMPLETED =========="
        echo "" >> "$LOG_FILE" 2>/dev/null || unset LOG_FILE
    fi
}

# Logging functions with colors for terminal and optional file logging
# All log output goes to stderr to prevent contamination of command substitution
log_info() {
    echo -e "${GREEN}[INFO]${NC} $*" >&2
    _log_to_file "[INFO] $*"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*" >&2
    _log_to_file "[WARN] $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
    _log_to_file "[ERROR] $*"
}

log_debug() {
    echo -e "${CYAN}[DEBUG]${NC} $*" >&2
    _log_to_file "[DEBUG] $*"
}

# Debug logging function that respects SSM_DEBUG environment variable
# This provides compatibility with SSM command runner's debug_log function
debug_log() {
    if [ "${SSM_DEBUG:-false}" = true ]; then
        echo -e "\n${CYAN}[DEBUG]${NC} $*" >&2
        _log_to_file "[DEBUG] $*"
    fi
}

# === NOTIFICATION UTILITIES ===

# Process webhook URL from environment variable (supports both base64 and plain text)
process_webhook_url() {
    # Parameter count validation
    if [[ $# -ne 1 ]]; then
        log_error "process_webhook_url requires exactly 1 parameter, got $#"
        return 1
    fi
    
    local webhook_var="$1"
    local webhook_url=""
    
    # Input validation
    if [[ -z "$webhook_var" ]]; then
        echo ""
        return 0
    fi
    
    # Try base64 decoding first, if it fails assume it's plain text
    if webhook_url=$(echo "$webhook_var" | base64 -d 2>/dev/null) && [[ "$webhook_url" =~ ^https://chat\.googleapis\.com/v1/spaces/.*/messages\?.*$ ]]; then
        # Successfully decoded base64 webhook URL (debug output can interfere with return value)
        :
    else
        # Using plain text webhook URL (not base64 encoded)
        webhook_url="$webhook_var"
    fi
    
    # Additional URL validation - validate Google Chat webhook format
    if [[ ! "$webhook_url" =~ ^https://chat\.googleapis\.com/v1/spaces/.*/messages\?.*$ ]]; then
        log_error "Invalid webhook URL format. Must be a Google Chat webhook URL with format:"
        log_error "  https://chat.googleapis.com/v1/spaces/SPACE_ID/messages?key=...&token=..."
        return 1
    fi
    
    echo "$webhook_url"
}

# Escape JSON special characters (comprehensive escaping)
escape_json() {
    # Parameter count validation
    if [[ $# -ne 1 ]]; then
        log_error "escape_json requires exactly 1 parameter, got $#"
        return 1
    fi
    
    local text="$1"
    
    # Input validation
    if [[ -z "$text" ]]; then
        echo ""
        return 0
    fi
    
    # Use printf for safer, more predictable escaping
    # Handle each character type systematically
    printf '%s' "$text" | sed \
        -e 's/\\/\\\\/g' \
        -e 's/"/\\"/g' \
        -e 's/	/\\t/g' \
        -e 's/\r/\\r/g' | \
    tr '\n' '\001' | sed 's/\001/\\n/g'
}

# Send webhook with error handling
send_webhook() {
    # Parameter count validation
    if [[ $# -ne 2 ]]; then
        log_error "send_webhook requires exactly 2 parameters, got $#"
        return 1
    fi
    
    local webhook_url="$1"
    local payload="$2"
    
    # Input validation
    if [[ -z "$webhook_url" ]]; then
        log_error "webhook_url parameter cannot be empty"
        return 1
    fi
    
    if [[ -z "$payload" ]]; then
        log_error "payload parameter cannot be empty"
        return 1
    fi
    
    # Debug output to stderr to avoid interfering with return values
    log_debug "Sending webhook request"
    log_debug "Webhook URL: $webhook_url"
    log_debug "Payload length: ${#payload} characters"
    
    # Create temp file with error handling
    local temp_file
    if ! temp_file=$(mktemp); then
        log_error "Failed to create temporary file"
        return 1
    fi
    
    local http_code
    local curl_exit_code
    
    # Perform curl request with comprehensive error handling
    http_code=$(curl -s -o "$temp_file" -w "%{http_code}" \
                     -X POST "$webhook_url" \
                     -H "Content-Type: application/json" \
                     -d "$payload" 2>/dev/null)
    curl_exit_code=$?
    
    # Read response body if temp file exists and is readable
    local response_body=""
    if [[ -f "$temp_file" && -r "$temp_file" ]]; then
        response_body=$(cat "$temp_file" 2>/dev/null || echo "Unable to read response")
    fi
    
    # Clean up temp file
    rm -f "$temp_file" 2>/dev/null
    
    log_debug "HTTP response code: $http_code"
    log_debug "Response body: $response_body"
    
    # Check curl exit code first
    if [[ $curl_exit_code -ne 0 ]]; then
        log_error "❌ Curl failed with exit code: $curl_exit_code"
        case $curl_exit_code in
            3) log_error "   Reason: URL malformed or invalid format" ;;
            6) log_error "   Reason: Could not resolve host" ;;
            7) log_error "   Reason: Failed to connect to host" ;;
            28) log_error "   Reason: Timeout occurred" ;;
            35) log_error "   Reason: SSL connect error" ;;
            *) log_error "   Reason: Unknown curl error (see curl manual)" ;;
        esac
        log_error "   Webhook URL: $webhook_url"
        return 1
    fi
    
    # Check HTTP response code
    if [[ "$http_code" =~ ^2[0-9][0-9]$ ]]; then
        return 0
    else
        log_error "❌ Failed to send notification"
        log_error "   HTTP Status: $http_code"
        log_error "   Response: $response_body"
        return 1
    fi
}

# Check required dependencies for notification scripts
check_notification_dependencies() {
    local -a required_commands=("curl" "base64" "sed" "tr")
    local missing_commands=()
    
    log_debug "Checking notification dependencies"
    
    for cmd in "${required_commands[@]}"; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            missing_commands+=("$cmd")
        fi
    done
    
    if [[ ${#missing_commands[@]} -gt 0 ]]; then
        log_error "Missing required commands: ${missing_commands[*]}"
        log_error "Please install the missing dependencies and try again."
        return 1
    fi
    
    log_debug "All notification dependencies are available"
    return 0
}

# Validate required parameters for notification scripts
# Usage: validate_notification_params webhook_url pr_title pr_number pr_url author repository
validate_notification_params() {
    # Parameter count validation
    if [[ $# -ne 6 ]]; then
        log_error "validate_notification_params requires exactly 6 parameters, got $#"
        return 1
    fi
    
    local webhook_url="$1"
    local pr_title="$2"
    local pr_number="$3"
    local pr_url="$4"
    local author="$5"
    local repository="$6"
    
    local missing_params=()
    
    [[ -z "$webhook_url" ]] && missing_params+=("webhook-url")
    [[ -z "$pr_title" ]] && missing_params+=("pr-title")
    [[ -z "$pr_number" ]] && missing_params+=("pr-number")
    [[ -z "$pr_url" ]] && missing_params+=("pr-url")
    [[ -z "$author" ]] && missing_params+=("author")
    [[ -z "$repository" ]] && missing_params+=("repository")
    
    if [[ ${#missing_params[@]} -gt 0 ]]; then
        log_error "Missing required parameters: ${missing_params[*]}"
        return 1
    fi
    
    log_debug "All required parameters validated successfully"
    return 0
}

# Validate release notification parameters
validate_release_params() {
    # Parameter count validation
    if [[ $# -ne 4 ]]; then
        log_error "validate_release_params requires exactly 4 parameters, got $#"
        return 1
    fi
    
    local webhook_url="$1"
    local version="$2"
    local release_url="$3"
    local repository="$4"
    
    local missing_params=()
    
    [[ -z "$webhook_url" ]] && missing_params+=("webhook-url")
    [[ -z "$version" ]] && missing_params+=("version")
    [[ -z "$release_url" ]] && missing_params+=("release-url")
    [[ -z "$repository" ]] && missing_params+=("repository")
    
    if [[ ${#missing_params[@]} -gt 0 ]]; then
        log_error "Missing required parameters: ${missing_params[*]}"
        return 1
    fi
    
    log_debug "All required parameters validated successfully"
    return 0
}
