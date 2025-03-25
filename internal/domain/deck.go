// internal/domain/deck.go
package domain

import (
	"path/filepath"
	"strings"
)

// Deck represents a collection of cards
type Deck struct {
	Path       string // Directory path (serves as identifier)
	Name       string // Directory name for display (derived from path)
	ParentPath string // Path to parent directory or empty for root
}

// NewDeck creates a new Deck from a directory path
func NewDeck(dirPath string) *Deck {
	// Clean and normalize the path
	path := filepath.Clean(dirPath)

	// Get the base name of the directory for display
	name := filepath.Base(path)

	// Get parent path
	parentPath := filepath.Dir(path)
	// If we're at the root, set parent path to empty
	if parentPath == path || parentPath == "." {
		parentPath = ""
	}

	return &Deck{
		Path:       path,
		Name:       name,
		ParentPath: parentPath,
	}
}

// GetRelativePath returns a path relative to a root directory
func (d *Deck) GetRelativePath(rootDir string) string {
	rel, err := filepath.Rel(rootDir, d.Path)
	if err != nil {
		return d.Name
	}
	return rel
}

// GetHierarchyPath returns a path representation showing the full hierarchy
func (d *Deck) GetHierarchyPath(rootDir string) string {
	rel := d.GetRelativePath(rootDir)
	if rel == "." {
		return d.Name
	}
	return strings.ReplaceAll(rel, string(filepath.Separator), " > ")
}
