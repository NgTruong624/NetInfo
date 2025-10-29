package display

import (
	"github.com/fatih/color"
)

// Color schemes for different types of output
var (
	// Header colors
	Header = color.New(color.FgCyan, color.Bold).SprintFunc()
	Title  = color.New(color.FgMagenta, color.Bold, color.Underline).SprintFunc()
	
	// Status colors
	Success = color.New(color.FgGreen, color.Bold).SprintFunc()
	Warning = color.New(color.FgYellow, color.Bold).SprintFunc()
	Error   = color.New(color.FgRed, color.Bold).SprintFunc()
	Info    = color.New(color.FgBlue, color.Bold).SprintFunc()
	
	// Data colors
	Primary   = color.New(color.FgWhite, color.Bold).SprintFunc()
	Secondary = color.New(color.FgHiWhite).SprintFunc()
	Muted     = color.New(color.FgHiBlack).SprintFunc()
	
	// Special colors
	Highlight = color.New(color.BgYellow, color.FgBlack, color.Bold).SprintFunc()
	URL       = color.New(color.FgBlue, color.Underline).SprintFunc()
	IP        = color.New(color.FgGreen).SprintFunc()
	MAC       = color.New(color.FgYellow).SprintFunc()
)

// PrintBanner prints the application banner in 8-bit style
func PrintBanner() {
	banner := `
███╗   ██╗███████╗████████╗██╗███╗   ██╗██╗     ███████╗ ██████╗ 
████╗  ██║██╔════╝╚══██╔══╝██║████╗  ██║██║     ██╔════╝██╔═══██╗
██╔██╗ ██║█████╗     ██║   ██║██╔██╗ ██║██║     █████╗  ██║   ██║
██║╚██╗██║██╔══╝     ██║   ██║██║╚██╗██║██║     ██╔══╝  ██║   ██║
██║ ╚████║███████╗   ██║   ██║██║ ╚████║███████╗███████╗╚██████╔╝
╚═╝  ╚═══╝╚══════╝   ╚═╝   ╚═╝╚═╝  ╚═══╝╚══════╝╚══════╝ ╚═════╝ 
                                                                  
███████╗ ██████╗  ██████╗ ██╗                                    
╚══██╔══╝██╔═══██╗██╔═══██╗██║                                    
   ██║   ██║   ██║██║   ██║██║                                    
   ██║   ██║   ██║██║   ██║██║                                    
   ██║   ╚██████╔╝╚██████╔╝███████╗                               
   ╚═╝    ╚═════╝  ╚═════╝ ╚══════╝                               
                                                                  
                   DIY by Etouuu                                  
`
	// 8-bit color scheme: bright cyan with some magenta accents
	color.New(color.FgCyan, color.Bold).Print(banner)
}

// PrintSeparator prints a separator line
func PrintSeparator() {
	color.New(color.FgHiBlack).Println("─────────────────────────────────────────────────────────")
}

// PrintSuccess prints a success message
func PrintSuccess(msg string) {
	color.New(color.FgGreen, color.Bold).Printf("✓ %s\n", msg)
}

// PrintWarning prints a warning message
func PrintWarning(msg string) {
	color.New(color.FgYellow, color.Bold).Printf("⚠ %s\n", msg)
}

// PrintError prints an error message
func PrintError(msg string) {
	color.New(color.FgRed, color.Bold).Printf("✗ %s\n", msg)
}

// PrintInfo prints an info message
func PrintInfo(msg string) {
	color.New(color.FgBlue, color.Bold).Printf("ℹ %s\n", msg)
}