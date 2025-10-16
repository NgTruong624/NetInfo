package display

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

// MenuItem represents a menu option
type MenuItem struct {
	Label string
	Value string
	Desc  string
}

// MenuConfig holds configuration for the interactive menu
type MenuConfig struct {
	Label    string
	Items    []MenuItem
	Size     int
	Selected string
}

// Main menu items for NetInfo
var MainMenuItems = []MenuItem{
	{
		Label: "Network Interfaces",
		Value: "interfaces",
		Desc:  "Show all network interfaces (IP, MAC, MTU, status)",
	},
	{
		Label: "IP Information",
		Value: "ip",
		Desc:  "Show local and public IP addresses",
	},
	{
		Label: "DNS Servers",
		Value: "dns",
		Desc:  "Show configured DNS servers",
	},
	{
		Label: "Default Gateway",
		Value: "gateway",
		Desc:  "Show default gateway information",
	},
	{
		Label: "Routing Table",
		Value: "routes",
		Desc:  "Show routing table",
	},
	{
		Label: "Active Connections",
		Value: "connections",
		Desc:  "Show active network connections",
	},
	{
		Label: "Ping Test",
		Value: "ping",
		Desc:  "Test connectivity to a host",
	},
	{
		Label: "Exit",
		Value: "exit",
		Desc:  "Exit NetInfo",
	},
}

var ConnectionsMenuItems = []MenuItem{
	{
		Label: "All Connections",
		Value: "all",
		Desc:  "Show all active network connections",
	},
	{
		Label: "Listening Ports",
		Value: "listening",
		Desc:  "Show only listening ports",
	},
	{
		Label: "By Process",
		Value: "by_process",
		Desc:  "Group connections by process",
	},
	{
		Label: "Back to Main Menu",
		Value: "back",
		Desc:  "Return to main menu",
	},
}

// ShowMainMenu displays the main interactive menu
func ShowMainMenu() (string, error) {
	config := &MenuConfig{
		Label:    "Select an option",
		Items:    MainMenuItems,
		Size:     8,
		Selected: "",
	}
	
	return ShowMenu(config)
}

// ShowConnectionsMenu displays the connections submenu
func ShowConnectionsMenu() (string, error) {
	config := &MenuConfig{
		Label:    "Select connection view",
		Items:    ConnectionsMenuItems,
		Size:     4,
		Selected: "",
	}
	
	return ShowMenu(config)
}

// ShowMenu displays an interactive menu using promptui
func ShowMenu(config *MenuConfig) (string, error) {
	// Create select prompt
	prompt := promptui.Select{
		Label:        config.Label,
		Items:        config.Items,
		Size:         config.Size,
		HideSelected: true,
		Templates: &promptui.SelectTemplates{
			Active:   "▸ {{ .Label | cyan }} {{ .Desc | faint }}",
			Inactive: "  {{ .Label }} {{ .Desc | faint }}",
			Selected: "{{ .Label | green }}",
			Details:  "{{ .Desc }}",
		},
		Searcher: func(input string, index int) bool {
			item := config.Items[index]
			return contains(item.Label, input) || contains(item.Desc, input)
		},
	}

	// Show prompt and get result
	index, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	// Return the selected value
	selectedItem := config.Items[index]
	return selectedItem.Value, nil
}

// ShowInput prompts for user input
func ShowInput(label, defaultValue string) (string, error) {
	prompt := promptui.Prompt{
		Label:   label,
		Default: defaultValue,
		Validate: func(input string) error {
			if len(input) == 0 {
				return fmt.Errorf("input cannot be empty")
			}
			return nil
		},
	}

	return prompt.Run()
}

// ShowConfirm prompts for yes/no confirmation
func ShowConfirm(label string) (bool, error) {
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
		Default:   "y",
	}

	result, err := prompt.Run()
	if err != nil {
		return false, err
	}

	return result == "y" || result == "Y" || result == "yes", nil
}

// ShowPassword prompts for password input (hidden)
func ShowPassword(label string) (string, error) {
	prompt := promptui.Prompt{
		Label: label,
		Mask:  '*',
		Validate: func(input string) error {
			if len(input) == 0 {
				return fmt.Errorf("password cannot be empty")
			}
			return nil
		},
	}

	return prompt.Run()
}

// PauseForUser waits for user to press Enter
func PauseForUser(message string) {
	if message == "" {
		message = "Press Enter to continue..."
	}
	
	fmt.Print(Info(message))
	fmt.Scanln()
}

// ClearScreen clears the terminal screen
func ClearScreen() {
	// ANSI escape sequence to clear screen
	fmt.Print("\033[2J\033[H")
}

// ShowHeader displays the application header
func ShowHeader() {
	ClearScreen()
	PrintBanner()
	fmt.Println()
	PrintInfo("Welcome to NetInfo - Network Information & Diagnostics Tool")
	PrintSeparator()
}

// ShowGoodbye displays the goodbye message
func ShowGoodbye() {
	fmt.Println()
	PrintSuccess("Thank you for using NetInfo!")
	PrintInfo("Network diagnostics completed successfully.")
	fmt.Println()
}

// Helper function to check if string contains substring (case insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (s[:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr || 
		     containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}