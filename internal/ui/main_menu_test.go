// File: internal/ui/menu_test.go

package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestMainMenuView tests that the menu renders correctly
func TestMainMenuView(t *testing.T) {
	// Create a new instance of our menu
	menu := NewMainMenu()

	// Test the initial view rendering
	view := menu.View()

	// Check that the view contains all expected menu items
	expectedItems := []string{"Study", "Browse Decks", "Statistics", "Quit"}
	for _, item := range expectedItems {
		if !strings.Contains(view, item) {
			t.Errorf("Expected view to contain menu item '%s', but it didn't", item)
		}
	}

	// Check that the view contains the title
	if !strings.Contains(view, "GoCard") {
		t.Errorf("Expected view to contain the title 'GoCard', but it didn't")
	}

	// Check that the view contains the subtitle
	if !strings.Contains(view, "Terminal Flashcards") {
		t.Errorf("Expected view to contain the subtitle 'Terminal Flashcards', but it didn't")
	}
}

// TestMainMenuUpdate tests cursor movement and selection
func TestMainMenuUpdate(t *testing.T) {
	// Create a new instance of our menu
	menu := NewMainMenu()

	// Test cursor movement with up/down keys
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	upMsg := tea.KeyMsg{Type: tea.KeyUp}

	// Initial cursor position should be 0
	if menu.cursor != 0 {
		t.Errorf("Expected initial cursor position to be 0, got %d", menu.cursor)
	}

	// Move cursor down once
	updatedModel, _ := menu.Update(downMsg)
	updatedMenu, ok := updatedModel.(MainMenu)
	if !ok {
		t.Fatalf("Expected MainMenu, got %T", updatedModel)
	}

	if updatedMenu.cursor != 1 {
		t.Errorf("Expected cursor position to be 1 after down key, got %d", updatedMenu.cursor)
	}

	// Move cursor up
	updatedModel, _ = updatedMenu.Update(upMsg)
	updatedMenu, ok = updatedModel.(MainMenu)
	if !ok {
		t.Fatalf("Expected MainMenu, got %T", updatedModel)
	}

	if updatedMenu.cursor != 0 {
		t.Errorf("Expected cursor position to be 0 after up key, got %d", updatedMenu.cursor)
	}
}
