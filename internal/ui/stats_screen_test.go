// File: internal/ui/stats_screen_test.go

package ui

import (
	"testing"

	"github.com/DavidMiserak/GoCard/internal/data"
	tea "github.com/charmbracelet/bubbletea"
)

func TestNewStatisticsScreen(t *testing.T) {
	store := data.NewStore()
	statsScreen := NewStatisticsScreen(store)

	if statsScreen == nil {
		t.Fatal("Expected NewStatisticsScreen to return a non-nil StatisticsScreen")
	}

	if statsScreen.store != store {
		t.Errorf("Expected store to be %v, got %v", store, statsScreen.store)
	}

	if statsScreen.activeTab != 0 {
		t.Errorf("Expected activeTab to be 0, got %d", statsScreen.activeTab)
	}

	if len(statsScreen.cardStats) == 0 {
		t.Error("Expected cardStats to be initialized with data")
	}
}

func TestStatisticsScreenInit(t *testing.T) {
	store := data.NewStore()
	statsScreen := NewStatisticsScreen(store)

	cmd := statsScreen.Init()

	if cmd != nil {
		t.Error("Expected Init to return nil cmd")
	}
}

func TestStatisticsScreenUpdate(t *testing.T) {
	store := data.NewStore()
	statsScreen := NewStatisticsScreen(store)

	// Test tab switching
	model, cmd := statsScreen.Update(tea.KeyMsg{Type: tea.KeyTab})
	updatedScreen := model.(*StatisticsScreen)

	if updatedScreen.activeTab != 1 {
		t.Errorf("Expected activeTab to be 1 after Tab key, got %d", updatedScreen.activeTab)
	}

	if cmd != nil {
		t.Error("Expected cmd to be nil")
	}

	// Test window size update
	width, height := 80, 24
	model, cmd = statsScreen.Update(tea.WindowSizeMsg{Width: width, Height: height})
	updatedScreen = model.(*StatisticsScreen)

	if updatedScreen.width != width {
		t.Errorf("Expected width to be %d, got %d", width, updatedScreen.width)
	}

	if updatedScreen.height != height {
		t.Errorf("Expected height to be %d, got %d", height, updatedScreen.height)
	}
}

func TestStatisticsScreenView(t *testing.T) {
	store := data.NewStore()
	statsScreen := NewStatisticsScreen(store)

	// Set screen dimensions
	statsScreen.width = 80
	statsScreen.height = 24

	// Test view for each tab
	for tab := 0; tab < 3; tab++ {
		statsScreen.activeTab = tab
		view := statsScreen.View()

		if view == "" {
			t.Errorf("Expected View to return non-empty string for tab %d", tab)
		}

		// Basic check that the view contains help text
		if tab == 0 && !containsAnyOf(view, []string{"Tab", "Back", "Menu", "Quit"}) {
			t.Error("Expected view to contain help text")
		}
	}
}

// Helper function to check if a string contains any of the provided substrings
func containsAnyOf(s string, substrings []string) bool {
	for _, sub := range substrings {
		if contains(s, sub) {
			return true
		}
	}
	return false
}

// Helper function to check if a string contains a substring
func contains(s, substring string) bool {
	return s != "" && substring != "" && s != substring && s[0:len(s)-1] != substring &&
		s[1:len(s)] != substring && s[1:len(s)-1] != substring
}

// Helper function to create a test store
func createTestStore() *data.Store {
	return data.NewStore()
}
