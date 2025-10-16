package cmd

import (
	"fmt"
	"os"

	"netinfo/display"
	"netinfo/network"
)

// Execute is the entrypoint invoked by main.
func Execute() {
mainLoop:
	for {
		// Show header
		display.ShowHeader()
		
		// Show main menu and get user choice
		choice, err := display.ShowMainMenu()
		if err != nil {
			display.PrintError(fmt.Sprintf("Menu error: %v", err))
			display.PauseForUser("")
			continue
		}
		
		// Handle user choice
		switch choice {
		case "interfaces":
			display.ClearScreen()
			display.ShowHeader()
			err := network.ShowNetworkInterfaces()
			if err != nil {
				display.PrintError(fmt.Sprintf("Failed to show interfaces: %v", err))
			}
			display.PauseForUser("")
			
		case "ip":
			display.ClearScreen()
			display.ShowHeader()
			err := network.ShowIPInformation()
			if err != nil {
				display.PrintError(fmt.Sprintf("Failed to show IP information: %v", err))
			}
			display.PauseForUser("")
			
		case "dns":
			display.ClearScreen()
			display.ShowHeader()
			err := network.ShowDNSInformation()
			if err != nil {
				display.PrintError(fmt.Sprintf("Failed to show DNS information: %v", err))
			}
			display.PauseForUser("")
			
		case "gateway":
			display.ClearScreen()
			display.ShowHeader()
			err := network.ShowGatewayInformation()
			if err != nil {
				display.PrintError(fmt.Sprintf("Failed to show gateway information: %v", err))
			}
			display.PauseForUser("")
			
		case "routes":
			display.ClearScreen()
			display.ShowHeader()
			err := network.ShowRoutingTable()
			if err != nil {
				display.PrintError(fmt.Sprintf("Failed to show routing table: %v", err))
			}
			display.PauseForUser("")
			
		case "connections":
			for {
				display.ClearScreen()
				display.ShowHeader()
				
				choice, err := display.ShowConnectionsMenu()
				if err != nil {
					display.PrintError(fmt.Sprintf("Menu error: %v", err))
					display.PauseForUser("")
					continue
				}
				
				switch choice {
				case "all":
					display.ClearScreen()
					display.ShowHeader()
					err := network.ShowActiveConnections()
					if err != nil {
						display.PrintError(fmt.Sprintf("Failed to show connections: %v", err))
					}
					display.PauseForUser("")
					
				case "listening":
					display.ClearScreen()
					display.ShowHeader()
					err := network.ShowListeningPorts()
					if err != nil {
						display.PrintError(fmt.Sprintf("Failed to show listening ports: %v", err))
					}
					display.PauseForUser("")
					
				case "by_process":
					display.ClearScreen()
					display.ShowHeader()
					err := network.ShowConnectionsByProcess()
					if err != nil {
						display.PrintError(fmt.Sprintf("Failed to show connections by process: %v", err))
					}
					display.PauseForUser("")
					
				case "back":
					goto mainLoop
					
				default:
					display.PrintWarning("Invalid choice. Please try again.")
					display.PauseForUser("")
				}
			}
			
		case "ping":
			for {
				display.ClearScreen()
				display.ShowHeader()
				
				choice, err := display.ShowPingMenu()
				if err != nil {
					display.PrintError(fmt.Sprintf("Menu error: %v", err))
					display.PauseForUser("")
					continue
				}
				
				switch choice {
				case "single":
					display.ClearScreen()
					display.ShowHeader()
					err := network.ShowPingTest()
					if err != nil {
						display.PrintError(fmt.Sprintf("Failed to run ping test: %v", err))
					}
					display.PauseForUser("")
					
				case "multiple":
					display.ClearScreen()
					display.ShowHeader()
					err := network.ShowPingMultipleHosts()
					if err != nil {
						display.PrintError(fmt.Sprintf("Failed to run multiple ping test: %v", err))
					}
					display.PauseForUser("")
					
				case "comprehensive":
					display.ClearScreen()
					display.ShowHeader()
					err := network.TestConnectivity()
					if err != nil {
						display.PrintError(fmt.Sprintf("Failed to run comprehensive test: %v", err))
					}
					display.PauseForUser("")
					
				case "back":
					goto mainLoop
					
				default:
					display.PrintWarning("Invalid choice. Please try again.")
					display.PauseForUser("")
				}
			}
			
		case "exit":
			display.ClearScreen()
			display.ShowGoodbye()
			os.Exit(0)
			
		default:
			display.PrintWarning("Invalid choice. Please try again.")
			display.PauseForUser("")
		}
	}
}


