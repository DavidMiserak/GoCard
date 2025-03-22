// Package input handles user input and key bindings for the UI.
package input

import (
	"github.com/charmbracelet/bubbles/key"
)

// KeyMap defines the keybindings for the TUI
type KeyMap struct {
	ShowAnswer key.Binding
	Rate0      key.Binding
	Rate1      key.Binding
	Rate2      key.Binding
	Rate3      key.Binding
	Rate4      key.Binding
	Rate5      key.Binding
	Edit       key.Binding
	New        key.Binding
	Delete     key.Binding
	Tags       key.Binding
	Search     key.Binding
	ChangeDeck key.Binding
	CreateDeck key.Binding
	RenameDeck key.Binding
	DeleteDeck key.Binding
	MoveToDeck key.Binding
	Quit       key.Binding
	Help       key.Binding
	Back       key.Binding
}

// ShortHelp returns keybinding help
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.ShowAnswer, k.ChangeDeck, k.Edit, k.New, k.Quit, k.Help}
}

// FullHelp returns the full set of keybindings
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ShowAnswer, k.Rate0, k.Rate1, k.Rate2, k.Rate3, k.Rate4, k.Rate5},
		{k.ChangeDeck, k.CreateDeck, k.RenameDeck, k.DeleteDeck, k.MoveToDeck},
		{k.Edit, k.New, k.Delete, k.Tags, k.Search},
		{k.Back, k.Quit, k.Help},
	}
}

// NewKeyMap creates the default keybindings
func NewKeyMap() KeyMap {
	return KeyMap{
		ShowAnswer: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "show answer"),
		),
		Rate0: key.NewBinding(
			key.WithKeys("0"),
			key.WithHelp("0", "rate: again"),
		),
		Rate1: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "rate: hard"),
		),
		Rate2: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "rate: difficult"),
		),
		Rate3: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "rate: good"),
		),
		Rate4: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "rate: easy"),
		),
		Rate5: key.NewBinding(
			key.WithKeys("5"),
			key.WithHelp("5", "rate: very easy"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit card"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new card"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete card"),
		),
		Tags: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "edit tags"),
		),
		Search: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "search"),
		),
		ChangeDeck: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "change deck"),
		),
		CreateDeck: key.NewBinding(
			key.WithKeys("C"),
			key.WithHelp("C", "create deck"),
		),
		RenameDeck: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "rename deck"),
		),
		DeleteDeck: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "delete deck"),
		),
		MoveToDeck: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "move to deck"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "backspace"),
			key.WithHelp("esc", "go back"),
		),
	}
}
