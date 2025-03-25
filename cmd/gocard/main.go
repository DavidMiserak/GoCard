// cmd/gocard/main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DavidMiserak/GoCard/internal/service/storage"
	"github.com/DavidMiserak/GoCard/pkg/algorithm"
)

// Default configuration values
const (
	defaultCardsDir = "~/GoCard"
)

func main() {
	// Parse command-line flags
	cardsDir := flag.String("d", defaultCardsDir, "Cards directory")
	flag.Parse()

	// Expand home directory if needed
	if *cardsDir == defaultCardsDir || strings.HasPrefix(*cardsDir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			os.Exit(1)
		}
		*cardsDir = filepath.Join(home, strings.TrimPrefix(*cardsDir, "~/"))
	}

	// Initialize storage
	storageService := storage.NewFileSystemStorage()
	if err := storageService.Initialize(*cardsDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
		os.Exit(1)
	}
	defer storageService.Close()

	// Create an SM2 algorithm instance
	sm2 := algorithm.NewSM2Algorithm()

	// Basic info display
	fmt.Println("GoCard - Spaced Repetition System")
	fmt.Println("=================================")
	fmt.Printf("Cards directory: %s\n\n", *cardsDir)

	// List decks
	deckPaths, err := storageService.ListDeckPaths(*cardsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing decks: %v\n", err)
		os.Exit(1)
	}

	if len(deckPaths) == 0 {
		fmt.Println("No decks found. Create a directory structure with markdown files to get started.")
		fmt.Println("Example structure:")
		fmt.Println("  ~/GoCard/")
		fmt.Println("  ├── Programming/")
		fmt.Println("  │   ├── Go/")
		fmt.Println("  │   │   ├── concurrency.md")
		fmt.Println("  │   │   └── interfaces.md")
		fmt.Println("  │   └── Python/")
		fmt.Println("  │       └── generators.md")
		fmt.Println("  └── Languages/")
		fmt.Println("      └── Spanish/")
		fmt.Println("          └── verbs.md")
		os.Exit(0)
	}

	fmt.Println("Available decks:")
	for i, deckPath := range deckPaths {
		deck, err := storageService.LoadDeck(deckPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading deck %s: %v\n", deckPath, err)
			continue
		}

		// Count cards in the deck
		cardPaths, err := storageService.ListCardPaths(deckPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing cards in deck %s: %v\n", deckPath, err)
			continue
		}

		// Count due cards
		dueCount := 0
		for _, cardPath := range cardPaths {
			card, err := storageService.LoadCard(cardPath)
			if err != nil {
				continue
			}
			if sm2.IsDue(card) {
				dueCount++
			}
		}

		fmt.Printf("%d. %s (%d cards, %d due)\n", i+1, deck.Name, len(cardPaths), dueCount)
	}

	fmt.Println("\nThis is a simple demo showing the basic structure.")
	fmt.Println("Run 'go build' to build the full application.")
}
