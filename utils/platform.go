package utils

import "runtime"

// IsWindows returns true when running on Windows.
func IsWindows() bool { return runtime.GOOS == "windows" }

// IsLinux returns true when running on Linux.
func IsLinux() bool { return runtime.GOOS == "linux" }


