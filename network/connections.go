package network

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"netinfo/display"
	"netinfo/utils"

	psnet "github.com/shirou/gopsutil/v3/net"
	psproc "github.com/shirou/gopsutil/v3/process"
)

// ConnectionInfo holds network connection information
type ConnectionInfo struct {
	Family     string `json:"family"`     // IPv4, IPv6
	Type       string `json:"type"`       // tcp, udp
	LocalAddr  string `json:"local_addr"` // local IP:port
	RemoteAddr string `json:"remote_addr"` // remote IP:port
	Status     string `json:"status"`     // ESTABLISHED, LISTEN, etc.
	PID        int32  `json:"pid"`        // process ID
	Process    string `json:"process"`    // process name
	User       string `json:"user"`       // process user
	CreateTime int64  `json:"create_time"` // connection creation time
}

// ConnectionConfig holds system connection configuration
type ConnectionConfig struct {
	Connections []ConnectionInfo `json:"connections"`
	TotalCount  int             `json:"total_count"`
	TCPCount    int             `json:"tcp_count"`
	UDPCount    int             `json:"udp_count"`
	ListenCount int             `json:"listen_count"`
	EstablishedCount int       `json:"established_count"`
}

// ShowActiveConnections displays active network connections
func ShowActiveConnections() error {
	display.PrintInfo("Gathering active network connections...")
	
	connectionConfig, err := getActiveConnections()
	if err != nil {
		display.PrintError(fmt.Sprintf("Failed to get connections: %v", err))
		return err
	}
	
	if len(connectionConfig.Connections) == 0 {
		display.PrintWarning("No active connections found")
		return nil
	}
	
	display.PrintSuccess(fmt.Sprintf("Found %d active connections", connectionConfig.TotalCount))
	
	// Sort connections by status, then by local address
	sort.Slice(connectionConfig.Connections, func(i, j int) bool {
		if connectionConfig.Connections[i].Status != connectionConfig.Connections[j].Status {
			return connectionConfig.Connections[i].Status < connectionConfig.Connections[j].Status
		}
		return connectionConfig.Connections[i].LocalAddr < connectionConfig.Connections[j].LocalAddr
	})
	
	// Create table data
	var tableData [][]string
	for _, conn := range connectionConfig.Connections {
		// Format local address
		localAddr := conn.LocalAddr
		if len(localAddr) > 25 {
			localAddr = utils.TruncateString(localAddr, 25)
		}
		
		// Format remote address
		remoteAddr := conn.RemoteAddr
		if remoteAddr == "" {
			remoteAddr = "-"
		} else if len(remoteAddr) > 25 {
			remoteAddr = utils.TruncateString(remoteAddr, 25)
		}
		
		// Format process name
		process := conn.Process
		if process == "" {
			process = "Unknown"
		} else if len(process) > 20 {
			process = utils.TruncateString(process, 20)
		}
		
		// Status with color
		status := conn.Status
		if status == "ESTABLISHED" {
			status = display.Success(status)
		} else if status == "LISTEN" {
			status = display.Info(status)
		} else {
			status = display.Warning(status)
		}
		
		row := []string{
			strings.ToUpper(conn.Type),
			conn.Family,
			localAddr,
			remoteAddr,
			status,
			fmt.Sprintf("%d", conn.PID),
			process,
		}
		
		tableData = append(tableData, row)
	}
	
	tableConfig := display.NewTableConfig()
	tableConfig.Title = "Active Network Connections"
	tableConfig.Headers = []string{"Type", "Family", "Local Address", "Remote Address", "Status", "PID", "Process"}
	tableConfig.Data = tableData
	tableConfig.MaxWidth = 100
	
	display.PrintTable(tableConfig)
	
	// Show summary statistics
	display.PrintSeparator()
	display.PrintInfo("Connection Summary:")
	display.PrintInfo(fmt.Sprintf("  • Total connections: %d", connectionConfig.TotalCount))
	display.PrintInfo(fmt.Sprintf("  • TCP connections: %d", connectionConfig.TCPCount))
	display.PrintInfo(fmt.Sprintf("  • UDP connections: %d", connectionConfig.UDPCount))
	display.PrintInfo(fmt.Sprintf("  • Listening connections: %d", connectionConfig.ListenCount))
	display.PrintInfo(fmt.Sprintf("  • Established connections: %d", connectionConfig.EstablishedCount))
	
	return nil
}

