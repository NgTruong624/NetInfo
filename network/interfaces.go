package network

import (
	"fmt"
	"net"
	"strings"

	"netinfo/display"
	"netinfo/utils"

	psnet "github.com/shirou/gopsutil/v3/net"
)

// InterfaceInfo holds network interface information
type InterfaceInfo struct {
	Name         string   `json:"name"`
	Index        int      `json:"index"`
	MTU          int      `json:"mtu"`
	HardwareAddr string   `json:"hardware_addr"`
	Flags        string   `json:"flags"`
	Addrs        []string `json:"addrs"`
	Status       string   `json:"status"`
}

// GetNetworkInterfaces retrieves all network interfaces information
func GetNetworkInterfaces() ([]InterfaceInfo, error) {
	var interfaces []InterfaceInfo
	
	// Get interfaces using gopsutil
	psInterfaces, err := psnet.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %v", err)
	}
	
	for _, psInterface := range psInterfaces {
		interfaceInfo := InterfaceInfo{
			Name:         psInterface.Name,
			Index:        int(psInterface.Index),
			MTU:          int(psInterface.MTU),
			HardwareAddr: psInterface.HardwareAddr,
			Flags:        strings.Join(psInterface.Flags, ","),
		}
		
		// Determine status based on flags
		if strings.Contains(strings.Join(psInterface.Flags, ","), "up") {
			interfaceInfo.Status = "UP"
		} else {
			interfaceInfo.Status = "DOWN"
		}
		
		// Get addresses for this interface
		for _, addr := range psInterface.Addrs {
			interfaceInfo.Addrs = append(interfaceInfo.Addrs, addr.Addr)
		}
		
		interfaces = append(interfaces, interfaceInfo)
	}
	
	return interfaces, nil
}

// GetNetworkInterfacesDetailed gets detailed interface information using Go standard library
func GetNetworkInterfacesDetailed() ([]InterfaceInfo, error) {
	var interfaces []InterfaceInfo
	
	// Get system interfaces using Go standard library
	systemInterfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get system interfaces: %v", err)
	}
	
	for _, iface := range systemInterfaces {
		// Get interface addresses
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		
		// Convert flags to string slice
		var flagStrings []string
		if iface.Flags&net.FlagUp != 0 {
			flagStrings = append(flagStrings, "up")
		}
		if iface.Flags&net.FlagLoopback != 0 {
			flagStrings = append(flagStrings, "loopback")
		}
		if iface.Flags&net.FlagBroadcast != 0 {
			flagStrings = append(flagStrings, "broadcast")
		}
		if iface.Flags&net.FlagPointToPoint != 0 {
			flagStrings = append(flagStrings, "pointtopoint")
		}
		if iface.Flags&net.FlagMulticast != 0 {
			flagStrings = append(flagStrings, "multicast")
		}
		
		interfaceInfo := InterfaceInfo{
			Name:         iface.Name,
			Index:        iface.Index,
			MTU:          iface.MTU,
			HardwareAddr: iface.HardwareAddr.String(),
			Flags:        strings.Join(flagStrings, ","),
		}
		
		// Determine status
		if iface.Flags&net.FlagUp != 0 {
			interfaceInfo.Status = "UP"
		} else {
			interfaceInfo.Status = "DOWN"
		}
		
		// Add addresses
		for _, addr := range addrs {
			interfaceInfo.Addrs = append(interfaceInfo.Addrs, addr.String())
		}
		
		interfaces = append(interfaces, interfaceInfo)
	}
	
	return interfaces, nil
}

// GetInterfaceByName gets information for a specific interface
func GetInterfaceByName(name string) (*InterfaceInfo, error) {
	interfaces, err := GetNetworkInterfaces()
	if err != nil {
		return nil, err
	}
	
	for _, iface := range interfaces {
		if iface.Name == name {
			return &iface, nil
		}
	}
	
	return nil, fmt.Errorf("interface %s not found", name)
}

// GetActiveInterfaces returns only interfaces that are up and have IP addresses
func GetActiveInterfaces() ([]InterfaceInfo, error) {
	allInterfaces, err := GetNetworkInterfaces()
	if err != nil {
		return nil, err
	}
	
	var activeInterfaces []InterfaceInfo
	for _, iface := range allInterfaces {
		if iface.Status == "UP" && len(iface.Addrs) > 0 {
			activeInterfaces = append(activeInterfaces, iface)
		}
	}
	
	return activeInterfaces, nil
}

// ShowNetworkInterfaces displays network interfaces in a formatted table
func ShowNetworkInterfaces() error {
	display.PrintInfo("Gathering network interface information...")
	
	interfaces, err := GetNetworkInterfaces()
	if err != nil {
		display.PrintError(fmt.Sprintf("Failed to get network interfaces: %v", err))
		return err
	}
	
	if len(interfaces) == 0 {
		display.PrintWarning("No network interfaces found")
		return nil
	}
	
	// Create table data
	var tableData [][]string
	for _, iface := range interfaces {
		// Format addresses
		addresses := strings.Join(iface.Addrs, ", ")
		if addresses == "" {
			addresses = "No IP"
		}
		
		// Truncate long addresses
		if len(addresses) > 30 {
			addresses = utils.TruncateString(addresses, 30)
		}
		
		// Format MAC address
		macAddr := iface.HardwareAddr
		if macAddr == "" {
			macAddr = "N/A"
		}
		
		// Status with color
		status := iface.Status
		if status == "UP" {
			status = display.Success(status)
		} else {
			status = display.Error(status)
		}
		
		row := []string{
			iface.Name,
			addresses,
			macAddr,
			fmt.Sprintf("%d", iface.MTU),
			status,
		}
		
		tableData = append(tableData, row)
	}
	
	// Display table
	tableConfig := display.NewTableConfig()
	tableConfig.Title = "Network Interfaces"
	tableConfig.Headers = []string{"Interface", "IP Addresses", "MAC Address", "MTU", "Status"}
	tableConfig.Data = tableData
	tableConfig.MaxWidth = 50
	
	display.PrintTable(tableConfig)
	
	// Show summary
	activeCount := 0
	for _, iface := range interfaces {
		if iface.Status == "UP" {
			activeCount++
		}
	}
	
	display.PrintInfo(fmt.Sprintf("Found %d total interfaces, %d active", len(interfaces), activeCount))
	
	return nil
}