// File: /internal/ui/tui.go

// Package ui contains the terminal user interface for GoCard.
package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"github.com/DavidMiserak/GoCard/internal/storage"
	"github.com/DavidMiserak/GoCard/internal/ui/input"
	"github.com/DavidMiserak/GoCard/internal/ui/views"
)

// TUIModel is the main model for the Terminal UI
type TUIModel struct {
	store        *storage.CardStore
	keys         input.KeyMap
	help         help.Model
	currentView  views.View
	previousView views.View
	showHelp     bool
	width        int
	height       int
	// Add a flag to track if we're in card edit mode
	editingCard bool
}

// initModel initializes the TUI model
func initModel(store *storage.CardStore, startWithTutorial bool) (TUIModel, error) {
	// Get terminal width and height
	width, height, err := getTerminalSize()
	if err != nil {
		return TUIModel{}, err
	}

	// Initialize keymap with enhanced editor keys
	keys := input.NewKeyMap()

	// Initialize help model
	helpModel := help.New()
	helpModel.ShowAll = false

	var initialView views.View

	if startWithTutorial {
		// Start with tutorial view for first-time users
		tutorialView, err := views.NewTutorialView(store, width, height)
		if err != nil {
			return TUIModel{}, fmt.Errorf("failed to create tutorial view: %w", err)
		}
		initialView = tutorialView
	} else {
		// Default to deck browser view for returning users
		deckView, err := views.NewDeckBrowserView(store, "", width, height)
		if err != nil {
			return TUIModel{}, fmt.Errorf("failed to create deck browser view: %w", err)
		}
		initialView = deckView
	}

	// Initialize the model
	m := TUIModel{
		store:        store,
		keys:         keys,
		help:         helpModel,
		currentView:  initialView,
		previousView: nil,
		showHelp:     false,
		width:        width,
		height:       height,
		editingCard:  false,
	}

	return m, nil
}

// getTerminalSize returns the terminal dimensions
func getTerminalSize() (width, height int, err error) {
	// Default fallback dimensions
	width, height = 80, 24

	// Try to get actual terminal size
	if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width, height = w, h
	}

	return width, height, nil
}

// Init initializes the TUI
func (m TUIModel) Init() tea.Cmd {
	return m.currentView.Init()
}

// Update handles events and updates the model
func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Check if we're currently editing a card
		if m.editingCard {
			// Let the card edit view handle all keyboard input
			// This ensures that our auto-save timer and other special handling works
			newView, cmd := m.currentView.Update(msg, m.keys)
			cmds = append(cmds, cmd)

			// If the view has changed (e.g., saved and exited), update our state
			if newView != m.currentView {
				m.previousView = m.currentView
				m.currentView = newView

				// Check if we're no longer in the edit view
				_, isEditView := m.currentView.(*views.CardEditView)
				m.editingCard = isEditView
			}

			return m, tea.Batch(cmds...)
		}

		// Global key handlers for non-editing views
		switch {
		case input.KeyMatches(msg, m.keys.Quit):
			return m, tea.Quit

		case input.KeyMatches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil

		case input.KeyMatches(msg, m.keys.Back):
			// Handle going back to previous view if we're not in the main view
			if m.currentView.State() != views.ViewDeckBrowser && m.previousView != nil {
				// Special handling for deck list view to ensure we return to proper deck browser
				if m.currentView.State() == views.ViewDeckList {
					// We should go back to deck browser for current deck
					// The deck list view will handle this internally
					// The view update below will take care of it
				} else {
					// For other views, switch back to previous view
					m.currentView = m.previousView
					m.previousView = nil
					return m, nil
				}
			}
		}

	case tea.WindowSizeMsg:
		// Handle window resize
		m.width = msg.Width
		m.height = msg.Height
	}

	// Let the current view handle the message
	var cmd tea.Cmd
	newView, cmd := m.currentView.Update(msg, m.keys)
	cmds = append(cmds, cmd)

	// If the view has changed, update our model
	if newView != m.currentView {
		m.previousView = m.currentView
		m.currentView = newView

		// Check if we're entering or leaving the edit view
		_, isEditView := m.currentView.(*views.CardEditView)
		m.editingCard = isEditView
	}

	return m, tea.Batch(cmds...)
}

// View renders the TUI
func (m TUIModel) View() string {
	// Render the current view
	content := m.currentView.Render(m.width, m.height)

	// If help is enabled, append the help view
	if m.showHelp {
		helpView := m.help.View(m.keys)
		content += "\n" + helpView
	}

	return content
}

// RunTUI starts the terminal UI
func RunTUI(store *storage.CardStore, startWithTutorial bool) error {
	m, err := initModel(store, startWithTutorial)
	if err != nil {
		return fmt.Errorf("failed to initialize TUI model: %w", err)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	_, err = p.Run()
	return err
}
