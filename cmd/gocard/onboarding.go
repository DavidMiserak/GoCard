// cmd/gocard/onboarding.go - Fixed by uncommenting required functions
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DavidMiserak/GoCard/internal/config"
	"github.com/DavidMiserak/GoCard/internal/storage"
	"github.com/DavidMiserak/GoCard/internal/storage/io"
)

// isFirstRun checks if this appears to be the first time running GoCard
func isFirstRun(cardsDir string, cfg *config.Config) bool {
	// If first run flag is explicitly set in config
	if cfg.FirstRun {
		return true
	}

	// Check if cards directory exists and is empty
	dirExists, err := io.DirectoryExists(cardsDir)
	if err != nil || !dirExists {
		return true
	}

	// Check if directory is empty (no card files)
	isEmpty, _ := isDirEmpty(cardsDir)
	return isEmpty
}

// isDirEmpty returns true if directory is empty or contains only hidden files
func isDirEmpty(dirPath string) (bool, error) {
	f, err := os.Open(dirPath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// Read directory entries
	files, err := f.Readdir(-1)
	if err != nil {
		return false, err
	}

	// Check for any non-hidden files
	for _, file := range files {
		// Skip hidden files and directories
		if !strings.HasPrefix(file.Name(), ".") {
			return false, nil
		}
	}

	return true, nil
}

// handleFirstRun handles the first run experience
func handleFirstRun(store storage.CardStoreInterface, cfg *config.Config, useTUI bool) {
	// Create onboarding content
	createOnboardingContent(store)

	// Show welcome message and tutorial only in CLI mode or if explicitly requested
	if !useTUI {
		showWelcomeMessage()
	}
	// In TUI mode, we'll start with the tutorial view directly

	// Update config to indicate that first run has been completed
	cfg.FirstRun = false
	if err := config.Save(cfg); err != nil {
		fmt.Printf("Warning: Failed to save configuration: %v\n", err)
	}
}

// showWelcomeMessage displays a welcome message for first-time users
func showWelcomeMessage() {
	fmt.Println(`
		┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
		┃                Welcome to GoCard!                   ┃
		┃                                                     ┃
		┃ The file-based spaced repetition system for         ┃
		┃ developers and text-oriented learners.              ┃
		┃                                                     ┃
		┃ We've created sample cards to help you get started. ┃
		┃ Use arrow keys to navigate and press ? for help.    ┃
		┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛`)

	// Ask if user wants to see a quick tutorial
	fmt.Println("Would you like to see a quick tutorial? (y/n)")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading input:", err)
		// Depending on context, you might want to handle this differently
	}

	input = strings.TrimSpace(strings.ToLower(input))

	if input == "y" || input == "yes" {
		showQuickTutorial()
	} else {
		fmt.Println("You can always access help by pressing ? during use.")
		fmt.Println("Starting GoCard...")
		time.Sleep(1 * time.Second)
	}
}

// showQuickTutorial displays a quick tutorial for first-time users
func showQuickTutorial() {
	tutorials := []struct {
		title   string
		content string
	}{
		{
			"Basic Navigation",
			"- Use arrow keys to navigate\n- Press Enter to select\n- Press Esc to go back\n- Press q to quit",
		},
		{
			"Reviewing Cards",
			"- Press space to reveal answer\n- Rate your recall from 0-5\n  (0: Completely forgot, 5: Perfect recall)\n- Cards will reappear based on your performance",
		},
		{
			"Creating Cards",
			"- Press n to create a new card\n- Cards are saved as Markdown files\n- You can edit cards with any text editor",
		},
		{
			"Directory Organization",
			"- Directories = Decks\n- Press c to change decks\n- Press C to create a new deck",
		},
	}

	fmt.Println("\nQuick Tutorial:")

	for i, tutorial := range tutorials {
		fmt.Printf("\n[%d/%d] %s\n", i+1, len(tutorials), tutorial.title)
		fmt.Println(tutorial.content)

		if i < len(tutorials)-1 {
			fmt.Println("\nPress Enter to continue...")
			reader := bufio.NewReader(os.Stdin)
			if _, err := reader.ReadString('\n'); err != nil {
				fmt.Println("Error reading input:", err)
				// Depending on context, you might want to handle this differently
			}
		}
	}

	fmt.Println("\nTutorial complete! Starting GoCard...")
	time.Sleep(1 * time.Second)
}

