package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// sudoCommand is a variable to allow mocking in tests
var sudoCommand = "sudo"

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate and install shell completion scripts",
	Long: `Generate and install shell completion scripts for ztictl.

This command helps you set up auto-completion for ztictl commands in your shell.
Running without arguments will detect your shell and provide setup instructions.`,
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	Args:      cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		install, _ := cmd.Flags().GetBool("install")

		// Detect shell if not specified
		shell := ""
		if len(args) > 0 {
			shell = args[0]
		} else {
			shell = detectShell()
			if shell == "" {
				fmt.Println("❌ Could not detect your shell automatically")
				fmt.Println("\n🔍 Please specify your shell explicitly:")
				fmt.Println("   ztictl completion bash")
				fmt.Println("   ztictl completion zsh")
				fmt.Println("   ztictl completion fish")
				fmt.Println("   ztictl completion powershell")
				return
			}
			fmt.Printf("🔍 Detected shell: %s\n\n", shell)
		}

		// Handle installation if requested
		if install {
			if err := installCompletion(shell); err != nil {
				logger.Error("Failed to install completion", "error", err)
				os.Exit(1)
			}
			return
		}

		// Otherwise show instructions
		showCompletionInstructions(shell)
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
	completionCmd.Flags().BoolP("install", "i", false, "Automatically install completion to the appropriate location")
}

func detectShell() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		shellName := filepath.Base(shell)
		switch {
		case strings.Contains(shellName, "bash"):
			return "bash"
		case strings.Contains(shellName, "zsh"):
			return "zsh"
		case strings.Contains(shellName, "fish"):
			return "fish"
		}
	}

	if runtime.GOOS == "windows" {
		if os.Getenv("PSModulePath") != "" {
			return "powershell"
		}
	}

	if runtime.GOOS != "windows" {
		ppid := os.Getppid()
		if ppid > 0 {
			cmd := exec.Command("ps", "-p", fmt.Sprintf("%d", ppid), "-o", "comm=") // #nosec G204
			if output, err := cmd.Output(); err == nil {
				shellName := strings.TrimSpace(string(output))
				switch {
				case strings.Contains(shellName, "bash"):
					return "bash"
				case strings.Contains(shellName, "zsh"):
					return "zsh"
				case strings.Contains(shellName, "fish"):
					return "fish"
				case strings.Contains(shellName, "pwsh"), strings.Contains(shellName, "powershell"):
					return "powershell"
				}
			}
		}
	}

	return ""
}

func showCompletionInstructions(shell string) {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("🚀 Setting up %s completion for ztictl\n", shell)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	switch shell {
	case "bash":
		showBashInstructions()
	case "zsh":
		showZshInstructions()
	case "fish":
		showFishInstructions()
	case "powershell":
		showPowerShellInstructions()
	default:
		fmt.Printf("❌ Unknown shell: %s\n", shell)
		fmt.Println("\nSupported shells: bash, zsh, fish, powershell")
	}
}

