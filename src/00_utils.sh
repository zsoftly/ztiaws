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
        echo "$message" >> "$LOG_FILE"
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
        mkdir -p "$LOG_DIR"
        LOG_FILE="${LOG_DIR}/${script_name%-*}-$(date +%Y-%m-%d).log"
        export LOG_FILE
        
        # Add spacer and start marker to log file
        echo "" >> "$LOG_FILE"
        _log_to_file "========== NEW $(echo "$script_name" | tr '[:lower:]' '[:upper:]') SCRIPT EXECUTION =========="
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
        echo "" >> "$LOG_FILE"
    fi
}

# Logging functions with colors for terminal and optional file logging
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
    _log_to_file "[INFO] $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
    _log_to_file "[WARN] $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    _log_to_file "[ERROR] $1"
}

log_debug() {
    echo -e "${CYAN}[DEBUG]${NC} $1"
    _log_to_file "[DEBUG] $1"
}

# Debug logging function that respects SSM_DEBUG environment variable
# This provides compatibility with SSM command runner's debug_log function
debug_log() {
    if [ "${SSM_DEBUG:-false}" = true ]; then
        echo -e "\n${CYAN}[DEBUG]${NC} $*" >&2
        _log_to_file "[DEBUG] $*"
    fi
}
