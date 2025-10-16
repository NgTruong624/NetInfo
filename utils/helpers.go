package utils

import (
	"fmt"
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
// (removed) FormatBytes: unused helper

// CleanString removes extra whitespace and newlines from a string.
// (removed) CleanString: unused helper

// ContainsAny checks if the string contains any of the given substrings.
// (removed) ContainsAny: unused helper