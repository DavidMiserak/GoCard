// File: internal/ui/browse_decks_test.go

package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/DavidMiserak/GoCard/internal/data"
)

func TestBrowseScreenView(t *testing.T) {
	// Create a store with dummy data
	store := data.NewStore()

	// Create the browse screen
	browse := NewBrowseScreen(store)

	// Test the view rendering
	view := browse.View()

	// Check that the view contains expected elements
	expectedElements := []string{
		"Browse Decks",
		"DECK NAME",
		"CARDS",
		"DUE",
		"LAST STUDIED",
		"Page 1 of",
	}

	for _, element := range expectedElements {
		if !strings.Contains(view, element) {
			t.Errorf("Expected view to contain '%s', but it didn't", element)
		}
	}

	// Check that the first deck is selected (has cursor)
	lines := strings.Split(view, "\n")
	foundSelected := false

	// Look for a line that contains both ">" and the first deck's name
	for _, line := range lines {
		if strings.Contains(line, ">") && strings.Contains(line, store.GetDecks()[0].Name) {
			foundSelected = true
			break
		}
	}

	if !foundSelected {
		t.Errorf("Expected first deck to be selected, but it wasn't")
	}
}

func TestBrowseScreenNavigation(t *testing.T) {
	// Create a store with dummy data
	store := data.NewStore()

	// Create the browse screen
	browse := NewBrowseScreen(store)

	// Test navigation
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	upMsg := tea.KeyMsg{Type: tea.KeyUp}

	// Move cursor down
	updatedModel, _ := browse.Update(downMsg)
	updatedBrowse, ok := updatedModel.(BrowseScreen)
	if !ok {
		t.Fatalf("Expected BrowseScreen, got %T", updatedModel)
	}

	if updatedBrowse.cursor != 1 {
		t.Errorf("Expected cursor position to be 1 after down key, got %d", updatedBrowse.cursor)
	}

	// Move cursor back up
	updatedModel, _ = updatedBrowse.Update(upMsg)
	updatedBrowse, ok = updatedModel.(BrowseScreen)
	if !ok {
		t.Fatalf("Expected BrowseScreen, got %T", updatedModel)
	}

	if updatedBrowse.cursor != 0 {
		t.Errorf("Expected cursor position to be 0 after up key, got %d", updatedBrowse.cursor)
	}
}

func TestBrowseScreenPagination(t *testing.T) {
	// Create a store with dummy data
	store := data.NewStore()

	// Ensure we have enough decks for pagination
	if len(store.GetDecks()) <= decksPerPage {
		t.Skip("Not enough decks to test pagination")
	}

	// Create the browse screen
	browse := NewBrowseScreen(store)

	// Test pagination
	nextPageMsg := tea.KeyMsg{Type: tea.KeyRight}
	prevPageMsg := tea.KeyMsg{Type: tea.KeyLeft}

	// Move to next page
	updatedModel, _ := browse.Update(nextPageMsg)
	updatedBrowse, ok := updatedModel.(BrowseScreen)
	if !ok {
		t.Fatalf("Expected BrowseScreen, got %T", updatedModel)
	}

	if updatedBrowse.page != 1 {
		t.Errorf("Expected page to be 1 after next page key, got %d", updatedBrowse.page)
	}

	// Move back to previous page
	updatedModel, _ = updatedBrowse.Update(prevPageMsg)
	updatedBrowse, ok = updatedModel.(BrowseScreen)
	if !ok {
		t.Fatalf("Expected BrowseScreen, got %T", updatedModel)
	}

	if updatedBrowse.page != 0 {
		t.Errorf("Expected page to be 0 after prev page key, got %d", updatedBrowse.page)
	}
}

func TestBrowseScreenBackButton(t *testing.T) {
	// Create a store with dummy data
	store := data.NewStore()

	// Create the browse screen
	browse := NewBrowseScreen(store)

	// Test back button
	backMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}

	// Press back button
	updatedModel, _ := browse.Update(backMsg)

	// Verify we got a MainMenu model back
	_, ok := updatedModel.(*MainMenu)
	if !ok {
		t.Fatalf("Expected *MainMenu after back key, got %T", updatedModel)
	}
}
