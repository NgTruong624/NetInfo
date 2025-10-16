package network

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"netinfo/display"
	"netinfo/utils"
)

// PingResult holds ping test results
type PingResult struct {
	Host        string        `json:"host"`
	Success     bool          `json:"success"`
	PacketLoss  float64       `json:"packet_loss"`
	MinRTT      time.Duration `json:"min_rtt"`
	MaxRTT      time.Duration `json:"max_rtt"`
	AvgRTT      time.Duration `json:"avg_rtt"`
	PacketsSent int           `json:"packets_sent"`
	PacketsRecv int           `json:"packets_recv"`
	RawOutput   string        `json:"raw_output"`
	Error       string        `json:"error,omitempty"`
}

// PingConfig holds ping configuration
type PingConfig struct {
	Host    string
	Count   int
	Timeout time.Duration
	Size    int // packet size in bytes
}

// DefaultPingConfig returns default ping configuration
func DefaultPingConfig(host string) *PingConfig {
	return &PingConfig{
		Host:    host,
		Count:   4,
		Timeout: 10 * time.Second,
		Size:    32,
	}
}

// ShowPingTest displays ping test interface
func ShowPingTest() error {
	display.PrintInfo("Ping Test Utility")
	display.PrintSeparator()
	
	// Get host from user input
	host, err := display.ShowInput("Enter host to ping", "google.com")
	if err != nil {
		display.PrintError(fmt.Sprintf("Input error: %v", err))
		return err
	}
	
	// Get count from user input
	countStr, err := display.ShowInput("Number of packets", "4")
	if err != nil {
		display.PrintError(fmt.Sprintf("Input error: %v", err))
		return err
	}
	
	count, err := strconv.Atoi(countStr)
	if err != nil || count <= 0 || count > 100 {
		display.PrintWarning("Invalid count, using default: 4")
		count = 4
	}
	
	// Create ping config
	config := &PingConfig{
		Host:    host,
		Count:   count,
		Timeout: 10 * time.Second,
		Size:    32,
	}
	
	// Execute ping test
	display.PrintInfo(fmt.Sprintf("Pinging %s with %d packets...", host, count))
	display.PrintSeparator()
	
	result, err := PingHost(config)
	if err != nil {
		display.PrintError(fmt.Sprintf("Ping failed: %v", err))
		return err
	}
	
	// Display results
	displayPingResult(result)
	
	return nil
}

// PingHost executes ping test on the specified host
func PingHost(config *PingConfig) (*PingResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()
	
	var cmd *exec.Cmd
	
	if utils.IsWindows() {
		// Windows: ping -n count host
		cmd = exec.CommandContext(ctx, "ping", "-n", strconv.Itoa(config.Count), config.Host)
	} else {
		// Linux: ping -c count -W timeout host
		timeoutSec := int(config.Timeout.Seconds())
		cmd = exec.CommandContext(ctx, "ping", "-c", strconv.Itoa(config.Count), "-W", strconv.Itoa(timeoutSec), config.Host)
	}
	
	output, err := cmd.Output()
	if err != nil {
		return &PingResult{
			Host:    config.Host,
			Success: false,
			Error:   err.Error(),
			RawOutput: string(output),
		}, nil
	}
	
	// Parse output based on platform
	result := &PingResult{
		Host:      config.Host,
		Success:   true,
		RawOutput: string(output),
	}
	
	if utils.IsWindows() {
		err = parseWindowsPingOutput(result, string(output))
	} else {
		err = parseLinuxPingOutput(result, string(output))
	}
	
	if err != nil {
		result.Error = err.Error()
		result.Success = false
	}
	
	return result, nil
}

// parseWindowsPingOutput parses Windows ping output
func parseWindowsPingOutput(result *PingResult, output string) error {
	lines := strings.Split(output, "\n")
	
	// Look for packet loss line
	packetLossRegex := regexp.MustCompile(`\((\d+)% loss\)`)
	// Look for RTT statistics line
	rttRegex := regexp.MustCompile(`Minimum = (\d+)ms, Maximum = (\d+)ms, Average = (\d+)ms`)
	
	var packetLoss float64
	var minRTT, maxRTT, avgRTT time.Duration
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Parse packet loss
		if matches := packetLossRegex.FindStringSubmatch(line); len(matches) > 1 {
			if loss, err := strconv.ParseFloat(matches[1], 64); err == nil {
				packetLoss = loss
			}
		}
		
		// Parse RTT statistics
		if matches := rttRegex.FindStringSubmatch(line); len(matches) > 3 {
			if min, err := strconv.Atoi(matches[1]); err == nil {
				minRTT = time.Duration(min) * time.Millisecond
			}
			if max, err := strconv.Atoi(matches[2]); err == nil {
				maxRTT = time.Duration(max) * time.Millisecond
			}
			if avg, err := strconv.Atoi(matches[3]); err == nil {
				avgRTT = time.Duration(avg) * time.Millisecond
			}
		}
	}
	
	// Calculate packet counts
	packetsSent := result.PacketsSent
	if packetsSent == 0 {
		packetsSent = 4 // default
	}
	
	packetsRecv := int(float64(packetsSent) * (100 - packetLoss) / 100)
	
	result.PacketLoss = packetLoss
	result.MinRTT = minRTT
	result.MaxRTT = maxRTT
	result.AvgRTT = avgRTT
	result.PacketsSent = packetsSent
	result.PacketsRecv = packetsRecv
	
	return nil
}

