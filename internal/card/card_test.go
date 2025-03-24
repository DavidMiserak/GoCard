// File: internal/card/card_test.go
package card

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewCard(t *testing.T) {
	title := "Test Card"
	question := "Test Question"
	answer := "Test Answer"
	tags := []string{"tag1", "tag2"}

	c := NewCard(title, question, answer, tags)

	// Basic field testing
	if c.Title != title {
		t.Errorf("Expected title %q, got %q", title, c.Title)
	}
	if c.Question != question {
		t.Errorf("Expected question %q, got %q", question, c.Question)
	}
	if c.Answer != answer {
		t.Errorf("Expected answer %q, got %q", answer, c.Answer)
	}

	// Tags testing
	if len(c.Tags) != len(tags) {
		t.Errorf("Expected %d tags, got %d", len(tags), len(c.Tags))
	}
	for i, tag := range tags {
		if c.Tags[i] != tag {
			t.Errorf("Expected tag %q, got %q", tag, c.Tags[i])
		}
	}

	// Default values for review fields
	if c.ReviewInterval != 0 {
		t.Errorf("Expected initial ReviewInterval to be 0, got %d", c.ReviewInterval)
	}
	if c.Difficulty != 0 {
		t.Errorf("Expected initial Difficulty to be 0, got %d", c.Difficulty)
	}
	if !c.LastReviewed.IsZero() {
		t.Errorf("Expected initial LastReviewed to be zero time, got %v", c.LastReviewed)
	}
	if c.Created.IsZero() {
		t.Error("Expected Created time to be set, got zero time")
	}
}

// Test for JSON serialization/deserialization
func TestCardSerialization(t *testing.T) {
	now := time.Now().Truncate(time.Second) // Truncate to avoid precision issues
	lastReviewTime := now.Add(-24 * time.Hour)

	original := &Card{
		FilePath:       "/path/to/card.md",
		Title:          "Serialization Test",
		Tags:           []string{"json", "test"},
		Created:        now,
		LastReviewed:   lastReviewTime,
		ReviewInterval: 3,
		Difficulty:     4,
		Question:       "Does JSON serialization work?",
		Answer:         "Yes, it works!",
	}

	// Serialize to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to serialize card: %v", err)
	}

	// Deserialize back to card
	var deserialized Card
	if err := json.Unmarshal(data, &deserialized); err != nil {
		t.Fatalf("Failed to deserialize card: %v", err)
	}

	// Compare fields
	if original.Title != deserialized.Title {
		t.Errorf("Title mismatch: got %q, want %q", deserialized.Title, original.Title)
	}
	if original.FilePath != deserialized.FilePath {
		t.Errorf("FilePath mismatch: got %q, want %q", deserialized.FilePath, original.FilePath)
	}
	if original.Question != deserialized.Question {
		t.Errorf("Question mismatch: got %q, want %q", deserialized.Question, original.Question)
	}
	if original.Answer != deserialized.Answer {
		t.Errorf("Answer mismatch: got %q, want %q", deserialized.Answer, original.Answer)
	}
	if original.ReviewInterval != deserialized.ReviewInterval {
		t.Errorf("ReviewInterval mismatch: got %d, want %d", deserialized.ReviewInterval, original.ReviewInterval)
	}
	if original.Difficulty != deserialized.Difficulty {
		t.Errorf("Difficulty mismatch: got %d, want %d", deserialized.Difficulty, original.Difficulty)
	}

	// Compare time fields (allowing minor tolerance due to serialization)
	if !original.Created.Equal(deserialized.Created) {
		t.Errorf("Created time mismatch: got %v, want %v", deserialized.Created, original.Created)
	}
	if !original.LastReviewed.Equal(deserialized.LastReviewed) {
		t.Errorf("LastReviewed time mismatch: got %v, want %v", deserialized.LastReviewed, original.LastReviewed)
	}

	// Compare tags
	if len(original.Tags) != len(deserialized.Tags) {
		t.Errorf("Tags length mismatch: got %d, want %d", len(deserialized.Tags), len(original.Tags))
	} else {
		for i, tag := range original.Tags {
			if tag != deserialized.Tags[i] {
				t.Errorf("Tag mismatch at index %d: got %q, want %q", i, deserialized.Tags[i], tag)
			}
		}
	}
}
