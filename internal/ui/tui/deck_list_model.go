// internal/ui/tui/deck_list_model.go

package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DavidMiserak/GoCard/internal/service/interfaces"
	tea "github.com/charmbracelet/bubbletea"
)

type DeckItem struct {
	Path       string
	Name       string
	TotalCards int
	DueCards   int
}

type DeckListModel struct {
	DeckService    interfaces.DeckService
	StorageService interfaces.StorageService
	RootDir        string
	Decks          []DeckItem
	Cursor         int
	Breadcrumbs    []string
	Keys           DeckListKeyMap
	TerminalWidth  int
	TerminalHeight int
}

func NewDeckListModel(
	deckService interfaces.DeckService,
	storageService interfaces.StorageService,
	rootDir string,
) *DeckListModel {
	return &DeckListModel{
		DeckService:    deckService,
		StorageService: storageService,
		RootDir:        rootDir,
		Breadcrumbs:    []string{"Home"},
		Keys:           DefaultDeckListKeyMap(),
	}
}

func (m *DeckListModel) createDefaultDecks() tea.Cmd {
	return func() tea.Msg {
		// Programming Deck
		programmingDeckPath := filepath.Join(m.RootDir, "Programming")
		goDeckPath := filepath.Join(programmingDeckPath, "Go")
		pythonDeckPath := filepath.Join(programmingDeckPath, "Python")
		languagesDeckPath := filepath.Join(m.RootDir, "Languages")
		spanishDeckPath := filepath.Join(languagesDeckPath, "Spanish")

		// Collect all paths to create
		pathsToCreate := []string{
			programmingDeckPath,
			goDeckPath,
			pythonDeckPath,
			languagesDeckPath,
			spanishDeckPath,
		}

		// Collect any errors during directory creation
		var errs []error
		for _, deck := range pathsToCreate {
			if err := os.MkdirAll(deck, 0755); err != nil {
				errs = append(errs, fmt.Errorf("failed to create directory %s: %w", deck, err))
			}
		}

		// If any errors occurred, return a composite error
		if len(errs) > 0 {
			return errMsg{fmt.Errorf("errors creating default deck directories: %v", errs)}
		}

		decksToCreate := []struct {
			path    string
			content []byte
		}{
			{
				path: filepath.Join(goDeckPath, "concurrency.md"),
				content: []byte(`---
title: Go Concurrency
tags:
  - go
  - programming
difficulty: 3
---
# What is a goroutine?

---

A goroutine is a lightweight thread managed by the Go runtime. It allows concurrent execution of functions.
`),
			},
			{
				path: filepath.Join(pythonDeckPath, "list_comprehensions.md"),
				content: []byte(`---
title: Python List Comprehensions
tags:
  - python
  - programming
difficulty: 2
---
# What is a list comprehension?

---

A list comprehension is a concise way to create lists in Python, offering a compact alternative to using for loops.
`),
			},
			{
				path: filepath.Join(spanishDeckPath, "basic_verbs.md"),
				content: []byte(`---
title: Basic Spanish Verbs
tags:
  - spanish
  - languages
difficulty: 1
---
# What are the basic forms of "ser" and "estar"?

---

"Ser" and "estar" are both forms of "to be" in Spanish, but they are used differently:
- "Ser" is used for permanent characteristics
- "Estar" is used for temporary states or locations
`),
			},
		}

		// Collect any errors during card creation
		errs = []error{}
		for _, cardInfo := range decksToCreate {
			if err := os.WriteFile(cardInfo.path, cardInfo.content, 0644); err != nil {
				errs = append(errs, fmt.Errorf("failed to create card file %s: %w", cardInfo.path, err))
			}
		}

		// If any errors occurred during card creation, return a composite error
		if len(errs) > 0 {
			return errMsg{fmt.Errorf("errors creating default deck cards: %v", errs)}
		}

		return nil
	}
}

func (m *DeckListModel) loadDecks() tea.Cmd {
	return func() tea.Msg {
		currentPath := m.RootDir
		if len(m.Breadcrumbs) > 1 {
			currentPath = m.Breadcrumbs[len(m.Breadcrumbs)-1]
		}

		subdecks, err := m.DeckService.GetSubdecks(currentPath)
		if err != nil {
			return errMsg{err}
		}

		var deckItems []DeckItem
		for _, deck := range subdecks {
			stats, err := m.DeckService.GetCardStats(deck.Path)
			if err != nil {
				continue // Skip decks with stat retrieval errors
			}

			deckItems = append(deckItems, DeckItem{
				Path:       deck.Path,
				Name:       deck.Name,
				TotalCards: stats["total"],
				DueCards:   stats["due"],
			})
		}

		m.Decks = deckItems
		return nil
	}
}

func (m *DeckListModel) Init() tea.Cmd {
	// Check if decks exist, if not, create default decks
	decks, err := m.DeckService.GetSubdecks(m.RootDir)
	if err != nil {
		return func() tea.Msg {
			return errMsg{err}
		}
	}
	if len(decks) == 0 {
		return m.createDefaultDecks()
	}
	return m.loadDecks()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Ensure the model implements tea.Model
var _ tea.Model = (*DeckListModel)(nil)
