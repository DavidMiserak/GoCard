// test/integration/deck_card_relationships_test.go
package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// createComplexDeckStructure sets up a more intricate deck and card hierarchy
func createComplexDeckStructure(rootDir string) (map[string][]string, error) {
	// Define a more complex deck structure with nested decks and multiple cards
	hierarchyStructure := map[string][]string{
		"Computer Science": {
			"Algorithms/Sorting",
			"Algorithms/Searching",
			"Programming/Go",
			"Programming/Python",
			"Programming/Web/Frontend",
			"Programming/Web/Backend",
		},
		"Languages": {
			"Human/Romance/Spanish",
			"Human/Romance/French",
			"Human/Germanic/German",
			"Programming/Markup/Markdown",
			"Programming/Scripting/Python",
		},
	}

	// Create deck hierarchy and sample cards
	for rootDeck, subdecks := range hierarchyStructure {
		for _, subdeck := range subdecks {
			// Full path of the subdeck
			fullSubdeckPath := filepath.Join(rootDir, rootDeck, subdeck)

			// Create deck directory
			if err := os.MkdirAll(fullSubdeckPath, 0755); err != nil {
				return nil, fmt.Errorf("failed to create subdeck %s: %v", subdeck, err)
			}

			// Prepare tags based on the full path
			pathParts := strings.Split(subdeck, string(filepath.Separator))
			baseName := pathParts[len(pathParts)-1]

			// Determine additional tags based on path structure
			additionalTags := []string{
				rootDeck,
				baseName,
			}

			// Add parent path components as tags
			for _, part := range pathParts {
				if part != baseName {
					additionalTags = append(additionalTags, part)
				}
			}

			// Add specific context tags based on deck structure
			if strings.Contains(subdeck, "Programming") {
				additionalTags = append(additionalTags, "Programming")
			}
			if strings.Contains(subdeck, "Scripting") {
				additionalTags = append(additionalTags, "Scripting")
			}

			// Create multiple cards for each subdeck
			cardTemplates := []struct {
				title      string
				tags       []string
				difficulty int
				content    string
			}{
				{
					title:      fmt.Sprintf("%s Fundamental Concept", baseName),
					tags:       append(additionalTags, "fundamental"),
					difficulty: 2,
					content:    "# Basic Concept\n\n---\n\nDetailed explanation of fundamental concepts.",
				},
				{
					title:      fmt.Sprintf("%s Advanced Topic", baseName),
					tags:       append(additionalTags, "advanced"),
					difficulty: 4,
					content:    "# Advanced Exploration\n\n---\n\nIn-depth analysis of advanced topics.",
				},
			}

			for i, cardTemplate := range cardTemplates {
				cardFilename := fmt.Sprintf("%s_card_%d.md", baseName, i+1)
				cardPath := filepath.Join(fullSubdeckPath, cardFilename)

				// Create YAML frontmatter with dynamic content
				cardContent := fmt.Sprintf(`---
title: %s
tags:
%s
difficulty: %d
---
%s
`,
					cardTemplate.title,
					func() string {
						var tagStr string
						for _, tag := range cardTemplate.tags {
							tagStr += fmt.Sprintf("  - %s\n", tag)
						}
						return tagStr
					}(),
					cardTemplate.difficulty,
					cardTemplate.content,
				)

				if err := os.WriteFile(cardPath, []byte(cardContent), 0644); err != nil {
					return nil, fmt.Errorf("failed to create card %s: %v", cardFilename, err)
				}
			}
		}
	}

	return hierarchyStructure, nil
}

