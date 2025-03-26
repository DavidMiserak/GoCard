// internal/ui/tui/deck_list_keys.go

package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

type DeckListKeyMap struct {
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
	Back  key.Binding
	Quit  key.Binding
}

func DefaultDeckListKeyMap() DeckListKeyMap {
	return DeckListKeyMap{
		Up:    key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
		Down:  key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
		Enter: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select deck")),
		Back:  key.NewBinding(key.WithKeys("backspace", "h"), key.WithHelp("backspace/h", "go back")),
		Quit:  key.NewBinding(key.WithKeys("ctrl+c", "q"), key.WithHelp("q/ctrl+c", "quit")),
	}
}
