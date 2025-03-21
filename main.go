// Filename: main.go
// Version: 0.0.0
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	// Create or use the default directory for flashcards
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get user home directory: %v", err)
	}

	cardsDir := filepath.Join(homeDir, "GoCard")

	// Initialize our card store
	store, err := NewCardStore(cardsDir)
	if err != nil {
		log.Fatalf("Failed to initialize card store: %v", err)
	}

	// Example: Create a new flashcard
	exampleCard, err := store.CreateCard(
		"Two-Pointer Technique",
		"What is the two-pointer technique in algorithms and when should it be used?",
		`The two-pointer technique uses two pointers to iterate through a data structure simultaneously.

It's particularly useful for:
- Sorted array operations
- Finding pairs with certain conditions
- String manipulation (palindromes)
- Linked list cycle detection

Example (Two Sum in sorted array):
`+"```python\ndef two_sum(nums, target):\n    left, right = 0, len(nums) - 1\n    while left < right:\n        current_sum = nums[left] + nums[right]\n        if current_sum == target:\n            return [left, right]\n        elif current_sum < target:\n            left += 1\n        else:\n            right -= 1\n    return [-1, -1]  # No solution\n```",
		[]string{"algorithms", "techniques", "arrays"},
	)
	if err != nil {
		log.Fatalf("Failed to create example card: %v", err)
	}

	fmt.Printf("Created new card: %s at %s\n", exampleCard.Title, exampleCard.FilePath)

	// Example: Load all cards and print them
	fmt.Println("\nAll cards in the store:")
	for path, card := range store.Cards {
		fmt.Printf("- %s (%s)\n", card.Title, path)
	}

	// Example: Get due cards
	dueCards := store.GetDueCards()
	fmt.Printf("\nFound %d cards due for review\n", len(dueCards))

	// Example: Update a card after review
	if len(dueCards) > 0 {
		card := dueCards[0]
		card.LastReviewed = time.Now()
		card.Difficulty = 3     // Medium difficulty
		card.ReviewInterval = 2 // Review again in 2 days

		if err := store.SaveCard(card); err != nil {
			log.Fatalf("Failed to save card: %v", err)
		}

		fmt.Printf("Updated card: %s\n", card.Title)
	}

	// Implementation note: In a real application, you would:
	// 1. Connect this to a GUI for user interaction
	// 2. Add proper file watching for external changes
	// 3. Implement a review scheduler based on the SM-2 algorithm
	// 4. Add handling for card organization into decks (directories)
}
