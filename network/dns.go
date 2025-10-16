package network

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"netinfo/display"
	"netinfo/utils"
)

// DNSInfo holds DNS server information
type DNSInfo struct {
	Interface string   `json:"interface"`
	IPv4      []string `json:"ipv4"`
	IPv6      []string `json:"ipv6"`
	All       []string `json:"all"`
}

// DNSConfig holds system DNS configuration
type DNSConfig struct {
	Servers    []DNSInfo `json:"servers"`
	SearchList []string  `json:"search_list"`
}

// PowerShell DNS command for Windows
const windowsDNSCmd = `Get-DnsClientServerAddress -AddressFamily IPv4,IPv6 | Select-Object InterfaceAlias, ServerAddresses | ConvertTo-Json`

// ShowDNSInformation displays DNS server information
func ShowDNSInformation() error {
	display.PrintInfo("Gathering DNS server information...")
	
	var dnsConfig *DNSConfig
	var err error
	
	if utils.IsWindows() {
		dnsConfig, err = getWindowsDNS()
	} else {
		dnsConfig, err = getLinuxDNS()
	}
	
	if err != nil {
		display.PrintError(fmt.Sprintf("Failed to get DNS information: %v", err))
		return err
	}
	
	if len(dnsConfig.Servers) == 0 {
		display.PrintWarning("No DNS servers found")
		return nil
	}
	
	display.PrintSuccess(fmt.Sprintf("Found %d DNS configurations", len(dnsConfig.Servers)))
	
	// Create table for DNS servers
	var tableData [][]string
	for _, dnsInfo := range dnsConfig.Servers {
		// Format IPv4 servers
		ipv4Str := strings.Join(dnsInfo.IPv4, ", ")
		if ipv4Str == "" {
			ipv4Str = "None"
		}
		
		// Format IPv6 servers
		ipv6Str := strings.Join(dnsInfo.IPv6, ", ")
		if ipv6Str == "" {
			ipv6Str = "None"
		}
		
		// Format all servers
		allStr := strings.Join(dnsInfo.All, ", ")
		if len(allStr) > 50 {
			allStr = utils.TruncateString(allStr, 50)
		}
		
		row := []string{
			dnsInfo.Interface,
			ipv4Str,
			ipv6Str,
			allStr,
		}
		tableData = append(tableData, row)
	}
	
	tableConfig := display.NewTableConfig()
	tableConfig.Title = "DNS Servers"
	tableConfig.Headers = []string{"Interface", "IPv4 DNS", "IPv6 DNS", "All DNS Servers"}
	tableConfig.Data = tableData
	tableConfig.MaxWidth = 60
	
	display.PrintTable(tableConfig)
	
	// Show search list if available
	if len(dnsConfig.SearchList) > 0 {
		display.PrintList(dnsConfig.SearchList, "DNS Search List")
	}
	
	// Show summary
	display.PrintSeparator()
	display.PrintInfo("DNS Information Summary:")
	display.PrintInfo(fmt.Sprintf("  • Total interfaces: %d", len(dnsConfig.Servers)))
	
	totalIPv4 := 0
	totalIPv6 := 0
	for _, dnsInfo := range dnsConfig.Servers {
		totalIPv4 += len(dnsInfo.IPv4)
		totalIPv6 += len(dnsInfo.IPv6)
	}
	
	display.PrintInfo(fmt.Sprintf("  • Total IPv4 DNS servers: %d", totalIPv4))
	display.PrintInfo(fmt.Sprintf("  • Total IPv6 DNS servers: %d", totalIPv6))
	display.PrintInfo(fmt.Sprintf("  • DNS search domains: %d", len(dnsConfig.SearchList)))
	
	return nil
}

// getWindowsDNS retrieves DNS information on Windows using PowerShell
func getWindowsDNS() (*DNSConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Execute PowerShell command
	output, err := utils.CommandWithTimeout(ctx, 8*time.Second, "powershell", "-NoProfile", "-Command", windowsDNSCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to execute PowerShell command: %v", err)
	}
	
	// Parse JSON output
	var dnsEntries []map[string]interface{}
	err = json.Unmarshal(output, &dnsEntries)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PowerShell output: %v", err)
	}
	
	var dnsConfig DNSConfig
	
	for _, entry := range dnsEntries {
		interfaceName, ok := entry["InterfaceAlias"].(string)
		if !ok {
			continue
		}
		
		dnsInfo := DNSInfo{
			Interface: interfaceName,
			IPv4:      []string{},
			IPv6:      []string{},
			All:       []string{},
		}
		
		// Parse server addresses
		if servers, ok := entry["ServerAddresses"].([]interface{}); ok {
			for _, server := range servers {
				if serverStr, ok := server.(string); ok {
					dnsInfo.All = append(dnsInfo.All, serverStr)
					
					// Check if it's IPv4 or IPv6
					if ip := net.ParseIP(serverStr); ip != nil {
						if ip.To4() != nil {
							dnsInfo.IPv4 = append(dnsInfo.IPv4, serverStr)
						} else {
							dnsInfo.IPv6 = append(dnsInfo.IPv6, serverStr)
						}
					}
				}
			}
		}
		
		if len(dnsInfo.All) > 0 {
			dnsConfig.Servers = append(dnsConfig.Servers, dnsInfo)
		}
	}
	
	return &dnsConfig, nil
}

