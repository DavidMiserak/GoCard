// File: internal/ui/stats_screen.go

package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DavidMiserak/GoCard/internal/data"
)

// StatisticsScreen represents the statistics view
type StatisticsScreen struct {
	store     *data.Store
	width     int
	height    int
	activeTab int
	cardStats []int // Cards studied per day for the last 5 days
}

// NewStatisticsScreen creates a new statistics screen
func NewStatisticsScreen(store *data.Store) *StatisticsScreen {
	return &StatisticsScreen{
		store:     store,
		activeTab: 0,
		cardStats: calculateCardStudiedPerDay(store),
	}
}

// calculateCardStudiedPerDay calculates cards studied per day for the last 5 days
func calculateCardStudiedPerDay(store *data.Store) []int {
	// This is a placeholder implementation
	// In a real app, you'd track actual study history
	return []int{35, 15, 25, 40, 20}
}

// Init initializes the statistics screen
func (s *StatisticsScreen) Init() tea.Cmd {
	return nil
}

// Update handles user input for the statistics screen
func (s *StatisticsScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit
		case "b":
			// Return to main menu
			return NewMainMenu(), nil
		case "tab":
			// Cycle through tabs
			s.activeTab = (s.activeTab + 1) % 3
		}

	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
	}

	return s, nil
}

// View renders the statistics screen
func (s *StatisticsScreen) View() string {
	var sb strings.Builder

	// Title
	sb.WriteString(statTitleStyle.Render("Statistics"))
	sb.WriteString("\n\n")

	// Tabs
	tabs := []string{"Summary", "Deck Review", "Review Forecast"}
	tabRow := ""
	for i, tab := range tabs {
		if i == s.activeTab {
			tabRow += activeTabStyle.Render(tab) + " "
		} else {
			tabRow += tabStyle.Render(tab) + " "
		}
	}
	sb.WriteString(tabRow)
	sb.WriteString("\n\n")

	// Render the active tab
	switch s.activeTab {
	case 0:
		sb.WriteString(renderSummaryStats(s.store))
	case 1:
		sb.WriteString(renderDeckReviewStats(s.store))
	case 2:
		sb.WriteString(renderReviewForecastStats(s.store))
	}

	sb.WriteString("\n\n")

	// Help text
	helpText := statLabelStyle.Render("Tab: Switch View" + "\tb: Back to Main Menu" + "\tq: Quit")
	sb.WriteString(helpText)

	return sb.String()
}
