#!/usr/bin/env bash

# ZTiAWS Release Notification Script
# Sends rich Google Chat App Card notifications for Release events
# Based on zsoftly-services notification pattern

set -euo pipefail

# Script configuration
SCRIPT_NAME="$(basename "$0")"
SCRIPT_VERSION="1.0.0"

# Color definitions for terminal output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly PURPLE='\033[0;35m'
readonly CYAN='\033[0;36m'
readonly NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $*" >&2
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*" >&2
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

log_debug() {
    if [[ "${DEBUG:-false}" == "true" ]]; then
        echo -e "${CYAN}[DEBUG]${NC} $*" >&2
    fi
}

# Usage information
usage() {
    cat << EOF
${PURPLE}ZTiAWS Release Notification Script v${SCRIPT_VERSION}${NC}

${BLUE}DESCRIPTION:${NC}
    Sends rich Google Chat App Card notifications for Release events.
    Follows zsoftly-services notification patterns with embedded styling.

${BLUE}USAGE:${NC}
    $SCRIPT_NAME [OPTIONS]

${BLUE}OPTIONS:${NC}
    --webhook-url URL      Google Chat webhook URL (required)
    --version VERSION      Release version (required, e.g., v1.0.0)
    --release-url URL      Release URL (required)
    --repository REPO      Repository name (required)
    --changelog-url URL    Changelog URL (optional)
    --debug                Enable debug output
    --help                 Show this help message

${BLUE}ENVIRONMENT VARIABLES:${NC}
    GOOGLE_CHAT_WEBHOOK    Base64 encoded webhook URL (alternative to --webhook-url)
    DEBUG                  Enable debug mode (true/false)

${BLUE}EXAMPLES:${NC}
    # Using command line parameters
    $SCRIPT_NAME \\
      --webhook-url "https://chat.googleapis.com/v1/spaces/..." \\
      --version "v1.2.0" \\
      --release-url "https://github.com/org/repo/releases/tag/v1.2.0" \\
      --repository "org/repo" \\
      --changelog-url "https://github.com/org/repo/blob/main/CHANGELOG.md"

    # Using environment variable for webhook
    export GOOGLE_CHAT_WEBHOOK=\$(echo -n "https://chat.googleapis.com/..." | base64)
    $SCRIPT_NAME --version "v1.2.0" --release-url "..." --repository "org/repo"

EOF
}

