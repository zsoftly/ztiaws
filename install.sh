#!/bin/bash

# ZTiAWS Installation Script
# Simple installation for end users (no make required)

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

# Function to check if running as root/sudo
check_permissions() {
    if [[ $EUID -eq 0 ]]; then
        print_color "$YELLOW" "Warning: Running as root. This is not recommended for security reasons."
        print_color "$YELLOW" "Consider running without sudo and entering password when prompted."
        echo
    fi
}

# Function to check system requirements
check_requirements() {
    print_color "$BLUE" "Checking system requirements..."
    
    # Check if AWS CLI is installed
    if ! command -v aws &> /dev/null; then
        print_color "$RED" "[ERROR] AWS CLI is not installed"
        print_color "$YELLOW" "Please install AWS CLI first:"
        print_color "$YELLOW" "https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html"
        exit 1
    else
        local aws_version
        aws_version=$(aws --version 2>&1 | cut -d/ -f2 | cut -d' ' -f1)
        print_color "$GREEN" "[OK] AWS CLI found (version: $aws_version)"
    fi
    
    # Check if we have write permissions to /usr/local/bin
    if [[ ! -w "/usr/local/bin" ]]; then
        print_color "$YELLOW" "[WARNING] /usr/local/bin is not writable. Will need sudo for installation."
        USE_SUDO=true
    else
        USE_SUDO=false
    fi
    
    echo
}

# Function to install files
install_files() {
    print_color "$BLUE" "Installing ZTiAWS files..."
    
    local install_cmd=""
    if [[ $USE_SUDO == true ]]; then
        install_cmd="sudo"
        print_color "$YELLOW" "Administrator privileges required for installation..."
    fi
    
    # Create src directory if it doesn't exist
    $install_cmd mkdir -p /usr/local/bin/src
    
    # Copy main scripts
    print_color "$BLUE" "  * Installing authaws..."
    $install_cmd cp authaws /usr/local/bin/
    $install_cmd chmod +x /usr/local/bin/authaws
    
    print_color "$BLUE" "  * Installing ssm..."
    $install_cmd cp ssm /usr/local/bin/
    $install_cmd chmod +x /usr/local/bin/ssm
    
    # Copy source modules
    print_color "$BLUE" "  * Installing source modules..."
    $install_cmd cp src/*.sh /usr/local/bin/src/
    
    print_color "$GREEN" "[SUCCESS] Files installed successfully!"
    echo
}

# Function to verify installation
verify_installation() {
    print_color "$BLUE" "Verifying installation..."
    
    # Test authaws
    if command -v authaws &> /dev/null; then
        print_color "$GREEN" "[OK] authaws command available"
        local authaws_version
        authaws_version=$(authaws --version 2>/dev/null | head -n1)
        print_color "$BLUE" "   Version: $authaws_version"
    else
        print_color "$RED" "[ERROR] authaws command not found"
        return 1
    fi
    
    # Test ssm
    if command -v ssm &> /dev/null; then
        print_color "$GREEN" "[OK] ssm command available" 
        local ssm_version
        ssm_version=$(ssm --version 2>/dev/null | head -n1)
        print_color "$BLUE" "   Version: $ssm_version"
    else
        print_color "$RED" "[ERROR] ssm command not found"
        return 1
    fi
    
    echo
    print_color "$GREEN" "Installation verification successful!"
    return 0
}

# Function to show next steps
show_next_steps() {
    print_color "$BLUE" "Next Steps:"
    echo
    print_color "$YELLOW" "1. Verify your AWS configuration:"
    echo "   authaws --check"
    echo
    print_color "$YELLOW" "2. Check SSM requirements:"
    echo "   ssm --check"
    echo
    print_color "$YELLOW" "3. Get help:"
    echo "   authaws --help"
    echo "   ssm --help"
    echo
    print_color "$YELLOW" "4. Quick start example:"
    echo "   # Authenticate with AWS SSO"
    echo "   authaws your-profile-name"
    echo
    echo "   # List EC2 instances in Canada Central"
    echo "   ssm --region cac1 --list"
    echo
    print_color "$BLUE" "Documentation: https://github.com/zsoftly/ztiaws"
    echo
}

# Function to show uninstall instructions
show_uninstall_info() {
    print_color "$BLUE" "To uninstall ZTiAWS in the future:"
    echo "   sudo rm -f /usr/local/bin/authaws"
    echo "   sudo rm -f /usr/local/bin/ssm" 
    echo "   sudo rm -rf /usr/local/bin/src"
    echo
}

# Main installation function
main() {
    print_color "$GREEN" "ZTiAWS Installation"
    print_color "$BLUE" "======================================"
    echo
    print_color "$BLUE" "Installing ZTiAWS (ZSoftly Tools for AWS)"
    print_color "$BLUE" "Legacy bash tools for AWS SSM and SSO management"
    echo
    
    # Check if we're in the right directory
    if [[ ! -f "authaws" ]] || [[ ! -f "ssm" ]] || [[ ! -d "src" ]]; then
        print_color "$RED" "[ERROR] Installation files not found in current directory"
        print_color "$YELLOW" "Please run this script from the ZTiAWS project directory"
        print_color "$YELLOW" "Expected files: authaws, ssm, src/"
        exit 1
    fi
    
    check_permissions
    check_requirements
    install_files
    
    if verify_installation; then
        show_next_steps
        show_uninstall_info
        print_color "$GREEN" "[SUCCESS] ZTiAWS installation completed successfully!"
    else
        print_color "$RED" "[ERROR] Installation verification failed"
        print_color "$YELLOW" "Please check the error messages above and try again"
        exit 1
    fi
}

# Show help if requested
if [[ "$1" == "--help" ]] || [[ "$1" == "-h" ]]; then
    print_color "$GREEN" "ZTiAWS Installation Script"
    echo
    print_color "$BLUE" "USAGE:"
    echo "  ./install.sh         # Install ZTiAWS"
    echo "  ./install.sh --help  # Show this help"
    echo
    print_color "$BLUE" "DESCRIPTION:"
    echo "  Installs ZTiAWS bash tools (authaws, ssm) to /usr/local/bin"
    echo "  for global access. No build tools or dependencies required."
    echo
    print_color "$BLUE" "REQUIREMENTS:"
    echo "  * AWS CLI (https://aws.amazon.com/cli/)"
    echo "  * Write access to /usr/local/bin (may require sudo)"
    echo
    print_color "$BLUE" "FOR DEVELOPERS:"
    echo "  Use 'make dev' instead for development environment setup"
    echo
    exit 0
fi

# Run main installation
main "$@"
