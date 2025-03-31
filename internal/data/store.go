// File: internal/data/store.go

package data

import (
	"github.com/DavidMiserak/GoCard/internal/model"
	"time"
)

// Store manages all data for the application
type Store struct {
	Decks []model.Deck
}

// NewStore creates a new data store with dummy data
func NewStore() *Store {
	store := &Store{
		Decks: []model.Deck{},
	}

	// Add dummy data
	store.addDummyData()

	return store
}

// addDummyData adds sample decks and cards
func (s *Store) addDummyData() {
	// Go Programming Deck
	goCards := []model.Card{
		{
			ID:           "go-1",
			Question:     "What is the purpose of the \"defer\" keyword in Go?",
			Answer:       "The \"defer\" keyword in Go schedules a function call to be executed just before the function returns. This is often used for cleanup actions, ensuring they will be executed even if the function panics.",
			DeckID:       "go-programming",
			LastReviewed: time.Now(),
			NextReview:   time.Now().Add(24 * time.Hour),
			Ease:         2.5,
			Interval:     1,
			Rating:       4,
		},
		{
			ID:           "go-2",
			Question:     "What are goroutines in Go?",
			Answer:       "Goroutines are lightweight threads managed by the Go runtime. They allow concurrent execution of functions without the overhead of traditional OS threads.",
			DeckID:       "go-programming",
			LastReviewed: time.Now().Add(-12 * time.Hour),
			NextReview:   time.Now().Add(36 * time.Hour),
			Ease:         2.3,
			Interval:     2,
			Rating:       3,
		},
		{
			ID:           "go-3",
			Question:     "How does a slice differ from an array in Go?",
			Answer:       "A slice is a reference to a contiguous segment of an array. Unlike arrays, slices are dynamic in size and don't carry their length as part of their type.",
			DeckID:       "go-programming",
			LastReviewed: time.Now().Add(-24 * time.Hour),
			NextReview:   time.Now().Add(48 * time.Hour),
			Ease:         2.7,
			Interval:     3,
			Rating:       5,
		},
	}

	goDeck := model.Deck{
		ID:          "go-programming",
		Name:        "Go Programming",
		Description: "Basic Go programming concepts",
		Cards:       goCards,
		CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		LastStudied: time.Now(),
	}

	s.Decks = append(s.Decks, goDeck)

	// Computer Science Deck
	csCards := []model.Card{
		{
			ID:           "cs-1",
			Question:     "What is a compiler?",
			Answer:       "A compiler is a program that translates source code written in a high-level programming language into machine code or another lower-level form.",
			DeckID:       "computer-science",
			LastReviewed: time.Now().Add(-24 * time.Hour),
			NextReview:   time.Now().Add(48 * time.Hour),
			Ease:         2.3,
			Interval:     2,
			Rating:       3,
		},
		{
			ID:           "cs-2",
			Question:     "What is the difference between process and thread?",
			Answer:       "A process is an instance of a program execution that has its own memory space. A thread is the smallest unit of execution within a process, and multiple threads share the memory space of the process.",
			DeckID:       "computer-science",
			LastReviewed: time.Now().Add(-36 * time.Hour),
			NextReview:   time.Now().Add(72 * time.Hour),
			Ease:         2.4,
			Interval:     3,
			Rating:       4,
		},
		{
			ID:           "cs-3",
			Question:     "What is cache memory?",
			Answer:       "Cache memory is a small, fast memory that stores frequently accessed data to reduce the time needed to access it from slower main memory.",
			DeckID:       "computer-science",
			LastReviewed: time.Now().Add(-48 * time.Hour),
			NextReview:   time.Now().Add(96 * time.Hour),
			Ease:         2.2,
			Interval:     4,
			Rating:       3,
		},
	}

	csDeck := model.Deck{
		ID:          "computer-science",
		Name:        "Computer Science",
		Description: "General computer science concepts",
		Cards:       csCards,
		CreatedAt:   time.Now().Add(-45 * 24 * time.Hour),
		LastStudied: time.Now().Add(-24 * time.Hour),
	}

	s.Decks = append(s.Decks, csDeck)

	// Data Structures Deck
	dsCards := []model.Card{
		{
			ID:           "ds-1",
			Question:     "What is a stack data structure?",
			Answer:       "A stack is a linear data structure that follows the Last In First Out (LIFO) principle, where elements are added and removed from the same end, called the top.",
			DeckID:       "data-structures",
			LastReviewed: time.Now().Add(-72 * time.Hour),
			NextReview:   time.Now().Add(15 * 24 * time.Hour),
			Ease:         2.6,
			Interval:     15,
			Rating:       4,
		},
		{
			ID:           "ds-2",
			Question:     "What is a queue data structure?",
			Answer:       "A queue is a linear data structure that follows the First In First Out (FIFO) principle, where elements are added at the rear and removed from the front.",
			DeckID:       "data-structures",
			LastReviewed: time.Now().Add(-84 * time.Hour),
			NextReview:   time.Now().Add(20 * 24 * time.Hour),
			Ease:         2.5,
			Interval:     20,
			Rating:       4,
		},
		{
			ID:           "ds-3",
			Question:     "What is a binary search tree?",
			Answer:       "A binary search tree is a tree data structure where each node has at most two children, and for each node, all elements in the left subtree are less than the node, and all elements in the right subtree are greater.",
			DeckID:       "data-structures",
			LastReviewed: time.Now().Add(-96 * time.Hour),
			NextReview:   time.Now().Add(25 * 24 * time.Hour),
			Ease:         2.7,
			Interval:     25,
			Rating:       5,
		},
	}

	dsDeck := model.Deck{
		ID:          "data-structures",
		Name:        "Data Structures",
		Description: "Common data structures and operations",
		Cards:       dsCards,
		CreatedAt:   time.Now().Add(-60 * 24 * time.Hour),
		LastStudied: time.Now().Add(-72 * time.Hour),
	}

	s.Decks = append(s.Decks, dsDeck)

	// Algorithms Deck
	algoCards := []model.Card{
		{
			ID:           "algo-1",
			Question:     "What is the time complexity of quicksort in the average case?",
			Answer:       "The average time complexity of quicksort is O(n log n), where n is the number of elements to sort.",
			DeckID:       "algorithms",
			LastReviewed: time.Now().Add(-48 * time.Hour),
			NextReview:   time.Now().Add(18 * 24 * time.Hour),
			Ease:         2.4,
			Interval:     18,
			Rating:       3,
		},
		{
			ID:           "algo-2",
			Question:     "What is dynamic programming?",
			Answer:       "Dynamic programming is a method for solving complex problems by breaking them down into simpler subproblems and storing the results of these subproblems to avoid redundant calculations.",
			DeckID:       "algorithms",
			LastReviewed: time.Now().Add(-60 * time.Hour),
			NextReview:   time.Now().Add(22 * 24 * time.Hour),
			Ease:         2.3,
			Interval:     22,
			Rating:       3,
		},
		{
			ID:           "algo-3",
			Question:     "What is breadth-first search?",
			Answer:       "Breadth-first search is a graph traversal algorithm that explores all neighbors at the present depth before moving on to nodes at the next depth level.",
			DeckID:       "algorithms",
			LastReviewed: time.Now().Add(-72 * time.Hour),
			NextReview:   time.Now().Add(26 * 24 * time.Hour),
			Ease:         2.6,
			Interval:     26,
			Rating:       4,
		},
	}

	algoDeck := model.Deck{
		ID:          "algorithms",
		Name:        "Algorithms",
		Description: "Common algorithms and their analysis",
		Cards:       algoCards,
		CreatedAt:   time.Now().Add(-50 * 24 * time.Hour),
		LastStudied: time.Now().Add(-48 * time.Hour),
	}

	s.Decks = append(s.Decks, algoDeck)

	// Bubble Tea UI Deck
	btCards := []model.Card{
		{
			ID:           "bt-1",
			Question:     "What is the Elm Architecture used by Bubble Tea?",
			Answer:       "The Elm Architecture is a design pattern consisting of three main components: Model (application state), View (renders the UI based on the state), and Update (handles events and updates the state).",
			DeckID:       "bubble-tea-ui",
			LastReviewed: time.Now().Add(-7 * 24 * time.Hour),
			NextReview:   time.Now().Add(10 * 24 * time.Hour),
			Ease:         2.1,
			Interval:     10,
			Rating:       4,
		},
		{
			ID:           "bt-2",
			Question:     "What is Lipgloss in the context of Bubble Tea?",
			Answer:       "Lipgloss is a styling library for terminal applications, often used with Bubble Tea to create visually appealing terminal UIs with colors, borders, and alignment.",
			DeckID:       "bubble-tea-ui",
			LastReviewed: time.Now().Add(-9 * 24 * time.Hour),
			NextReview:   time.Now().Add(12 * 24 * time.Hour),
			Ease:         2.2,
			Interval:     12,
			Rating:       4,
		},
	}

	btDeck := model.Deck{
		ID:          "bubble-tea-ui",
		Name:        "Bubble Tea UI",
		Description: "Bubble Tea TUI framework concepts",
		Cards:       btCards,
		CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		LastStudied: time.Now().Add(-7 * 24 * time.Hour),
	}

	s.Decks = append(s.Decks, btDeck)
}

// GetDecks returns all decks
func (s *Store) GetDecks() []model.Deck {
	return s.Decks
}

// GetDeck returns a deck by ID
func (s *Store) GetDeck(id string) (model.Deck, bool) {
	for _, deck := range s.Decks {
		if deck.ID == id {
			return deck, true
		}
	}
	return model.Deck{}, false
}

// GetDueCards returns cards due for review
func (s *Store) GetDueCards() []model.Card {
	var dueCards []model.Card
	now := time.Now()

	for _, deck := range s.Decks {
		for _, card := range deck.Cards {
			if card.NextReview.Before(now) {
				dueCards = append(dueCards, card)
			}
		}
	}

	return dueCards
}

// GetDueCardsForDeck returns cards due for review in a specific deck
func (s *Store) GetDueCardsForDeck(deckID string) []model.Card {
	var dueCards []model.Card
	now := time.Now()

	for _, deck := range s.Decks {
		if deck.ID == deckID {
			for _, card := range deck.Cards {
				if card.NextReview.Before(now) {
					dueCards = append(dueCards, card)
				}
			}
			break
		}
	}

	return dueCards
}
