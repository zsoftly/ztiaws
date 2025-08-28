package colors

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestColorVariables(t *testing.T) {
	// Force colors enabled for testing
	originalNoColor := color.NoColor
	defer func() { color.NoColor = originalNoColor }()
	color.NoColor = false

	tests := []struct {
		name       string
		colorVar   *color.Color
		shouldBold bool
	}{
		{"Header", Header, true},
		{"Data", Data, false},
		{"Success", Success, true},
		{"Error", Error, true},
		{"Warning", Warning, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.colorVar == nil {
				t.Errorf("%s color variable should not be nil", tt.name)
			}

			// Test that the color variable produces output
			result := tt.colorVar.Sprint("test")
			if result == "" {
				t.Errorf("%s color should produce output", tt.name)
			}

			// For colors that should be bold, ensure they contain ANSI bold codes when colors are enabled
			if tt.shouldBold {
				// Bold ANSI code contains ";1m"
				if !strings.Contains(result, ";1m") {
					t.Errorf("%s should be bold but doesn't contain bold ANSI codes, got: %q", tt.name, result)
				}
			}
		})
	}
}

func TestPrintFunctions(t *testing.T) {
	// Force colors enabled and capture stdout for testing print functions
	originalNoColor := color.NoColor
	originalStdout := os.Stdout
	defer func() {
		color.NoColor = originalNoColor
		os.Stdout = originalStdout
	}()
	color.NoColor = false

	tests := []struct {
		name      string
		printFunc func(string, ...interface{})
		format    string
		args      []interface{}
		expected  string
	}{
		{"PrintHeader", PrintHeader, "Test %s", []interface{}{"header"}, "Test header"},
		{"PrintData", PrintData, "Test %s", []interface{}{"data"}, "Test data"},
		{"PrintSuccess", PrintSuccess, "Test %s", []interface{}{"success"}, "Test success"},
		{"PrintError", PrintError, "Test %s", []interface{}{"error"}, "Test error"},
		{"PrintWarning", PrintWarning, "Test %s", []interface{}{"warning"}, "Test warning"},
		{"PrintHeader no args", PrintHeader, "Simple header", nil, "Simple header"},
		{"PrintData no args", PrintData, "Simple data", nil, "Simple data"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture output
			var buf bytes.Buffer

			// Temporarily redirect color output to our buffer
			color.Output = &buf
			defer func() {
				color.Output = os.Stdout
			}()

			// Call the print function
			if tt.args != nil {
				tt.printFunc(tt.format, tt.args...)
			} else {
				tt.printFunc(tt.format)
			}

			output := buf.String()

			// Check that the expected text is present (ignoring ANSI codes)
			if !strings.Contains(output, tt.expected) {
				t.Errorf("%s output should contain %q, got %q", tt.name, tt.expected, output)
			}
		})
	}
}

func TestColorFormattingFunctions(t *testing.T) {
	// Force colors enabled for testing
	originalNoColor := color.NoColor
	defer func() { color.NoColor = originalNoColor }()
	color.NoColor = false

	tests := []struct {
		name           string
		colorFunc      func(string, ...interface{}) string
		format         string
		args           []interface{}
		expected       string
		shouldHaveANSI bool
	}{
		{"ColorHeader", ColorHeader, "Test %s", []interface{}{"header"}, "Test header", true},
		{"ColorData", ColorData, "Test %s", []interface{}{"data"}, "Test data", true},
		{"ColorSuccess", ColorSuccess, "Test %s", []interface{}{"success"}, "Test success", true},
		{"ColorError", ColorError, "Test %s", []interface{}{"error"}, "Test error", true},
		{"ColorWarning", ColorWarning, "Test %s", []interface{}{"warning"}, "Test warning", true},
		{"ColorHeader no args", ColorHeader, "Simple header", nil, "Simple header", true},
		{"ColorData no args", ColorData, "Simple data", nil, "Simple data", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			if tt.args != nil {
				result = tt.colorFunc(tt.format, tt.args...)
			} else {
				result = tt.colorFunc(tt.format)
			}

			// Check that the expected text is present
			if !strings.Contains(result, tt.expected) {
				t.Errorf("%s should contain %q, got %q", tt.name, tt.expected, result)
			}

			// Check for ANSI escape codes if expected
			if tt.shouldHaveANSI {
				if !strings.Contains(result, "\033[") && !strings.Contains(result, "\x1b[") {
					t.Errorf("%s should contain ANSI escape codes, got %q", tt.name, result)
				}
			}
		})
	}
}

func TestColorFormattingEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		colorFunc func(string, ...interface{}) string
		format    string
		args      []interface{}
	}{
		{"Empty format", ColorData, "", nil},
		{"Format with no placeholders", ColorSuccess, "No placeholders", []interface{}{"unused"}},
		{"Multiple placeholders", ColorError, "Error %d: %s occurred at %s", []interface{}{404, "Not Found", "endpoint"}},
		{"Special characters", ColorWarning, "Warning: Special chars !@#$%^&*()", nil},
		{"Unicode characters", ColorHeader, "Unicode: 你好 世界 🌍", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			if tt.args != nil {
				result = tt.colorFunc(tt.format, tt.args...)
			} else {
				result = tt.colorFunc(tt.format)
			}

			// Should not panic and should return a string
			if result == "" && tt.format != "" {
				t.Errorf("%s should return non-empty string for non-empty format", tt.name)
			}
		})
	}
}

func TestColorDisabling(t *testing.T) {
	// Test behavior when colors are disabled
	originalNoColor := color.NoColor
	defer func() { color.NoColor = originalNoColor }()

	// Disable colors
	color.NoColor = true

	tests := []struct {
		name      string
		colorFunc func(string, ...interface{}) string
		format    string
		expected  string
	}{
		{"ColorHeader disabled", ColorHeader, "Header text", "Header text"},
		{"ColorData disabled", ColorData, "Data text", "Data text"},
		{"ColorSuccess disabled", ColorSuccess, "Success text", "Success text"},
		{"ColorError disabled", ColorError, "Error text", "Error text"},
		{"ColorWarning disabled", ColorWarning, "Warning text", "Warning text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.colorFunc(tt.format)

			// Should contain the expected text
			if !strings.Contains(result, tt.expected) {
				t.Errorf("%s should contain %q, got %q", tt.name, tt.expected, result)
			}

			// Should NOT contain ANSI escape codes when disabled
			if strings.Contains(result, "\033[") || strings.Contains(result, "\x1b[") {
				t.Errorf("%s should not contain ANSI codes when colors disabled, got %q", tt.name, result)
			}
		})
	}
}

func TestPrintFunctionsWithDisabledColors(t *testing.T) {
	// Test print functions with colors disabled
	originalNoColor := color.NoColor
	defer func() {
		color.NoColor = originalNoColor
	}()

	// Disable colors
	color.NoColor = true

	tests := []struct {
		name      string
		printFunc func(string, ...interface{})
		text      string
	}{
		{"PrintHeader disabled", PrintHeader, "Header without colors"},
		{"PrintSuccess disabled", PrintSuccess, "Success without colors"},
		{"PrintError disabled", PrintError, "Error without colors"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture output
			var buf bytes.Buffer

			// Temporarily redirect color output to our buffer
			color.Output = &buf
			defer func() {
				color.Output = os.Stdout
			}()

			tt.printFunc(tt.text)
			output := buf.String()

			// Should contain the expected text
			if !strings.Contains(output, tt.text) {
				t.Errorf("%s should contain %q, got %q", tt.name, tt.text, output)
			}

			// Should NOT contain ANSI codes when disabled
			if strings.Contains(output, "\033[") || strings.Contains(output, "\x1b[") {
				t.Errorf("%s should not contain ANSI codes when disabled, got %q", tt.name, output)
			}
		})
	}
}

