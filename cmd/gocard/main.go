// File: internal/main.go

package main

import (
	"fmt"
	"os"

	"github.com/DavidMiserak/GoCard/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Initialize the main menu
	p := tea.NewProgram(ui.NewMainMenu(), tea.WithAltScreen())

	// Start the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
