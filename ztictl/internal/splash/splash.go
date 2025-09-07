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
	// Clean ASCII art banner
	banner = `
               .    o8o                .   oooo  
             .o8    ` + "`" + `"'              .o8   ` + "`" + `888  
  oooooooo .o888oo oooo   .ooooo.  .o888oo  888  
 d'""7d8P    888   ` + "`" + `888  d88' ` + "`" + `"Y8   888    888  
   .d8P'     888    888  888         888    888  
 .d8P'  .P   888 .  888  888   .o8   888 .  888  
d8888888P    "888" o888o ` + "`" + `Y8bod8P'   "888" o888o 
                                                 
                          Z S o f t l y
                 AWS SSO & Systems Manager CLI
                  Small commands, powerful results
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
				"⚡ EC2 Power Management - Start/Stop/Reboot instances individually or in bulk",
				"🏷️  Advanced Tag-Based Operations - Target multiple instances with flexible filtering",
				"🚀 Parallel Execution Engine - Process multiple instances concurrently for speed",
				"🔒 Enhanced Security - Command injection protection and input validation",
				"🔐 AWS SSO Authentication with interactive selection",
				"📁 File transfer through SSM with intelligent S3 routing for large files",
				"🌐 Port forwarding and remote command execution via SSM",
				"🌍 Multi-region support with comprehensive logging",
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
	// High-intensity color scheme for better visibility on dark terminals
	titleColor := color.New(color.FgHiWhite, color.Bold)    // Bright white for main title
	versionColor := color.New(color.FgHiGreen, color.Bold)  // Bright green for version
	descColor := color.New(color.FgWhite)                   // Normal white for description
	featureColor := color.New(color.FgHiCyan)               // Bright cyan for feature list
	headerColor := color.New(color.FgHiYellow, color.Bold)  // Bright yellow for section headers
	accentColor := color.New(color.FgHiMagenta, color.Bold) // Bright magenta for accents or links
	butterflyColor := color.New(color.FgHiBlue, color.Bold) // Bright blue for butterfly elements

	// Clear screen for better presentation
	fmt.Print("\033[2J\033[H")

	// Display banner with butterfly theme colors
	butterflyColor.Print(banner)

	// Version and title
	fmt.Print("\n")
	titleColor.Printf("  %s ", config.AppName)
	versionColor.Printf("v%s\n", config.AppVersion)
	descColor.Printf("  %s\n\n", config.Description)

	// Welcome message
	if config.IsFirstRun {
		headerColor.Println("  🎉 Welcome to ztictl!")
		descColor.Println("  Small commands, powerful AWS transformations.")
		fmt.Println("  Let's get you set up with everything you need.")
	} else if config.IsNewVersion {
		headerColor.Printf("  ✨ ztictl v%s is ready!\n", config.AppVersion)
		descColor.Println("  Small updates, big improvements.")
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

	// Animated separator
	fmt.Println()
	animateMessage("  🎯" + strings.Repeat("═", 56) + "🎯")
	fmt.Println()

	// Pause for user to read
	if config.IsFirstRun {
		headerColor.Print("  🚀 Press Enter to continue...")
		fmt.Scanln()
	} else {
		time.Sleep(3 * time.Second)
		descColor.Println("  🚀 Starting ztictl...")
		time.Sleep(1 * time.Second)
	}
}

// animateMessage displays a message with a simple animation effect
func animateMessage(message string) {
	color := color.New(color.FgHiCyan)
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
	accentColor := color.New(color.FgHiBlue, color.Bold)
	titleColor := color.New(color.FgHiWhite, color.Bold)

	accentColor.Print("🎯 ")
	titleColor.Printf("ztictl v%s", version)
	fmt.Println(" - Transform your AWS workflow")
	fmt.Println("Type 'ztictl --help' for usage information.")
	fmt.Println()
}