func showBashInstructions() {
	fmt.Println("\n📋 BASH COMPLETION SETUP")
	fmt.Println("========================")

	if _, err := exec.LookPath("bash-completion"); err != nil {
		fmt.Println("\n⚠️  Prerequisites:")
		fmt.Println("   bash-completion package is required but not installed")
		fmt.Println("\n   Install it first:")
		switch runtime.GOOS {
		case "darwin":
			fmt.Println("   brew install bash-completion")
		case "linux":
			fmt.Println("   # Ubuntu/Debian:")
			fmt.Println("   sudo apt-get install bash-completion")
			fmt.Println("\n   # RHEL/CentOS/Fedora:")
			fmt.Println("   sudo yum install bash-completion")
		}
		fmt.Println()
	}

	fmt.Println("\n🔧 OPTION 1: Quick Install (Recommended)")
	fmt.Println("─────────────────────────────────────────")
	fmt.Println("Run this command:")
	fmt.Println("   ztictl completion bash --install")

	fmt.Println("\n🔧 OPTION 2: Manual Installation")
	fmt.Println("─────────────────────────────────────────")

	fmt.Println("\nFor current session only:")
	fmt.Println("   source <(ztictl completion bash)")

	fmt.Println("\nFor permanent installation:")

	switch runtime.GOOS {
	case "darwin":
		fmt.Println("   # macOS:")
		fmt.Println("   ztictl completion bash > $(brew --prefix)/etc/bash_completion.d/ztictl")
	case "linux":
		fmt.Println("   # Linux (system-wide):")
		fmt.Println("   sudo ztictl completion bash > /etc/bash_completion.d/ztictl")
		fmt.Println("\n   # Linux (user-specific):")
		fmt.Println("   mkdir -p ~/.local/share/bash-completion/completions")
		fmt.Println("   ztictl completion bash > ~/.local/share/bash-completion/completions/ztictl")
	default:
		fmt.Println("   # Add to your ~/.bashrc or ~/.bash_profile:")
		fmt.Println("   echo 'source <(ztictl completion bash)' >> ~/.bashrc")
	}

	fmt.Println("\n✅ FINAL STEP:")
	fmt.Println("──────────────")
	fmt.Println("   Restart your shell or run: source ~/.bashrc")

	fmt.Println("\n📝 TEST IT:")
	fmt.Println("────────────")
	fmt.Println("   Type 'ztictl ' and press TAB twice to see available commands")
}

func showZshInstructions() {
	fmt.Println("\n📋 ZSH COMPLETION SETUP")
	fmt.Println("=======================")

	fmt.Println("\n🔧 OPTION 1: Quick Install (Recommended)")
	fmt.Println("─────────────────────────────────────────")
	fmt.Println("Run this command:")
	fmt.Println("   ztictl completion zsh --install")

	fmt.Println("\n🔧 OPTION 2: Oh My Zsh Users")
	fmt.Println("─────────────────────────────────────────")
	fmt.Println("   mkdir -p ~/.oh-my-zsh/custom/plugins/ztictl")
	fmt.Println("   ztictl completion zsh > ~/.oh-my-zsh/custom/plugins/ztictl/_ztictl")
	fmt.Println("\n   ⚠️  IMPORTANT: Add 'ztictl' to your plugins in ~/.zshrc:")
	fmt.Println("   plugins=(git zsh-autosuggestions ztictl)  # Add ztictl here")
	fmt.Println("\n   Then reload: source ~/.zshrc")

	fmt.Println("\n🔧 OPTION 3: Manual Installation")
	fmt.Println("─────────────────────────────────────────")

	fmt.Println("\nFor current session only:")
	fmt.Println("   source <(ztictl completion zsh)")

	fmt.Println("\nFor permanent installation:")
	fmt.Println("   # Add to your ~/.zshrc:")
	fmt.Println("   echo 'source <(ztictl completion zsh)' >> ~/.zshrc")

	fmt.Println("\n   # Or use the completions directory:")
	fmt.Println("   ztictl completion zsh > ~/.zsh/completions/_ztictl")
	fmt.Println("   # Make sure ~/.zsh/completions is in your fpath")

	fmt.Println("\n✅ FINAL STEP:")
	fmt.Println("──────────────")
	fmt.Println("   Restart your shell or run: source ~/.zshrc")

	fmt.Println("\n📝 HOW TO USE TAB COMPLETION:")
	fmt.Println("────────────────────────────────")
	fmt.Println("   1. Type: ztictl ")
	fmt.Println("   2. Press TAB once to see available commands")
	fmt.Println("   3. Start typing a command and press TAB to autocomplete")
	fmt.Println("   4. Press TAB twice to see all options")
	fmt.Println("\n   Examples:")
	fmt.Println("   • ztictl s[TAB]        → completes to 'ssm'")
	fmt.Println("   • ztictl ssm [TAB]     → shows: list, connect, exec, etc.")
	fmt.Println("   • ztictl ssm con[TAB]  → completes to 'connect'")

	fmt.Println("\n💡 TROUBLESHOOTING:")
	fmt.Println("───────────────────")
	fmt.Println("   If completion doesn't work:")
	fmt.Println("   1. Ensure compinit is loaded:")
	fmt.Println("      echo 'autoload -Uz compinit && compinit' >> ~/.zshrc")
	fmt.Println("   2. For Oh My Zsh users, verify plugin is in list:")
	fmt.Println("      grep 'plugins=' ~/.zshrc")
	fmt.Println("   3. Check if completion file exists:")
	fmt.Println("      ls ~/.oh-my-zsh/custom/plugins/ztictl/_ztictl")
}

