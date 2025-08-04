#!/usr/bin/env bash

# Release Documentation Generator - Simplified and Fixed
# Generates CHANGELOG.md and RELEASE_NOTES.txt for releases

set -e

# Script configuration
VERSION=""
LATEST_TAG=""
FORCE_REGENERATE=false
DEBUG_MODE=false

usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Generate CHANGELOG.md and RELEASE_NOTES.txt for a release version.

OPTIONS:
    -v, --version VERSION    Release version (required, format: v1.2.3 or 1.2.3)
    -t, --latest-tag TAG     Latest release tag for comparison (auto-detected if not provided)
    -f, --force             Force regeneration even if files exist
    -d, --debug             Enable debug mode with verbose output
    -h, --help              Show this help message

EXAMPLES:
    $0 --version v2.1.0
    $0 --version 2.1.0 --latest-tag v2.0.5 --debug

EOF
}

debug_log() {
    if [[ "$DEBUG_MODE" == true ]]; then
        echo "[DEBUG] $*" >&2
    fi
}

log_info() {
    echo "[INFO] $*" >&2
}

log_error() {
    echo "[ERROR] $*" >&2
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--version) VERSION="$2"; shift 2 ;;
            -t|--latest-tag) LATEST_TAG="$2"; shift 2 ;;
            -f|--force) FORCE_REGENERATE=true; shift ;;
            -d|--debug) DEBUG_MODE=true; shift ;;
            -h|--help) usage; exit 0 ;;
            *) log_error "Unknown option: $1"; usage; exit 1 ;;
        esac
    done

    if [[ -z "$VERSION" ]]; then
        log_error "Version is required. Use --version or -v to specify."
        usage
        exit 1
    fi
}

