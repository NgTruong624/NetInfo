package network

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"netinfo/display"
	"netinfo/utils"
)

// RouteInfo holds routing table entry information
type RouteInfo struct {
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Interface   string `json:"interface"`
	Metric      int    `json:"metric"`
	Protocol    string `json:"protocol"`
	Source      string `json:"source"`
	Type        string `json:"type"`
}

// RouteConfig holds system routing configuration
type RouteConfig struct {
	Routes []RouteInfo `json:"routes"`
}

// PowerShell commands for Windows
const (
	windowsRoutesCmd = `Get-NetRoute | Select-Object DestinationPrefix, NextHop, InterfaceAlias, RouteMetric, Protocol | ConvertTo-Json`
)

// ShowRoutingTable displays the complete routing table
func ShowRoutingTable() error {
	display.PrintInfo("Gathering routing table information...")
	
	var routeConfig *RouteConfig
	var err error
	
	if utils.IsWindows() {
		routeConfig, err = getWindowsRoutes()
	} else {
		routeConfig, err = getLinuxRoutes()
	}
	
	if err != nil {
		display.PrintError(fmt.Sprintf("Failed to get routing table: %v", err))
		return err
	}
	
	if len(routeConfig.Routes) == 0 {
		display.PrintWarning("No routing table entries found")
		return nil
	}
	
	display.PrintSuccess(fmt.Sprintf("Found %d routing table entries", len(routeConfig.Routes)))
	
	// Sort routes by destination for better readability
	sort.Slice(routeConfig.Routes, func(i, j int) bool {
		return routeConfig.Routes[i].Destination < routeConfig.Routes[j].Destination
	})
	
	// Create table data
	var tableData [][]string
	for _, route := range routeConfig.Routes {
		// Format gateway
		gateway := route.Gateway
		if gateway == "" {
			gateway = "On-link"
		}
		
		// Truncate long destinations
		destination := route.Destination
		if len(destination) > 25 {
			destination = utils.TruncateString(destination, 25)
		}
		
		row := []string{
			destination,
			gateway,
			route.Interface,
			fmt.Sprintf("%d", route.Metric),
			route.Protocol,
			route.Source,
		}
		
		tableData = append(tableData, row)
	}
	
	tableConfig := display.NewTableConfig()
	tableConfig.Title = "Routing Table"
	tableConfig.Headers = []string{"Destination", "Gateway", "Interface", "Metric", "Protocol", "Source"}
	tableConfig.Data = tableData
	tableConfig.MaxWidth = 80
	
	display.PrintTable(tableConfig)
	
	// Show summary statistics
	display.PrintSeparator()
	display.PrintInfo("Routing Table Summary:")
	
	// Count by interface
	interfaceCount := make(map[string]int)
	for _, route := range routeConfig.Routes {
		interfaceCount[route.Interface]++
	}
	
	display.PrintInfo(fmt.Sprintf("  • Total routes: %d", len(routeConfig.Routes)))
	display.PrintInfo(fmt.Sprintf("  • Routes by interface:"))
	
	// Sort interfaces by route count
	var interfaces []string
	for iface := range interfaceCount {
		interfaces = append(interfaces, iface)
	}
	sort.Slice(interfaces, func(i, j int) bool {
		return interfaceCount[interfaces[i]] > interfaceCount[interfaces[j]]
	})
	
	for _, iface := range interfaces {
		display.PrintInfo(fmt.Sprintf("    - %s: %d routes", iface, interfaceCount[iface]))
	}
	
	// Count default routes
	defaultRoutes := 0
	for _, route := range routeConfig.Routes {
		if route.Destination == "0.0.0.0/0" || route.Destination == "::/0" || route.Destination == "default" {
			defaultRoutes++
		}
	}
	
	if defaultRoutes > 0 {
		display.PrintInfo(fmt.Sprintf("  • Default routes: %d", defaultRoutes))
	}
	
	return nil
}