// parseLinuxPingOutput parses Linux ping output
func parseLinuxPingOutput(result *PingResult, output string) error {
	lines := strings.Split(output, "\n")
	
	// Look for packet loss line
	packetLossRegex := regexp.MustCompile(`(\d+)% packet loss`)
	// Look for RTT statistics line
	rttRegex := regexp.MustCompile(`rtt min/avg/max/mdev = ([\d.]+)/([\d.]+)/([\d.]+)/([\d.]+) ms`)
	// Look for packets transmitted/received line
	packetsRegex := regexp.MustCompile(`(\d+) packets transmitted, (\d+) received`)
	
	var packetLoss float64
	var minRTT, maxRTT, avgRTT time.Duration
	var packetsSent, packetsRecv int
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Parse packet loss
		if matches := packetLossRegex.FindStringSubmatch(line); len(matches) > 1 {
			if loss, err := strconv.ParseFloat(matches[1], 64); err == nil {
				packetLoss = loss
			}
		}
		
		// Parse RTT statistics
		if matches := rttRegex.FindStringSubmatch(line); len(matches) > 4 {
			if min, err := strconv.ParseFloat(matches[1], 64); err == nil {
				minRTT = time.Duration(min*1000) * time.Microsecond
			}
			if avg, err := strconv.ParseFloat(matches[2], 64); err == nil {
				avgRTT = time.Duration(avg*1000) * time.Microsecond
			}
			if max, err := strconv.ParseFloat(matches[3], 64); err == nil {
				maxRTT = time.Duration(max*1000) * time.Microsecond
			}
		}
		
		// Parse packet counts
		if matches := packetsRegex.FindStringSubmatch(line); len(matches) > 2 {
			if sent, err := strconv.Atoi(matches[1]); err == nil {
				packetsSent = sent
			}
			if recv, err := strconv.Atoi(matches[2]); err == nil {
				packetsRecv = recv
			}
		}
	}
	
	result.PacketLoss = packetLoss
	result.MinRTT = minRTT
	result.MaxRTT = maxRTT
	result.AvgRTT = avgRTT
	result.PacketsSent = packetsSent
	result.PacketsRecv = packetsRecv
	
	return nil
}

// displayPingResult displays ping test results in a formatted table
func displayPingResult(result *PingResult) {
	if !result.Success {
		display.PrintError(fmt.Sprintf("Ping to %s failed", result.Host))
		if result.Error != "" {
			display.PrintError(fmt.Sprintf("Error: %s", result.Error))
		}
		display.PrintJSON(result.RawOutput, "Raw Output")
		return
	}
	
	display.PrintSuccess(fmt.Sprintf("Ping to %s successful", result.Host))
	
	// Create summary table
	summaryData := map[string]string{
		"Host":         result.Host,
		"Packets Sent": fmt.Sprintf("%d", result.PacketsSent),
		"Packets Received": fmt.Sprintf("%d", result.PacketsRecv),
		"Packet Loss":  fmt.Sprintf("%.1f%%", result.PacketLoss),
		"Min RTT":      utils.FormatDuration(result.MinRTT),
		"Max RTT":      utils.FormatDuration(result.MaxRTT),
		"Avg RTT":      utils.FormatDuration(result.AvgRTT),
	}
	
	display.PrintKeyValue(summaryData, "Ping Statistics")
	
	// Show raw output for debugging
	if result.RawOutput != "" {
		display.PrintJSON(result.RawOutput, "Raw Ping Output")
	}
	
	// Performance assessment
	display.PrintSeparator()
	display.PrintInfo("Performance Assessment:")
	
	if result.PacketLoss == 0 {
		display.PrintSuccess("✓ No packet loss - Excellent connectivity")
	} else if result.PacketLoss < 5 {
		display.PrintWarning(fmt.Sprintf("⚠ %.1f%% packet loss - Good connectivity", result.PacketLoss))
	} else {
		display.PrintError(fmt.Sprintf("✗ %.1f%% packet loss - Poor connectivity", result.PacketLoss))
	}
	
	if result.AvgRTT < 50*time.Millisecond {
		display.PrintSuccess("✓ Low latency - Excellent response time")
	} else if result.AvgRTT < 200*time.Millisecond {
		display.PrintInfo("ℹ Moderate latency - Good response time")
	} else {
		display.PrintWarning(fmt.Sprintf("⚠ High latency (%.1fms) - Consider network optimization", float64(result.AvgRTT.Nanoseconds())/1e6))
	}
}

