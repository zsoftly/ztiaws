#!/usr/bin/env bash

# Release Documentation Generator
# Generates CHANGELOG.md and RELEASE_NOTES.txt for releases
# Repository: https://github.com/ZSoftly/ztiaws

set -e  # Exit on any error

# Source utility functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
source "$PROJECT_ROOT/src/00_utils.sh"

# Initialize logging
init_logging "release_docs_generator"

# Script configuration
REPO_ROOT="${REPO_ROOT:-$PROJECT_ROOT}"
VERSION=""
LATEST_TAG=""
FORCE_REGENERATE=false

# Usage information
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Generate CHANGELOG.md and RELEASE_NOTES.txt for a release version.

OPTIONS:
    -v, --version VERSION    Release version (required, format: v1.2.3 or 1.2.3)
    -t, --latest-tag TAG     Latest release tag for comparison (auto-detected if not provided)
    -f, --force             Force regeneration even if files exist
    -h, --help              Show this help message

EXAMPLES:
    $0 --version v2.1.0
    $0 --version 2.1.0 --latest-tag v2.0.5
    $0 -v v2.1.0 -f

ENVIRONMENT VARIABLES:
    REPO_ROOT               Repository root directory (default: auto-detected)
    LOG_DIR                 Directory for log files (default: \$HOME/logs)

EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -t|--latest-tag)
                LATEST_TAG="$2"
                shift 2
                ;;
            -f|--force)
                FORCE_REGENERATE=true
                shift
                ;;
            -h|--help)
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

    # Validate required arguments
    if [[ -z "$VERSION" ]]; then
        log_error "Version is required. Use --version or -v to specify."
        usage
        exit 1
    fi
}

