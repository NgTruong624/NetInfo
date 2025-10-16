# NetInfo

NetInfo is a cross-platform terminal tool (Windows/Linux) for viewing network information and running common diagnostics from an interactive menu.

## Features
- Network Interfaces: list interface name, IPs, MAC, MTU, status
- IP Information: local IPv4/IPv6 per interface and public IP lookup
- DNS: show configured DNS servers (Windows via PowerShell, Linux via /etc/resolv.conf)
- Default Gateway: show IPv4/IPv6 gateways and metrics
- Routing Table: display routes with interface, gateway, metric, protocol
- Active Connections: list connections (TCP/UDP), listening ports, group by process
- Ping: single host, multiple common hosts, and a simple connectivity test

## Requirements
- Go 1.20+ (recommended)
- Windows: PowerShell available in PATH (default)
- Linux: `iproute2` package for route/gateway info

## Install / Build

### Windows (PowerShell)
```powershell
# Clone
git clone https://github.com/<your-username>/NetInfo.git
cd NetInfo

# Build executable
go build -o netinfo.exe .

# Run
./netinfo.exe

# Or run without building
go run .

# Optional: install to GOPATH/bin to call from anywhere
go install .
$env:Path += ";$env:USERPROFILE\go\bin"
netinfo
```

### Ubuntu / Debian
```bash
sudo apt update
sudo apt install -y git golang-go iproute2
# (If the Go version is too old, consider: sudo snap install go --classic)

# Clone and build
git clone https://github.com/<your-username>/NetInfo.git
cd NetInfo
go build -o netinfo .

# Run
./netinfo
# Some views (e.g., Active Connections / Listening Ports) may need sudo
sudo ./netinfo

# Optional: install to PATH
sudo mv netinfo /usr/local/bin/
netinfo

# Or install to GOPATH/bin
go install .
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bashrc
source ~/.bashrc
netinfo
```

## Usage
Run the program, then use the arrow keys and Enter to select menu items.

Main menu options include:
- Network Interfaces
- IP Information
- DNS Servers
- Default Gateway
- Routing Table
- Active Connections
- Ping Test
- Exit

## Notes & Troubleshooting
- Linux: ensure `iproute2` is installed for route/gateway features.
- Public IP lookup uses multiple endpoints with timeouts; if the network is restricted, this may fail gracefully.
- Some features (e.g., listing active connections/processes) might require elevated privileges on certain systems.

## Dependencies (Go modules)
- github.com/shirou/gopsutil/v3 (system/network info)
- github.com/manifoldco/promptui (interactive prompts)
- github.com/olekukonko/tablewriter (table output)
- github.com/fatih/color (colored output)

## License
MIT