// cmd/gocard/main.go - Slim entry point
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

	// If TUI mode is enabled, launch the terminal UI
	if opts.UseTUI {
		fmt.Printf("Starting GoCard terminal UI with cards from: %s\n", cardsDir)

		// Start with tutorial if this is first run
		startWithTutorial := isFirstRun(cardsDir, cfg)

		if err := ui.RunTUI(store, startWithTutorial); err != nil {
			log.Fatalf("Error running terminal UI: %v", err)
		}
		return
	}

	// Otherwise, run the CLI mode
	runCLIMode(store)
}
