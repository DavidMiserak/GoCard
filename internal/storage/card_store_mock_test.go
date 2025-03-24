// File: internal/storage/card_store_mock_test.go

package storage

import (
	"fmt"
	"sync"
	"testing"

	"github.com/DavidMiserak/GoCard/internal/storage/io"
)

// TestConcurrentCardCreationWithMock tests creating cards concurrently using a mock filesystem
func TestConcurrentCardCreationWithMock(t *testing.T) {
	// Create a mock filesystem
	mockFS := io.NewMockFileSystem()
	originalFS := io.SetDefaultFS(mockFS)
	defer io.SetDefaultFS(originalFS)

	// Set up root directory - use a path that doesn't need special permissions
	mockFS.SetupDir("/tmp/gocard")

	// Initialize a card store with the mock filesystem
	store, err := NewCardStore("/tmp/gocard")
	if err != nil {
		t.Fatalf("Failed to create card store: %v", err)
	}

	// Number of goroutines and cards per goroutine
	numGoroutines := 5
	cardsPerGoroutine := 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch goroutines to create cards concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < cardsPerGoroutine; j++ {
				title := fmt.Sprintf("Card %d-%d", id, j)
				question := fmt.Sprintf("Question %d-%d", id, j)
				answer := fmt.Sprintf("Answer %d-%d", id, j)
				tags := []string{fmt.Sprintf("tag%d", id), "test"}

				_, err := store.CreateCard(title, question, answer, tags)
				if err != nil {
					t.Errorf("Failed to create card in goroutine %d: %v", id, err)
				}
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Verify that all cards were created
	expectedCount := numGoroutines * cardsPerGoroutine
	if store.GetCardCount() != expectedCount {
		t.Errorf("Expected %d cards, got %d", expectedCount, store.GetCardCount())
	}
}
