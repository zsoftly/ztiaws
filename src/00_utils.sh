#!/usr/bin/env bash

# Common Utility Functions and Definitions
# Repository: https://github.com/ZSoftly/ztiaws
# src/00_utils.sh
# Color definitions
# Export these variables so they're available to other scripts that source this file
export RED='\033[0;31m'
export GREEN='\033[0;32m'
export YELLOW='\033[1;33m'
export CYAN='\033[0;36m'
export NC='\033[0m' # No Color
# Basic logging functions
log_info() {
    echo -e "${CYAN}[INFO]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

# Optional: Enable or disable detailed logging
init_logging() {
    local log_name="$1"
    local enable_logging="${2:-false}"

    if [ "$enable_logging" = "true" ]; then
        exec > >(tee -a "${log_name}.log") 2>&1
        log_info "Logging enabled to ${log_name}.log"
    fi
}
