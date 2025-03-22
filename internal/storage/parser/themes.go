// File: internal/storage/parser/themes.go
package parser

import (
	"github.com/alecthomas/chroma/v2/styles"
)

// AvailableThemes returns a list of available syntax highlighting themes
func AvailableThemes() []string {
	return styles.Names()
}

// The following are common themes that are useful for various scenarios:
const (
	// Light themes
	SyntaxThemeGithub         = "github"
	SyntaxThemeFriendly       = "friendly"
	SyntaxThemeParaisoLight   = "paraiso-light"
	SyntaxThemeSolarizedLight = "solarized-light"

	// Dark themes
	SyntaxThemeMonokai       = "monokai"
	SyntaxThemeDracula       = "dracula"
	SyntaxThemeSolarizedDark = "solarized-dark"
	SyntaxThemeNord          = "nord"
	SyntaxThemeVS            = "vs"

	// High contrast themes
	SyntaxThemeAbap   = "abap"   // High contrast light
	SyntaxThemeNative = "native" // High contrast dark
)

// GetThemeForTerminal returns an appropriate syntax theme for the terminal
// based on terminal background color (light or dark)
func GetThemeForTerminal(isDarkBackground bool) string {
	if isDarkBackground {
		return SyntaxThemeMonokai
	}
	return SyntaxThemeGithub
}

// IsDarkTheme returns true if the given theme is dark
func IsDarkTheme(theme string) bool {
	switch theme {
	case SyntaxThemeMonokai, SyntaxThemeDracula, SyntaxThemeSolarizedDark, SyntaxThemeNord:
		return true
	default:
		return false
	}
}

// GetHighContrastTheme returns a high contrast theme
func GetHighContrastTheme(isDarkBackground bool) string {
	if isDarkBackground {
		return SyntaxThemeNative
	}
	return SyntaxThemeAbap
}

// ThemeDescription returns a description of the given theme
func ThemeDescription(theme string) string {
	switch theme {
	case SyntaxThemeGithub:
		return "Github's light theme (light background)"
	case SyntaxThemeFriendly:
		return "Friendly light theme with good contrast"
	case SyntaxThemeParaisoLight:
		return "Paraiso light theme with pastel colors"
	case SyntaxThemeSolarizedLight:
		return "Solarized light theme with low contrast"
	case SyntaxThemeMonokai:
		return "Monokai dark theme with bright colors"
	case SyntaxThemeDracula:
		return "Dracula dark theme with purple accents"
	case SyntaxThemeSolarizedDark:
		return "Solarized dark theme with low contrast"
	case SyntaxThemeNord:
		return "Nord dark theme with cool blue colors"
	case SyntaxThemeVS:
		return "Visual Studio-inspired theme"
	case SyntaxThemeAbap:
		return "ABAP high contrast light theme"
	case SyntaxThemeNative:
		return "Native high contrast dark theme"
	default:
		return "Custom theme"
	}
}
