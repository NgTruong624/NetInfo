package utils

import (
	"fmt"
	"strings"
	"time"
)

// ErrorType represents different types of errors
type ErrorType int

const (
	ErrorTypeNetwork ErrorType = iota
	ErrorTypeTimeout
	ErrorTypePermission
	ErrorTypeCommand
	ErrorTypeParse
	ErrorTypeValidation
	ErrorTypeUnknown
)

// NetworkError represents a network-related error with additional context
type NetworkError struct {
	Type        ErrorType
	Message     string
	OriginalErr error
	Context     map[string]interface{}
	Timestamp   time.Time
}

func (e *NetworkError) Error() string {
	if e.OriginalErr != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.OriginalErr)
	}
	return e.Message
}

// NewNetworkError creates a new network error
func NewNetworkError(errType ErrorType, message string, originalErr error) *NetworkError {
	return &NetworkError{
		Type:        errType,
		Message:     message,
		OriginalErr: originalErr,
		Context:     make(map[string]interface{}),
		Timestamp:   time.Now(),
	}
}

// AddContext adds context information to the error
func (e *NetworkError) AddContext(key string, value interface{}) {
	e.Context[key] = value
}

// IsTimeoutError checks if the error is a timeout error
func IsTimeoutError(err error) bool {
	if netErr, ok := err.(*NetworkError); ok {
		return netErr.Type == ErrorTypeTimeout
	}
	return strings.Contains(err.Error(), "timeout") || 
		   strings.Contains(err.Error(), "deadline exceeded")
}

// IsPermissionError checks if the error is a permission error
func IsPermissionError(err error) bool {
	if netErr, ok := err.(*NetworkError); ok {
		return netErr.Type == ErrorTypePermission
	}
	return strings.Contains(err.Error(), "permission denied") ||
		   strings.Contains(err.Error(), "access denied") ||
		   strings.Contains(err.Error(), "not permitted")
}

// IsNetworkError checks if the error is a network-related error
func IsNetworkError(err error) bool {
	if netErr, ok := err.(*NetworkError); ok {
		return netErr.Type == ErrorTypeNetwork
	}
	return strings.Contains(err.Error(), "network") ||
		   strings.Contains(err.Error(), "connection") ||
		   strings.Contains(err.Error(), "unreachable")
}

// WrapError wraps an error with additional context
func WrapError(err error, message string, errType ErrorType) *NetworkError {
	return NewNetworkError(errType, message, err)
}

// GetUserFriendlyMessage returns a user-friendly error message
func GetUserFriendlyMessage(err error) string {
	if netErr, ok := err.(*NetworkError); ok {
		switch netErr.Type {
		case ErrorTypeTimeout:
			return MsgTryAgain + " (Operation timed out)"
		case ErrorTypePermission:
			return ErrPermissionDenied
		case ErrorTypeNetwork:
			return ErrNetworkUnavailable
		case ErrorTypeCommand:
			return MsgFeatureUnavailable
		case ErrorTypeParse:
			return "Data parsing error - please try again"
		case ErrorTypeValidation:
			return ErrInvalidInput
		default:
			return MsgTryAgain
		}
	}
	
	// Handle common error patterns
	errStr := err.Error()
	switch {
	case IsTimeoutError(err):
		return MsgTryAgain + " (Operation timed out)"
	case IsPermissionError(err):
		return ErrPermissionDenied
	case IsNetworkError(err):
		return ErrNetworkUnavailable
	case strings.Contains(errStr, "no such host"):
		return "Host not found - check the hostname or DNS settings"
	case strings.Contains(errStr, "connection refused"):
		return "Connection refused - the service may not be running"
	case strings.Contains(errStr, "network unreachable"):
		return "Network unreachable - check your network connection"
	default:
		return MsgTryAgain
	}
}

// ShouldRetry determines if an operation should be retried
func ShouldRetry(err error) bool {
	if IsTimeoutError(err) {
		return true
	}
	if IsNetworkError(err) {
		return true
	}
	// Don't retry permission errors or validation errors
	if IsPermissionError(err) {
		return false
	}
	return false
}

// GetRetryDelay returns the appropriate delay for retrying
func GetRetryDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return RetryDelay
	}
	// Exponential backoff with jitter
	delay := time.Duration(attempt) * RetryDelay
	if delay > 10*time.Second {
		delay = 10 * time.Second
	}
	return delay
}