// TestDeckCardRelationships validates complex deck and card interactions
func TestDeckCardRelationships(t *testing.T) {
	// Setup test environment
	rootDir, storageService, _, deckService, _, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Create complex deck structure
	hierarchyStructure, err := createComplexDeckStructure(rootDir)
	if err != nil {
		t.Fatalf("Failed to create deck structure: %v", err)
	}

	// Subtest 1: Verify Deck Hierarchy
	t.Run("DeckHierarchyValidation", func(t *testing.T) {
		// Test each root deck
		for rootDeck := range hierarchyStructure {
			rootPath := filepath.Join(rootDir, rootDeck)

			// Get subdecks
			subdecks, err := deckService.GetSubdecks(rootPath)
			if err != nil {
				t.Fatalf("Failed to get subdecks for %s: %v", rootDeck, err)
			}

			// Verify number of direct subdecks
			expectedSubdeckCount := len(getDirectSubdecks(hierarchyStructure[rootDeck]))
			if len(subdecks) != expectedSubdeckCount {
				t.Errorf("Expected %d direct subdecks for %s, got %d",
					expectedSubdeckCount, rootDeck, len(subdecks))
			}

			// Verify subdeck properties
			for _, subdeck := range subdecks {
				// Check parent path is correct
				if filepath.Dir(subdeck.Path) != rootPath {
					t.Errorf("Subdeck %s does not have correct parent path", subdeck.Name)
				}
			}
		}
	})

	// Subtest 2: Card Distribution Across Decks
	t.Run("CardDistributionValidation", func(t *testing.T) {
		// Verify card distribution in various decks
		for rootDeck, subdecks := range hierarchyStructure {
			for _, subdeck := range subdecks {
				fullSubdeckPath := filepath.Join(rootDir, rootDeck, subdeck)

				// Get cards in this subdeck
				cards, err := deckService.GetCards(fullSubdeckPath)
				if err != nil {
					t.Fatalf("Failed to get cards for %s: %v", subdeck, err)
				}

				// Each subdeck should have 2 cards
				expectedCardCount := 2
				if len(cards) != expectedCardCount {
					t.Errorf("Expected %d cards in %s, got %d",
						expectedCardCount, subdeck, len(cards))
				}

				// Verify card properties
				for _, card := range cards {
					// Validate tags include deck hierarchy
					hasRootTag := false
					hasSubdeckTag := false
					for _, tag := range card.Tags {
						if tag == rootDeck {
							hasRootTag = true
						}
						if tag == filepath.Base(subdeck) {
							hasSubdeckTag = true
						}
					}

					if !hasRootTag {
						t.Errorf("Card %s missing root deck tag %s", card.Title, rootDeck)
					}
					if !hasSubdeckTag {
						t.Errorf("Card %s missing subdeck tag %s", card.Title, filepath.Base(subdeck))
					}
				}
			}
		}
	})

	// Subtest 3: Nested Deck Card Search
	t.Run("NestedDeckCardSearch", func(t *testing.T) {
		// Test searching for cards across nested decks
		searchQueries := []struct {
			query         string
			minResults    int
			maxResults    int
			expectedDecks []string
		}{
			{
				query:         "fundamental",
				minResults:    10,
				maxResults:    12,
				expectedDecks: nil,
			},
			{
				query:         "advanced",
				minResults:    15,
				maxResults:    20,
				expectedDecks: nil,
			},
			{
				query:         "Programming",
				minResults:    10,
				maxResults:    15,
				expectedDecks: nil,
			},
		}

		for _, testCase := range searchQueries {
			t.Run(fmt.Sprintf("Search_%s", testCase.query), func(t *testing.T) {
				// Perform search
				searchResults, err := storageService.SearchCards(testCase.query)
				if err != nil {
					t.Fatalf("Failed to search cards: %v", err)
				}

				// Verify number of results
				if len(searchResults) < testCase.minResults || len(searchResults) > testCase.maxResults {
					t.Errorf("Expected between %d and %d results for query '%s', got %d",
						testCase.minResults, testCase.maxResults, testCase.query, len(searchResults))
				}

				// Log detailed results
				uniqueResults := make(map[string]bool)
				t.Logf("Search Results for '%s':", testCase.query)
				for _, result := range searchResults {
					if !uniqueResults[result.Title] {
						t.Logf("  - %s (Tags: %v, Path: %s)",
							result.Title, result.Tags, result.FilePath)
						uniqueResults[result.Title] = true
					}
				}
			})
		}
	})

	// Subtest 4: Tag-based Card Retrieval
	t.Run("TagBasedCardRetrieval", func(t *testing.T) {
		// Test finding cards by specific tags
		tagQueries := []struct {
			tag        string
			minResults int
			maxResults int
		}{
			{"fundamental", 5, 7},
			{"advanced", 15, 20},
			{"Programming", 10, 15},
			{"Scripting", 1, 3},
		}

		for _, tagQuery := range tagQueries {
			t.Run(fmt.Sprintf("Tag_%s", tagQuery.tag), func(t *testing.T) {
				// Find cards by tag
				taggedCards, err := storageService.FindCardsByTag(tagQuery.tag)
				if err != nil {
					t.Fatalf("Failed to find cards by tag %s: %v", tagQuery.tag, err)
				}

				// Verify number of results with flexibility
				if len(taggedCards) < tagQuery.minResults || len(taggedCards) > tagQuery.maxResults {
					t.Errorf("Expected between %d and %d cards with tag '%s', got %d",
						tagQuery.minResults, tagQuery.maxResults, tagQuery.tag, len(taggedCards))
				}

				// Log unique results
				uniqueResults := make(map[string]bool)
				t.Logf("Tagged Cards for '%s':", tagQuery.tag)
				for _, card := range taggedCards {
					if !uniqueResults[card.Title] {
						t.Logf("  - %s (Path: %s, Tags: %v)",
							card.Title, card.FilePath, card.Tags)
						uniqueResults[card.Title] = true
					}
				}
			})
		}
	})
}

// getDirectSubdecks extracts only the first-level subdeck names
func getDirectSubdecks(subdecks []string) []string {
	directSubdecks := make(map[string]bool)
	for _, subdeck := range subdecks {
		// Split the path and take the first component
		parts := strings.Split(subdeck, string(filepath.Separator))
		directSubdecks[parts[0]] = true
	}

	// Convert map keys to slice
	result := make([]string, 0, len(directSubdecks))
	for k := range directSubdecks {
		result = append(result, k)
	}
	return result
}
