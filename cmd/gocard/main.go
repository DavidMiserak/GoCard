// File: cmd/gocard/main.go
package main

import (
	"fmt"
	"log"

	"github.com/DavidMiserak/GoCard/internal/config"
	"github.com/DavidMiserak/GoCard/internal/storage"
	"github.com/DavidMiserak/GoCard/internal/ui"
)

func main() {
	// Parse command line flags
	opts, err := parseFlags()
	if err != nil {
		log.Fatalf("Error parsing command-line flags: %v", err)
	}

	// Initialize configuration
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		// Non-fatal, we'll use defaults
		fmt.Printf("Warning: Failed to load configuration: %v\n", err)
		cfg = config.Default()
	}

	// Apply command-line overrides to config
	applyFlagOverrides(cfg, opts)

	// Create or use the directory for flashcards
	cardsDir := getCardsDirectory(opts, cfg)
	fmt.Printf("Using cards directory: %s\n", cardsDir)

	// Initialize our card store
	store, err := storage.NewCardStore(cardsDir)
	if err != nil {
		log.Fatalf("Failed to initialize card store: %v", err)
	}

	// Configure logging
	configureLogging(store, opts.Verbose, cfg.Logging.Level)

	// Ensure we clean up resources when the program exits
	defer store.Close()

	// Check if this is the first run
	isFirstRunApp := isFirstRun(cardsDir, cfg)

	// If it's the first run or example mode is enabled, create example content
	if isFirstRunApp || opts.ExampleMode {
		// Create example content
		fmt.Println("Creating example content...")
		if err := createExampleContent(store); err != nil {
			fmt.Printf("Warning: Failed to create example content: %v\n", err)
		}

		// Handle first run onboarding
		if isFirstRunApp {
			handleFirstRun(store, cfg, opts.UseTUI)
		}
	}

	// If TUI mode is enabled, launch the terminal UI
	if opts.UseTUI {
		fmt.Printf("Starting GoCard terminal UI with cards from: %s\n", cardsDir)

		// Start with tutorial if this is first run
		startWithTutorial := isFirstRunApp

		if err := ui.RunTUI(store, startWithTutorial); err != nil {
			log.Fatalf("Error running terminal UI: %v", err)
		}
		return
	}

	// Otherwise, run the CLI mode
	runCLIMode(store)
}
