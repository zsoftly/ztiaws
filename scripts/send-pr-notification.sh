#!/usr/bin/env bash

# ZTiAWS PR Notification Script
# Sends Google Chat App Card notifications for Pull Request events

set -euo pipefail

# --- Configuration ---
WEBHOOK_URL=""
PR_TITLE=""
PR_NUMBER=""
PR_URL=""
AUTHOR=""
REPOSITORY=""
STATUS=""
MESSAGE=""

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
init_logging "$(basename "${BASH_SOURCE[0]}")" false

# --- Usage Function ---
usage() {
    echo -e "${GREEN}ZTiAWS PR Notification Script v1.1.0${NC}"
    echo
    echo -e "${CYAN}PURPOSE:${NC}"
    echo "  Sends Google Chat App Card notifications for Pull Request events."
    echo
    echo -e "${CYAN}USAGE:${NC}"
    echo "  $0 --pr-title TITLE --pr-number NUMBER --pr-url URL --author USER --repository REPO [--status STATUS] [--message MESSAGE]"
    echo
    echo -e "${CYAN}REQUIRED:${NC}"
    echo "  --pr-title TITLE       Pull request title"
    echo "  --pr-number NUMBER     Pull request number" 
    echo "  --pr-url URL           Pull request URL"
    echo "  --author USERNAME      PR author username"
    echo "  --repository REPO      Repository name (org/repo)"
    echo
    echo -e "${CYAN}OPTIONAL:${NC}"
    echo "  --status STATUS        PR status (success/failure)"
    echo "  --message MESSAGE      Status message"
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
            --pr-title) 
                [[ $# -lt 2 ]] && { log_error "Option $1 requires a value"; usage; }
                PR_TITLE="$2"; shift 2 ;;
            --pr-number) 
                [[ $# -lt 2 ]] && { log_error "Option $1 requires a value"; usage; }
                PR_NUMBER="$2"; shift 2 ;;
            --pr-url) 
                [[ $# -lt 2 ]] && { log_error "Option $1 requires a value"; usage; }
                PR_URL="$2"; shift 2 ;;
            --author) 
                [[ $# -lt 2 ]] && { log_error "Option $1 requires a value"; usage; }
                AUTHOR="$2"; shift 2 ;;
            --repository) 
                [[ $# -lt 2 ]] && { log_error "Option $1 requires a value"; usage; }
                REPOSITORY="$2"; shift 2 ;;
            --status) 
                [[ $# -lt 2 ]] && { log_error "Option $1 requires a value"; usage; }
                STATUS="$2"; shift 2 ;;
            --message) 
                [[ $# -lt 2 ]] && { log_error "Option $1 requires a value"; usage; }
                MESSAGE="$2"; shift 2 ;;
            --help) usage ;;
            *) log_error "Unknown option: $1"; shift ;;
        esac
    done

    # Process webhook URL from environment if not provided via argument
    if [[ -z "$WEBHOOK_URL" && -n "${GOOGLE_CHAT_WEBHOOK:-}" ]]; then
        WEBHOOK_URL=$(process_webhook_url "$GOOGLE_CHAT_WEBHOOK")
    fi

    # Validate required parameters using centralized function
    validate_notification_params "$WEBHOOK_URL" "$PR_TITLE" "$PR_NUMBER" "$PR_URL" "$AUTHOR" "$REPOSITORY" || { usage; exit 1; }
}

# --- Create Google Chat App Card JSON ---
create_chat_payload() {
    log_debug "Creating Google Chat App Card payload"
    
    local escaped_title=$(escape_json "$PR_TITLE")
    local escaped_author=$(escape_json "$AUTHOR")
    local escaped_repository=$(escape_json "$REPOSITORY")
    local escaped_pr_number=$(escape_json "$PR_NUMBER")
    local files_url="${PR_URL}/files"
    local escaped_pr_url=$(escape_json "$PR_URL")
    local escaped_files_url=$(escape_json "$files_url")
    local escaped_message=$(escape_json "${MESSAGE:-New pull request opened}")
    
    # Determine status icon and header
    local status_icon="NOTIFICATION_ICON"
    local header_title="New Pull Request"
    local status_color=""
    
    if [[ "$STATUS" == "success" ]]; then
        status_icon="STAR"
        header_title="PR Ready for Review"
        status_color=""
    elif [[ "$STATUS" == "failure" ]]; then
        status_icon="ERROR"
        header_title="PR Tests Failed"
        status_color=""
    fi
    
    cat << EOF
{
  "cards": [
    {
      "header": {
        "title": "$header_title",
        "subtitle": "$escaped_repository Repository",
        "imageUrl": "https://github.com/fluidicon.png",
        "imageStyle": "AVATAR"
      },
      "sections": [
        {
          "widgets": [
            {
              "keyValue": {
                "topLabel": "Pull Request",
                "content": "$escaped_title",
                "contentMultiline": true,
                "icon": "DESCRIPTION"
              }
            },
            {
              "keyValue": {
                "topLabel": "Author",
                "content": "$escaped_author",
                "icon": "PERSON"
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
                "topLabel": "PR Number",
                "content": "#$escaped_pr_number",
                "icon": "CONFIRMATION_NUMBER_ICON"
              }
            },
            {
              "keyValue": {
                "topLabel": "Status",
                "content": "$escaped_message",
                "contentMultiline": true,
                "icon": "$status_icon"
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
                    "text": "Review PR",
                    "onClick": {
                      "openLink": {
                        "url": "$escaped_pr_url"
                      }
                    }
                  }
                },
                {
                  "textButton": {
                    "text": "View Files",
                    "onClick": {
                      "openLink": {
                        "url": "$escaped_files_url"
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
    
    log_info "Sending PR notification to Google Chat"
    log_debug "PR: #$PR_NUMBER by $AUTHOR"
    
    local payload
    if ! payload=$(create_chat_payload); then
        log_error "Failed to create chat payload"
        exit 1
    fi
    
    if send_webhook "$WEBHOOK_URL" "$payload"; then
        log_info "[OK] PR notification sent successfully!"
        log_info "     PR: $PR_TITLE (#$PR_NUMBER)"
        log_info "     Author: $AUTHOR"
        log_info "     Repository: $REPOSITORY"
    else
        exit 1
    fi
}

main "$@"