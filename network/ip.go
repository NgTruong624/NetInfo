package network

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"netinfo/display"
	"netinfo/utils"
)

// IPInfo holds IP address information
type IPInfo struct {
	LocalIPs  []string `json:"local_ips"`
	PublicIP  string   `json:"public_ip"`
	IPv4      string   `json:"ipv4"`
	IPv6      string   `json:"ipv6"`
	Interface string   `json:"interface"`
}

// PublicIPResponse represents response from public IP services
type PublicIPResponse struct {
	IP string `json:"ip"`
}

// Public IP service endpoints with fallback
var PublicIPEndpoints = []string{
	"https://api.ipify.org",
	"https://ifconfig.me/ip",
	"https://icanhazip.com",
	"https://ident.me",
	"https://ipecho.net/plain",
}

// GetLocalIPs retrieves all local IP addresses
func GetLocalIPs() ([]IPInfo, error) {
	var ipInfos []IPInfo
	
	// Get all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %v", err)
	}
	
	for _, iface := range interfaces {
		// Skip down interfaces
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		
		// Get addresses for this interface
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		
		if len(addrs) == 0 {
			continue
		}
		
		var localIPs []string
		var ipv4, ipv6 string
		
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			
			ip := ipNet.IP
			if ip.IsLoopback() {
				continue
			}
			
			ipStr := ip.String()
			localIPs = append(localIPs, ipStr)
			
			// Separate IPv4 and IPv6
			if ip.To4() != nil {
				ipv4 = ipStr
			} else {
				ipv6 = ipStr
			}
		}
		
		if len(localIPs) > 0 {
			ipInfo := IPInfo{
				LocalIPs:  localIPs,
				IPv4:      ipv4,
				IPv6:      ipv6,
				Interface: iface.Name,
			}
			ipInfos = append(ipInfos, ipInfo)
		}
	}
	
	return ipInfos, nil
}

// GetPublicIP retrieves public IP address with fallback endpoints
func GetPublicIP() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), utils.PublicIPTimout)
	defer cancel()
	
	client := &http.Client{
		Timeout: utils.HTTPTimeout,
	}
	
	var lastErr error
	for _, endpoint := range PublicIPEndpoints {
		select {
		case <-ctx.Done():
			return "", utils.WrapError(nil, utils.ErrPublicIP, utils.ErrorTypeTimeout)
		default:
		}
		
		req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
		if err != nil {
			lastErr = utils.WrapError(err, "Failed to create request", utils.ErrorTypeNetwork)
			continue
		}
		
		resp, err := client.Do(req)
		if err != nil {
			lastErr = utils.WrapError(err, "Request failed", utils.ErrorTypeNetwork)
			continue
		}
		
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = utils.WrapError(nil, fmt.Sprintf("HTTP %d", resp.StatusCode), utils.ErrorTypeNetwork)
			continue
		}
		
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		
		if err != nil {
			lastErr = utils.WrapError(err, "Failed to read response", utils.ErrorTypeParse)
			continue
		}
		
		// Clean the response
		publicIP := strings.TrimSpace(string(body))
		
		// Validate IP address
		if net.ParseIP(publicIP) != nil {
			return publicIP, nil
		}
		
		lastErr = utils.WrapError(nil, "Invalid IP address received", utils.ErrorTypeParse)
	}
	
	// If we get here, all endpoints failed
	if lastErr != nil {
		return "", utils.WrapError(lastErr, utils.ErrPublicIP, utils.ErrorTypeNetwork)
	}
	
	return "", utils.WrapError(nil, utils.ErrPublicIP, utils.ErrorTypeNetwork)
}

