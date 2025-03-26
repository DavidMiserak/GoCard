// internal/ui/tui/deck_list_model.go

package tui

import (
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
