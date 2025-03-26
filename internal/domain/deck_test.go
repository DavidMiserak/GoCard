// internal/domain/deck_test.go
package domain

import (
	"path/filepath"
	"testing"
)

func TestNewDeck(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		dirPath        string
		expectedName   string
		expectedParent string
	}{
		{
			name:           "simple path",
			dirPath:        "/test/cards/programming",
			expectedName:   "programming",
			expectedParent: "/test/cards",
		},
		{
			name:           "root path",
			dirPath:        "/",
			expectedName:   "/",
			expectedParent: "",
		},
		{
			name:           "current directory",
			dirPath:        ".",
			expectedName:   ".",
			expectedParent: "",
		},
		{
			name:           "nested path",
			dirPath:        "/test/cards/programming/go/concurrency",
			expectedName:   "concurrency",
			expectedParent: "/test/cards/programming/go",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Handle path separators for cross-platform testing
			dirPath := filepath.FromSlash(tc.dirPath)
			expectedParent := filepath.FromSlash(tc.expectedParent)

			// Create a new deck
			deck := NewDeck(dirPath)

			// Check path
			if deck.Path != dirPath {
				t.Errorf("expected Path to be %s, got %s", dirPath, deck.Path)
			}

			// Check name
			if deck.Name != tc.expectedName {
				t.Errorf("expected Name to be %s, got %s", tc.expectedName, deck.Name)
			}

			// Check parent path
			if deck.ParentPath != expectedParent {
				t.Errorf("expected ParentPath to be %s, got %s", expectedParent, deck.ParentPath)
			}
		})
	}
}

func TestGetRelativePath(t *testing.T) {
	// Test cases
	testCases := []struct {
		name     string
		deckPath string
		rootDir  string
		expected string
	}{
		{
			name:     "deck in root",
			deckPath: "/test/cards/deck1",
			rootDir:  "/test/cards",
			expected: "deck1",
		},
		{
			name:     "nested deck",
			deckPath: "/test/cards/category/subcategory/deck2",
			rootDir:  "/test/cards",
			expected: "category/subcategory/deck2",
		},
		{
			name:     "deck is root",
			deckPath: "/test/cards",
			rootDir:  "/test/cards",
			expected: ".",
		},
		{
			name:     "invalid root (error case)",
			deckPath: "/test/cards/deck1",
			rootDir:  "invalid://path",
			expected: "deck1", // Should fall back to the Name if Rel fails
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Handle path separators for cross-platform testing
			deckPath := filepath.FromSlash(tc.deckPath)
			rootDir := filepath.FromSlash(tc.rootDir)
			expected := filepath.FromSlash(tc.expected)

			// Create a deck with the test path
			deck := NewDeck(deckPath)

			// Get the relative path
			result := deck.GetRelativePath(rootDir)

			// Check result
			if result != expected {
				t.Errorf("expected relative path %s, got %s", expected, result)
			}
		})
	}
}

func TestGetHierarchyPath(t *testing.T) {
	// Test cases
	testCases := []struct {
		name     string
		deckPath string
		rootDir  string
		expected string
	}{
		{
			name:     "deck in root",
			deckPath: "/test/cards/deck1",
			rootDir:  "/test/cards",
			expected: "deck1",
		},
		{
			name:     "nested deck",
			deckPath: "/test/cards/category/subcategory/deck2",
			rootDir:  "/test/cards",
			expected: "category > subcategory > deck2",
		},
		{
			name:     "deck is root",
			deckPath: "/test/cards",
			rootDir:  "/test/cards",
			expected: "cards", // Deck's name
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a deck with the test path
			deckPath := filepath.FromSlash(tc.deckPath)
			rootDir := filepath.FromSlash(tc.rootDir)

			deck := NewDeck(deckPath)

			// Get the hierarchy path
			result := deck.GetHierarchyPath(rootDir)

			// For Windows, use the right separator pattern
			expected := tc.expected
			// No special handling needed for Windows paths here since
			// we use platform-independent separator in GetHierarchyPath

			// Check result
			if result != expected {
				t.Errorf("expected hierarchy path %s, got %s", expected, result)
			}
		})
	}
}
