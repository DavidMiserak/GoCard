// internal/ui/tui/deck_list_keys.go

package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

type DeckListKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	Back    key.Binding
	Study   key.Binding
	Refresh key.Binding
	Help    key.Binding
	Quit    key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k DeckListKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Back, k.Study, k.Quit}
}

// FullHelp returns keybindings for the expanded help view.
func (k DeckListKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.Back, k.Study, k.Refresh},
		{k.Help, k.Quit},
	}
}

func DefaultDeckListKeyMap() DeckListKeyMap {
	return DeckListKeyMap{
		Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
		Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
		Enter:   key.NewBinding(key.WithKeys("enter", "right"), key.WithHelp("enter/→", "open deck")),
		Back:    key.NewBinding(key.WithKeys("backspace", "left", "h"), key.WithHelp("←/h/backspace", "go back")),
		Study:   key.NewBinding(key.WithKeys("s", "space"), key.WithHelp("s/space", "study deck")),
		Refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
		Quit:    key.NewBinding(key.WithKeys("ctrl+c", "q"), key.WithHelp("q/ctrl+c", "quit")),
	}
}