// getLinuxDNS retrieves DNS information on Linux by reading /etc/resolv.conf
func getLinuxDNS() (*DNSConfig, error) {
	dnsConfig := &DNSConfig{
		Servers:    []DNSInfo{},
		SearchList: []string{},
	}
	
	// Read /etc/resolv.conf
	content, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to read /etc/resolv.conf: %v", err)
	}
	
	lines := strings.Split(string(content), "\n")
	
	// Create a single DNS info entry for Linux
	dnsInfo := DNSInfo{
		Interface: "system",
		IPv4:      []string{},
		IPv6:      []string{},
		All:       []string{},
	}
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		
		switch fields[0] {
		case "nameserver":
			if len(fields) >= 2 {
				server := fields[1]
				dnsInfo.All = append(dnsInfo.All, server)
				
				// Check if it's IPv4 or IPv6
				if ip := net.ParseIP(server); ip != nil {
					if ip.To4() != nil {
						dnsInfo.IPv4 = append(dnsInfo.IPv4, server)
					} else {
						dnsInfo.IPv6 = append(dnsInfo.IPv6, server)
					}
				}
			}
		case "search":
			if len(fields) >= 2 {
				for i := 1; i < len(fields); i++ {
					dnsConfig.SearchList = append(dnsConfig.SearchList, fields[i])
				}
			}
		}
	}
	
	if len(dnsInfo.All) > 0 {
		dnsConfig.Servers = append(dnsConfig.Servers, dnsInfo)
	}
	
	return dnsConfig, nil
}

// TestDNSResolution tests DNS resolution for a given hostname
func TestDNSResolution(hostname string) error {
	display.PrintInfo(fmt.Sprintf("Testing DNS resolution for: %s", hostname))
	
	// Test A record (IPv4)
	ips, err := net.LookupIP(hostname)
	if err != nil {
		display.PrintError(fmt.Sprintf("DNS resolution failed: %v", err))
		return err
	}
	
	var ipv4s, ipv6s []string
	for _, ip := range ips {
		if ip.To4() != nil {
			ipv4s = append(ipv4s, ip.String())
		} else {
			ipv6s = append(ipv6s, ip.String())
		}
	}
	
	display.PrintSuccess(fmt.Sprintf("DNS resolution successful for: %s", hostname))
	
	// Display results
	if len(ipv4s) > 0 {
		display.PrintInfo(fmt.Sprintf("IPv4 addresses: %s", strings.Join(ipv4s, ", ")))
	}
	if len(ipv6s) > 0 {
		display.PrintInfo(fmt.Sprintf("IPv6 addresses: %s", strings.Join(ipv6s, ", ")))
	}
	
	// Test CNAME if available
	cnames, err := net.LookupCNAME(hostname)
	if err == nil && cnames != hostname+"." {
		display.PrintInfo(fmt.Sprintf("CNAME: %s", cnames))
	}
	
	// Test MX records
	mxRecords, err := net.LookupMX(hostname)
	if err == nil && len(mxRecords) > 0 {
		var mxStrings []string
		for _, mx := range mxRecords {
			mxStrings = append(mxStrings, fmt.Sprintf("%s (priority: %d)", mx.Host, mx.Pref))
		}
		display.PrintList(mxStrings, "MX Records")
	}
	
	return nil
}

// GetDNSServersByInterface returns DNS servers for a specific interface
func GetDNSServersByInterface(interfaceName string) (*DNSInfo, error) {
	dnsConfig, err := getWindowsDNS()
	if utils.IsLinux() {
		dnsConfig, err = getLinuxDNS()
	}
	
	if err != nil {
		return nil, err
	}
	
	for _, dnsInfo := range dnsConfig.Servers {
		if dnsInfo.Interface == interfaceName {
			return &dnsInfo, nil
		}
	}
	
	return nil, fmt.Errorf("no DNS servers found for interface: %s", interfaceName)
}

// IsValidDNS checks if an IP address is a valid DNS server
func IsValidDNS(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	
	// Check if it's a private IP (common for local DNS servers)
	if IsPrivateIP(ip) {
		return true
	}
	
	// Check if it's a well-known public DNS server
	wellKnownDNS := []string{
		"8.8.8.8",     // Google DNS
		"8.8.4.4",     // Google DNS
		"1.1.1.1",     // Cloudflare DNS
		"1.0.0.1",     // Cloudflare DNS
		"208.67.222.222", // OpenDNS
		"208.67.220.220", // OpenDNS
	}
	
	for _, dns := range wellKnownDNS {
		if dns == ip {
			return true
		}
	}
	
	return true // Assume valid if it's a valid IP
}

// ShowDNSResolutionTest provides an interactive DNS resolution test
// (removed) ShowDNSResolutionTest: unused in application