// getWindowsRoutes retrieves routing table on Windows using PowerShell
func getWindowsRoutes() (*RouteConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	routeConfig := &RouteConfig{
		Routes: []RouteInfo{},
	}
	
	// Execute PowerShell command
	output, err := utils.CommandWithTimeout(ctx, 10*time.Second, "powershell", "-NoProfile", "-Command", windowsRoutesCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to execute PowerShell command: %v", err)
	}
	
	// Parse JSON output
	var routes []map[string]interface{}
	err = json.Unmarshal(output, &routes)
	if err != nil {
		// Try parsing as single object first
		var singleRoute map[string]interface{}
		err = json.Unmarshal(output, &singleRoute)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PowerShell output: %v", err)
		}
		routes = []map[string]interface{}{singleRoute}
	}
	
	for _, route := range routes {
		destination, _ := route["DestinationPrefix"].(string)
		nextHop, _ := route["NextHop"].(string)
		interfaceName, _ := route["InterfaceAlias"].(string)
		metric, _ := route["RouteMetric"].(float64)
		protocol, _ := route["Protocol"].(string)
		
		// Skip empty destinations
		if destination == "" {
			continue
		}
		
		// Determine route type
		routeType := "Host"
		if strings.Contains(destination, "/") {
			if destination == "0.0.0.0/0" || destination == "::/0" {
				routeType = "Default"
			} else {
				routeType = "Network"
			}
		}
		
		routeInfo := RouteInfo{
			Destination: destination,
			Gateway:     nextHop,
			Interface:   interfaceName,
			Metric:      int(metric),
			Protocol:    protocol,
			Source:      "PowerShell",
			Type:        routeType,
		}
		
		routeConfig.Routes = append(routeConfig.Routes, routeInfo)
	}
	
	return routeConfig, nil
}

// getLinuxRoutes retrieves routing table on Linux using ip command
func getLinuxRoutes() (*RouteConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	routeConfig := &RouteConfig{
		Routes: []RouteInfo{},
	}
	
	// Try ip -j route first (preferred method)
	output, err := utils.CommandWithTimeout(ctx, 5*time.Second, "ip", "-j", "route", "show", "all")
	if err == nil {
		var routes []map[string]interface{}
		err = json.Unmarshal(output, &routes)
		if err == nil {
			for _, route := range routes {
				destination, _ := route["dst"].(string)
				gateway, _ := route["gateway"].(string)
				interfaceName, _ := route["dev"].(string)
				metric, _ := route["metric"].(float64)
				protocol, _ := route["protocol"].(string)
				
				// Skip empty destinations
				if destination == "" {
					continue
				}
				
				// Determine route type
				routeType := "Host"
				if strings.Contains(destination, "/") {
					if destination == "default" || destination == "0.0.0.0/0" || destination == "::/0" {
						routeType = "Default"
					} else {
						routeType = "Network"
					}
				}
				
				routeInfo := RouteInfo{
					Destination: destination,
					Gateway:     gateway,
					Interface:   interfaceName,
					Metric:      int(metric),
					Protocol:    protocol,
					Source:      "ip -j route",
					Type:        routeType,
				}
				
				routeConfig.Routes = append(routeConfig.Routes, routeInfo)
			}
		}
	}
	
	// Fallback to parsing ip route output if JSON failed
	if len(routeConfig.Routes) == 0 {
		output, err = utils.CommandWithTimeout(ctx, 5*time.Second, "ip", "route", "show")
		if err == nil {
			routeConfig = parseLinuxIPRouteOutput(string(output))
		}
	}
	
	// Final fallback to /proc/net/route
	if len(routeConfig.Routes) == 0 {
		routeConfig, err = parseLinuxProcNetRoute()
	}
	
	return routeConfig, err
}

