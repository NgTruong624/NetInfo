package network

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"netinfo/display"
	"netinfo/utils"
)

// GatewayInfo holds gateway information
type GatewayInfo struct {
	Interface string `json:"interface"`
	Gateway   string `json:"gateway"`
	IPVersion string `json:"ip_version"`
	Metric    int    `json:"metric"`
	Source    string `json:"source"`
}

// GatewayConfig holds system gateway configuration
type GatewayConfig struct {
	DefaultIPv4 *GatewayInfo `json:"default_ipv4"`
	DefaultIPv6 *GatewayInfo `json:"default_ipv6"`
	AllGateways []GatewayInfo `json:"all_gateways"`
}

// PowerShell commands for Windows
const (
	windowsGatewayCmd = `Get-NetRoute -DestinationPrefix "0.0.0.0/0" | Select-Object InterfaceAlias, NextHop, RouteMetric | ConvertTo-Json`
	windowsGatewayIPv6Cmd = `Get-NetRoute -DestinationPrefix "::/0" | Select-Object InterfaceAlias, NextHop, RouteMetric | ConvertTo-Json`
)

// ShowGatewayInformation displays gateway information
func ShowGatewayInformation() error {
	display.PrintInfo("Gathering gateway information...")
	
	var gatewayConfig *GatewayConfig
	var err error
	
	if utils.IsWindows() {
		gatewayConfig, err = getWindowsGateway()
	} else {
		gatewayConfig, err = getLinuxGateway()
	}
	
	if err != nil {
		display.PrintError(fmt.Sprintf("Failed to get gateway information: %v", err))
		return err
	}
	
	// Display default gateways
	if gatewayConfig.DefaultIPv4 != nil || gatewayConfig.DefaultIPv6 != nil {
		display.PrintSuccess("Found default gateway information")
		
		var tableData [][]string
		
		if gatewayConfig.DefaultIPv4 != nil {
			row := []string{
				"Default IPv4",
				gatewayConfig.DefaultIPv4.Interface,
				gatewayConfig.DefaultIPv4.Gateway,
				fmt.Sprintf("%d", gatewayConfig.DefaultIPv4.Metric),
				gatewayConfig.DefaultIPv4.Source,
			}
			tableData = append(tableData, row)
		}
		
		if gatewayConfig.DefaultIPv6 != nil {
			row := []string{
				"Default IPv6",
				gatewayConfig.DefaultIPv6.Interface,
				gatewayConfig.DefaultIPv6.Gateway,
				fmt.Sprintf("%d", gatewayConfig.DefaultIPv6.Metric),
				gatewayConfig.DefaultIPv6.Source,
			}
			tableData = append(tableData, row)
		}
		
		tableConfig := display.NewTableConfig()
		tableConfig.Title = "Default Gateways"
		tableConfig.Headers = []string{"Type", "Interface", "Gateway", "Metric", "Source"}
		tableConfig.Data = tableData
		tableConfig.MaxWidth = 60
		
		display.PrintTable(tableConfig)
	} else {
		display.PrintWarning("No default gateway found")
	}
	
	// Display all gateways if available
	if len(gatewayConfig.AllGateways) > 0 {
		display.PrintInfo(fmt.Sprintf("Found %d total gateway routes", len(gatewayConfig.AllGateways)))
		
		var tableData [][]string
		for _, gateway := range gatewayConfig.AllGateways {
			row := []string{
				gateway.IPVersion,
				gateway.Interface,
				gateway.Gateway,
				fmt.Sprintf("%d", gateway.Metric),
				gateway.Source,
			}
			tableData = append(tableData, row)
		}
		
		tableConfig := display.NewTableConfig()
		tableConfig.Title = "All Gateway Routes"
		tableConfig.Headers = []string{"Version", "Interface", "Gateway", "Metric", "Source"}
		tableConfig.Data = tableData
		tableConfig.MaxWidth = 60
		
		display.PrintTable(tableConfig)
	}
	
	// Show summary
	display.PrintSeparator()
	display.PrintInfo("Gateway Information Summary:")
	
	if gatewayConfig.DefaultIPv4 != nil {
		display.PrintInfo(fmt.Sprintf("  • Default IPv4 Gateway: %s (%s)", 
			display.IP(gatewayConfig.DefaultIPv4.Gateway), 
			gatewayConfig.DefaultIPv4.Interface))
	} else {
		display.PrintInfo("  • Default IPv4 Gateway: Not found")
	}
	
	if gatewayConfig.DefaultIPv6 != nil {
		display.PrintInfo(fmt.Sprintf("  • Default IPv6 Gateway: %s (%s)", 
			display.IP(gatewayConfig.DefaultIPv6.Gateway), 
			gatewayConfig.DefaultIPv6.Interface))
	} else {
		display.PrintInfo("  • Default IPv6 Gateway: Not found")
	}
	
	display.PrintInfo(fmt.Sprintf("  • Total gateway routes: %d", len(gatewayConfig.AllGateways)))
	
	return nil
}

