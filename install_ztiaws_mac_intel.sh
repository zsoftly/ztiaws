#!/bin/bash

# ZTiAWS Installation Script for Mac Intel (macOS Sonoma)
# This script installs both the stable shell scripts and the new Go binary

set -e  # Exit on any error

echo "🚀 ZTiAWS Installation Script for Mac Intel"
echo "============================================="
echo
echo "This script will install:"
echo "1. Production stable shell scripts (ssm & authaws)"
echo "2. New Go binary (ztictl) - Preview version"
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if running on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    print_error "This script is designed for macOS. Current OS: $OSTYPE"
    exit 1
fi

# Check for required tools
check_requirements() {
    print_info "Checking requirements..."
    
    if ! command -v curl &> /dev/null; then
        print_error "curl is required but not installed."
        exit 1
    fi
    
    if ! command -v chmod &> /dev/null; then
        print_error "chmod is required but not installed."
        exit 1
    fi
    
    print_status "All requirements met"
}

# Create installation directory
create_install_dir() {
    print_info "Creating installation directory..."
    INSTALL_DIR="$HOME/ztiaws"
    mkdir -p "$INSTALL_DIR"
    cd "$INSTALL_DIR"
    print_status "Created directory: $INSTALL_DIR"
}

# Install shell scripts (production stable)
install_shell_scripts() {
    print_info "Installing shell scripts (Production Stable v1.4.x)..."
    
    # Download main scripts
    curl -L -o ssm https://raw.githubusercontent.com/zsoftly/ztiaws/main/ssm
    curl -L -o authaws https://raw.githubusercontent.com/zsoftly/ztiaws/main/authaws
    
    # Make executable
    chmod +x ssm authaws
    
    # Download supporting files
    mkdir -p src
    curl -L -o src/00_utils.sh https://raw.githubusercontent.com/zsoftly/ztiaws/main/src/00_utils.sh
    curl -L -o src/01_regions.sh https://raw.githubusercontent.com/zsoftly/ztiaws/main/src/01_regions.sh
    curl -L -o src/02_ssm_instance_resolver.sh https://raw.githubusercontent.com/zsoftly/ztiaws/main/src/02_ssm_instance_resolver.sh
    curl -L -o src/03_ssm_command_runner.sh https://raw.githubusercontent.com/zsoftly/ztiaws/main/src/03_ssm_command_runner.sh
    curl -L -o src/04_ssm_file_transfer.sh https://raw.githubusercontent.com/zsoftly/ztiaws/main/src/04_ssm_file_transfer.sh
    
    print_status "Shell scripts installed successfully"
}

# Install Go binary (preview version)
install_go_binary() {
    print_info "Installing Go binary (Preview v2.0.x)..."
    
    # Download ztictl for Mac Intel
    curl -L -o ztictl https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-darwin-amd64
    
    # Make executable
    chmod +x ztictl
    
    print_status "Go binary installed successfully"
}

# Test installations
test_installations() {
    print_info "Testing installations..."
    
    # Test shell scripts
    if ./ssm --help &> /dev/null; then
        print_status "Shell script 'ssm' is working"
    else
        print_warning "Shell script 'ssm' test failed"
    fi
    
    if ./authaws --help &> /dev/null; then
        print_status "Shell script 'authaws' is working"
    else
        print_warning "Shell script 'authaws' test failed"
    fi
    
    # Test Go binary
    if ./ztictl --version &> /dev/null; then
        print_status "Go binary 'ztictl' is working"
    else
        print_warning "Go binary 'ztictl' test failed"
    fi
}

# Offer to install to system PATH
offer_system_install() {
    echo
    print_info "Installation completed in: $INSTALL_DIR"
    echo
    echo "Would you like to install to system PATH (/usr/local/bin)?"
    echo "This will allow you to run the commands from anywhere."
    echo -n "Install to system PATH? (y/N): "
    read -r response
    
    if [[ "$response" =~ ^[Yy]$ ]]; then
        print_info "Installing to system PATH..."
        
        # Check if /usr/local/bin exists
        if [[ ! -d "/usr/local/bin" ]]; then
            print_info "Creating /usr/local/bin directory..."
            sudo mkdir -p /usr/local/bin
        fi
        
        # Copy files
        sudo cp ssm authaws ztictl /usr/local/bin/
        sudo cp -r src /usr/local/bin/
        
        print_status "Tools installed to system PATH"
        
        # Test system installation
        if command -v ssm &> /dev/null; then
            print_status "System installation successful - commands available globally"
        else
            print_warning "System installation may have issues - try running from $INSTALL_DIR"
        fi
    else
        print_info "Tools installed locally in $INSTALL_DIR"
        print_info "To use the tools, run: cd $INSTALL_DIR && ./ssm or ./ztictl"
    fi
}

# Show usage information
show_usage() {
    echo
    print_info "Usage Examples:"
    echo
    echo "Shell Scripts (Production Stable):"
    echo "  ./ssm --help                    # Show help"
    echo "  ./authaws configure             # Configure AWS SSO"
    echo "  ./ssm list                      # List SSM instances"
    echo "  ./ssm connect <instance-id>     # Connect to instance"
    echo
    echo "Go Binary (Preview):"
    echo "  ./ztictl --help                 # Show help"
    echo "  ./ztictl auth configure         # Configure AWS SSO"
    echo "  ./ztictl ssm list               # List SSM instances"
    echo "  ./ztictl ssm connect <id>       # Connect to instance"
    echo
    print_info "Next Steps:"
    echo "1. Configure AWS CLI: aws configure"
    echo "2. Configure AWS SSO: ./authaws configure OR ./ztictl auth configure"
    echo "3. Test with your AWS environment"
    echo "4. Compare both versions for your workflow"
    echo
    print_info "Documentation: https://github.com/zsoftly/ztiaws"
}

# Main installation process
main() {
    check_requirements
    create_install_dir
    install_shell_scripts
    install_go_binary
    test_installations
    offer_system_install
    show_usage
    
    echo
    print_status "Installation completed successfully!"
    print_info "Happy DevOps-ing with ZTiAWS! 🎉"
}

# Run main function
main