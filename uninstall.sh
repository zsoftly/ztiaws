#!/bin/bash

# ZTiAWS Uninstallation Script
# Remove ZTiAWS tools from system

set -e  # Exit on any error

# Color output for better UX
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_color() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to check what's installed
check_installation() {
    print_color "$BLUE" "Checking current installation..."
    
    local found_files=()
    
    # Check for authaws
    if [[ -f "/usr/local/bin/authaws" ]]; then
        found_files+=("/usr/local/bin/authaws")
        print_color "$YELLOW" "  * Found: /usr/local/bin/authaws"
    fi
    
    # Check for ssm
    if [[ -f "/usr/local/bin/ssm" ]]; then
        found_files+=("/usr/local/bin/ssm")
        print_color "$YELLOW" "  * Found: /usr/local/bin/ssm"
    fi
    
    # Check for src directory
    if [[ -d "/usr/local/bin/src" ]]; then
        local src_files
        src_files=$(find /usr/local/bin/src -name "*.sh" 2>/dev/null | wc -l)
        if [[ $src_files -gt 0 ]]; then
            found_files+=("/usr/local/bin/src")
            print_color "$YELLOW" "  * Found: /usr/local/bin/src (with $src_files .sh files)"
        fi
    fi
    
    if [[ ${#found_files[@]} -eq 0 ]]; then
        print_color "$GREEN" "[OK] No ZTiAWS installation found"
        print_color "$BLUE" "Nothing to uninstall."
        exit 0
    fi
    
    echo
    print_color "$YELLOW" "Found ${#found_files[@]} ZTiAWS component(s) to remove."
    return 0
}

# Function to confirm uninstallation
confirm_uninstall() {
    if [[ "$1" != "--yes" ]] && [[ "$1" != "-y" ]]; then
        print_color "$RED" "[WARNING] This will permanently remove ZTiAWS from your system."
        echo
        read -p "Are you sure you want to continue? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_color "$BLUE" "Uninstallation cancelled."
            exit 0
        fi
    fi
}

# Function to remove files
remove_files() {
    print_color "$BLUE" "Removing ZTiAWS files..."
    
    local removed_count=0
    
    # Remove authaws
    if [[ -f "/usr/local/bin/authaws" ]]; then
        print_color "$BLUE" "  * Removing authaws..."
        if sudo rm -f /usr/local/bin/authaws; then
            print_color "$GREEN" "    [OK] Removed /usr/local/bin/authaws"
            ((removed_count++))
        else
            print_color "$RED" "    [ERROR] Failed to remove /usr/local/bin/authaws"
        fi
    fi
    
    # Remove ssm
    if [[ -f "/usr/local/bin/ssm" ]]; then
        print_color "$BLUE" "  * Removing ssm..."
        if sudo rm -f /usr/local/bin/ssm; then
            print_color "$GREEN" "    [OK] Removed /usr/local/bin/ssm"
            ((removed_count++))
        else
            print_color "$RED" "    [ERROR] Failed to remove /usr/local/bin/ssm"
        fi
    fi
    
    # Remove src directory
    if [[ -d "/usr/local/bin/src" ]]; then
        print_color "$BLUE" "  * Removing source modules..."
        if sudo rm -rf /usr/local/bin/src; then
            print_color "$GREEN" "    [OK] Removed /usr/local/bin/src"
            ((removed_count++))
        else
            print_color "$RED" "    [ERROR] Failed to remove /usr/local/bin/src"
        fi
    fi
    
    echo
    if [[ $removed_count -gt 0 ]]; then
        print_color "$GREEN" "[SUCCESS] Removed $removed_count ZTiAWS component(s)"
    else
        print_color "$RED" "[ERROR] No files were removed"
        return 1
    fi
}

# Function to verify removal
verify_removal() {
    print_color "$BLUE" "Verifying removal..."
    
    local remaining_files=()
    
    # Check if anything is left
    if command -v authaws &> /dev/null; then
        remaining_files+=("authaws")
    fi
    
    if command -v ssm &> /dev/null; then
        remaining_files+=("ssm")
    fi
    
    if [[ ${#remaining_files[@]} -gt 0 ]]; then
        print_color "$RED" "[ERROR] Some files may still be accessible:"
        for file in "${remaining_files[@]}"; do
            print_color "$YELLOW" "  * $file: $(which "$file")"
        done
        print_color "$YELLOW" "Note: These may be from a different installation location"
        return 1
    else
        print_color "$GREEN" "[OK] ZTiAWS commands are no longer accessible"
        return 0
    fi
}

# Function to show cleanup suggestions
show_cleanup_suggestions() {
    print_color "$BLUE" "Additional Cleanup (Optional):"
    echo
    print_color "$YELLOW" "You may also want to remove:"
    echo "  * AWS CLI credentials: ~/.aws/"
    echo "  * ZTiAWS configuration: ~/.env files"
    echo "  * Any custom aliases or PATH modifications"
    echo
    print_color "$BLUE" "These are not removed automatically for safety."
}

# Main uninstallation function
main() {
    print_color "$RED" "ZTiAWS Uninstallation"
    print_color "$BLUE" "======================================"
    echo
    print_color "$BLUE" "Removing ZTiAWS (ZSoftly Tools for AWS) from your system"
    echo
    
    check_installation
    confirm_uninstall "$1"
    
    if remove_files; then
        if verify_removal; then
            show_cleanup_suggestions
            print_color "$GREEN" "[SUCCESS] ZTiAWS uninstallation completed successfully!"
        else
            print_color "$YELLOW" "[WARNING] Uninstallation completed with warnings (see above)"
        fi
    else
        print_color "$RED" "[ERROR] Uninstallation failed"
        print_color "$YELLOW" "Some files may still be present on your system"
        exit 1
    fi
}

# Show help if requested
if [[ "$1" == "--help" ]] || [[ "$1" == "-h" ]]; then
    print_color "$RED" "ZTiAWS Uninstallation Script"
    echo
    print_color "$BLUE" "USAGE:"
    echo "  ./uninstall.sh         # Remove ZTiAWS (with confirmation)"
    echo "  ./uninstall.sh --yes   # Remove ZTiAWS (no confirmation)"
    echo "  ./uninstall.sh --help  # Show this help"
    echo
    print_color "$BLUE" "DESCRIPTION:"
    echo "  Removes ZTiAWS bash tools (authaws, ssm) from /usr/local/bin"
    echo "  Requires sudo permissions for removal."
    echo
    print_color "$BLUE" "WHAT GETS REMOVED:"
    echo "  * /usr/local/bin/authaws"
    echo "  * /usr/local/bin/ssm"
    echo "  * /usr/local/bin/src/*.sh (source modules)"
    echo
    print_color "$BLUE" "WHAT IS PRESERVED:"
    echo "  * AWS CLI configuration (~/.aws/)"
    echo "  * Personal configuration files"
    echo "  * Custom aliases or PATH modifications"
    echo
    exit 0
fi

# Run main uninstallation
main "$@"
