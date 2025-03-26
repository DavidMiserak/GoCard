// test/integration/tag_management_test.go
package integration

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/DavidMiserak/GoCard/internal/domain"
)

// TestCardTagOperations tests various tag-related operations
func TestCardTagOperations(t *testing.T) {
	// Setup test environment
	rootDir, storageService, _, deckService, _, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Create a test deck with tagged cards
	deckPath := filepath.Join(rootDir, "TagTestDeck")
	if err := os.MkdirAll(deckPath, 0755); err != nil {
		t.Fatalf("Failed to create deck directory: %v", err)
	}

	// Prepare card contents with diverse tags
	taggedCardContents := []struct {
		filename string
		content  string
	}{
		{
			filename: "programming_go.md",
			content: `---
title: Go Concurrency
tags:
  - programming
  - go
  - concurrency
difficulty: 3
---
# Goroutines and Channels in Go Programming

---

Explanation of concurrent programming techniques in Go, showcasing goroutines and concurrency patterns.
`,
		},
		{
			filename: "programming_python.md",
			content: `---
title: Python Generators
tags:
  - programming
  - python
  - generators
difficulty: 2
---
# How Generators Work

---

Detailed explanation of Python generators.
`,
		},
		{
			filename: "languages_spanish.md",
			content: `---
title: Spanish Verb Conjugations
tags:
  - languages
  - spanish
  - grammar
difficulty: 4
---
# Verb Conjugation Rules

---

Comprehensive guide to Spanish verb conjugations.
`,
		},
		{
			filename: "computer_science_algorithm.md",
			content: `---
title: Sorting Algorithms
tags:
  - programming
  - computer-science
  - algorithms
difficulty: 5
---
# Sorting Algorithm Comparison

---

Detailed comparison of various sorting algorithms.
`,
		},
	}

	// Create the tagged cards
	for _, cardInfo := range taggedCardContents {
		cardPath := filepath.Join(deckPath, cardInfo.filename)
		if err := os.WriteFile(cardPath, []byte(cardInfo.content), 0644); err != nil {
			t.Fatalf("Failed to create card file %s: %v", cardInfo.filename, err)
		}
	}

	// Subtest 1: Finding cards by single tag
	t.Run("FindCardsByTag", func(t *testing.T) {
		// Test finding cards with 'programming' tag
		programmingCards, err := storageService.FindCardsByTag("programming")
		if err != nil {
			t.Fatalf("Failed to find cards by 'programming' tag: %v", err)
		}

		// Expected programming cards count
		expectedProgrammingCardCount := 3
		if len(programmingCards) != expectedProgrammingCardCount {
			t.Errorf("Expected %d cards with 'programming' tag, got %d",
				expectedProgrammingCardCount, len(programmingCards))
		}

		// Verify specific card titles
		expectedTitles := []string{
			"Go Concurrency",
			"Python Generators",
			"Sorting Algorithms",
		}

		// Create a slice of found card titles
		var foundTitles []string
		for _, card := range programmingCards {
			foundTitles = append(foundTitles, card.Title)
		}

		// Sort both lists for consistent comparison
		sort.Strings(expectedTitles)
		sort.Strings(foundTitles)

		// Compare titles
		for i, title := range expectedTitles {
			if foundTitles[i] != title {
				t.Errorf("Unexpected card title. Expected %s, got %s", title, foundTitles[i])
			}
		}
	})

	// Subtest 2: Complex tag search
	t.Run("MultiTagSearch", func(t *testing.T) {
		// Perform multiple search strategies
		searchQueries := []struct {
			query           string
			expectedTitle   string
			expectedResults int
		}{
			{
				query:           "programming concurrency",
				expectedTitle:   "Go Concurrency",
				expectedResults: 1,
			},
			{
				query:           "goroutines concurrent",
				expectedTitle:   "Go Concurrency",
				expectedResults: 1,
			},
			{
				query:           "go concurrency",
				expectedTitle:   "Go Concurrency",
				expectedResults: 1,
			},
		}

		for _, testCase := range searchQueries {
			t.Run(testCase.query, func(t *testing.T) {
				// Perform the search
				searchResults, err := storageService.SearchCards(testCase.query)
				if err != nil {
					t.Fatalf("Failed to search cards with query '%s': %v", testCase.query, err)
				}

				// Log all card details for debugging
				t.Logf("Search query: %q", testCase.query)
				t.Logf("Found %d cards", len(searchResults))
				for _, card := range searchResults {
					t.Logf("Found card: Title='%s', Tags=%v, Question='%s'",
						card.Title, card.Tags, card.Question)
				}

				// Check number of results
				if len(searchResults) != testCase.expectedResults {
					t.Errorf("Expected %d card matching '%s', got %d",
						testCase.expectedResults, testCase.query, len(searchResults))
				}

				// Check card title if results found
				if len(searchResults) > 0 && searchResults[0].Title != testCase.expectedTitle {
					t.Errorf("Expected '%s' card, got %s",
						testCase.expectedTitle, searchResults[0].Title)
				}
			})
		}
	})

	// Subtest 3: Edge Cases
	t.Run("TagEdgeCases", func(t *testing.T) {
		// Test non-existent tag
		nonExistentCards, err := storageService.FindCardsByTag("nonexistent-tag")
		if err != nil {
			t.Fatalf("Error searching for non-existent tag: %v", err)
		}
		if len(nonExistentCards) != 0 {
			t.Errorf("Expected 0 cards for non-existent tag, got %d", len(nonExistentCards))
		}

		// Test case-sensitivity and partial matches
		caseInsensitiveCards, err := storageService.FindCardsByTag("GO")
		if err != nil {
			t.Fatalf("Error searching for case-insensitive tag: %v", err)
		}
		if len(caseInsensitiveCards) != 0 {
			t.Errorf("Tag search should be case-sensitive, got %d results", len(caseInsensitiveCards))
		}
	})

	// Subtest 4: Deck-Level Tag Filtering
	t.Run("DeckLevelTagFiltering", func(t *testing.T) {
		// Get cards in the test deck
		deckCards, err := deckService.GetCards(deckPath)
		if err != nil {
			t.Fatalf("Failed to get deck cards: %v", err)
		}

		// Filter cards with 'programming' tag manually
		var programmingCards []domain.Card
		for _, card := range deckCards {
			for _, tag := range card.Tags {
				if tag == "programming" {
					programmingCards = append(programmingCards, card)
					break
				}
			}
		}

		// Verify programming card count
		expectedProgrammingCardCount := 3
		if len(programmingCards) != expectedProgrammingCardCount {
			t.Errorf("Expected %d programming cards in deck, got %d",
				expectedProgrammingCardCount, len(programmingCards))
		}
	})
}
