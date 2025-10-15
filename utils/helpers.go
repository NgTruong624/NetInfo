package utils

import (
	"fmt"
	"strings"
	"time"
)

// TruncateString truncates a string to the specified length and adds ellipsis if needed.
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// FormatDuration formats a duration into a human-readable string.
func FormatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%.0fns", float64(d.Nanoseconds()))
	} else if d < time.Millisecond {
		return fmt.Sprintf("%.1fμs", float64(d.Microseconds()))
	} else if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d.Milliseconds()))
	} else {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}

// FormatBytes formats bytes into a human-readable string.
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// CleanString removes extra whitespace and newlines from a string.
func CleanString(s string) string {
	// Replace multiple whitespace with single space
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	
	// Split by spaces, remove empty strings, join back
	parts := strings.Fields(s)
	return strings.Join(parts, " ")
}

// ContainsAny checks if the string contains any of the given substrings.
func ContainsAny(s string, substrings ...string) bool {
	for _, substr := range substrings {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}