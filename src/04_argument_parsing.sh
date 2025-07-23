#!/usr/bin/env bash

# AWS SSO Authentication Helper - Argument Parsing Module
# Functions for handling command-line arguments and displaying help/version info

# Global variables that will be set by parse_parameters
PARSED_COMMAND=""
PARSED_PROFILE=""
PARSED_REGION=""
PARSED_EXPORT_FORMAT=""

parse_parameters() {
  local profile=""
  local command=""
  local region=""
  local export_format=""
  
  while [[ $# -gt 0 ]]; do
    case $1 in
      --profile|-p)
        if [[ -z "$2" || "$2" == --* ]]; then
          log_error "Error: --profile requires a profile name"
          exit 1
        fi
        profile="$2"
        shift 2
        ;;
      --check)
        command="check"
        shift
        ;;
      --creds)
        command="creds"
        shift
        ;;
      --help|-h)
        command="help"
        shift
        ;;
      --version|-v)
        command="version"
        shift
        ;;
      --export)
        export_format="env"  # Default to env format, could be extended later
        shift
        ;;
      --region|-r)
        if [[ -z "$2" || "$2" == --* ]]; then
          log_error "Error: --region requires a region name"
          exit 1
        fi
        region="$2"
        shift 2
        ;;
      # Handle old-style commands
      check|creds|help|version)
       # Only set command if not already set by flags
        if [[ -z "$command" ]]; then
          command="$1"
        else
          log_error "Multiple commands specified: '$command' and '$1'"
          exit 1
        fi
        shift
        ;;
      # Handle unknown flags
      --*)
        log_error "Unknown option: $1"
        log_info "Run 'authaws help' for usage information"
        exit 1
        ;;
      # Handle positional arguments (profile names)
      *)
        if [[ -z "$profile" ]]; then
          profile="$1"
        else
          # Check if this might be a second positional argument for old creds syntax
          if [[ "$command" == "creds" && -z "$PARSED_PROFILE" ]]; then
            # Handle: authaws creds profile-name
            profile="$1"
          else
            log_error "Multiple profile names specified: '$profile' and '$1'"
            exit 1
          fi
        fi
        shift
        ;;
    esac
  done
  
  # Validate command-flag combinations
  case "$command" in
    "check")
      if [[ -n "$profile" ]]; then
        log_warn "Profile specified with --check will be ignored"
      fi
      ;;
    "creds")
      # --creds requires a profile (either via flag or positional)
      if [[ -z "$profile" ]]; then
        # Try to get from environment or .env as fallback
        profile="${AWS_PROFILE:-}"
        if [[ -z "$profile" && -f "${ENV_FILE}" ]]; then
          # shellcheck source=.env
            # shellcheck disable=SC1091
          source "${ENV_FILE}"
          profile="$DEFAULT_PROFILE"
        fi
        if [[ -z "$profile" ]]; then
          log_error "No profile specified for credentials display"
          log_info "Usage: authaws --creds --profile <name> or authaws creds <name>"
          exit 1
        fi
      fi
      ;;
    "help"|"version")
      if [[ -n "$profile" ]]; then
        log_warn "Profile specified with --$command will be ignored"
      fi
      ;;
    "")
      # No command specified, this is authentication mode
      # Profile handling will be done in main()
      ;;
  esac

  # Set global variables for use in main function
  PARSED_COMMAND="$command"
  PARSED_PROFILE="$profile"
  PARSED_REGION="$region"
  PARSED_EXPORT_FORMAT="$export_format"
  : "${PARSED_EXPORT_FORMAT}"
}

# Display help information
show_help() {
  cat << 'EOL'
AWS SSO Authentication Helper

USAGE:
  authaws [PROFILE]                    # Positional syntax (current)
  authaws --profile PROFILE           # Flag syntax (new)

COMMANDS:
  authaws help, --help, -h            # Show this help message
  authaws version, --version, -v      # Show version information  
  authaws check, --check              # Check system requirements
  authaws creds [PROFILE]             # Show credentials (positional)
  authaws --creds --profile PROFILE   # Show credentials (flag)

SETUP:
  Before first use, create a .env file in the script directory with:
  SSO_START_URL="https://your-sso-url.awsapps.com/start"
  SSO_REGION="your-region"
  DEFAULT_PROFILE="your-default-profile"

EXAMPLES:
  # Authentication (both styles work identically)
  authaws dev-profile                  # Quick login (current style)
  authaws --profile dev-profile        # Self-documenting (new style)
  
  # System checks
  authaws check                        # Check dependencies
  authaws --check                      # Check dependencies (flag style)
  
  # Credential management
  authaws creds                        # Show default profile credentials
  authaws creds dev-profile            # Show specific profile credentials
  authaws --creds --profile dev-profile # Show credentials (flag style)

FLAGS:
  -p, --profile PROFILE               # Specify AWS SSO profile name
  -h, --help                          # Show this help message
  --check                             # Check system requirements
  --creds                             # Display profile credentials
  -v, --version                       # Show version information
  --export                            # Export credentials in environment format
  -r, --region                        # Override AWS region for the profile

Both positional and flag syntaxes are fully supported and produce identical results.
Choose the style that works best for your workflow and team preferences.
EOL
  exit 0
}

# Display version information
show_version() {
  cat << 'EOL'
authaws - AWS SSO Authentication Helper Script
Version: 1.4.1
Repository: https://github.com/ZSoftly/quickssm

Dependencies:
  - AWS CLI (required)
  - jq (required) 
  - fzf (required)

Environment:
  - Bash version: $BASH_VERSION
  - Operating System: $(uname -s)

For more information, visit: https://github.com/ZSoftly/quickssm
EOL
  exit 0
}