// GetPublicIPWithService tries to get public IP from a specific service
func GetPublicIPWithService(endpoint string) (string, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	resp, err := client.Get(endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	publicIP := strings.TrimSpace(string(body))
	
	// Validate IP address
	if net.ParseIP(publicIP) == nil {
		return "", fmt.Errorf("invalid IP address: %s", publicIP)
	}
	
	return publicIP, nil
}

// GetIPLocation gets approximate location for an IP (using ipapi.co)
func GetIPLocation(ip string) (map[string]interface{}, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	url := fmt.Sprintf("https://ipapi.co/%s/json/", ip)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var location map[string]interface{}
	err = json.Unmarshal(body, &location)
	if err != nil {
		return nil, err
	}
	
	return location, nil
}

// ShowIPInformation displays comprehensive IP information
func ShowIPInformation() error {
	display.PrintInfo(utils.MsgGatheringInfo)
	
	// Get local IPs
	localIPs, err := GetLocalIPs()
	if err != nil {
		userMsg := utils.GetUserFriendlyMessage(err)
		display.PrintError(userMsg)
		return err
	}
	
	// Display local IPs
	if len(localIPs) == 0 {
		display.PrintWarning(utils.MsgNoIPs)
	} else {
		display.PrintSuccess(fmt.Sprintf("Found %d network interfaces with IP addresses", len(localIPs)))
		
		// Create table for local IPs
		var tableData [][]string
		for _, ipInfo := range localIPs {
			// Join all local IPs
			allIPs := strings.Join(ipInfo.LocalIPs, ", ")
			if len(allIPs) > 40 {
				allIPs = utils.TruncateString(allIPs, 40)
			}
			
			row := []string{
				ipInfo.Interface,
				ipInfo.IPv4,
				ipInfo.IPv6,
				allIPs,
			}
			tableData = append(tableData, row)
		}
		
		tableConfig := display.NewTableConfig()
		tableConfig.Title = "Local IP Addresses"
		tableConfig.Headers = []string{"Interface", "IPv4", "IPv6", "All IPs"}
		tableConfig.Data = tableData
		tableConfig.MaxWidth = 60
		
		display.PrintTable(tableConfig)
	}
	
	// Get public IP
	display.PrintInfo("Getting public IP address...")
	publicIP, err := GetPublicIP()
	if err != nil {
		userMsg := utils.GetUserFriendlyMessage(err)
		display.PrintError(userMsg)
		display.PrintWarning(utils.MsgTryAgain)
	} else {
		display.PrintSuccess(fmt.Sprintf("Public IP: %s", display.IP(publicIP)))
		
		// Try to get location information
		display.PrintInfo("Getting location information...")
		location, err := GetIPLocation(publicIP)
		if err != nil {
			display.PrintWarning("Could not retrieve location information")
		} else {
			// Display location information
			locationInfo := map[string]string{
				"Country":    getString(location, "country_name"),
				"Region":     getString(location, "region"),
				"City":       getString(location, "city"),
				"ISP":        getString(location, "org"),
				"Timezone":   getString(location, "timezone"),
			}
			
			display.PrintKeyValue(locationInfo, "Public IP Location")
		}
	}
	
	// Show summary
	display.PrintSeparator()
	display.PrintInfo("IP Information Summary:")
	display.PrintInfo(fmt.Sprintf("  • Local interfaces: %d", len(localIPs)))
	if publicIP != "" {
		display.PrintInfo(fmt.Sprintf("  • Public IP: %s", publicIP))
	} else {
		display.PrintInfo("  • Public IP: Not available")
	}
	
	return nil
}

// Helper function to safely get string from map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return "N/A"
}

// GetPrimaryLocalIP returns the primary local IP (usually the first non-loopback IPv4)
func GetPrimaryLocalIP() (string, error) {
	localIPs, err := GetLocalIPs()
	if err != nil {
		return "", err
	}
	
	for _, ipInfo := range localIPs {
		if ipInfo.IPv4 != "" {
			return ipInfo.IPv4, nil
		}
	}
	
	return "", fmt.Errorf("no IPv4 address found")
}

// IsPrivateIP checks if an IP address is private
func IsPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	
	private := []*net.IPNet{
		{IP: net.IPv4(10, 0, 0, 0), Mask: net.IPv4Mask(255, 0, 0, 0)},
		{IP: net.IPv4(172, 16, 0, 0), Mask: net.IPv4Mask(255, 240, 0, 0)},
		{IP: net.IPv4(192, 168, 0, 0), Mask: net.IPv4Mask(255, 255, 0, 0)},
		{IP: net.IPv4(127, 0, 0, 0), Mask: net.IPv4Mask(255, 0, 0, 0)},
	}
	
	for _, p := range private {
		if p.Contains(parsedIP) {
			return true
		}
	}
	
	return false
}