func showFishInstructions() {
	fmt.Println("\n📋 FISH COMPLETION SETUP")
	fmt.Println("========================")

	fmt.Println("\n🔧 OPTION 1: Quick Install (Recommended)")
	fmt.Println("─────────────────────────────────────────")
	fmt.Println("Run this command:")
	fmt.Println("   ztictl completion fish --install")

	fmt.Println("\n🔧 OPTION 2: Manual Installation")
	fmt.Println("─────────────────────────────────────────")

	fmt.Println("\nFor current session only:")
	fmt.Println("   ztictl completion fish | source")

	fmt.Println("\nFor permanent installation:")
	fmt.Println("   ztictl completion fish > ~/.config/fish/completions/ztictl.fish")

	fmt.Println("\n✅ FINAL STEP:")
	fmt.Println("──────────────")
	fmt.Println("   Completions will be available immediately in new shells")

	fmt.Println("\n📝 TEST IT:")
	fmt.Println("────────────")
	fmt.Println("   Type 'ztictl ' and press TAB to see available commands")
}

func showPowerShellInstructions() {
	fmt.Println("\n📋 POWERSHELL COMPLETION SETUP")
	fmt.Println("===============================")

	fmt.Println("\n📦 PowerShell Versions:")
	fmt.Println("─────────────────────────")
	fmt.Println("• Windows PowerShell (5.1) - Built into Windows")
	fmt.Println("  Profile: ~/Documents/WindowsPowerShell/Microsoft.PowerShell_profile.ps1")
	fmt.Println("• PowerShell Core (7+) - Cross-platform version")
	fmt.Println("  Profile: ~/Documents/PowerShell/Microsoft.PowerShell_profile.ps1")

	fmt.Println("\n🔧 OPTION 1: Quick Install (Recommended)")
	fmt.Println("─────────────────────────────────────────")
	fmt.Println("Run this command:")
	fmt.Println("   ztictl completion powershell --install")
	fmt.Println("\nThis will auto-detect your PowerShell version and install to the correct location.")

	fmt.Println("\n🔧 OPTION 2: Manual Installation")
	fmt.Println("─────────────────────────────────────────")

	fmt.Println("\nFor current session only:")
	fmt.Println("   ztictl completion powershell | Out-String | Invoke-Expression")

	fmt.Println("\nFor permanent installation:")
	fmt.Println("   1. First, check your PowerShell profile path:")
	fmt.Println("      echo $PROFILE")

	fmt.Println("\n   2. Check which PowerShell you're using:")
	fmt.Println("      $PSVersionTable.PSVersion")

	fmt.Println("\n   3. Create the profile if it doesn't exist:")
	fmt.Println("      if (!(Test-Path -Path $PROFILE)) { New-Item -ItemType File -Path $PROFILE -Force }")

	fmt.Println("\n   4. Add the completion to your profile:")
	fmt.Println("      ztictl completion powershell >> $PROFILE")

	fmt.Println("\n✅ FINAL STEP:")
	fmt.Println("──────────────")
	fmt.Println("   Restart PowerShell or run: . $PROFILE")

	fmt.Println("\n📝 HOW TO TEST:")
	fmt.Println("────────────────")
	fmt.Println("   1. Type: ztictl [SPACE][TAB]")
	fmt.Println("      Result: Shows available commands")
	fmt.Println("   2. Type: ztictl ssm [SPACE][TAB]")
	fmt.Println("      Result: Shows ssm subcommands")

	fmt.Println("\n🔍 VERIFY INSTALLATION:")
	fmt.Println("──────────────────────")
	fmt.Println("   # Check if profile exists and contains ztictl:")
	fmt.Println("   Test-Path $PROFILE")
	fmt.Println("   Select-String -Path $PROFILE -Pattern 'ztictl'")

	fmt.Println("\n⚠️  TROUBLESHOOTING:")
	fmt.Println("────────────────────")
	fmt.Println("   If you get an execution policy error:")
	fmt.Println("   Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser")
	fmt.Println("\n   If completion doesn't work after restart:")
	fmt.Println("   1. Check if profile is loaded: echo $PROFILE")
	fmt.Println("   2. Manually load it: . $PROFILE")
	fmt.Println("   3. Check for errors: Get-Error")
}

