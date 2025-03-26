// cmd/gocard/main.go

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DavidMiserak/GoCard/internal/service/card"
	"github.com/DavidMiserak/GoCard/internal/service/deck"
	"github.com/DavidMiserak/GoCard/internal/service/review"
	"github.com/DavidMiserak/GoCard/internal/service/storage"
	"github.com/DavidMiserak/GoCard/internal/ui/tui"
	"github.com/DavidMiserak/GoCard/pkg/algorithm"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Initialize storage service
	storageService := storage.NewFileSystemStorage()

	// Use default cards directory or a configurable path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting user home directory: %v\n", err)
		os.Exit(1)
	}
	cardsDir := filepath.Join(homeDir, "GoCard")

	if err := storageService.Initialize(cardsDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
		os.Exit(1)
	}
	defer storageService.Close()

	// Create dependencies
	alg := algorithm.NewSM2Algorithm()
	cardService := card.NewCardService(storageService, alg)
	deckService := deck.NewDeckService(storageService, cardService)
	reviewService := review.NewReviewService(storageService, cardService, deckService, alg)

	// Create the application model
	model := tui.NewAppModel(deckService, cardService, reviewService, storageService, cardsDir)

	// Start the Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
