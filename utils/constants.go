package utils

import "time"

// Timeout constants for various operations
const (
	// Network operation timeouts
	NetworkTimeout     = 10 * time.Second
	DNSQueryTimeout    = 5 * time.Second
	PingTimeout        = 10 * time.Second
	QuickPingTimeout   = 5 * time.Second
	
	// Process operation timeouts
	ProcessListTimeout = 15 * time.Second
	ConnectionTimeout  = 15 * time.Second
	
	// Command execution timeouts
	CommandTimeout     = 30 * time.Second
	PowerShellTimeout  = 15 * time.Second
	LinuxCommandTimeout = 10 * time.Second
	
	// HTTP operation timeouts
	HTTPTimeout        = 5 * time.Second
	PublicIPTimout     = 10 * time.Second
	
	// Retry configuration
	MaxRetries         = 3
	RetryDelay         = 1 * time.Second
)

// Error messages
const (
	ErrNetworkInterfaces = "Failed to retrieve network interfaces"
	ErrIPInformation     = "Failed to retrieve IP information"
	ErrDNSInformation    = "Failed to retrieve DNS information"
	ErrGatewayInfo       = "Failed to retrieve gateway information"
	ErrRoutingTable      = "Failed to retrieve routing table"
	ErrConnections       = "Failed to retrieve network connections"
	ErrPingTest          = "Failed to execute ping test"
	ErrPublicIP          = "Failed to retrieve public IP address"
	ErrCommandTimeout    = "Command execution timed out"
	ErrCommandFailed     = "Command execution failed"
	ErrPermissionDenied  = "Permission denied - some features may require elevated privileges"
	ErrNetworkUnavailable = "Network is not available"
	ErrInvalidInput      = "Invalid input provided"
)

// User-friendly error messages
const (
	MsgNoInterfaces      = "No network interfaces found"
	MsgNoIPs             = "No IP addresses found"
	MsgNoDNSServers      = "No DNS servers configured"
	MsgNoGateway         = "No default gateway found"
	MsgNoRoutes          = "No routing information available"
	MsgNoConnections     = "No active network connections found"
	MsgPingFailed        = "Ping test failed - check network connectivity"
	MsgPublicIPFailed    = "Could not determine public IP - check internet connection"
	MsgFeatureUnavailable = "This feature is not available on your system"
	MsgTryAgain          = "Please try again or check your network configuration"
)

// Status messages
const (
	MsgGatheringInfo     = "Gathering network information..."
	MsgProcessingData    = "Processing data..."
	MsgConnecting        = "Connecting to network services..."
	MsgCheckingConnectivity = "Checking network connectivity..."
	MsgRetrying          = "Retrying operation..."
	MsgOperationComplete = "Operation completed successfully"
)
