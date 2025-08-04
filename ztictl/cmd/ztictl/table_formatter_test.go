package main

import (
	"strings"
	"testing"
)

func TestTableFormatter(t *testing.T) {
	formatter := NewTableFormatter(2)

	// Test data with varying lengths to simulate the real scenario
	names := []string{
		"short",
		"very-long-instance-name-that-would-break-fixed-width",
		"medium-name",
	}

	instanceIDs := []string{
		"i-1234567890abcdef0",
		"i-abcdef1234567890",
		"i-1111222233334444",
	}

	ipAddresses := []string{
		"10.0.1.100",
		"192.168.1.1",
		"172.16.0.250",
	}

	states := []string{
		"running",
		"stopped",
		"pending",
	}

	// Simulate colored SSM statuses
	ssmStatuses := []string{
		"\033[32m✓ Online\033[0m",   // Green
		"\033[33m⚠ Lost\033[0m",     // Yellow
		"\033[31m✗ No Agent\033[0m", // Red
	}

	platforms := []string{
		"Linux/UNIX",
		"Windows",
		"Linux/UNIX",
	}

	// Add columns
	formatter.AddColumn("Name", names, 8)
	formatter.AddColumn("Instance ID", instanceIDs, 12)
	formatter.AddColumn("IP Address", ipAddresses, 10)
	formatter.AddColumn("State", states, 8)
	formatter.AddColumn("SSM Status", ssmStatuses, 10)
	formatter.AddColumn("Platform", platforms, 8)

	// Test header formatting
	header := formatter.FormatHeader()
	if header == "" {
		t.Error("Header should not be empty")
	}

	// Verify header contains all column names
	for _, col := range formatter.Columns {
		if !strings.Contains(header, col.Header) {
			t.Errorf("Header should contain '%s'", col.Header)
		}
	}

	// Test row formatting
	rowCount := formatter.GetRowCount()
	if rowCount != 3 {
		t.Errorf("Expected 3 rows, got %d", rowCount)
	}

	// Test each row
	for i := 0; i < rowCount; i++ {
		row := formatter.FormatRow(i)
		if row == "" {
			t.Errorf("Row %d should not be empty", i)
		}

		// Verify the row contains the expected data
		if !strings.Contains(row, names[i]) {
			t.Errorf("Row %d should contain name '%s'", i, names[i])
		}
		if !strings.Contains(row, instanceIDs[i]) {
			t.Errorf("Row %d should contain instance ID '%s'", i, instanceIDs[i])
		}
	}
}

func TestStripAnsiCodes(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "\033[32m✓ Online\033[0m",
			expected: "✓ Online",
		},
		{
			input:    "\033[33m⚠ Lost\033[0m",
			expected: "⚠ Lost",
		},
		{
			input:    "\033[31m✗ No Agent\033[0m",
			expected: "✗ No Agent",
		},
		{
			input:    "Plain text",
			expected: "Plain text",
		},
	}

	for _, tc := range testCases {
		result := stripAnsiCodes(tc.input)
		if result != tc.expected {
			t.Errorf("stripAnsiCodes(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

func TestColumnWidthCalculation(t *testing.T) {
	formatter := NewTableFormatter(1)

	// Add a column with varying widths
	values := []string{
		"short",
		"very-very-long-content-that-exceeds-header",
		"medium",
	}

	formatter.AddColumn("Header", values, 5)

	widths := formatter.calculateColumnWidths()
	if len(widths) != 1 {
		t.Errorf("Expected 1 width, got %d", len(widths))
	}

	// Width should be based on the longest content
	expectedWidth := len("very-very-long-content-that-exceeds-header")
	if widths[0] != expectedWidth {
		t.Errorf("Expected width %d, got %d", expectedWidth, widths[0])
	}
}
