// File: internal/ui/stats_screen_test.go

package ui

import (
	"strings"
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

	// Now we expect activeTab to be initialized to 1 (Deck Review tab)
	if statsScreen.activeTab != 1 {
		t.Errorf("Expected activeTab to be 1, got %d", statsScreen.activeTab)
	}

	if len(statsScreen.cardStats) == 0 {
		t.Error("Expected cardStats to be initialized with data")
	}

	// Check that lastDeckID is initialized to empty string
	if statsScreen.lastDeckID != "" {
		t.Errorf("Expected lastDeckID to be empty, got %s", statsScreen.lastDeckID)
	}
}

func TestNewStatisticsScreenWithDeck(t *testing.T) {
	store := data.NewStore()
	deckID := "test-deck-id"
	statsScreen := NewStatisticsScreenWithDeck(store, deckID)

	if statsScreen == nil {
		t.Fatal("Expected NewStatisticsScreenWithDeck to return a non-nil StatisticsScreen")
	}

	if statsScreen.store != store {
		t.Errorf("Expected store to be %v, got %v", store, statsScreen.store)
	}

	if statsScreen.activeTab != 1 {
		t.Errorf("Expected activeTab to be 1 (Deck Review tab), got %d", statsScreen.activeTab)
	}

	if len(statsScreen.cardStats) == 0 {
		t.Error("Expected cardStats to be initialized with data")
	}

	// Check that lastDeckID is set to the provided deckID
	if statsScreen.lastDeckID != deckID {
		t.Errorf("Expected lastDeckID to be %s, got %s", deckID, statsScreen.lastDeckID)
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

	// Since activeTab is now initialized to 1, we expect it to cycle to 2
	model, cmd := statsScreen.Update(tea.KeyMsg{Type: tea.KeyTab})
	updatedScreen := model.(*StatisticsScreen)

	if updatedScreen.activeTab != 2 {
		t.Errorf("Expected activeTab to be 2 after Tab key, got %d", updatedScreen.activeTab)
	}

	if cmd != nil {
		t.Error("Expected cmd to be nil")
	}

	// Test cycling back to 0
	model, cmd = updatedScreen.Update(tea.KeyMsg{Type: tea.KeyTab})
	updatedScreen = model.(*StatisticsScreen)

	if updatedScreen.activeTab != 0 {
		t.Errorf("Expected activeTab to be 0 after second Tab key, got %d", updatedScreen.activeTab)
	}

	// Test back to main menu
	model, cmd = updatedScreen.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	_, ok := model.(*MainMenu)
	if !ok {
		t.Errorf("Expected model to be *MainMenu after 'b' key, got %T", model)
	}

	// Test window size update
	width, height := 80, 24
	statsScreen = NewStatisticsScreen(store) // Reset stats screen
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
		if !containsAnyOf(view, []string{"Tab", "Back", "Menu", "Quit"}) {
			t.Error("Expected view to contain help text")
		}

		// Check that the tab headings are present
		if !containsAnyOf(view, []string{"Summary", "Deck Review", "Review Forecast"}) {
			t.Error("Expected view to contain tab headings")
		}
	}
}

func TestStatisticsScreenWithDeckView(t *testing.T) {
	// Create a store with some test data
	store := data.NewStore()

	// Assuming GetDecks returns at least one deck with an ID and Name
	decks := store.GetDecks()
	if len(decks) == 0 {
		t.Skip("No decks available for testing")
		return
	}

	deckID := decks[0].ID

	// Create stats screen with a specific deck
	statsScreen := NewStatisticsScreenWithDeck(store, deckID)

	// Set screen dimensions
	statsScreen.width = 80
	statsScreen.height = 24

	// Make sure we're on the Deck Review tab
	statsScreen.activeTab = 1

	// Get the view
	view := statsScreen.View()

	// Check that the view is not empty
	if view == "" {
		t.Error("Expected view to return non-empty string")
	}

	// We should see the Deck Review tab content
	if !containsAnyOf(view, []string{"Deck Review", "Ratings Distribution"}) {
		t.Error("Expected view to contain Deck Review content")
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
	// This current implementation looks a bit unusual and might not work as expected
	// Let's use a more straightforward approach
	return s != "" && substring != "" && s != substring && strings.Contains(s, substring)
}

// Helper function to create a test store
func createTestStore() *data.Store {
	return data.NewStore()
}
