#!/usr/bin/env bash

# ZTiAWS Release Notification Script
# Sends Google Chat App Card notifications for Release events

set -euo pipefail

# --- Configuration ---
WEBHOOK_URL=""
VERSION=""
RELEASE_URL=""
REPOSITORY=""

# --- Load Shared Utilities ---
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)" || {
    echo "Error: Failed to determine script directory" >&2
    exit 1
}

if [ -f "${SCRIPT_DIR}/../src/00_utils.sh" ]; then
    # shellcheck source=../src/00_utils.sh
    source "${SCRIPT_DIR}/../src/00_utils.sh" || {
        echo "Error: Failed to source utilities file" >&2
        exit 1
    }
else
    echo "Error: Required utilities file not found at ${SCRIPT_DIR}/../src/00_utils.sh" >&2
    exit 1
fi

# Initialize logging for this script
init_logging "$(basename "$0")" false

# --- Usage Function ---
usage() {
    echo -e "${GREEN}ZTiAWS Release Notification Script v1.1.0${NC}"
    echo
    echo -e "${CYAN}PURPOSE:${NC}"
    echo "  Sends Google Chat App Card notifications for Release events."
    echo
    echo -e "${CYAN}USAGE:${NC}"
    echo "  $0 --version VERSION --release-url URL --repository REPO"
    echo
    echo -e "${CYAN}REQUIRED:${NC}"
    echo "  --version VERSION      Release version (e.g., v1.0.0)"
    echo "  --release-url URL      GitHub release URL"
    echo "  --repository REPO      Repository name (org/repo)"
    echo
    echo -e "${CYAN}WEBHOOK:${NC}"
    echo "  --webhook-url URL      Google Chat webhook URL"
    echo "  OR set GOOGLE_CHAT_WEBHOOK environment variable"
    echo
    echo -e "${CYAN}OPTIONS:${NC}"
    echo "  --help                 Show this help"
    echo
    exit 0
}

# --- Argument Parsing ---
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --webhook-url) 
                [[ $# -lt 2 ]] && { log_error "Option $1 requires a value"; usage; }
                WEBHOOK_URL="$2"; shift 2 ;;
            --version) 
                [[ $# -lt 2 ]] && { log_error "Option $1 requires a value"; usage; }
                VERSION="$2"; shift 2 ;;
            --release-url) 
                [[ $# -lt 2 ]] && { log_error "Option $1 requires a value"; usage; }
                RELEASE_URL="$2"; shift 2 ;;
            --repository) 
                [[ $# -lt 2 ]] && { log_error "Option $1 requires a value"; usage; }
                REPOSITORY="$2"; shift 2 ;;
            --help) usage ;;
            *) log_error "Unknown option: $1"; shift ;;
        esac
    done

    # Process webhook URL from environment if not provided via argument
    if [[ -z "$WEBHOOK_URL" && -n "${GOOGLE_CHAT_WEBHOOK:-}" ]]; then
        WEBHOOK_URL=$(process_webhook_url "$GOOGLE_CHAT_WEBHOOK")
    fi

    # Validate required parameters using centralized function
    validate_release_params "$WEBHOOK_URL" "$VERSION" "$RELEASE_URL" "$REPOSITORY" || { usage; exit 1; }
}

# --- Create Google Chat App Card JSON ---
create_chat_payload() {
    log_debug "Creating Google Chat App Card payload"
    
    local escaped_version=$(escape_json "$VERSION")
    local escaped_repository=$(escape_json "$REPOSITORY")
    local escaped_release_url=$(escape_json "$RELEASE_URL")
    local changelog_url="https://github.com/$REPOSITORY/blob/main/CHANGELOG.md"
    local escaped_changelog_url=$(escape_json "$changelog_url")
    
    cat << EOF
{
  "cards": [
    {
      "header": {
        "title": "🚀 New Release Available",
        "subtitle": "$escaped_repository Repository",
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
                "icon": "CONFIRMATION_NUMBER_ICON"
              }
            }
          ]
        },
        {
          "widgets": [
            {
              "buttons": [
                {
                  "textButton": {
                    "text": "📋 View Release",
                    "onClick": {
                      "openLink": {
                        "url": "$escaped_release_url"
                      }
                    }
                  }
                },
                {
                  "textButton": {
                    "text": "⬇️ Download",
                    "onClick": {
                      "openLink": {
                        "url": "$escaped_release_url"
                      }
                    }
                  }
                },
                {
                  "textButton": {
                    "text": "📝 Changelog",
                    "onClick": {
                      "openLink": {
                        "url": "$escaped_changelog_url"
                      }
                    }
                  }
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}
EOF
}

# --- Main Logic ---
main() {
    parse_arguments "$@"
    check_notification_dependencies || exit 1
    
    log_info "Sending Release notification to Google Chat"
    log_debug "Version: $VERSION"
    
    local payload
    if ! payload=$(create_chat_payload); then
        log_error "Failed to create chat payload"
        exit 1
    fi
    
    if send_webhook "$WEBHOOK_URL" "$payload"; then
        log_info "✅ Release notification sent successfully!"
        log_info "   Version: $VERSION"
        log_info "   Repository: $REPOSITORY"
        log_info "   Release URL: $RELEASE_URL"
    else
        exit 1
    fi
}

main "$@"