# Validate version format
validate_version() {
    if [[ ! "$VERSION" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        log_error "Invalid version format: $VERSION"
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
    
    LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    
    if [[ -n "$LATEST_TAG" ]]; then
        log_info "Found latest tag: $LATEST_TAG"
        debug_log "Tag date: $(git log -1 --format=%ai "$LATEST_TAG" 2>/dev/null || echo "unknown")"
    else
        log_info "No previous tags found - will include all commits"
    fi
}

# Get all commits and generate both files
generate_documentation() {
    local commit_range
    
    if [[ -n "$LATEST_TAG" ]]; then
        commit_range="${LATEST_TAG}..HEAD"
    else
        commit_range="HEAD"
    fi
    
    log_info "Getting commits using range: $commit_range"
    
    # Get raw commits and save to temporary file for processing
    local commits_file
    commits_file=$(mktemp)
    
    git log --pretty=format:"%s" --no-merges "$commit_range" 2>/dev/null > "$commits_file" || {
        log_error "Failed to get git log"
        rm -f "$commits_file"
        exit 1
    }
    
    local total_commits
    total_commits=$(wc -l < "$commits_file")
    debug_log "Total raw commits: $total_commits"
    
    if [[ $total_commits -eq 0 ]]; then
        log_info "No commits found"
        rm -f "$commits_file"
        return
    fi
    
    # Filter out automation commits
    local filtered_file
    filtered_file=$(mktemp)
    
    grep -v -E "^(Auto-generate changelog|Restore RELEASE_NOTES|Restore CHANGELOG|Merge (branch|pull request))" "$commits_file" > "$filtered_file" || cp "$commits_file" "$filtered_file"
    
    local filtered_commits
    filtered_commits=$(wc -l < "$filtered_file")
    debug_log "Commits after filtering: $filtered_commits"
    
    if [[ $filtered_commits -eq 0 ]]; then
        log_info "No notable changes after filtering"
        rm -f "$commits_file" "$filtered_file"
        return
    fi
    
    # Show first few commits for debugging
    if [[ "$DEBUG_MODE" == true ]]; then
        debug_log "First 10 filtered commits:"
        head -10 "$filtered_file" | nl >&2
    fi
    
    # Create temporary files for each category
    local features_file fixes_file other_file
    features_file=$(mktemp)
    fixes_file=$(mktemp)
    other_file=$(mktemp)
    
    # Categorize commits
    while IFS= read -r commit; do
        [[ -z "$commit" ]] && continue
        
        debug_log "Processing: $commit"
        
        if [[ "$commit" =~ ^(feat|feature)(\(.*\))?:?.* ]] || [[ "$commit" =~ ^(Enhance|Add|Implement|Create) ]]; then
            echo "$commit" >> "$features_file"
            debug_log "  -> FEATURE"
        elif [[ "$commit" =~ ^(fix|bug|hotfix)(\(.*\))?:?.* ]] || [[ "$commit" =~ (fix|bug|resolve|correct) ]]; then
            echo "$commit" >> "$fixes_file"
            debug_log "  -> FIX"
        else
            echo "$commit" >> "$other_file"
            debug_log "  -> OTHER"
        fi
    done < "$filtered_file"
    
    local feature_count fix_count other_count
    feature_count=$(wc -l < "$features_file")
    fix_count=$(wc -l < "$fixes_file")
    other_count=$(wc -l < "$other_file")
    
    debug_log "Final counts: Features=$feature_count, Fixes=$fix_count, Other=$other_count"
    
    # Generate CHANGELOG.md
    generate_changelog "$features_file" "$fixes_file" "$other_file"
    
    # Generate RELEASE_NOTES.txt
    generate_release_notes "$features_file" "$fixes_file" "$other_file"
    
    # Cleanup
    rm -f "$commits_file" "$filtered_file" "$features_file" "$fixes_file" "$other_file"
}

generate_changelog() {
    local features_file="$1"
    local fixes_file="$2"
    local other_file="$3"
    
    log_info "Generating CHANGELOG.md..."
    
    local temp_entry
    temp_entry=$(mktemp)
    
    echo "## [$VERSION] - $(date +%Y-%m-%d)" > "$temp_entry"
    echo "" >> "$temp_entry"
    
    # Add features section
    if [[ -s "$features_file" ]]; then
        echo "### Added" >> "$temp_entry"
        while IFS= read -r commit; do
            echo "- $commit" >> "$temp_entry"
        done < "$features_file"
        echo "" >> "$temp_entry"
        debug_log "Added $(wc -l < "$features_file") features to changelog"
    fi
    
    # Add fixes section
    if [[ -s "$fixes_file" ]]; then
        echo "### Fixed" >> "$temp_entry"
        while IFS= read -r commit; do
            echo "- $commit" >> "$temp_entry"
        done < "$fixes_file"
        echo "" >> "$temp_entry"
        debug_log "Added $(wc -l < "$fixes_file") fixes to changelog"
    fi
    
    # Add other changes section
    if [[ -s "$other_file" ]]; then
        echo "### Changed" >> "$temp_entry"
        while IFS= read -r commit; do
            echo "- $commit" >> "$temp_entry"
        done < "$other_file"
        echo "" >> "$temp_entry"
        debug_log "Added $(wc -l < "$other_file") other changes to changelog"
    fi
    
    # Update CHANGELOG.md
    if [[ -f CHANGELOG.md ]]; then
        local temp_changelog
        temp_changelog=$(mktemp)
        head -n 1 CHANGELOG.md > "$temp_changelog"
        echo "" >> "$temp_changelog"
        cat "$temp_entry" >> "$temp_changelog"
        tail -n +2 CHANGELOG.md >> "$temp_changelog"
        mv "$temp_changelog" CHANGELOG.md
        log_info "Updated existing CHANGELOG.md"
    else
        echo "# Changelog" > CHANGELOG.md
        echo "" >> CHANGELOG.md
        cat "$temp_entry" >> CHANGELOG.md
        log_info "Created new CHANGELOG.md"
    fi
    
    rm -f "$temp_entry"
}

generate_release_notes() {
    local features_file="$1"
    local fixes_file="$2"
    local other_file="$3"
    
    log_info "Generating RELEASE_NOTES.txt..."
    
    # Get GitHub repo info
    local repo_url
    repo_url=$(git config --get remote.origin.url 2>/dev/null || echo "")
    local github_repo=""
    
    if [[ "$repo_url" =~ github\.com[:/]([^/]+/[^/]+)(\.git)?$ ]]; then
        github_repo="${BASH_REMATCH[1]%.git}"
    fi
    
    # Create release notes header
    cat > RELEASE_NOTES.txt << EOF
# ztictl $VERSION Release Notes

**Installation:** [Installation Guide](https://github.com/${github_repo:-your-org/your-repo}/blob/release/$VERSION/INSTALLATION.md)

**Release Date:** $(date '+%B %d, %Y')

EOF
    
    # Add features section
    echo "## 🚀 New Features" >> RELEASE_NOTES.txt
    if [[ -s "$features_file" ]]; then
        while IFS= read -r commit; do
            echo "• $commit" >> RELEASE_NOTES.txt
        done < "$features_file"
        debug_log "Added $(wc -l < "$features_file") features to release notes"
    else
        echo "• No new features in this release" >> RELEASE_NOTES.txt
    fi
    echo "" >> RELEASE_NOTES.txt
    
    # Add fixes section
    echo "## 🐛 Bug Fixes" >> RELEASE_NOTES.txt
    if [[ -s "$fixes_file" ]]; then
        while IFS= read -r commit; do
            echo "• $commit" >> RELEASE_NOTES.txt
        done < "$fixes_file"
        debug_log "Added $(wc -l < "$fixes_file") fixes to release notes"
    else
        echo "• No bug fixes in this release" >> RELEASE_NOTES.txt
    fi
    echo "" >> RELEASE_NOTES.txt
    
    # Add other changes section
    echo "## 📝 Other Changes" >> RELEASE_NOTES.txt
    if [[ -s "$other_file" ]]; then
        while IFS= read -r commit; do
            echo "• $commit" >> RELEASE_NOTES.txt
        done < "$other_file"
        debug_log "Added $(wc -l < "$other_file") other changes to release notes"
    else
        echo "• No other changes in this release" >> RELEASE_NOTES.txt
    fi
    
    log_info "RELEASE_NOTES.txt generated successfully"
}

# Check if files already exist
check_existing_files() {
    if [[ -f CHANGELOG.md ]] && [[ "$FORCE_REGENERATE" != true ]]; then
        log_error "CHANGELOG.md already exists. Use --force to regenerate."
        exit 1
    fi
    
    if [[ -f RELEASE_NOTES.txt ]] && [[ "$FORCE_REGENERATE" != true ]]; then
        log_error "RELEASE_NOTES.txt already exists. Use --force to regenerate."
        exit 1
    fi
}

# Main execution
main() {
    log_info "Starting release documentation generation"
    
    parse_args "$@"
    validate_version
    get_latest_tag
    check_existing_files
    
    generate_documentation
    
    log_info "Release documentation generation completed successfully!"
}

# Execute main function
main "$@"