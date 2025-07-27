package splash

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
)

const (
	// ASCII art banner for ztictl with butterfly theme
	banner = `
  ╭─────────────────────────────────────────────╮
  │    🦋 ztictl - Transform Your AWS Workflow   │
  │         ╭─╮   Small commands,              │
  │      ╭─╯   ╰─╮ Big impact                  │
  │   ╭─╯  ◦ ◦  ╰─╮                            │
  │  ╰─╮    ◦    ╱─╯  🔐 SSO • 🖥️ SSM • ⚡ More │
  │     ╰─╮     ╱                              │
  │       ╰─────╯                              │
  ╰─────────────────────────────────────────────╯
`

	// Version tracking file
	versionTrackingFile = ".ztictl_version"
)

// SplashConfig contains configuration for the splash screen
type SplashConfig struct {
	AppVersion   string
	AppName      string
	Description  string
	Features     []string
	IsFirstRun   bool
	IsNewVersion bool
}

// ShowSplash displays the welcome splash screen if appropriate
func ShowSplash(version string) (bool, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("failed to get home directory: %w", err)
	}

	versionFile := filepath.Join(homeDir, versionTrackingFile)

	// Check if this is first run or new version
	isFirstRun := false
	isNewVersion := false

	if _, err := os.Stat(versionFile); os.IsNotExist(err) {
		isFirstRun = true
	} else {
		// Read the last version
		lastVersionBytes, err := os.ReadFile(versionFile)
		if err != nil {
			return false, fmt.Errorf("failed to read version file: %w", err)
		}

		lastVersion := strings.TrimSpace(string(lastVersionBytes))
		if lastVersion != version {
			isNewVersion = true
		}
	}

	// Show splash if first run or new version
	if isFirstRun || isNewVersion {
		config := SplashConfig{
			AppVersion:   version,
			AppName:      "ztictl",
			Description:  "Unified AWS SSO & Systems Manager CLI",
			IsFirstRun:   isFirstRun,
			IsNewVersion: isNewVersion,
			Features: []string{
				"🔐 AWS SSO Authentication with interactive selection",
				"🖥️  SSM Session Manager connections",
				"⚡ Remote command execution via SSM",
				"📁 File transfer through SSM (with S3 for large files)",
				"🌐 Port forwarding through SSM tunnels",
				"🌍 Multi-region support",
				"📊 Comprehensive logging and configuration",
			},
		}

		displaySplash(config)

		// Update version tracking file
		if err := os.WriteFile(versionFile, []byte(version), 0644); err != nil {
			return false, fmt.Errorf("failed to write version file: %w", err)
		}

		return true, nil
	}

	return false, nil
}

// displaySplash renders the colored splash screen
func displaySplash(config SplashConfig) {
	// Blue and white color scheme for butterfly theme
	titleColor := color.New(color.FgBlue, color.Bold)        // Blue for main elements
	versionColor := color.New(color.FgCyan, color.Bold)      // Cyan for version
	descColor := color.New(color.FgWhite)                    // White for description
	featureColor := color.New(color.FgBlue)                  // Blue for features
	headerColor := color.New(color.FgBlue, color.Bold)       // Bold blue for headers
	accentColor := color.New(color.FgCyan, color.Bold)       // Cyan for accents
	butterflyColor := color.New(color.FgMagenta, color.Bold) // Special color for butterfly elements

	// Clear screen for better presentation
	fmt.Print("\033[2J\033[H")

	// Display banner with butterfly theme colors
	butterflyColor.Print(banner)

	// Version and title
	fmt.Print("\n")
	titleColor.Printf("  %s ", config.AppName)
	versionColor.Printf("v%s\n", config.AppVersion)
	descColor.Printf("  %s\n\n", config.Description)

	// Welcome message based on run type with butterfly theme
	if config.IsFirstRun {
		headerColor.Println("  🦋 Welcome to ztictl! Ready to transform your AWS workflow?")
		descColor.Println("  Like a butterfly effect, small commands create powerful changes.")
		fmt.Println("  Let's get you set up with everything you need.")
	} else if config.IsNewVersion {
		headerColor.Printf("  🦋 Welcome back! Your workflow just got more powerful with v%s\n", config.AppVersion)
		descColor.Println("  New features await - small updates, big improvements.")
	}

	// Feature showcase
	fmt.Println()
	headerColor.Println("  ✨ Features & Capabilities:")
	headerColor.Println("  " + strings.Repeat("═", 40))

	for _, feature := range config.Features {
		featureColor.Printf("    %s\n", feature)
	}

	// Quick start guide
	fmt.Println()
	headerColor.Println("  🚀 Quick Start Guide:")
	headerColor.Println("  " + strings.Repeat("═", 25))

	if config.IsFirstRun {
		accentColor.Println("    1. Configure your settings:")
		fmt.Println("       ztictl config init")
		fmt.Println()
		accentColor.Println("    2. Check system requirements:")
		fmt.Println("       ztictl config check")
		fmt.Println()
		accentColor.Println("    3. Authenticate with AWS SSO:")
		fmt.Println("       ztictl auth login")
		fmt.Println()
		accentColor.Println("    4. List your EC2 instances:")
		fmt.Println("       ztictl ssm list")
	} else {
		accentColor.Println("    • View help:           ztictl --help")
		accentColor.Println("    • Check configuration: ztictl config show")
		accentColor.Println("    • Login to AWS:        ztictl auth login")
		accentColor.Println("    • List instances:      ztictl ssm list")
	}

	// Footer
	fmt.Println()
	headerColor.Println("  📚 Documentation & Support:")
	headerColor.Println("  " + strings.Repeat("═", 35))
	featureColor.Println("    • GitHub: https://github.com/zsoftly/ztiaws")
	featureColor.Println("    • Help:   ztictl --help")
	featureColor.Println("    • Config: ztictl config --help")

	// Animated separator with butterfly theme
	fmt.Println()
	animateMessage("  🦋" + strings.Repeat("═", 56) + "🦋")
	fmt.Println()

	// Pause for user to read
	if config.IsFirstRun {
		butterflyColor.Print("  🦋 Press Enter to begin your transformation...")
		fmt.Scanln()
	} else {
		time.Sleep(3 * time.Second)
		descColor.Println("  🦋 Spreading wings... Starting ztictl...")
		time.Sleep(1 * time.Second)
	}
}

// animateMessage displays a message with a simple animation effect
func animateMessage(message string) {
	color := color.New(color.FgCyan)
	for i, char := range message {
		color.Printf("%c", char)
		if i%10 == 0 {
			time.Sleep(50 * time.Millisecond)
		}
	}
	fmt.Println()
}

// ShowBriefWelcome shows a minimal welcome for subsequent runs
func ShowBriefWelcome(version string) {
	butterflyColor := color.New(color.FgMagenta, color.Bold)
	titleColor := color.New(color.FgBlue, color.Bold)

	butterflyColor.Print("🦋 ")
	titleColor.Printf("ztictl v%s", version)
	fmt.Println(" - Transform your AWS workflow")
	fmt.Println("Type 'ztictl --help' for usage information.")
	fmt.Println()
}