// parseLinuxIPRouteOutput parses the output of 'ip route show'
func parseLinuxIPRouteOutput(output string) *RouteConfig {
	routeConfig := &RouteConfig{
		Routes: []RouteInfo{},
	}
	
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Parse line like: "192.168.1.0/24 via 192.168.1.1 dev wlan0 proto dhcp metric 600"
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		
		destination := parts[0]
		gateway := ""
		interfaceName := ""
		metric := 0
		protocol := "static"
		
		// Find gateway
		for i, part := range parts {
			if part == "via" && i+1 < len(parts) {
				gateway = parts[i+1]
				break
			}
		}
		
		// Find interface
		for i, part := range parts {
			if part == "dev" && i+1 < len(parts) {
				interfaceName = parts[i+1]
				break
			}
		}
		
		// Find protocol
		for i, part := range parts {
			if part == "proto" && i+1 < len(parts) {
				protocol = parts[i+1]
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
		
		// Determine route type
		routeType := "Host"
		if strings.Contains(destination, "/") {
			if destination == "default" || destination == "0.0.0.0/0" || destination == "::/0" {
				routeType = "Default"
			} else {
				routeType = "Network"
			}
		}
		
		routeInfo := RouteInfo{
			Destination: destination,
			Gateway:     gateway,
			Interface:   interfaceName,
			Metric:      metric,
			Protocol:    protocol,
			Source:      "ip route",
			Type:        routeType,
		}
		
		routeConfig.Routes = append(routeConfig.Routes, routeInfo)
	}
	
	return routeConfig
}

// parseLinuxProcNetRoute parses /proc/net/route for IPv4 routes
func parseLinuxProcNetRoute() (*RouteConfig, error) {
	// This is a simplified implementation for /proc/net/route
	// In a real implementation, you would need to parse the binary format
	routeConfig := &RouteConfig{
		Routes: []RouteInfo{},
	}
	
	// For now, return empty config
	// A full implementation would read and parse /proc/net/route
	return routeConfig, fmt.Errorf("parsing /proc/net/route not implemented")
}

// GetRoutesByInterface returns routes for a specific interface
func GetRoutesByInterface(interfaceName string) ([]RouteInfo, error) {
	routeConfig, err := getWindowsRoutes()
	if utils.IsLinux() {
		routeConfig, err = getLinuxRoutes()
	}
	
	if err != nil {
		return nil, err
	}
	
	var interfaceRoutes []RouteInfo
	for _, route := range routeConfig.Routes {
		if route.Interface == interfaceName {
			interfaceRoutes = append(interfaceRoutes, route)
		}
	}
	
	return interfaceRoutes, nil
}

// GetDefaultRoutes returns all default routes
func GetDefaultRoutes() ([]RouteInfo, error) {
	routeConfig, err := getWindowsRoutes()
	if utils.IsLinux() {
		routeConfig, err = getLinuxRoutes()
	}
	
	if err != nil {
		return nil, err
	}
	
	var defaultRoutes []RouteInfo
	for _, route := range routeConfig.Routes {
		if route.Type == "Default" {
			defaultRoutes = append(defaultRoutes, route)
		}
	}
	
	return defaultRoutes, nil
}

// ShowRouteSummary displays a summary of the routing table
func ShowRouteSummary() error {
	display.PrintInfo("Generating routing table summary...")
	
	routeConfig, err := getWindowsRoutes()
	if utils.IsLinux() {
		routeConfig, err = getLinuxRoutes()
	}
	
	if err != nil {
		return err
	}
	
	if len(routeConfig.Routes) == 0 {
		display.PrintWarning("No routes found")
		return nil
	}
	
	// Count by type
	typeCount := make(map[string]int)
	for _, route := range routeConfig.Routes {
		typeCount[route.Type]++
	}
	
	// Count by protocol
	protocolCount := make(map[string]int)
	for _, route := range routeConfig.Routes {
		protocolCount[route.Protocol]++
	}
	
	display.PrintSuccess("Routing Table Summary")
	display.PrintSeparator()
	
	display.PrintInfo(fmt.Sprintf("Total routes: %d", len(routeConfig.Routes)))
	
	display.PrintInfo("Routes by type:")
	for routeType, count := range typeCount {
		display.PrintInfo(fmt.Sprintf("  • %s: %d", routeType, count))
	}
	
	display.PrintInfo("Routes by protocol:")
	for protocol, count := range protocolCount {
		display.PrintInfo(fmt.Sprintf("  • %s: %d", protocol, count))
	}
	
	return nil
}