// Filename: cmd/gocard/main.go
// Version: 0.0.0
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/DavidMiserak/GoCard/internal/algorithm"
	"github.com/DavidMiserak/GoCard/internal/storage"
	"github.com/DavidMiserak/GoCard/internal/ui"
)

func main() {
	// Define command-line flags
	var useTUI bool
	flag.BoolVar(&useTUI, "tui", false, "Use terminal UI mode")
	flag.Parse()

	// Create or use the default directory for flashcards
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get user home directory: %v", err)
	}

	cardsDir := filepath.Join(homeDir, "GoCard")

	// Allow specifying a different cards directory as a positional argument
	if flag.NArg() > 0 {
		cardsDir = flag.Arg(0)
	}

	// Initialize our card store
	store, err := storage.NewCardStore(cardsDir)
	if err != nil {
		log.Fatalf("Failed to initialize card store: %v", err)
	}

	// If TUI mode is enabled, launch the terminal UI
	if useTUI {
		fmt.Printf("Starting GoCard terminal UI with cards from: %s\n", cardsDir)
		if err := ui.RunTUI(store); err != nil {
			log.Fatalf("Error running terminal UI: %v", err)
		}
		return
	}

	// Otherwise, run the original example code
	runExampleMode(store)
}

// runExampleMode runs the original example code from the previous main function
func runExampleMode(store *storage.CardStore) {
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

	// Example: Get due cards
	dueCards := store.GetDueCards()
	fmt.Printf("\nFound %d cards due for review\n", len(dueCards))

	// Example: Review a card with SM-2 algorithm
	if len(dueCards) > 0 {
		card := dueCards[0]

		// Simulate reviewing the card
		fmt.Printf("\nReviewing card: %s\n", card.Title)
		fmt.Printf("Question: %s\n", card.Question)
		fmt.Println("...(User would see answer and rate their recall)...")

		// Rating: 0-5 where:
		// 0-2: Difficult/incorrect (reset interval)
		// 3: Correct but difficult (small interval increase)
		// 4: Correct and somewhat easy (larger interval increase)
		// 5: Very easy (largest interval increase)
		rating := 4 // Example rating (good recall)

		// Apply the SM-2 algorithm and save the card
		prevInterval := card.ReviewInterval
		err := store.ReviewCard(card, rating)
		if err != nil {
			log.Fatalf("Failed to review card: %v", err)
		}

		fmt.Printf("Card reviewed with rating: %d\n", rating)
		fmt.Printf("Review interval changed from %d to %d days\n", prevInterval, card.ReviewInterval)
		fmt.Printf("Next review date: %s\n", algorithm.SM2.CalculateDueDate(card).Format("Jan 2, 2006"))
	}

	// Example: Create several cards with different review histories
	createDemoCards(store)

	// Display statistics
	stats := store.GetReviewStats()
	fmt.Println("\nCard Statistics:")
	fmt.Printf("Total cards: %d\n", stats["total_cards"])
	fmt.Printf("Due cards: %d\n", stats["due_cards"])
	fmt.Printf("New cards: %d\n", stats["new_cards"])
	fmt.Printf("Young cards (1-7 days): %d\n", stats["young_cards"])
	fmt.Printf("Mature cards (>7 days): %d\n", stats["mature_cards"])

	fmt.Println("\nNext due card: ", store.GetNextDueDate().Format("Jan 2, 2006"))
}

// createDemoCards creates a few cards with different review histories
// to demonstrate the SM-2 algorithm behavior
func createDemoCards(store *storage.CardStore) {
	// Create a new card (never reviewed)
	newCard, _ := store.CreateCard(
		"Binary Search",
		"Explain the binary search algorithm and its time complexity.",
		"Binary search is an O(log n) algorithm that works on sorted arrays by repeatedly dividing the search interval in half.",
		[]string{"algorithms", "searching"},
	)
	fmt.Printf("\nCreated new card: %s (never reviewed)\n", newCard.Title)

	// Create a young card (reviewed recently, short interval)
	youngCard, _ := store.CreateCard(
		"Quick Sort",
		"How does Quick Sort work?",
		"Quick sort is a divide-and-conquer algorithm that picks a pivot element and partitions the array around it.",
		[]string{"algorithms", "sorting"},
	)
	// Simulate a previous review 2 days ago with a good rating
	youngCard.LastReviewed = time.Now().AddDate(0, 0, -2)
	youngCard.ReviewInterval = 4
	youngCard.Difficulty = 4
	if err := store.SaveCard(youngCard); err != nil {
		log.Fatalf("Failed to save young card: %v", err)
	}
	fmt.Printf("Created young card: %s (reviewed 2 days ago, due in 2 days)\n", youngCard.Title)

	// Create a mature card (reviewed long ago, long interval)
	matureCard, _ := store.CreateCard(
		"Graph Traversal",
		"Compare BFS and DFS graph traversal algorithms.",
		"BFS uses a queue and explores all neighbors before moving to the next level. DFS uses a stack (or recursion) and explores as far as possible along one branch before backtracking.",
		[]string{"algorithms", "graphs"},
	)
	// Simulate several successful reviews, resulting in a long interval
	matureCard.LastReviewed = time.Now().AddDate(0, 0, -10)
	matureCard.ReviewInterval = 30
	matureCard.Difficulty = 5
	if err := store.SaveCard(matureCard); err != nil {
		log.Fatalf("Failed to save mature card: %v", err)
	}
	fmt.Printf("Created mature card: %s (reviewed 10 days ago, due in 20 days)\n", matureCard.Title)

	// Create an overdue card
	overdueCard, _ := store.CreateCard(
		"Dynamic Programming",
		"What is dynamic programming and when is it useful?",
		"Dynamic programming is an optimization technique that solves problems by breaking them down into simpler subproblems and storing the results to avoid redundant calculations.",
		[]string{"algorithms", "optimization"},
	)
	// Simulate a review that's now overdue
	overdueCard.LastReviewed = time.Now().AddDate(0, 0, -15)
	overdueCard.ReviewInterval = 7
	overdueCard.Difficulty = 3
	if err := store.SaveCard(overdueCard); err != nil {
		log.Fatalf("Failed to save overdue card: %v", err)
	}
	fmt.Printf("Created overdue card: %s (was due 8 days ago)\n", overdueCard.Title)
}