# Validate version format
validate_version() {
    log_info "Validating version format: $VERSION"
    
    if [[ ! "$VERSION" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        log_error "Invalid version format: $VERSION"
        log_error "Expected format: v1.2.3 or 1.2.3"
        exit 1
    fi
    
    log_info "Version format is valid: $VERSION"
}

# Get latest release tag if not provided
get_latest_tag() {
    if [[ -n "$LATEST_TAG" ]]; then
        log_info "Using provided latest tag: $LATEST_TAG"
        return
    fi
    
    log_info "Auto-detecting latest release tag..."
    
    cd "$REPO_ROOT"
    LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    
    if [[ -n "$LATEST_TAG" ]]; then
        log_info "Found latest tag: $LATEST_TAG"
    else
        log_warn "No previous tags found - will include all commits"
    fi
}

# Get commits for changelog
get_commits() {
    local commit_range
    
    cd "$REPO_ROOT"
    
    if [[ -n "$LATEST_TAG" ]]; then
        commit_range="${LATEST_TAG}..HEAD"
    else
        commit_range="HEAD"
    fi
    
    # Get commits with proper error handling and log to stderr to avoid capture
    log_info "Getting commits for changelog generation..." >&2
    log_info "Using commit range: $commit_range" >&2
    
    local commits
    commits=$(git log --pretty=format:"%s" --no-merges "$commit_range" 2>/dev/null || echo "")
    
    if [[ -z "$commits" ]]; then
        log_warn "No commits found since last tag" >&2
        echo "No notable changes"
    else
        # Filter out automated commits, color codes, and log patterns
        commits=$(echo "$commits" | grep -v -E "(Auto-generate changelog|Restore RELEASE_NOTES|Restore CHANGELOG|\[0;[0-9]+m|\[INFO\]|\[WARN\]|\[ERROR\]|\[DEBUG\])" || echo "")
        
        # Remove any remaining ANSI color codes
        commits=$(echo "$commits" | sed 's/\x1b\[[0-9;]*m//g' || echo "")
        
        if [[ -z "$commits" ]]; then
            echo "No notable changes"
        else
            echo "$commits"
        fi
    fi
}

# Categorize commits based on conventional commit patterns
categorize_commits() {
    local commits="$1"
    local features=""
    local fixes=""
    local other=""
    
    log_info "Categorizing commits using conventional commit patterns..."
    
    # Process commits line by line safely
    while IFS= read -r commit; do
        if [[ -z "$commit" ]]; then
            continue
        fi
        
        # Escape commit message for safe processing
        commit=$(echo "$commit" | sed 's/["`$\\]/\\&/g')
        
        # Improved regex patterns for conventional commits
        if [[ "$commit" =~ ^(feat|feature|add)(\(.*\))?:?.* ]]; then
            features="${features}- ${commit}"$'\n'
        elif [[ "$commit" =~ ^(fix|bug|hotfix)(\(.*\))?:?.* ]]; then
            fixes="${fixes}- ${commit}"$'\n'
        else
            other="${other}- ${commit}"$'\n'
        fi
    done <<< "$commits"
    
    # Return categorized commits as associative array-like output
    echo "FEATURES|$features"
    echo "FIXES|$fixes"
    echo "OTHER|$other"
}

# Generate CHANGELOG.md
generate_changelog() {
    log_info "Generating CHANGELOG.md for $VERSION..."
    
    cd "$REPO_ROOT"
    
    # Get and categorize commits
    local commits
    commits=$(get_commits)
    
    local categorized
    categorized=$(categorize_commits "$commits")
    
    # Parse categorized output
    local features="" fixes="" other=""
    while IFS='|' read -r category content; do
        case "$category" in
            FEATURES) features="$content" ;;
            FIXES) fixes="$content" ;;
            OTHER) other="$content" ;;
        esac
    done <<< "$categorized"
    
    # Build new version entry
    local new_version="## [$VERSION] - $(date +%Y-%m-%d)"
    
    # Create temp file with proper error handling
    local temp_entry
    temp_entry=$(mktemp) || {
        log_error "Failed to create temp file"
        exit 1
    }
    
    echo "$new_version" > "$temp_entry"
    echo "" >> "$temp_entry"
    
    # Add categorized sections only if they have content
    if [[ -n "$features" ]]; then
        log_debug "Adding features section to changelog"
        echo "### Added" >> "$temp_entry"
        echo -e "$features" >> "$temp_entry"
    fi
    
    if [[ -n "$fixes" ]]; then
        log_debug "Adding fixes section to changelog"
        echo "### Fixed" >> "$temp_entry"
        echo -e "$fixes" >> "$temp_entry"
    fi
    
    if [[ -n "$other" ]]; then
        log_debug "Adding other changes section to changelog"
        echo "### Changed" >> "$temp_entry"
        echo -e "$other" >> "$temp_entry"
    fi
    
    echo "" >> "$temp_entry"
    
    # Safely update CHANGELOG.md
    local temp_changelog
    temp_changelog=$(mktemp) || {
        log_error "Failed to create temp changelog"
        exit 1
    }
    
    if [[ -f CHANGELOG.md ]]; then
        log_info "Updating existing CHANGELOG.md"
        # Insert after first line (# Changelog)
        head -n 1 CHANGELOG.md > "$temp_changelog"
        echo "" >> "$temp_changelog"
        cat "$temp_entry" >> "$temp_changelog"
        tail -n +2 CHANGELOG.md >> "$temp_changelog"
    else
        log_info "Creating new CHANGELOG.md"
        echo "# Changelog" > "$temp_changelog"
        echo "" >> "$temp_changelog"
        cat "$temp_entry" >> "$temp_changelog"
    fi
    
    # Atomic update
    mv "$temp_changelog" CHANGELOG.md || {
        log_error "Failed to update CHANGELOG.md"
        exit 1
    }
    
    # Cleanup
    rm -f "$temp_entry"
    
    log_info "CHANGELOG.md updated successfully"
}

