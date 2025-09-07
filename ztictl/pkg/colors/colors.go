package colors

import "github.com/fatih/color"

// Standardized color definitions for ztictl
// These colors are used consistently across all SSM commands and other UI elements

var (
	// Header/Section colors - bright yellow with bold for headers and section titles
	Header = color.New(color.FgHiYellow, color.Bold)

	// Data/Results colors - bright cyan for data output like instance IDs, IP addresses, etc.
	Data = color.New(color.FgHiCyan)

	// Success message colors - bright green with bold for positive feedback
	Success = color.New(color.FgHiGreen, color.Bold)

	// Error message colors - bright red with bold for error messages
	Error = color.New(color.FgHiRed, color.Bold)

	// Warning message colors - bright yellow with bold for warnings
	Warning = color.New(color.FgHiYellow, color.Bold)
)

// Convenience functions for common color operations
func PrintHeader(format string, args ...interface{}) {
	_, _ = Header.Printf(format, args...) // #nosec G104
}

func PrintData(format string, args ...interface{}) {
	_, _ = Data.Printf(format, args...) // #nosec G104
}

func PrintSuccess(format string, args ...interface{}) {
	_, _ = Success.Printf(format, args...) // #nosec G104
}

func PrintError(format string, args ...interface{}) {
	_, _ = Error.Printf(format, args...) // #nosec G104
}

func PrintWarning(format string, args ...interface{}) {
	_, _ = Warning.Printf(format, args...) // #nosec G104
}

// Color formatting functions that return colored strings
func ColorHeader(format string, args ...interface{}) string {
	return Header.Sprintf(format, args...)
}

func ColorData(format string, args ...interface{}) string {
	return Data.Sprintf(format, args...)
}

func ColorSuccess(format string, args ...interface{}) string {
	return Success.Sprintf(format, args...)
}

func ColorError(format string, args ...interface{}) string {
	return Error.Sprintf(format, args...)
}

func ColorWarning(format string, args ...interface{}) string {
	return Warning.Sprintf(format, args...)
}