// QuickPing performs a quick ping test with default settings
func QuickPing(host string) (*PingResult, error) {
	config := DefaultPingConfig(host)
	config.Count = 3 // Quick test with 3 packets
	config.Timeout = 5 * time.Second
	
	return PingHost(config)
}

// PingMultipleHosts tests connectivity to multiple hosts
func PingMultipleHosts(hosts []string) ([]*PingResult, error) {
	var results []*PingResult
	
	display.PrintInfo(fmt.Sprintf("Testing connectivity to %d hosts...", len(hosts)))
	
	for i, host := range hosts {
		display.PrintProgress(i+1, len(hosts), fmt.Sprintf("Pinging %s", host))
		
		result, err := QuickPing(host)
		if err != nil {
			display.PrintWarning(fmt.Sprintf("Failed to ping %s: %v", host, err))
		}
		
		results = append(results, result)
		
		// Small delay between pings
		time.Sleep(500 * time.Millisecond)
	}
	
	display.PrintSuccess("All ping tests completed")
	
	return results, nil
}

// ShowPingMultipleHosts displays ping results for multiple hosts
func ShowPingMultipleHosts() error {
	// Common hosts to test
	defaultHosts := []string{
		"google.com",
		"cloudflare.com",
		"microsoft.com",
		"1.1.1.1",
		"8.8.8.8",
	}
	
	display.PrintInfo("Multiple Host Ping Test")
	display.PrintSeparator()
	display.PrintInfo("Testing connectivity to common hosts...")
	
	results, err := PingMultipleHosts(defaultHosts)
	if err != nil {
		display.PrintError(fmt.Sprintf("Failed to ping hosts: %v", err))
		return err
	}
	
	// Create summary table
	var tableData [][]string
	for _, result := range results {
		status := "FAILED"
		if result.Success {
			status = display.Success("OK")
		} else {
			status = display.Error("FAILED")
		}
		
		packetLoss := fmt.Sprintf("%.1f%%", result.PacketLoss)
		avgRTT := utils.FormatDuration(result.AvgRTT)
		
		row := []string{
			result.Host,
			status,
			packetLoss,
			avgRTT,
			fmt.Sprintf("%d/%d", result.PacketsRecv, result.PacketsSent),
		}
		
		tableData = append(tableData, row)
	}
	
	tableConfig := display.NewTableConfig()
	tableConfig.Title = "Multiple Host Ping Results"
	tableConfig.Headers = []string{"Host", "Status", "Packet Loss", "Avg RTT", "Packets"}
	tableConfig.Data = tableData
	tableConfig.MaxWidth = 80
	
	display.PrintTable(tableConfig)
	
	return nil
}

// TestConnectivity performs comprehensive connectivity tests
func TestConnectivity() error {
	display.PrintInfo("Comprehensive Connectivity Test")
	display.PrintSeparator()
	
	// Test local connectivity first
	display.PrintInfo("1. Testing local connectivity...")
	_, err := QuickPing("127.0.0.1")
	if err != nil {
		display.PrintError("Local ping failed - system issue")
	} else {
		display.PrintSuccess("Local connectivity: OK")
	}
	
	// Test default gateway
	display.PrintInfo("2. Testing gateway connectivity...")
	gateway, err := GetDefaultGateway("IPv4")
	if err == nil {
		gatewayResult, err := QuickPing(gateway.Gateway)
		if err != nil {
			display.PrintWarning(fmt.Sprintf("Gateway ping failed: %v", err))
		} else {
			display.PrintSuccess(fmt.Sprintf("Gateway connectivity: OK (%s)", utils.FormatDuration(gatewayResult.AvgRTT)))
		}
	} else {
		display.PrintWarning("No gateway found")
	}
	
	// Test DNS servers
	display.PrintInfo("3. Testing DNS connectivity...")
	dnsConfig, err := getWindowsDNS()
	if utils.IsLinux() {
		dnsConfig, err = getLinuxDNS()
	}
	
	if err == nil && len(dnsConfig.Servers) > 0 {
		for _, dnsInfo := range dnsConfig.Servers {
			if len(dnsInfo.IPv4) > 0 {
				dnsResult, err := QuickPing(dnsInfo.IPv4[0])
				if err != nil {
					display.PrintWarning(fmt.Sprintf("DNS server %s ping failed", dnsInfo.IPv4[0]))
				} else {
					display.PrintSuccess(fmt.Sprintf("DNS server %s: OK (%s)", dnsInfo.IPv4[0], utils.FormatDuration(dnsResult.AvgRTT)))
				}
				break // Test only first DNS server
			}
		}
	}
	
	// Test internet connectivity
	display.PrintInfo("4. Testing internet connectivity...")
	internetResult, err := QuickPing("8.8.8.8")
	if err != nil {
		display.PrintError("Internet connectivity: FAILED")
	} else {
		display.PrintSuccess(fmt.Sprintf("Internet connectivity: OK (%s)", utils.FormatDuration(internetResult.AvgRTT)))
	}
	
	display.PrintSeparator()
	display.PrintSuccess("Connectivity test completed")
	
	return nil
}