// getWindowsGateway retrieves gateway information on Windows using PowerShell
func getWindowsGateway() (*GatewayConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	gatewayConfig := &GatewayConfig{
		AllGateways: []GatewayInfo{},
	}
	
	// Get IPv4 default gateway
	output, err := utils.CommandWithTimeout(ctx, 8*time.Second, "powershell", "-NoProfile", "-Command", windowsGatewayCmd)
	if err == nil {
		// Try parsing as single object first
		var singleRoute map[string]interface{}
		err = json.Unmarshal(output, &singleRoute)
		if err == nil {
			// Single route object
			interfaceName, _ := singleRoute["InterfaceAlias"].(string)
			nextHop, _ := singleRoute["NextHop"].(string)
			metric, _ := singleRoute["RouteMetric"].(float64)
			
			if interfaceName != "" && nextHop != "" {
				gatewayInfo := GatewayInfo{
					Interface: interfaceName,
					Gateway:   nextHop,
					IPVersion: "IPv4",
					Metric:    int(metric),
					Source:    "PowerShell",
				}
				
				gatewayConfig.AllGateways = append(gatewayConfig.AllGateways, gatewayInfo)
				
				// Set as default IPv4 if not set yet
				if gatewayConfig.DefaultIPv4 == nil {
					gatewayConfig.DefaultIPv4 = &gatewayInfo
				}
			}
		} else {
			// Try parsing as array
			var routes []map[string]interface{}
			err = json.Unmarshal(output, &routes)
			if err == nil {
				for _, route := range routes {
					interfaceName, _ := route["InterfaceAlias"].(string)
					nextHop, _ := route["NextHop"].(string)
					metric, _ := route["RouteMetric"].(float64)
					
					if interfaceName != "" && nextHop != "" {
						gatewayInfo := GatewayInfo{
							Interface: interfaceName,
							Gateway:   nextHop,
							IPVersion: "IPv4",
							Metric:    int(metric),
							Source:    "PowerShell",
						}
						
						gatewayConfig.AllGateways = append(gatewayConfig.AllGateways, gatewayInfo)
						
						// Set as default IPv4 if not set yet
						if gatewayConfig.DefaultIPv4 == nil {
							gatewayConfig.DefaultIPv4 = &gatewayInfo
						}
					}
				}
			}
		}
	}
	
	// Get IPv6 default gateway
	output, err = utils.CommandWithTimeout(ctx, 8*time.Second, "powershell", "-NoProfile", "-Command", windowsGatewayIPv6Cmd)
	if err == nil {
		// Try parsing as single object first
		var singleRoute map[string]interface{}
		err = json.Unmarshal(output, &singleRoute)
		if err == nil {
			// Single route object
			interfaceName, _ := singleRoute["InterfaceAlias"].(string)
			nextHop, _ := singleRoute["NextHop"].(string)
			metric, _ := singleRoute["RouteMetric"].(float64)
			
			if interfaceName != "" && nextHop != "" {
				gatewayInfo := GatewayInfo{
					Interface: interfaceName,
					Gateway:   nextHop,
					IPVersion: "IPv6",
					Metric:    int(metric),
					Source:    "PowerShell",
				}
				
				gatewayConfig.AllGateways = append(gatewayConfig.AllGateways, gatewayInfo)
				
				// Set as default IPv6 if not set yet
				if gatewayConfig.DefaultIPv6 == nil {
					gatewayConfig.DefaultIPv6 = &gatewayInfo
				}
			}
		} else {
			// Try parsing as array
			var routes []map[string]interface{}
			err = json.Unmarshal(output, &routes)
			if err == nil {
				for _, route := range routes {
					interfaceName, _ := route["InterfaceAlias"].(string)
					nextHop, _ := route["NextHop"].(string)
					metric, _ := route["RouteMetric"].(float64)
					
					if interfaceName != "" && nextHop != "" {
						gatewayInfo := GatewayInfo{
							Interface: interfaceName,
							Gateway:   nextHop,
							IPVersion: "IPv6",
							Metric:    int(metric),
							Source:    "PowerShell",
						}
						
						gatewayConfig.AllGateways = append(gatewayConfig.AllGateways, gatewayInfo)
						
						// Set as default IPv6 if not set yet
						if gatewayConfig.DefaultIPv6 == nil {
							gatewayConfig.DefaultIPv6 = &gatewayInfo
						}
					}
				}
			}
		}
	}
	
	return gatewayConfig, nil
}