func installCompletion(shell string) error {
	fmt.Printf("🔧 Installing %s completion...\n\n", shell)

	switch shell {
	case "bash":
		return installBashCompletion()
	case "zsh":
		return installZshCompletion()
	case "fish":
		return installFishCompletion()
	case "powershell":
		return installPowerShellCompletion()
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}
}

func installBashCompletion() error {
	// Generate the completion script
	var completionScript strings.Builder
	if err := rootCmd.GenBashCompletion(&completionScript); err != nil {
		return fmt.Errorf("failed to generate bash completion: %w", err)
	}

	// Determine installation path
	var installPath string

	switch runtime.GOOS {
	case "darwin":
		if brewPrefix, err := exec.Command("brew", "--prefix").Output(); err == nil {
			installPath = filepath.Join(strings.TrimSpace(string(brewPrefix)), "etc", "bash_completion.d", "ztictl")
		}
	case "linux":
		// Try system-wide first
		if _, err := os.Stat("/etc/bash_completion.d"); err == nil {
			installPath = "/etc/bash_completion.d/ztictl"
		} else {
			// Fall back to user directory
			home, _ := os.UserHomeDir()
			completionDir := filepath.Join(home, ".local", "share", "bash-completion", "completions")
			if err := os.MkdirAll(completionDir, 0755); err != nil {
				return fmt.Errorf("failed to create bash completion directory: %w", err)
			}
			installPath = filepath.Join(completionDir, "ztictl")
		}
	}

	if installPath == "" {
		// Fallback: add to bashrc
		home, _ := os.UserHomeDir()
		bashrc := filepath.Join(home, ".bashrc")

		if content, err := os.ReadFile(bashrc); err == nil {
			contentStr := string(content)
			if strings.Contains(contentStr, "ztictl completion bash") ||
				strings.Contains(contentStr, "ztictl  completion  bash") ||
				strings.Contains(contentStr, "$(ztictl completion bash)") ||
				strings.Contains(contentStr, "`ztictl completion bash`") {
				fmt.Println("✅ Completion already configured in ~/.bashrc")
				return nil
			}
		}

		// Add to bashrc
		f, err := os.OpenFile(bashrc, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("failed to open ~/.bashrc: %w", err)
		}
		defer f.Close()

		fmt.Fprintf(f, "\n# ztictl completion\nsource <(ztictl completion bash)\n")
		fmt.Println("✅ Added completion to ~/.bashrc")
		fmt.Println("🔄 Restart your shell or run: source ~/.bashrc")
		return nil
	}

	if strings.Contains(installPath, "..") {
		return fmt.Errorf("path traversal detected in install path: %s", installPath)
	}

	needsSudo := false
	if strings.HasPrefix(installPath, "/etc") || strings.HasPrefix(installPath, "/usr") {
		validSystemPaths := []string{
			"/etc/bash_completion.d/",
			"/usr/share/bash-completion/completions/",
			"/usr/local/share/bash-completion/completions/",
		}

		isValid := false
		for _, validPath := range validSystemPaths {
			if strings.HasPrefix(installPath, validPath) {
				isValid = true
				needsSudo = true
				break
			}
		}

		if !isValid {
			return fmt.Errorf("refusing to install to non-standard system path: %s", installPath)
		}
	}

	if needsSudo {
		fmt.Printf("📝 Installing to %s (requires sudo)...\n", installPath)

		// Create temp file in user's temp directory
		tempFile, err := os.CreateTemp("", "ztictl-completion-*.sh")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		tempPath := tempFile.Name()
		defer os.Remove(tempPath) // Clean up temp file

		// Write completion script to temp file
		if _, err := tempFile.WriteString(completionScript.String()); err != nil {
			tempFile.Close()
			return fmt.Errorf("failed to write temp file: %w", err)
		}
		tempFile.Close()

		cmd := exec.Command(sudoCommand, "cp", tempPath, installPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install completion with sudo: %w", err)
		}

		cmd = exec.Command(sudoCommand, "chmod", "644", installPath)
		if err := cmd.Run(); err != nil {
			fmt.Printf("⚠️  Warning: Could not set permissions on %s\n", installPath)
		}
	} else {
		if err := os.WriteFile(installPath, []byte(completionScript.String()), 0600); err != nil {
			return fmt.Errorf("failed to write completion file: %w", err)
		}
	}

	fmt.Printf("✅ Completion installed to %s\n", installPath)
	fmt.Println("🔄 Restart your shell for changes to take effect")
	return nil
}

