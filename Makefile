# ZTiAWS Makefile
# Simple installation and development management

.PHONY: install uninstall dev clean test help

# Default target
help:
	@echo "ZTiAWS - AWS Tools Installation"
	@echo ""
	@echo "Available targets:"
	@echo "  install    - Install authaws and ssm to /usr/local/bin (global access)"
	@echo "  uninstall  - Remove installed tools from /usr/local/bin"
	@echo "  dev        - Set up development environment (adds local PATH)"
	@echo "  test       - Run tests and validation"
	@echo "  clean      - Clean up temporary files"
	@echo "  help       - Show this help message"
	@echo ""
	@echo "Quick start:"
	@echo "  make install    # Install globally (recommended)"
	@echo "  authaws --check # Test installation"

# Install tools globally to /usr/local/bin
install:
	@echo "Installing ZTiAWS tools to /usr/local/bin..."
	sudo cp authaws /usr/local/bin/authaws
	sudo cp ssm /usr/local/bin/ssm
	sudo chmod +x /usr/local/bin/authaws /usr/local/bin/ssm
	@echo "Creating source directory..."
	sudo mkdir -p /usr/local/bin/src
	sudo cp src/*.sh /usr/local/bin/src/
	sudo chmod +x /usr/local/bin/src/*.sh
	@echo ""
	@echo "✅ Installation complete!"
	@echo "You can now run 'authaws --check' and 'ssm --help' from anywhere."

# Remove installed tools
uninstall:
	@echo "Removing ZTiAWS tools from /usr/local/bin..."
	sudo rm -f /usr/local/bin/authaws
	sudo rm -f /usr/local/bin/ssm
	sudo rm -rf /usr/local/bin/src
	@echo "✅ Uninstallation complete!"

# Development environment setup
dev:
	@echo "Setting up development environment..."
	@echo "Adding current directory to PATH for this session..."
	@grep -qxF 'export PATH="$(PWD):$$PATH"' ~/.zshrc || echo 'export PATH="$(PWD):$$PATH"' >> ~/.zshrc
	@grep -qxF 'export PATH="$(PWD):$$PATH"' ~/.bashrc || echo 'export PATH="$(PWD):$$PATH"' >> ~/.bashrc
	@echo ""
	@echo "✅ Development environment configured!"
	@echo "Reload your shell or run: source ~/.zshrc"
	@echo "Then you can use 'authaws --check' directly in development."

# Run tests
test:
	@echo "Running shellcheck..."
	shellcheck -x authaws ssm src/*.sh
	@echo "Running development functionality tests..."
	@echo "  (Using ./command syntax to test local development versions)"
	./authaws --help > /dev/null
	./ssm --help > /dev/null
	@echo "✅ Tests passed!"

# Clean up
clean:
	@echo "Cleaning up temporary files..."
	find . -name "*.log" -delete
	find . -name ".DS_Store" -delete
	@echo "✅ Cleanup complete!"
