// internal/service/storage/filesystem.go
package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DavidMiserak/GoCard/internal/domain"
	"github.com/DavidMiserak/GoCard/internal/service/interfaces"

	"gopkg.in/yaml.v3"
)

// FileSystemStorage implements StorageService using the local filesystem
type FileSystemStorage struct {
	rootDir   string
	cardCache map[string]domain.Card // Path -> Card
	deckCache map[string]domain.Deck // Path -> Deck
}

// NewFileSystemStorage creates a new filesystem-based storage service
func NewFileSystemStorage() *FileSystemStorage {
	return &FileSystemStorage{
		cardCache: make(map[string]domain.Card),
		deckCache: make(map[string]domain.Deck),
	}
}

// Initialize sets up the storage with the root directory
func (fs *FileSystemStorage) Initialize(rootDir string) error {
	// Make sure the directory exists
	info, err := os.Stat(rootDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Try to create the directory
			if err := os.MkdirAll(rootDir, 0755); err != nil {
				return fmt.Errorf("failed to create root directory: %w", err)
			}
		} else {
			return fmt.Errorf("failed to access root directory: %w", err)
		}
	} else if !info.IsDir() {
		return fmt.Errorf("specified path is not a directory: %s", rootDir)
	}

	fs.rootDir = rootDir
	return nil
}

// Close cleans up any resources
func (fs *FileSystemStorage) Close() error {
	// Clear caches
	fs.cardCache = make(map[string]domain.Card)
	fs.deckCache = make(map[string]domain.Deck)
	return nil
}

// ForceCardIntoCache adds a card directly to the cache (for testing)
func (fs *FileSystemStorage) ForceCardIntoCache(card domain.Card) {
	fs.cardCache[card.FilePath] = card
}

// LoadCard loads a card from a file
func (fs *FileSystemStorage) LoadCard(filePath string) (domain.Card, error) {
	// Check cache first
	if card, ok := fs.cardCache[filePath]; ok {
		return card, nil
	}

	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return domain.Card{}, fmt.Errorf("failed to read card file: %w", err)
	}

	// Parse frontmatter and content
	frontmatter, markdown, err := fs.ParseFrontmatter(content)
	if err != nil {
		return domain.Card{}, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Create a new card
	card := domain.NewCard(filePath)
	card.RawContent = string(content)
	card.Frontmatter = frontmatter

	// Extract title, tags, and other metadata from frontmatter
	if title, ok := frontmatter["title"].(string); ok {
		card.Title = title
	} else {
		// Use filename as title if not specified
		card.Title = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	}

	// Extract tags
	if tags, ok := frontmatter["tags"].([]interface{}); ok {
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				card.Tags = append(card.Tags, tagStr)
			}
		}
	}

	// Parse markdown content for question and answer
	parts := strings.Split(string(markdown), "---")
	if len(parts) >= 1 {
		card.Question = strings.TrimSpace(parts[0])
	}
	if len(parts) >= 2 {
		card.Answer = strings.TrimSpace(parts[1])
	}

	// Extract review metadata
	if created, ok := frontmatter["created"].(string); ok {
		if t, err := time.Parse("2006-01-02", created); err == nil {
			card.Created = t
		}
	}

	// Enhanced date parsing for last_reviewed
	if lastReviewed, ok := frontmatter["last_reviewed"].(string); ok {
		// This will explicitly parse the date string
		t, err := time.Parse("2006-01-02", lastReviewed)
		if err == nil {
			card.LastReviewed = t
			fmt.Printf("DEBUG: Successfully parsed last_reviewed: %s -> %v\n",
				lastReviewed, t.Format("2006-01-02"))
		} else {
			fmt.Printf("DEBUG: Failed to parse last_reviewed: %s, error: %v\n",
				lastReviewed, err)
		}
	} else {
		// Check what type the value actually is
		if lastReviewed, exists := frontmatter["last_reviewed"]; exists {
			fmt.Printf("DEBUG: last_reviewed exists but is type: %T, value: %v\n",
				lastReviewed, lastReviewed)
		}
	}

	// Handle numeric values from YAML
	if interval, ok := frontmatter["review_interval"].(int); ok {
		card.ReviewInterval = interval
	} else if intervalFloat, ok := frontmatter["review_interval"].(float64); ok {
		// YAML parser often converts numbers to float64
		card.ReviewInterval = int(intervalFloat)
	}
	if difficulty, ok := frontmatter["difficulty"].(int); ok {
		card.Difficulty = difficulty
	} else if difficultyFloat, ok := frontmatter["difficulty"].(float64); ok {
		// YAML parser often converts numbers to float64
		card.Difficulty = int(difficultyFloat)
	}

	// Cache the card
	fs.cardCache[filePath] = *card

	return *card, nil
}

// UpdateCardMetadata updates the frontmatter in a card file
func (fs *FileSystemStorage) UpdateCardMetadata(card domain.Card) error {
	// Read the current file content
	content, err := os.ReadFile(card.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read card file: %w", err)
	}

	// Prepare updated frontmatter
	updates := map[string]interface{}{
		"last_reviewed":   card.LastReviewed.Format("2006-01-02"),
		"review_interval": card.ReviewInterval,
		"difficulty":      card.Difficulty,
	}

	// Update the frontmatter in the content
	newContent, err := fs.UpdateFrontmatter(content, updates)
	if err != nil {
		return fmt.Errorf("failed to update frontmatter: %w", err)
	}

	// Write the updated content back to the file
	if err := os.WriteFile(card.FilePath, newContent, 0644); err != nil {
		return fmt.Errorf("failed to write updated card file: %w", err)
	}

	// Update cache
	fs.cardCache[card.FilePath] = card

	return nil
}

