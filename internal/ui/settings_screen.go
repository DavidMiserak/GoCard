// File: internal/ui/settings_screen.go

package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/DavidMiserak/GoCard/internal/data"
	"github.com/DavidMiserak/GoCard/internal/model"
)

// SettingsKeyMap defines key bindings for settings screen
type settingsKeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
	Back  key.Binding
	Quit  key.Binding
}

var settingsKeys = settingsKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "navigate"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "navigate"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// SettingsScreen represents the settings menu model
type SettingsScreen struct {
	store          *data.Store
	selectedDeck   *model.Deck
	cursor         int
	algorithms     []string
	confirmingToggle bool
	toggleMessage  string
	width          int
	height         int
}

// NewSettingsScreen creates a new settings screen
func NewSettingsScreen(store *data.Store) *SettingsScreen {
	// For now, show settings for the first deck (if available)
	var selectedDeck *model.Deck
	if len(store.Decks) > 0 {
		selectedDeck = &store.Decks[0]
	}

	return &SettingsScreen{
		store:        store,
		selectedDeck: selectedDeck,
		algorithms:   []string{"SM2", "FSRS"},
		cursor:       0,
	}
}

// Init initializes the settings screen
func (s SettingsScreen) Init() tea.Cmd {
	return nil
}

// Update handles user input
func (s SettingsScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if s.confirmingToggle {
			// Handle confirmation dialog
			switch {
			case key.Matches(msg, settingsKeys.Enter):
				// Confirm toggle
				if s.selectedDeck != nil {
					success, converted, err := s.store.ToggleDeckAlgorithm(s.selectedDeck.ID, s.algorithms[s.cursor])
					if success {
						s.toggleMessage = fmt.Sprintf("✓ Switched to %s. Converted %d cards.", s.algorithms[s.cursor], converted)
						s.confirmingToggle = false
					} else {
						s.toggleMessage = fmt.Sprintf("✗ Toggle failed: %v", err)
						s.confirmingToggle = false
					}
				}

			case key.Matches(msg, settingsKeys.Back):
				// Cancel toggle
				s.confirmingToggle = false
				s.toggleMessage = ""
			}
		} else {
			// Normal settings navigation
			switch {
			case key.Matches(msg, settingsKeys.Quit):
				return NewMainMenu(s.store), nil

			case key.Matches(msg, settingsKeys.Back):
				return NewMainMenu(s.store), nil

			case key.Matches(msg, settingsKeys.Up):
				if s.cursor > 0 {
					s.cursor--
				}

			case key.Matches(msg, settingsKeys.Down):
				if s.cursor < len(s.algorithms)-1 {
					s.cursor++
				}

			case key.Matches(msg, settingsKeys.Enter):
				// Show confirmation dialog
				if s.selectedDeck != nil {
					currentAlgorithm := s.selectedDeck.Algorithm
					if currentAlgorithm == "" {
						currentAlgorithm = "SM2"
					}

					if currentAlgorithm != s.algorithms[s.cursor] {
						s.confirmingToggle = true
						s.toggleMessage = fmt.Sprintf("Switch from %s to %s? (Existing data will be preserved.)\nEnter: Confirm, B: Cancel",
							currentAlgorithm, s.algorithms[s.cursor])
					}
				}
			}
		}

	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
	}

	return s, nil
}

// View renders the settings screen
func (s SettingsScreen) View() string {
	if s.selectedDeck == nil {
		return "No decks available. Please create a deck first.\n\n" +
			helpStyle.Render("b: Back | q: Quit")
	}

	output := ""
	output += titleStyle.Render("Settings")
	output += "\n\n"

	// Display current deck
	currentAlgorithm := s.selectedDeck.Algorithm
	if currentAlgorithm == "" {
		currentAlgorithm = "SM2"
	}
	output += fmt.Sprintf("Deck: %s\n", s.selectedDeck.Name)
	output += fmt.Sprintf("Current Algorithm: %s\n", currentAlgorithm)
	output += "\n"

	// Algorithm selection
	output += "Select Algorithm:\n"
	for i, algo := range s.algorithms {
		if i == s.cursor {
			output += selectedItemStyle.Render("  ● " + algo)
		} else {
			output += normalItemStyle.Render("  ○ " + algo)
		}
		output += "\n"
	}

	output += "\n"

	// Display messages
	if s.toggleMessage != "" {
		output += helpStyle.Render(s.toggleMessage)
		output += "\n\n"
	}

	if s.confirmingToggle {
		output += helpStyle.Render("↑/↓: Navigate | Enter: Confirm | b: Cancel | q: Quit")
	} else {
		output += helpStyle.Render("↑/↓: Navigate | Enter: Toggle | b: Back | q: Quit")
	}

	return output
}
