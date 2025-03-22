package input

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// KeyMatches checks if a key message matches a key binding
func KeyMatches(msg tea.KeyMsg, binding key.Binding) bool {
	return key.Matches(msg, binding)
}

// GetRatingFromKey extracts a rating (0-5) from a key message
func GetRatingFromKey(msg tea.KeyMsg, keys KeyMap) (int, bool) {
	switch {
	case KeyMatches(msg, keys.Rate0):
		return 0, true
	case KeyMatches(msg, keys.Rate1):
		return 1, true
	case KeyMatches(msg, keys.Rate2):
		return 2, true
	case KeyMatches(msg, keys.Rate3):
		return 3, true
	case KeyMatches(msg, keys.Rate4):
		return 4, true
	case KeyMatches(msg, keys.Rate5):
		return 5, true
	default:
		return -1, false
	}
}

// IsNavKey returns true if the key is a navigation key
func IsNavKey(msg tea.KeyMsg) bool {
	return msg.String() == "up" ||
		msg.String() == "down" ||
		msg.String() == "left" ||
		msg.String() == "right" ||
		msg.String() == "home" ||
		msg.String() == "end" ||
		msg.String() == "pgup" ||
		msg.String() == "pgdown"
}

// IsEnterKey returns true if the key is Enter
func IsEnterKey(msg tea.KeyMsg) bool {
	return msg.String() == "enter"
}

// IsEscapeKey returns true if the key is Escape
func IsEscapeKey(msg tea.KeyMsg) bool {
	return msg.String() == "esc"
}