func installZshCompletion() error {
	// Generate the completion script
	var completionScript strings.Builder
	if err := rootCmd.GenZshCompletion(&completionScript); err != nil {
		return fmt.Errorf("failed to generate zsh completion: %w", err)
	}

	home, _ := os.UserHomeDir()

	if _, err := os.Stat(filepath.Join(home, ".oh-my-zsh")); err == nil {
		// Install to Oh My Zsh custom plugins
		pluginDir := filepath.Join(home, ".oh-my-zsh", "custom", "plugins", "ztictl")
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			return fmt.Errorf("failed to create oh-my-zsh plugin directory: %w", err)
		}

		completionPath := filepath.Join(pluginDir, "_ztictl")
		if err := os.WriteFile(completionPath, []byte(completionScript.String()), 0600); err != nil {
			return fmt.Errorf("failed to write completion file: %w", err)
		}

		fmt.Printf("✅ Completion file installed to Oh My Zsh custom plugins\n")

		zshrc := filepath.Join(home, ".zshrc")
		content, err := os.ReadFile(zshrc)
		if err == nil {
			contentStr := string(content)
			if strings.Contains(contentStr, "plugins=(") && !strings.Contains(contentStr, "ztictl") {
				fmt.Println("\n⚠️  ACTION REQUIRED:")
				fmt.Println("────────────────────")
				fmt.Println("You need to add 'ztictl' to your plugins list in ~/.zshrc")
				fmt.Println("\n1. Edit ~/.zshrc and find the line starting with: plugins=(")
				fmt.Println("2. Add 'ztictl' to the list, for example:")
				fmt.Println("   plugins=(git zsh-autosuggestions ztictl)")
				fmt.Println("3. Save the file and reload: source ~/.zshrc")
				fmt.Println("\nTip: You can edit with your favorite editor:")
				fmt.Println("   nano ~/.zshrc")
				fmt.Println("   # or")
				fmt.Println("   vim ~/.zshrc")
			} else if strings.Contains(contentStr, "ztictl") {
				fmt.Println("✅ Plugin 'ztictl' already in ~/.zshrc")
				fmt.Println("🔄 Reload your shell: source ~/.zshrc")
			}
		}
		return nil
	}

	// Regular zsh installation
	zshrc := filepath.Join(home, ".zshrc")

	if content, err := os.ReadFile(zshrc); err == nil {
		contentStr := string(content)
		if strings.Contains(contentStr, "ztictl completion zsh") ||
			strings.Contains(contentStr, "ztictl  completion  zsh") ||
			strings.Contains(contentStr, "$(ztictl completion zsh)") ||
			strings.Contains(contentStr, "`ztictl completion zsh`") {
			fmt.Println("✅ Completion already configured in ~/.zshrc")
			return nil
		}
	}

	// Add to zshrc
	f, err := os.OpenFile(zshrc, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open ~/.zshrc: %w", err)
	}
	defer f.Close()

	fmt.Fprintf(f, "\n# ztictl completion\nsource <(ztictl completion zsh)\n")
	fmt.Println("✅ Added completion to ~/.zshrc")
	fmt.Println("🔄 Restart your shell or run: source ~/.zshrc")
	return nil
}

