// internal/ui/tui/welcome_keys.go

package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

// WelcomeKeyMap defines keybindings for the welcome screen
type WelcomeKeyMap struct {
	Enter key.Binding
	Quit  key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k WelcomeKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter, k.Quit}
}

// FullHelp returns keybindings for the expanded help view.
func (k WelcomeKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Enter, k.Quit},
	}
}

// DefaultWelcomeKeyMap returns the default keybindings for the welcome screen
func DefaultWelcomeKeyMap() WelcomeKeyMap {
	return WelcomeKeyMap{
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "browse decks"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "quit"),
		),
	}
}