func TestColorConsistency(t *testing.T) {
	// Test that Print and Color functions produce consistent text content
	tests := []struct {
		name      string
		printFunc func(string, ...interface{})
		colorFunc func(string, ...interface{}) string
		format    string
		args      []interface{}
	}{
		{"Header consistency", PrintHeader, ColorHeader, "Test %s", []interface{}{"header"}},
		{"Data consistency", PrintData, ColorData, "Test %s", []interface{}{"data"}},
		{"Success consistency", PrintSuccess, ColorSuccess, "Test %s", []interface{}{"success"}},
		{"Error consistency", PrintError, ColorError, "Test %s", []interface{}{"error"}},
		{"Warning consistency", PrintWarning, ColorWarning, "Test %s", []interface{}{"warning"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get the colored string
			var colorResult string
			if tt.args != nil {
				colorResult = tt.colorFunc(tt.format, tt.args...)
			} else {
				colorResult = tt.colorFunc(tt.format)
			}

			// Capture print output
			var buf bytes.Buffer
			color.Output = &buf
			defer func() {
				color.Output = os.Stdout
			}()

			if tt.args != nil {
				tt.printFunc(tt.format, tt.args...)
			} else {
				tt.printFunc(tt.format)
			}

			printResult := buf.String()

			// The content should be the same (ignoring potential differences in ANSI handling)
			// Extract plain text by removing ANSI codes
			stripANSI := func(s string) string {
				// Simple ANSI stripping for test purposes
				result := s
				for strings.Contains(result, "\033[") {
					start := strings.Index(result, "\033[")
					end := strings.Index(result[start:], "m")
					if end == -1 {
						break
					}
					result = result[:start] + result[start+end+1:]
				}
				return result
			}

			colorStripped := stripANSI(colorResult)
			printStripped := stripANSI(printResult)

			if colorStripped != printStripped {
				t.Errorf("%s: Print and Color functions should produce same content. Print: %q, Color: %q", tt.name, printStripped, colorStripped)
			}
		})
	}
}

func TestFormatStringValidation(t *testing.T) {
	// Test various format string scenarios
	tests := []struct {
		name        string
		colorFunc   func(string, ...interface{}) string
		format      string
		args        []interface{}
		shouldPanic bool
	}{
		{"Valid format", ColorData, "Hello %s", []interface{}{"world"}, false},
		{"Too few args", ColorData, "Hello %s %s", []interface{}{"world"}, false},        // Go fmt handles this gracefully
		{"Too many args", ColorData, "Hello %s", []interface{}{"world", "extra"}, false}, // Go fmt handles this gracefully
		{"Invalid format verb", ColorData, "Hello %z", []interface{}{"world"}, false},    // Go fmt handles this gracefully
		{"No format placeholders", ColorData, "Hello world", []interface{}{"unused"}, false},
		{"Multiple same placeholder", ColorData, "Hello %s and %s", []interface{}{"world", "universe"}, false},
		{"Mixed placeholders", ColorData, "Number %d, String %s, Float %.2f", []interface{}{42, "test", 3.14159}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("%s should not panic, but panicked with: %v", tt.name, r)
					}
				}
			}()

			result := tt.colorFunc(tt.format, tt.args...)

			if tt.shouldPanic {
				t.Errorf("%s should have panicked but didn't", tt.name)
			}

			// Basic validation that we got some result
			if result == "" && tt.format != "" {
				t.Errorf("%s should return non-empty result for non-empty format", tt.name)
			}
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	// Test that color functions are safe for concurrent use
	done := make(chan bool)

	// Start multiple goroutines using color functions
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				ColorSuccess("Goroutine %d iteration %d", id, j)
				ColorError("Error in goroutine %d iteration %d", id, j)
				ColorData("Data from goroutine %d iteration %d", id, j)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we reach here without deadlocks or race conditions, test passes
}

func BenchmarkColorFunctions(b *testing.B) {
	benchmarks := []struct {
		name      string
		colorFunc func(string, ...interface{}) string
	}{
		{"ColorHeader", ColorHeader},
		{"ColorData", ColorData},
		{"ColorSuccess", ColorSuccess},
		{"ColorError", ColorError},
		{"ColorWarning", ColorWarning},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = bm.colorFunc("Benchmark test %d", i)
			}
		})
	}
}

func BenchmarkPrintFunctions(b *testing.B) {
	// Redirect stdout to discard for benchmarking
	originalStdout := os.Stdout
	os.Stdout, _ = os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	defer func() { os.Stdout = originalStdout }()

	benchmarks := []struct {
		name      string
		printFunc func(string, ...interface{})
	}{
		{"PrintHeader", PrintHeader},
		{"PrintData", PrintData},
		{"PrintSuccess", PrintSuccess},
		{"PrintError", PrintError},
		{"PrintWarning", PrintWarning},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				bm.printFunc("Benchmark test %d", i)
			}
		})
	}
}