func installFishCompletion() error {
	// Generate the completion script
	var completionScript strings.Builder
	if err := rootCmd.GenFishCompletion(&completionScript, true); err != nil {
		return fmt.Errorf("failed to generate fish completion: %w", err)
	}

	home, _ := os.UserHomeDir()
	completionDir := filepath.Join(home, ".config", "fish", "completions")
	if err := os.MkdirAll(completionDir, 0755); err != nil {
		return fmt.Errorf("failed to create fish completion directory: %w", err)
	}

	completionPath := filepath.Join(completionDir, "ztictl.fish")
	if err := os.WriteFile(completionPath, []byte(completionScript.String()), 0600); err != nil {
		return fmt.Errorf("failed to write completion file: %w", err)
	}

	fmt.Printf("✅ Completion installed to %s\n", completionPath)
	fmt.Println("🎉 Completion is now active in new fish shells")
	return nil
}

func installPowerShellCompletion() error {
	// Generate the completion script
	var completionScript strings.Builder
	if err := rootCmd.GenPowerShellCompletion(&completionScript); err != nil {
		return fmt.Errorf("failed to generate PowerShell completion: %w", err)
	}

	// Try to detect which PowerShell version is being used
	var profilePath string
	var psVersion string

	// First try PowerShell Core (pwsh)
	if _, err := exec.LookPath("pwsh"); err == nil {
		cmd := exec.Command("pwsh", "-Command", "echo $PROFILE")
		if output, err := cmd.Output(); err == nil {
			profilePath = strings.TrimSpace(string(output))
			psVersion = "PowerShell Core (pwsh)"
		}
	}

	// Fall back to Windows PowerShell if Core not found or failed
	if profilePath == "" {
		cmd := exec.Command("powershell", "-Command", "echo $PROFILE")
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get PowerShell profile path: %w", err)
		}
		profilePath = strings.TrimSpace(string(output))
		psVersion = "Windows PowerShell"
	}

	fmt.Printf("🔍 Detected: %s\n", psVersion)
	fmt.Printf("📁 Profile location: %s\n", profilePath)

	// Create profile directory if it doesn't exist
	profileDir := filepath.Dir(profilePath)
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}

	if content, err := os.ReadFile(profilePath); err == nil {
		contentStr := string(content)
		if strings.Contains(contentStr, "ztictl completion powershell") ||
			strings.Contains(contentStr, "ztictl  completion  powershell") ||
			strings.Contains(contentStr, "Register-ArgumentCompleter") && strings.Contains(contentStr, "ztictl") {
			fmt.Println("✅ Completion already configured in PowerShell profile")
			return nil
		}
	}

	// Add to profile
	f, err := os.OpenFile(profilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open PowerShell profile: %w", err)
	}
	defer f.Close()

	fmt.Fprintf(f, "\n# ztictl completion\n%s\n", completionScript.String())
	fmt.Printf("✅ Completion installed to %s\n", profilePath)
	fmt.Println("🔄 Restart PowerShell or run: . $PROFILE")
	return nil
}
