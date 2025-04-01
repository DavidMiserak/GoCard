// File: internal/main.go

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DavidMiserak/GoCard/internal/data"
	"github.com/DavidMiserak/GoCard/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Parse command-line flags
	var deckDir string
	defaultDir := filepath.Join(os.Getenv("HOME"), "GoCard")
	flag.StringVar(&deckDir, "dir", defaultDir, "Directory containing flashcard decks")
	flag.Parse()

	// Resolve tilde in path if present
	if deckDir == "~/GoCard" || deckDir == "~/GoCard/" {
		deckDir = defaultDir
	}

	// Initialize the store
	var store *data.Store

	// Check if directory exists and load decks from it
	if _, err := os.Stat(deckDir); os.IsNotExist(err) {
		fmt.Printf("Warning: Directory '%s' does not exist. Using default decks.\n", deckDir)
		store = data.NewStore() // Use default store with dummy data
	} else {
		// Load decks from the specified directory
		var err error
		store, err = data.NewStoreFromDir(deckDir)
		if err != nil {
			fmt.Printf("Error loading decks: %v\nUsing default decks instead.\n", err)
			store = data.NewStore() // Fallback to default store with dummy data
		}
	}

	// Initialize the main menu with the store
	p := tea.NewProgram(ui.NewMainMenu(store), tea.WithAltScreen())

	// Start the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