// getActiveConnections retrieves active network connections using gopsutil
func getActiveConnections() (*ConnectionConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	connectionConfig := &ConnectionConfig{
		Connections: []ConnectionInfo{},
	}
	
	// Get connections using gopsutil
	connections, err := psnet.ConnectionsWithContext(ctx, "all")
	if err != nil {
		return nil, fmt.Errorf("failed to get connections: %v", err)
	}
	
	// Get process information for better display
	processMap := make(map[int32]string)
	processes, err := psproc.ProcessesWithContext(ctx)
	if err == nil {
		for _, proc := range processes {
			name, err := proc.NameWithContext(ctx)
			if err == nil {
				processMap[proc.Pid] = name
			}
		}
	}
	
	for _, conn := range connections {
		// Skip connections with invalid addresses
		if conn.Laddr.IP == "" {
			continue
		}
		
		// Determine family
		family := "IPv4"
		if net.ParseIP(conn.Laddr.IP).To4() == nil {
			family = "IPv6"
		}
		
		// Format local address
		localAddr := fmt.Sprintf("%s:%d", conn.Laddr.IP, conn.Laddr.Port)
		
		// Format remote address
		remoteAddr := ""
		if conn.Raddr.IP != "" && conn.Raddr.IP != "0.0.0.0" && conn.Raddr.IP != "::" {
			remoteAddr = fmt.Sprintf("%s:%d", conn.Raddr.IP, conn.Raddr.Port)
		}
		
		// Get process name
		processName := processMap[conn.Pid]
		
		// Map gopsutil status to readable status
		status := mapGopsutilStatus(conn.Status)
		
		connectionInfo := ConnectionInfo{
			Family:     family,
			Type:       getConnectionType(conn.Type),
			LocalAddr:  localAddr,
			RemoteAddr: remoteAddr,
			Status:     status,
			PID:        conn.Pid,
			Process:    processName,
			User:       "", // Not available in gopsutil
			CreateTime: 0,  // Not available in gopsutil
		}
		
		connectionConfig.Connections = append(connectionConfig.Connections, connectionInfo)
		
		// Update counters
		connectionConfig.TotalCount++
		if getConnectionType(conn.Type) == "tcp" {
			connectionConfig.TCPCount++
		} else if getConnectionType(conn.Type) == "udp" {
			connectionConfig.UDPCount++
		}
		
		if status == "LISTEN" {
			connectionConfig.ListenCount++
		} else if status == "ESTABLISHED" {
			connectionConfig.EstablishedCount++
		}
	}
	
	return connectionConfig, nil
}

// getConnectionType maps gopsutil connection type to string
func getConnectionType(connType uint32) string {
	switch connType {
	case 1: // TCP
		return "tcp"
	case 2: // UDP
		return "udp"
	case 3: // TCP6
		return "tcp6"
	case 4: // UDP6
		return "udp6"
	case 5: // UNIX
		return "unix"
	default:
		return "unknown"
	}
}

// mapGopsutilStatus maps gopsutil status to readable status
func mapGopsutilStatus(status string) string {
	switch strings.ToUpper(status) {
	case "ESTABLISHED":
		return "ESTABLISHED"
	case "LISTEN":
		return "LISTEN"
	case "SYN_SENT":
		return "SYN_SENT"
	case "SYN_RECV":
		return "SYN_RECV"
	case "FIN_WAIT1":
		return "FIN_WAIT1"
	case "FIN_WAIT2":
		return "FIN_WAIT2"
	case "TIME_WAIT":
		return "TIME_WAIT"
	case "CLOSE":
		return "CLOSE"
	case "CLOSE_WAIT":
		return "CLOSE_WAIT"
	case "LAST_ACK":
		return "LAST_ACK"
	case "LISTENING":
		return "LISTEN"
	default:
		return strings.ToUpper(status)
	}
}

