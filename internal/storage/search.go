// File: internal/storage/search.go

// Package storage implements the file-based storage system for GoCard.
// This file contains operations for searching and filtering cards.
package storage

import (
	"strings"

	"github.com/DavidMiserak/GoCard/internal/card"
)

// GetAllTags returns a list of all unique tags used in cards
func (s *CardStore) GetAllTags() []string {
	tagMap := make(map[string]bool)

	for _, cardObj := range s.Cards {
		for _, tag := range cardObj.Tags {
			tagMap[tag] = true
		}
	}

	tags := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		tags = append(tags, tag)
	}

	return tags
}

// GetCardsByTag returns all cards with a specific tag
func (s *CardStore) GetCardsByTag(tag string) []*card.Card {
	var result []*card.Card

	for _, cardObj := range s.Cards {
		for _, cardTag := range cardObj.Tags {
			if cardTag == tag {
				result = append(result, cardObj)
				break
			}
		}
	}

	return result
}

// SearchCards searches for cards matching the given text in title, question, or answer
func (s *CardStore) SearchCards(searchText string) []*card.Card {
	var result []*card.Card

	// Convert search text to lowercase for case-insensitive matching
	searchLower := strings.ToLower(searchText)

	for _, cardObj := range s.Cards {
		// Check title, question, and answer for the search text
		titleMatch := strings.Contains(strings.ToLower(cardObj.Title), searchLower)
		questionMatch := strings.Contains(strings.ToLower(cardObj.Question), searchLower)
		answerMatch := strings.Contains(strings.ToLower(cardObj.Answer), searchLower)

		if titleMatch || questionMatch || answerMatch {
			result = append(result, cardObj)
		}
	}

	return result
}