// createOnboardingContent creates initial content for first-time users
func createOnboardingContent(store storage.CardStoreInterface) {
	// Create Getting Started deck
	gettingStartedDeck, err := store.CreateDeck("Getting Started", nil)
	if err != nil {
		fmt.Printf("Warning: Failed to create Getting Started deck: %v\n", err)
		return
	}

	// Create Quick Start Guide card
	quickStartContent := `# GoCard Quick Start Guide

	Welcome to GoCard, the file-based spaced repetition system for developers and text-oriented learners!

	## Basic Concepts

	- **Cards** are individual flashcards stored as Markdown files
	- **Decks** are directories that organize your cards
	- **Spaced Repetition** schedules reviews based on how well you remember each card

	## Getting Started

	1. **Navigate Decks**
	- Press ctrl+o to browse decks
	- Use arrow keys (or j/k) to navigate
	- Press Enter to select a deck
	- Press Esc to go back

	2. **Review Cards**
	- Press Space to start reviewing cards
	- Press Space to reveal the answer
	- Rate your recall from 0-5

	3. **Create Cards**
	- Press ctrl+n to create a new card
	- Or create a markdown file directly in the deck directory

	4. **Get Help**
	- Press ctrl+h at any time to see all keyboard shortcuts

	Happy learning with GoCard!`

	_, err = store.CreateCardInDeck(
		"Quick Start Guide",
		"How do I get started with GoCard?",
		quickStartContent,
		[]string{"gocard", "tutorial"},
		gettingStartedDeck,
	)
	if err != nil {
		fmt.Printf("Warning: Failed to create Quick Start Guide card: %v\n", err)
	}

	// Create Keyboard Shortcuts card
	keyboardShortcutsContent := `# GoCard Keyboard Shortcuts

	| Key               | Action                     |
	|-------------------|----------------------------|
	| Space             | Show answer                |
	| 0-5               | Rate card difficulty       |
	| ctrl+o            | Change deck                |
	| ctrl+alt+n        | Create new deck            |
	| ctrl+n            | Create new card            |
	| ctrl+e            | Edit current card          |
	| ctrl+x d          | Delete current card        |
	| ctrl+f            | Search cards               |
	| ctrl+h/F1         | Toggle help                |
	| ctrl+q            | Quit                       |
	| ↑/k               | Move up in lists           |
	| ↓/j               | Move down in lists         |
	| Enter             | Select/move forward        |
	| Esc               | Go back                    |`

	_, err = store.CreateCardInDeck(
		"Keyboard Shortcuts",
		"What are the keyboard shortcuts in GoCard?",
		keyboardShortcutsContent,
		[]string{"gocard", "tutorial", "shortcuts"},
		gettingStartedDeck,
	)
	if err != nil {
		fmt.Printf("Warning: Failed to create Keyboard Shortcuts card: %v\n", err)
	}

	// Create Card Format card - Using string constants to handle the nested backticks
	const markdownStart = "```markdown"
	const markdownEnd = "```"

	cardFormatContent := `# Card Format

	GoCard uses plain Markdown files with YAML frontmatter:

	` + markdownStart + `
	---
	tags: tag1, tag2, tag3
	created: YYYY-MM-DD
	last_reviewed: YYYY-MM-DD
	review_interval: N
	difficulty: 0-5
	---

	# Card Title

	## Question

	Your question goes here. This can be multiline and include any markdown.

	## Answer

	Your answer goes here. This can include:
	- Lists
	- Code blocks
	- Images
	- Tables
	- And any other markdown formatting
	` + markdownEnd + `

	You can edit these files directly with any text editor, and GoCard will automatically detect changes.`

	_, err = store.CreateCardInDeck(
		"Card Format",
		"What is the file format for GoCard cards?",
		cardFormatContent,
		[]string{"gocard", "tutorial", "format"},
		gettingStartedDeck,
	)
	if err != nil {
		fmt.Printf("Warning: Failed to create Card Format card: %v\n", err)
	}

	// Create a sample .gocard.yaml file in the user's home directory if it doesn't exist
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(homeDir, ".gocard.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// Get default cards directory path properly using os.UserHomeDir
			defaultCardsDir := filepath.Join(homeDir, "GoCard")
			defaultLogPath := filepath.Join(homeDir, ".gocard.log")

			// Create the sample config with proper paths
			sampleConfig := fmt.Sprintf(`# GoCard Configuration File
				# This file controls the behavior of GoCard

				# Cards directory (default: %s)
				cards_dir: "%s"

				# Logging settings
				logging:
				# Log level: debug, info, warn, error
				level: "info"
				# Enable file logging
				file_enabled: false
				# Log file path
				file_path: "%s"

				# UI settings
				ui:
				# Theme: auto, light, dark
				theme: "auto"
				# Code highlighting theme
				highlight_theme: "monokai"
				# Show line numbers in code blocks
				show_line_numbers: true

				# Spaced repetition settings
				spaced_repetition:
				# Bonus for easy cards (higher = longer intervals)
				easy_bonus: 1.3
				# Global interval modifier (higher = longer intervals)
				interval_modifier: 1.0
				# Maximum interval in days
				max_interval: 365
				# Number of new cards per day
				new_cards_per_day: 20
				`,
				defaultCardsDir, defaultCardsDir, defaultLogPath)

			err = os.WriteFile(configPath, []byte(sampleConfig), 0644)
			if err != nil {
				fmt.Printf("Warning: Failed to create sample config file: %v\n", err)
			}
		}
	}
}