# Parse command line arguments
parse_arguments() {
    WEBHOOK_URL=""
    VERSION=""
    RELEASE_URL=""
    REPOSITORY=""
    CHANGELOG_URL=""

    while [[ $# -gt 0 ]]; do
        case $1 in
            --webhook-url)
                WEBHOOK_URL="$2"
                shift 2
                ;;
            --version)
                VERSION="$2"
                shift 2
                ;;
            --release-url)
                RELEASE_URL="$2"
                shift 2
                ;;
            --repository)
                REPOSITORY="$2"
                shift 2
                ;;
            --changelog-url)
                CHANGELOG_URL="$2"
                shift 2
                ;;
            --debug)
                export DEBUG=true
                shift
                ;;
            --help)
                usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done

    # If webhook URL not provided via argument, try environment variable
    if [[ -z "$WEBHOOK_URL" && -n "${GOOGLE_CHAT_WEBHOOK:-}" ]]; then
        log_debug "Processing webhook URL from environment variable"
        # Try base64 decoding first, if it fails assume it's plain text
        if WEBHOOK_URL=$(echo "$GOOGLE_CHAT_WEBHOOK" | base64 -d 2>/dev/null) && [[ "$WEBHOOK_URL" =~ ^https://chat\.googleapis\.com ]]; then
            log_debug "Successfully decoded base64 webhook URL"
        else
            log_debug "Using plain text webhook URL (not base64 encoded)"
            WEBHOOK_URL="$GOOGLE_CHAT_WEBHOOK"
        fi
    fi

    # Set default changelog URL if not provided
    if [[ -z "$CHANGELOG_URL" && -n "$REPOSITORY" ]]; then
        CHANGELOG_URL="https://github.com/$REPOSITORY/blob/main/CHANGELOG.md"
        log_debug "Using default changelog URL: $CHANGELOG_URL"
    fi

    # Validate required parameters
    local missing_params=()
    [[ -z "$WEBHOOK_URL" ]] && missing_params+=("webhook-url or GOOGLE_CHAT_WEBHOOK")
    [[ -z "$VERSION" ]] && missing_params+=("version")
    [[ -z "$RELEASE_URL" ]] && missing_params+=("release-url")
    [[ -z "$REPOSITORY" ]] && missing_params+=("repository")

    if [[ ${#missing_params[@]} -gt 0 ]]; then
        log_error "Missing required parameters: ${missing_params[*]}"
        usage
        exit 1
    fi

    log_debug "Parameters validated successfully"
}

# Create Google Chat App Card JSON payload
create_chat_payload() {
    log_debug "Creating Google Chat App Card payload"
    
    # Escape JSON special characters in text fields
    local escaped_version=$(echo "$VERSION" | sed 's/"/\\"/g')
    local escaped_repository=$(echo "$REPOSITORY" | sed 's/"/\\"/g')
    
    # Create the buttons section - always include view and download, conditionally include changelog
    local buttons_json='[
                {
                  "textButton": {
                    "text": "📋 View Release",
                    "onClick": {
                      "openLink": {
                        "url": "'$RELEASE_URL'"
                      }
                    }
                  }
                },
                {
                  "textButton": {
                    "text": "⬇️ Download",
                    "onClick": {
                      "openLink": {
                        "url": "'$RELEASE_URL'"
                      }
                    }
                  }
                }'
    
    # Add changelog button if URL is provided
    if [[ -n "$CHANGELOG_URL" ]]; then
        buttons_json+=',
                {
                  "textButton": {
                    "text": "📝 Changelog",
                    "onClick": {
                      "openLink": {
                        "url": "'$CHANGELOG_URL'"
                      }
                    }
                  }
                }'
    fi
    
    buttons_json+=']'
    
    cat << EOF
{
  "cards": [
    {
      "header": {
        "title": "🚀 New Release Available",
        "subtitle": "ztiaws Repository",
        "imageUrl": "https://github.com/fluidicon.png",
        "imageStyle": "AVATAR"
      },
      "sections": [
        {
          "widgets": [
            {
              "keyValue": {
                "topLabel": "Version",
                "content": "$escaped_version",
                "icon": "STAR"
              }
            },
            {
              "keyValue": {
                "topLabel": "Repository",
                "content": "$escaped_repository",
                "icon": "BOOKMARK"
              }
            },
            {
              "keyValue": {
                "topLabel": "Status",
                "content": "🎉 Ready for deployment!",
                "icon": "DONE"
              }
            }
          ]
        },
        {
          "widgets": [
            {
              "buttons": $buttons_json
            }
          ]
        }
      ]
    }
  ]
}
EOF
}

# Send notification to Google Chat
send_notification() {
    log_info "Sending release notification to Google Chat"
    log_debug "Webhook URL: ${WEBHOOK_URL:0:50}..."
    log_debug "Release: $VERSION"
    
    local payload
    payload=$(create_chat_payload)
    
    log_debug "Payload created, sending to webhook"
    
    # Send the notification
    local response
    local http_code
    
    response=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "$payload" \
        "$WEBHOOK_URL" 2>&1)
    
    # Extract HTTP status code (last line of response)
    http_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    log_debug "HTTP response code: $http_code"
    log_debug "Response body: $response_body"
    
    if [[ "$http_code" =~ ^2[0-9][0-9]$ ]]; then
        log_info "✅ Release notification sent successfully!"
        log_info "   Version: $VERSION"
        log_info "   Repository: $REPOSITORY"
        log_info "   Release URL: $RELEASE_URL"
        return 0
    else
        log_error "❌ Failed to send notification"
        log_error "   HTTP Status: $http_code"
        log_error "   Response: $response_body"
        return 1
    fi
}

# Validate dependencies
check_dependencies() {
    log_debug "Checking dependencies"
    
    local deps=("curl" "base64")
    local missing_deps=()
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing_deps+=("$dep")
        fi
    done
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_error "Missing required dependencies: ${missing_deps[*]}"
        log_error "Please install the missing dependencies and try again"
        exit 1
    fi
    
    log_debug "All dependencies satisfied"
}

# Main function
main() {
    log_debug "Starting $SCRIPT_NAME v$SCRIPT_VERSION"
    
    # Check dependencies
    check_dependencies
    
    # Parse arguments
    parse_arguments "$@"
    
    # Send notification
    if send_notification; then
        log_info "🎉 Release notification process completed successfully"
        exit 0
    else
        log_error "💥 Release notification process failed"
        exit 1
    fi
}

# Script entry point
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi