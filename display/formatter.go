package display

import (
	"fmt"
	"os"
	"strings"

	"netinfo/utils"

	"github.com/olekukonko/tablewriter"
)

// TableConfig holds configuration for table formatting
type TableConfig struct {
	Headers    []string
	Data       [][]string
	Title      string
	Border     bool
	CenterText bool
	AutoWrap   bool
	MaxWidth   int
}

// NewTableConfig creates a new table configuration with defaults
func NewTableConfig() *TableConfig {
	return &TableConfig{
		Headers:    []string{},
		Data:       [][]string{},
		Title:      "",
		Border:     true,
		CenterText: false,
		AutoWrap:   true,
		MaxWidth:   80,
	}
}

// PrintTable prints a formatted table to stdout
func PrintTable(config *TableConfig) {
	if len(config.Data) == 0 {
		PrintWarning("No data to display")
		return
	}

	// Print title if provided
	if config.Title != "" {
		fmt.Println()
		fmt.Println(Title(config.Title))
		PrintSeparator()
	}

	// Create table
	table := tablewriter.NewWriter(os.Stdout)
	
	// Set headers
	if len(config.Headers) > 0 {
		table.Header(config.Headers)
	}
	
	// Process and add data with truncation
	for _, row := range config.Data {
		processedRow := make([]string, len(row))
		for i, cell := range row {
			if config.AutoWrap && len(cell) > config.MaxWidth {
				processedRow[i] = utils.TruncateString(cell, config.MaxWidth)
			} else {
				processedRow[i] = cell
			}
		}
		table.Append(processedRow)
	}
	
	// Render table
	table.Render()
	fmt.Println()
}

// PrintKeyValue prints key-value pairs in a formatted way
func PrintKeyValue(pairs map[string]string, title string) {
	if len(pairs) == 0 {
		PrintWarning("No data to display")
		return
	}

	if title != "" {
		fmt.Println()
		fmt.Println(Title(title))
		PrintSeparator()
	}

	// Find the longest key for alignment
	maxKeyLen := 0
	for key := range pairs {
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}
	}

	// Print key-value pairs
	for key, value := range pairs {
		padding := strings.Repeat(" ", maxKeyLen-len(key))
		fmt.Printf("%s%s: %s\n", 
			Primary(key), 
			padding, 
			Secondary(value))
	}
	fmt.Println()
}

// PrintList prints a list of items with bullet points
func PrintList(items []string, title string) {
	if len(items) == 0 {
		PrintWarning("No items to display")
		return
	}

	if title != "" {
		fmt.Println()
		fmt.Println(Title(title))
		PrintSeparator()
	}

	for _, item := range items {
		fmt.Printf("  • %s\n", Secondary(item))
	}
	fmt.Println()
}

// PrintStatus prints a status with colored indicator
func PrintStatus(status string, success bool) {
	if success {
		fmt.Printf("%s %s\n", Success("✓"), status)
	} else {
		fmt.Printf("%s %s\n", Error("✗"), status)
	}
}

// PrintProgress prints a simple progress indicator
func PrintProgress(current, total int, description string) {
	percentage := float64(current) / float64(total) * 100
	fmt.Printf("\r%s %s (%d/%d) %.1f%%", 
		Info("⏳"), 
		description, 
		current, 
		total, 
		percentage)
	
	if current == total {
		fmt.Println()
	}
}

// PrintJSON prints JSON data in a formatted way (simple implementation)
func PrintJSON(data string, title string) {
	if title != "" {
		fmt.Println()
		fmt.Println(Title(title))
		PrintSeparator()
	}
	
	// Simple JSON formatting with indentation
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			fmt.Printf("  %s\n", Secondary(trimmed))
		}
	}
	fmt.Println()
}