# Generate RELEASE_NOTES.txt
generate_release_notes() {
    log_info "Generating RELEASE_NOTES.txt for $VERSION..."
    
    cd "$REPO_ROOT"
    
    # Get repository info for GitHub links
    local repo_url
    repo_url=$(git config --get remote.origin.url 2>/dev/null || echo "")
    local github_repo=""
    
    if [[ "$repo_url" =~ github\.com[:/]([^/]+/[^/]+)(\.git)?$ ]]; then
        github_repo="${BASH_REMATCH[1]}"
        github_repo="${github_repo%.git}"  # Remove .git suffix if present
        log_debug "Detected GitHub repository: $github_repo"
    fi
    
    # Create release notes header
    cat > RELEASE_NOTES.txt << EOF
# ztictl $VERSION Release Notes

EOF
    
    # Add installation link if GitHub repo detected
    if [[ -n "$github_repo" ]]; then
        echo "**Installation:** [Installation Guide](https://github.com/$github_repo/blob/main/INSTALLATION.md)" >> RELEASE_NOTES.txt
    else
        echo "**Installation:** See INSTALLATION.md in the repository" >> RELEASE_NOTES.txt
    fi
    
    echo "" >> RELEASE_NOTES.txt
    echo "**Release Date:** $(date '+%B %d, %Y')" >> RELEASE_NOTES.txt
    echo "" >> RELEASE_NOTES.txt
    
    # Get and categorize commits
    local commits
    commits=$(get_commits)
    
    if [[ "$commits" == "No notable changes" ]]; then
        echo "## 📝 Changes" >> RELEASE_NOTES.txt
        echo "• No notable changes since last release" >> RELEASE_NOTES.txt
    else
        local categorized
        categorized=$(categorize_commits "$commits")
        
        # Parse categorized output for release notes format
        local features="" fixes="" other=""
        while IFS='|' read -r category content; do
            case "$category" in
                FEATURES) features="$content" ;;
                FIXES) fixes="$content" ;;
                OTHER) other="$content" ;;
            esac
        done <<< "$categorized"
        
        # Convert to bullet points for release notes
        features=$(echo "$features" | sed 's/^- /• /')
        fixes=$(echo "$fixes" | sed 's/^- /• /')
        other=$(echo "$other" | sed 's/^- /• /')
        
        # Add sections with emojis only if they have content
        if [[ -n "$features" ]]; then
            log_debug "Adding features section to release notes"
            echo "## 🚀 New Features" >> RELEASE_NOTES.txt
            echo -e "$features" >> RELEASE_NOTES.txt
        fi
        
        if [[ -n "$fixes" ]]; then
            log_debug "Adding fixes section to release notes"
            echo "## 🐛 Bug Fixes" >> RELEASE_NOTES.txt
            echo -e "$fixes" >> RELEASE_NOTES.txt
        fi
        
        if [[ -n "$other" ]]; then
            log_debug "Adding other changes section to release notes"
            echo "## 📝 Other Changes" >> RELEASE_NOTES.txt
            echo -e "$other" >> RELEASE_NOTES.txt
        fi
    fi
    
    log_info "RELEASE_NOTES.txt generated successfully"
}

# Check if files already exist
check_existing_files() {
    local files_exist=false
    
    cd "$REPO_ROOT"
    
    if [[ -f CHANGELOG.md ]] && ! $FORCE_REGENERATE; then
        log_warn "CHANGELOG.md already exists. Use --force to regenerate."
        files_exist=true
    fi
    
    if [[ -f RELEASE_NOTES.txt ]] && ! $FORCE_REGENERATE; then
        log_warn "RELEASE_NOTES.txt already exists. Use --force to regenerate."
        files_exist=true
    fi
    
    if $files_exist; then
        log_error "Files already exist. Use --force flag to overwrite."
        exit 1
    fi
}

# Main execution
main() {
    log_info "Starting release documentation generation"
    log_info "Repository root: $REPO_ROOT"
    
    parse_args "$@"
    validate_version
    get_latest_tag
    check_existing_files
    
    generate_changelog
    generate_release_notes
    
    log_info "Release documentation generation completed successfully!"
    log_completion
}

# Execute main function with all arguments
main "$@"
