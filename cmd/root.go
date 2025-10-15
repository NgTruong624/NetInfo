package cmd

import (
	"fmt"
	"os"

	"netinfo/display"
	"netinfo/network"
)

// Execute is the entrypoint invoked by main.
func Execute() {
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
			display.PrintInfo("IP Information feature coming soon...")
			display.PauseForUser("")
			
		case "dns":
			display.ClearScreen()
			display.ShowHeader()
			display.PrintInfo("DNS Servers feature coming soon...")
			display.PauseForUser("")
			
		case "gateway":
			display.ClearScreen()
			display.ShowHeader()
			display.PrintInfo("Default Gateway feature coming soon...")
			display.PauseForUser("")
			
		case "routes":
			display.ClearScreen()
			display.ShowHeader()
			display.PrintInfo("Routing Table feature coming soon...")
			display.PauseForUser("")
			
		case "connections":
			display.ClearScreen()
			display.ShowHeader()
			display.PrintInfo("Active Connections feature coming soon...")
			display.PauseForUser("")
			
		case "ping":
			display.ClearScreen()
			display.ShowHeader()
			display.PrintInfo("Ping Test feature coming soon...")
			display.PauseForUser("")
			
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


