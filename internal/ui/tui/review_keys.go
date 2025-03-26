// internal/ui/tui/review_keys.go

package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

// ReviewKeyMap defines keybindings for the review screen
type ReviewKeyMap struct {
	Space key.Binding
	Again key.Binding
	Hard  key.Binding
	Good  key.Binding
	Easy  key.Binding
	Skip  key.Binding
	Help  key.Binding
	Quit  key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k ReviewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Space, k.Again, k.Hard, k.Good, k.Easy, k.Quit}
}

// FullHelp returns keybindings for the expanded help view.
func (k ReviewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Space},
		{k.Again, k.Hard, k.Good, k.Easy},
		{k.Skip, k.Help, k.Quit},
	}
}

// DefaultReviewKeyMap returns the default keybindings for the review screen
func DefaultReviewKeyMap() ReviewKeyMap {
	return ReviewKeyMap{
		Space: key.NewBinding(
			key.WithKeys("space"),
			key.WithHelp("space", "show answer / start rating"),
		),
		Again: key.NewBinding(
			key.WithKeys("0", "1"),
			key.WithHelp("0/1", "again/hard (forgot)"),
		),
		Hard: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "okay (recalled with difficulty)"),
		),
		Good: key.NewBinding(
			key.WithKeys("3", "4"),
			key.WithHelp("3/4", "good/easy (recalled)"),
		),
		Easy: key.NewBinding(
			key.WithKeys("5"),
			key.WithHelp("5", "perfect (effortless recall)"),
		),
		Skip: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "skip card"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc/q", "quit review"),
		),
	}
}
