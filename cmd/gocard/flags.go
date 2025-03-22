// cmd/gocard/flags.go - Command-line flag handling
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DavidMiserak/GoCard/internal/config"
	"github.com/DavidMiserak/GoCard/internal/storage"
	"github.com/DavidMiserak/GoCard/internal/storage/io"
)

// Options holds the command-line options
type Options struct {
	UseTUI      bool
	ExampleMode bool
	Verbose     bool
	ConfigPath  string
	CardsDir    string
	ShowHelp    bool
	ShowVersion bool
}

// Version information - should be set during build
var (
	Version   = "0.1.0"
	BuildDate = "unknown"
	GitCommit = "unknown"
)

// parseFlags parses command-line flags and returns the options
func parseFlags() (*Options, error) {
	opts := &Options{}

	// Define command-line flags
	flag.BoolVar(&opts.UseTUI, "tui", true, "Use terminal UI mode")
	flag.BoolVar(&opts.ExampleMode, "example", false, "Run in example mode with sample cards")
	flag.BoolVar(&opts.Verbose, "verbose", false, "Enable verbose logging")
	flag.StringVar(&opts.ConfigPath, "config", "", "Path to configuration file (default: ~/.gocard.yaml)")
	flag.BoolVar(&opts.ShowHelp, "h", false, "Show help information")
	flag.BoolVar(&opts.ShowHelp, "help", false, "Show help information")
	flag.BoolVar(&opts.ShowVersion, "version", false, "Show version information")

	// Parse the flags
	flag.Parse()

	// Show help if requested
	if opts.ShowHelp {
		printUsage()
		os.Exit(0)
	}

	// Show version if requested
	if opts.ShowVersion {
		printVersion()
		os.Exit(0)
	}

	// Handle positional arguments (cards directory)
	if flag.NArg() > 0 {
		opts.CardsDir = flag.Arg(0)
	}

	return opts, nil
}

// printUsage prints the usage information
func printUsage() {
	fmt.Println(`Usage: gocard [options] [cards_directory]

GoCard is a file-based spaced repetition system built in Go.

Options:
  -tui         Use terminal UI mode (default: true)
  -example     Run in example mode with sample cards
  -verbose     Enable detailed logging (useful for debugging)
  -config PATH Path to configuration file (default: ~/.gocard.yaml)
  -h, -help    Show help information
  -version     Show version information

Arguments:
  cards_directory  Directory for flashcards (default: ~/GoCard)

For more information, visit: https://github.com/DavidMiserak/GoCard`)
}

// printVersion prints the version information
func printVersion() {
	fmt.Printf("GoCard v%s\n", Version)
	fmt.Printf("Build Date: %s\n", BuildDate)
	fmt.Printf("Git Commit: %s\n", GitCommit)
}

// getCardsDirectory determines the cards directory based on options and config
func getCardsDirectory(opts *Options, cfg *config.Config) string {
	// Command-line option takes precedence
	if opts.CardsDir != "" {
		return opts.CardsDir
	}

	// Then config file
	if cfg.CardsDir != "" {
		return cfg.CardsDir
	}

	// Finally, use default
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// If we can't get the home directory, use the current directory
		return "GoCard"
	}
	return filepath.Join(homeDir, "GoCard")
}

// applyFlagOverrides applies command-line flag overrides to the config
func applyFlagOverrides(cfg *config.Config, opts *Options) {
	// Override verbose logging
	if opts.Verbose {
		cfg.Logging.Level = "debug"
	}
}

// configureLogging configures the logging based on options and config
func configureLogging(store storage.CardStoreInterface, verbose bool, logLevel string) {
	// Configure logging based on mode
	if verbose {
		// Enable debug logging in verbose mode
		store.SetLogLevel(io.DEBUG)
	} else {
		// Set log level from config
		switch logLevel {
		case "debug":
			store.SetLogLevel(io.DEBUG)
		case "info":
			store.SetLogLevel(io.INFO)
		case "warn":
			store.SetLogLevel(io.WARN)
		case "error":
			store.SetLogLevel(io.ERROR)
		default:
			store.SetLogLevel(io.INFO)
		}
	}
}
