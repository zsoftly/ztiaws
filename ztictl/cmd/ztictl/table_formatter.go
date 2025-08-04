package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// ColumnData represents data for a single column
type ColumnData struct {
	Header   string
	Values   []string
	MinWidth int
}

// TableFormatter handles dynamic column formatting
type TableFormatter struct {
	Columns []ColumnData
	Padding int
}

// NewTableFormatter creates a new table formatter
func NewTableFormatter(padding int) *TableFormatter {
	return &TableFormatter{
		Columns: make([]ColumnData, 0),
		Padding: padding,
	}
}

// AddColumn adds a column to the table
func (tf *TableFormatter) AddColumn(header string, values []string, minWidth int) {
	tf.Columns = append(tf.Columns, ColumnData{
		Header:   header,
		Values:   values,
		MinWidth: minWidth,
	})
}

// calculateColumnWidths determines the optimal width for each column
func (tf *TableFormatter) calculateColumnWidths() []int {
	widths := make([]int, len(tf.Columns))

	for i, col := range tf.Columns {
		// Start with minimum width or header width
		width := col.MinWidth
		headerWidth := utf8.RuneCountInString(col.Header)
		if headerWidth > width {
			width = headerWidth
		}

		// Check all values in the column
		for _, value := range col.Values {
			// Strip ANSI color codes for width calculation
			cleanValue := stripAnsiCodes(value)
			valueWidth := utf8.RuneCountInString(cleanValue)
			if valueWidth > width {
				width = valueWidth
			}
		}

		widths[i] = width
	}

	return widths
}

// stripAnsiCodes removes ANSI color codes from a string for accurate width calculation
func stripAnsiCodes(s string) string {
	// More robust ANSI escape sequence removal
	result := s

	// Remove ANSI CSI sequences (Control Sequence Introducer)
	// Pattern: ESC [ followed by parameters and a final byte
	for {
		start := strings.Index(result, "\033[")
		if start == -1 {
			break
		}

		// Find the end of the sequence (letter that ends the sequence)
		end := start + 2
		for end < len(result) {
			char := result[end]
			// ANSI sequences end with a letter (A-Z, a-z) or certain symbols
			if (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') ||
				char == 'm' || char == 'H' || char == 'J' || char == 'K' {
				end++
				break
			}
			end++
		}

		if end <= len(result) {
			result = result[:start] + result[end:]
		} else {
			break
		}
	}

	return result
}

// FormatHeader formats and returns the table header
func (tf *TableFormatter) FormatHeader() string {
	if len(tf.Columns) == 0 {
		return ""
	}

	widths := tf.calculateColumnWidths()
	var header strings.Builder
	var separator strings.Builder

	// Build header line
	for i, col := range tf.Columns {
		format := fmt.Sprintf("%%-%ds", widths[i])
		header.WriteString(fmt.Sprintf(format, col.Header))

		// Add separator line
		separator.WriteString(strings.Repeat("-", widths[i]))

		// Add padding between columns (except last)
		if i < len(tf.Columns)-1 {
			header.WriteString(strings.Repeat(" ", tf.Padding))
			separator.WriteString(strings.Repeat(" ", tf.Padding))
		}
	}

	return header.String() + "\n" + separator.String()
}

// FormatRow formats a single row of data
func (tf *TableFormatter) FormatRow(rowIndex int) string {
	if len(tf.Columns) == 0 || rowIndex < 0 {
		return ""
	}

	// Check if row index is valid for all columns
	for _, col := range tf.Columns {
		if rowIndex >= len(col.Values) {
			return ""
		}
	}

	widths := tf.calculateColumnWidths()
	var row strings.Builder

	for i, col := range tf.Columns {
		value := col.Values[rowIndex]
		cleanValue := stripAnsiCodes(value)
		valueWidth := utf8.RuneCountInString(cleanValue)

		// For the last column, don't pad (let it flow naturally)
		if i == len(tf.Columns)-1 {
			row.WriteString(value)
		} else {
			// Calculate padding needed after the colored text
			paddingNeeded := widths[i] - valueWidth
			row.WriteString(value)
			if paddingNeeded > 0 {
				row.WriteString(strings.Repeat(" ", paddingNeeded))
			}
			row.WriteString(strings.Repeat(" ", tf.Padding))
		}
	}

	return row.String()
}

// GetRowCount returns the number of rows (assumes all columns have same length)
func (tf *TableFormatter) GetRowCount() int {
	if len(tf.Columns) == 0 {
		return 0
	}
	return len(tf.Columns[0].Values)
}
