package splash

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"ztictl/pkg/security"
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

	// Validate file path to prevent directory traversal (G304 fix)
	if err := security.ValidateFilePath(versionFile, homeDir); err != nil {
		return false, fmt.Errorf("invalid version file path: %w", err)
	}

	// Check if this is first run or new version
	isFirstRun := false
	isNewVersion := false

	if _, err := os.Stat(versionFile); os.IsNotExist(err) {
		isFirstRun = true
	} else {
		// Read the last version (path already validated above)
		lastVersionBytes, err := os.ReadFile(versionFile) // #nosec G304
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
		if err := os.WriteFile(versionFile, []byte(version), 0600); err != nil {
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
	_, _ = butterflyColor.Print(banner) // #nosec G104

	// Version and title
	fmt.Print("\n")
	_, _ = titleColor.Printf("  %s ", config.AppName)       // #nosec G104
	_, _ = versionColor.Printf("v%s\n", config.AppVersion)  // #nosec G104
	_, _ = descColor.Printf("  %s\n\n", config.Description) // #nosec G104

	// Welcome message
	if config.IsFirstRun {
		_, _ = headerColor.Println("  🎉 Welcome to ztictl!")                        // #nosec G104
		_, _ = descColor.Println("  Small commands, powerful AWS transformations.") // #nosec G104
		fmt.Println("  Let's get you set up with everything you need.")
	} else if config.IsNewVersion {
		_, _ = headerColor.Printf("  ✨ ztictl v%s is ready!\n", config.AppVersion) // #nosec G104
		_, _ = descColor.Println("  Small updates, big improvements.")             // #nosec G104
	}

	// Feature showcase
	fmt.Println()
	_, _ = headerColor.Println("  ✨ Features & Capabilities:") // #nosec G104
	_, _ = headerColor.Println("  " + strings.Repeat("═", 40)) // #nosec G104

	for _, feature := range config.Features {
		_, _ = featureColor.Printf("    %s\n", feature) // #nosec G104
	}

	// Quick start guide
	fmt.Println()
	_, _ = headerColor.Println("  🚀 Quick Start Guide:")       // #nosec G104
	_, _ = headerColor.Println("  " + strings.Repeat("═", 25)) // #nosec G104

	if config.IsFirstRun {
		_, _ = accentColor.Println("    1. Configure your settings:") // #nosec G104
		fmt.Println("       ztictl config init")
		fmt.Println()
		_, _ = accentColor.Println("    2. Check system requirements:") // #nosec G104
		fmt.Println("       ztictl config check")
		fmt.Println()
		_, _ = accentColor.Println("    3. Authenticate with AWS SSO:") // #nosec G104
		fmt.Println("       ztictl auth login")
		fmt.Println()
		_, _ = accentColor.Println("    4. List your EC2 instances:") // #nosec G104
		fmt.Println("       ztictl ssm list")
	} else {
		_, _ = accentColor.Println("    • View help:           ztictl --help")      // #nosec G104
		_, _ = accentColor.Println("    • Check configuration: ztictl config show") // #nosec G104
		_, _ = accentColor.Println("    • Login to AWS:        ztictl auth login")  // #nosec G104
		_, _ = accentColor.Println("    • List instances:      ztictl ssm list")    // #nosec G104
	}

	// Footer
	fmt.Println()
	_, _ = headerColor.Println("  📚 Documentation & Support:")                     // #nosec G104
	_, _ = headerColor.Println("  " + strings.Repeat("═", 35))                     // #nosec G104
	_, _ = featureColor.Println("    • GitHub: https://github.com/zsoftly/ztiaws") // #nosec G104
	_, _ = featureColor.Println("    • Help:   ztictl --help")                     // #nosec G104
	_, _ = featureColor.Println("    • Config: ztictl config --help")              // #nosec G104

	// Animated separator
	fmt.Println()
	animateMessage("  🎯" + strings.Repeat("═", 56) + "🎯")
	fmt.Println()

	// Pause for user to read
	if config.IsFirstRun {
		_, _ = headerColor.Print("  🚀 Press Enter to continue...") // #nosec G104
		_, _ = fmt.Scanln()                                        // Ignore error as user input is not critical
	} else {
		time.Sleep(3 * time.Second)
		_, _ = descColor.Println("  🚀 Starting ztictl...") // #nosec G104
		time.Sleep(1 * time.Second)
	}
}

// animateMessage displays a message with a simple animation effect
func animateMessage(message string) {
	color := color.New(color.FgHiCyan)
	for i, char := range message {
		_, _ = color.Printf("%c", char) // #nosec G104
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

	_, _ = accentColor.Print("🎯 ")                  // #nosec G104
	_, _ = titleColor.Printf("ztictl v%s", version) // #nosec G104
	fmt.Println(" - Transform your AWS workflow")
	fmt.Println("Type 'ztictl --help' for usage information.")
	fmt.Println()
}