// getLinuxGateway retrieves gateway information on Linux using ip command
func getLinuxGateway() (*GatewayConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	gatewayConfig := &GatewayConfig{
		AllGateways: []GatewayInfo{},
	}
	
	// Try ip -j route first (preferred method)
	output, err := utils.CommandWithTimeout(ctx, 5*time.Second, "ip", "-j", "route", "show", "default")
	if err == nil {
		var routes []map[string]interface{}
		err = json.Unmarshal(output, &routes)
		if err == nil {
			for _, route := range routes {
				interfaceName, _ := route["dev"].(string)
				gateway, _ := route["gateway"].(string)
				metric, _ := route["metric"].(float64)
				
				if interfaceName != "" && gateway != "" {
					// Determine IP version
					ipVersion := "IPv4"
					if net.ParseIP(gateway).To4() == nil {
						ipVersion = "IPv6"
					}
					
					gatewayInfo := GatewayInfo{
						Interface: interfaceName,
						Gateway:   gateway,
						IPVersion: ipVersion,
						Metric:    int(metric),
						Source:    "ip -j route",
					}
					
					gatewayConfig.AllGateways = append(gatewayConfig.AllGateways, gatewayInfo)
					
					// Set as default based on IP version
					if ipVersion == "IPv4" && gatewayConfig.DefaultIPv4 == nil {
						gatewayConfig.DefaultIPv4 = &gatewayInfo
					} else if ipVersion == "IPv6" && gatewayConfig.DefaultIPv6 == nil {
						gatewayConfig.DefaultIPv6 = &gatewayInfo
					}
				}
			}
		}
	}
	
	// Fallback to parsing ip route output if JSON failed
	if len(gatewayConfig.AllGateways) == 0 {
		output, err = utils.CommandWithTimeout(ctx, 5*time.Second, "ip", "route", "show", "default")
		if err == nil {
			gatewayConfig = parseIPRouteOutput(string(output))
		}
	}
	
	// Final fallback to /proc/net/route
	if len(gatewayConfig.AllGateways) == 0 {
		gatewayConfig, err = parseProcNetRoute()
	}
	
	return gatewayConfig, err
}

// parseIPRouteOutput parses the output of 'ip route show default'
func parseIPRouteOutput(output string) *GatewayConfig {
	gatewayConfig := &GatewayConfig{
		AllGateways: []GatewayInfo{},
	}
	
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Parse line like: "default via 192.168.1.1 dev wlan0 proto dhcp metric 600"
		parts := strings.Fields(line)
		if len(parts) >= 4 && parts[0] == "default" && parts[1] == "via" {
			gateway := parts[2]
			interfaceName := ""
			metric := 0
			
			// Find interface name
			for i, part := range parts {
				if part == "dev" && i+1 < len(parts) {
					interfaceName = parts[i+1]
					break
				}
			}
			
			// Find metric
			for i, part := range parts {
				if part == "metric" && i+1 < len(parts) {
					if m, err := strconv.Atoi(parts[i+1]); err == nil {
						metric = m
					}
					break
				}
			}
			
			if interfaceName != "" && gateway != "" {
				// Determine IP version
				ipVersion := "IPv4"
				if net.ParseIP(gateway).To4() == nil {
					ipVersion = "IPv6"
				}
				
				gatewayInfo := GatewayInfo{
					Interface: interfaceName,
					Gateway:   gateway,
					IPVersion: ipVersion,
					Metric:    metric,
					Source:    "ip route",
				}
				
				gatewayConfig.AllGateways = append(gatewayConfig.AllGateways, gatewayInfo)
				
				// Set as default based on IP version
				if ipVersion == "IPv4" && gatewayConfig.DefaultIPv4 == nil {
					gatewayConfig.DefaultIPv4 = &gatewayInfo
				} else if ipVersion == "IPv6" && gatewayConfig.DefaultIPv6 == nil {
					gatewayConfig.DefaultIPv6 = &gatewayInfo
				}
			}
		}
	}
	
	return gatewayConfig
}

// parseProcNetRoute parses /proc/net/route for IPv4 default gateway
func parseProcNetRoute() (*GatewayConfig, error) {
	// This is a simplified implementation for /proc/net/route
	// In a real implementation, you would need to parse the binary format
	gatewayConfig := &GatewayConfig{
		AllGateways: []GatewayInfo{},
	}
	
	// For now, return empty config
	// A full implementation would read and parse /proc/net/route
	return gatewayConfig, fmt.Errorf("parsing /proc/net/route not implemented")
}

// GetDefaultGateway returns the default gateway for the specified IP version
func GetDefaultGateway(ipVersion string) (*GatewayInfo, error) {
	gatewayConfig, err := getWindowsGateway()
	if utils.IsLinux() {
		gatewayConfig, err = getLinuxGateway()
	}
	
	if err != nil {
		return nil, err
	}
	
	if ipVersion == "IPv4" && gatewayConfig.DefaultIPv4 != nil {
		return gatewayConfig.DefaultIPv4, nil
	}
	
	if ipVersion == "IPv6" && gatewayConfig.DefaultIPv6 != nil {
		return gatewayConfig.DefaultIPv6, nil
	}
	
	return nil, fmt.Errorf("no default %s gateway found", ipVersion)
}

// TestGatewayConnectivity tests connectivity to the default gateway
// (removed) TestGatewayConnectivity: unused in application

// IsValidGateway checks if an IP address is a valid gateway
func IsValidGateway(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	
	// Check if it's a private IP (common for local gateways)
	if IsPrivateIP(ip) {
		return true
	}
	
	// Check if it's a valid public IP
	return true
}