// ListCardPaths finds all markdown files in a directory
func (fs *FileSystemStorage) ListCardPaths(deckPath string) ([]string, error) {
	var cardPaths []string

	err := filepath.Walk(deckPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".markdown")) {
			cardPaths = append(cardPaths, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list card paths: %w", err)
	}

	return cardPaths, nil
}

// ParseFrontmatter extracts YAML frontmatter from content
func (fs *FileSystemStorage) ParseFrontmatter(content []byte) (map[string]interface{}, []byte, error) {
	const frontmatterDelimiter = "---"

	strContent := string(content)

	// Check if content starts with frontmatter delimiter
	if !strings.HasPrefix(strContent, frontmatterDelimiter) {
		// No frontmatter, return empty map and original content
		return make(map[string]interface{}), content, nil
	}

	// Find the end of the frontmatter
	restContent := strContent[len(frontmatterDelimiter):]
	endIndex := strings.Index(restContent, frontmatterDelimiter)
	if endIndex == -1 {
		// No closing delimiter, treat as if there's no frontmatter
		return make(map[string]interface{}), content, nil
	}

	// Extract the frontmatter content
	yamlContent := restContent[:endIndex]
	yamlContent = strings.TrimSpace(yamlContent) // Trim whitespace before parsing

	// Parse the YAML
	var frontmatter map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &frontmatter); err != nil {
		return nil, nil, fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	// Extract the remaining content after frontmatter
	markdownContent := restContent[endIndex+len(frontmatterDelimiter):]

	return frontmatter, []byte(markdownContent), nil
}

// UpdateFrontmatter updates or adds frontmatter to content
func (fs *FileSystemStorage) UpdateFrontmatter(content []byte, updates map[string]interface{}) ([]byte, error) {
	// Parse existing frontmatter
	frontmatter, markdown, err := fs.ParseFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse existing frontmatter: %w", err)
	}

	// Update frontmatter with new values
	for key, value := range updates {
		frontmatter[key] = value
	}

	// Convert updated frontmatter back to YAML
	yamlBytes, err := yaml.Marshal(frontmatter)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updated frontmatter: %w", err)
	}

	// Combine updated frontmatter with original markdown content
	result := fmt.Sprintf("---\n%s---\n%s", string(yamlBytes), string(markdown))

	return []byte(result), nil
}

// LoadDeck loads a deck from a directory
func (fs *FileSystemStorage) LoadDeck(dirPath string) (domain.Deck, error) {
	// Check cache first
	if deck, ok := fs.deckCache[dirPath]; ok {
		return deck, nil
	}

	// Check if directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		return domain.Deck{}, fmt.Errorf("failed to access deck directory: %w", err)
	}
	if !info.IsDir() {
		return domain.Deck{}, fmt.Errorf("specified path is not a directory: %s", dirPath)
	}

	// Create a new deck
	deck := domain.NewDeck(dirPath)

	// Cache the deck
	fs.deckCache[dirPath] = *deck

	return *deck, nil
}

// ListDeckPaths finds all subdirectories in a directory
func (fs *FileSystemStorage) ListDeckPaths(parentPath string) ([]string, error) {
	var deckPaths []string

	entries, err := os.ReadDir(parentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			deckPath := filepath.Join(parentPath, entry.Name())
			deckPaths = append(deckPaths, deckPath)
		}
	}

	return deckPaths, nil
}

// FindCardsByTag finds all cards with a specific tag
func (fs *FileSystemStorage) FindCardsByTag(tag string) ([]domain.Card, error) {
	var matchingCards []domain.Card

	err := filepath.Walk(fs.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".markdown")) {
			card, err := fs.LoadCard(path)
			if err != nil {
				return nil // Skip this card but continue
			}

			for _, cardTag := range card.Tags {
				if cardTag == tag {
					matchingCards = append(matchingCards, card)
					break
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error searching for cards by tag: %w", err)
	}

	return matchingCards, nil
}

// SearchCards finds cards matching a query string
func (fs *FileSystemStorage) SearchCards(query string) ([]domain.Card, error) {
	if query == "" {
		return nil, errors.New("search query cannot be empty")
	}

	query = strings.ToLower(query)
	var matchingCards []domain.Card

	err := filepath.Walk(fs.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".markdown")) {
			card, err := fs.LoadCard(path)
			if err != nil {
				return nil // Skip this card but continue
			}

			// Check if query matches title, question, or answer
			if strings.Contains(strings.ToLower(card.Title), query) ||
				strings.Contains(strings.ToLower(card.Question), query) ||
				strings.Contains(strings.ToLower(card.Answer), query) {
				matchingCards = append(matchingCards, card)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error searching for cards: %w", err)
	}

	return matchingCards, nil
}

// Ensure FileSystemStorage implements StorageService
var _ interfaces.StorageService = (*FileSystemStorage)(nil)
