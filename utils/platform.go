package utils

import (
	"context"
	"os/exec"
	"runtime"
	"time"
)

// IsWindows returns true when running on Windows.
func IsWindows() bool { return runtime.GOOS == "windows" }

// IsLinux returns true when running on Linux.
func IsLinux() bool { return runtime.GOOS == "linux" }

// CommandWithTimeout executes a command with timeout and returns the output.
func CommandWithTimeout(ctx context.Context, timeout time.Duration, name string, args ...string) ([]byte, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(ctx, name, args...)

	// Execute and return output
	return cmd.Output()
}

// ExecuteCommand is a convenience function that executes a command with default timeout.
// (removed) ExecuteCommand: unused helper

// GetShell returns the appropriate shell for the current platform.
// (removed) GetShell: unused helper


