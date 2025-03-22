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
}

// initModel initializes the TUI model
func initModel(store *storage.CardStore) (TUIModel, error) {
	// Get terminal width and height
	width, height, err := getTerminalSize()
	if err != nil {
		return TUIModel{}, err
	}

	// Initialize keymap
	keys := input.NewKeyMap()

	// Initialize help model
	helpModel := help.New()
	helpModel.ShowAll = false

	// Create the main deck browser view as the starting view
	deckView, err := views.NewDeckBrowserView(store, "", width, height)
	if err != nil {
		return TUIModel{}, fmt.Errorf("failed to create deck browser view: %w", err)
	}

	// Initialize the model
	m := TUIModel{
		store:        store,
		keys:         keys,
		help:         helpModel,
		currentView:  deckView,
		previousView: nil,
		showHelp:     false,
		width:        width,
		height:       height,
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
		// Global key handlers
		switch {
		case input.KeyMatches(msg, m.keys.Quit):
			return m, tea.Quit

		case input.KeyMatches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil

		case input.KeyMatches(msg, m.keys.Back):
			// Handle going back to previous view if we're not in the main view
			if m.currentView.State() != views.ViewDeckBrowser && m.previousView != nil {
				m.currentView = m.previousView
				m.previousView = nil
				return m, nil
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
func RunTUI(store *storage.CardStore) error {
	m, err := initModel(store)
	if err != nil {
		return fmt.Errorf("failed to initialize TUI model: %w", err)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	_, err = p.Run()
	return err
}