// ShowConnectionsByProcess shows connections grouped by process
func ShowConnectionsByProcess() error {
	display.PrintInfo("Grouping connections by process...")
	
	connectionConfig, err := getActiveConnections()
	if err != nil {
		display.PrintError(fmt.Sprintf("Failed to get connections: %v", err))
		return err
	}
	
	if len(connectionConfig.Connections) == 0 {
		display.PrintWarning("No active connections found")
		return nil
	}
	
	// Group connections by process
	processGroups := make(map[string][]ConnectionInfo)
	for _, conn := range connectionConfig.Connections {
		processName := conn.Process
		if processName == "" {
			processName = "Unknown"
		}
		processGroups[processName] = append(processGroups[processName], conn)
	}
	
	// Sort processes by connection count
	var processes []string
	for process := range processGroups {
		processes = append(processes, process)
	}
	sort.Slice(processes, func(i, j int) bool {
		return len(processGroups[processes[i]]) > len(processGroups[processes[j]])
	})
	
	display.PrintSuccess(fmt.Sprintf("Found %d processes with active connections", len(processes)))
	
	// Display process groups
	for _, process := range processes {
		connections := processGroups[process]
		
		// Count by type
		tcpCount := 0
		udpCount := 0
		listenCount := 0
		establishedCount := 0
		
		for _, conn := range connections {
			if conn.Type == "tcp" {
				tcpCount++
			} else if conn.Type == "udp" {
				udpCount++
			}
			
			if conn.Status == "LISTEN" {
				listenCount++
			} else if conn.Status == "ESTABLISHED" {
				establishedCount++
			}
		}
		
		// Display process info
		processInfo := map[string]string{
			"Process":     process,
			"PID":         fmt.Sprintf("%d", connections[0].PID),
			"Total":       fmt.Sprintf("%d", len(connections)),
			"TCP":         fmt.Sprintf("%d", tcpCount),
			"UDP":         fmt.Sprintf("%d", udpCount),
			"Listening":   fmt.Sprintf("%d", listenCount),
			"Established": fmt.Sprintf("%d", establishedCount),
		}
		
		display.PrintKeyValue(processInfo, "")
	}
	
	return nil
}

// ShowListeningPorts shows only listening ports
func ShowListeningPorts() error {
	display.PrintInfo("Gathering listening ports...")
	
	connectionConfig, err := getActiveConnections()
	if err != nil {
		display.PrintError(fmt.Sprintf("Failed to get connections: %v", err))
		return err
	}
	
	// Filter listening connections
	var listeningConnections []ConnectionInfo
	for _, conn := range connectionConfig.Connections {
		if conn.Status == "LISTEN" {
			listeningConnections = append(listeningConnections, conn)
		}
	}
	
	if len(listeningConnections) == 0 {
		display.PrintWarning("No listening ports found")
		return nil
	}
	
	display.PrintSuccess(fmt.Sprintf("Found %d listening ports", len(listeningConnections)))
	
	// Sort by port number
	sort.Slice(listeningConnections, func(i, j int) bool {
		port1, _ := strconv.Atoi(strings.Split(listeningConnections[i].LocalAddr, ":")[1])
		port2, _ := strconv.Atoi(strings.Split(listeningConnections[j].LocalAddr, ":")[1])
		return port1 < port2
	})
	
	// Create table data
	var tableData [][]string
	for _, conn := range listeningConnections {
		// Parse port
		port := strings.Split(conn.LocalAddr, ":")[1]
		
		// Get service name
		serviceName := getServiceName(port, conn.Type)
		
		// Format process name
		process := conn.Process
		if process == "" {
			process = "Unknown"
		}
		
		row := []string{
			port,
			strings.ToUpper(conn.Type),
			conn.Family,
			serviceName,
			fmt.Sprintf("%d", conn.PID),
			process,
		}
		
		tableData = append(tableData, row)
	}
	
	tableConfig := display.NewTableConfig()
	tableConfig.Title = "Listening Ports"
	tableConfig.Headers = []string{"Port", "Type", "Family", "Service", "PID", "Process"}
	tableConfig.Data = tableData
	tableConfig.MaxWidth = 80
	
	display.PrintTable(tableConfig)
	
	return nil
}

// getServiceName returns the service name for a given port and protocol
func getServiceName(port, protocol string) string {
	// Common port mappings
	commonPorts := map[string]string{
		"22":   "SSH",
		"23":   "Telnet",
		"25":   "SMTP",
		"53":   "DNS",
		"80":   "HTTP",
		"110":  "POP3",
		"143":  "IMAP",
		"443":  "HTTPS",
		"993":  "IMAPS",
		"995":  "POP3S",
		"1433": "MSSQL",
		"3306": "MySQL",
		"3389": "RDP",
		"5432": "PostgreSQL",
		"5900": "VNC",
		"6379": "Redis",
	}
	
	if service, exists := commonPorts[port]; exists {
		return service
	}
	
	return "Unknown"
}

// GetConnectionsByPort returns connections for a specific port
func GetConnectionsByPort(port int) ([]ConnectionInfo, error) {
	connectionConfig, err := getActiveConnections()
	if err != nil {
		return nil, err
	}
	
	var portConnections []ConnectionInfo
	for _, conn := range connectionConfig.Connections {
		// Check local port
		if strings.HasSuffix(conn.LocalAddr, fmt.Sprintf(":%d", port)) {
			portConnections = append(portConnections, conn)
		}
		// Check remote port
		if strings.HasSuffix(conn.RemoteAddr, fmt.Sprintf(":%d", port)) {
			portConnections = append(portConnections, conn)
		}
	}
	
	return portConnections, nil
}