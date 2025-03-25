# GoCard: Service-Oriented Spaced Repetition System

+ github: www.github.com/DavidMiserak/GoCard

## 1. Introduction

GoCard is a lightweight, file-based spaced repetition system (SRS) designed for developers and knowledge workers. It provides a distraction-free terminal interface for learning and retaining information through flashcards. This document outlines the architecture and design of GoCard with a focus on service-oriented principles and interface-based design for maintainability and testability.

### 1.1 Project Goals

+ Create a maintainable, testable flashcard application with clean separation of concerns
+ Provide a simple, intuitive terminal user interface for studying flashcards
+ Support Markdown formatting for rich content including code syntax highlighting
+ Implement effective spaced repetition algorithm (SM-2) for optimal learning
+ Design for a single user on a local machine
+ Allow pointing to any directory to read all markdown files as flashcards
+ Store all review state directly in the cards' YAML frontmatter

### 1.2 Non-Goals

+ Real-time file watching or external editor integration
+ In-app card or deck editing - cards should be created and edited with external text editors
+ Multi-user or distributed operation
+ Web or mobile interfaces
+ Synchronization with third-party services
+ Maintaining a separate state or database - all state is in the markdown files

## 2. System Architecture

GoCard follows a layered architecture with clean separation of concerns through well-defined interfaces:

```sh
┌──────────────────────────────────────────────┐
│                   UI Layer                   │
│    (Terminal UI, Command-Line Interface)     │
└───────────────────┬──────────────────────────┘
                    │
┌───────────────────▼──────────────────────────┐
│              Application Layer               │
│    (Use Cases, Application Orchestration)    │
└───────────────────┬──────────────────────────┘
                    │
┌───────────────────▼──────────────────────────┐
│                Domain Layer                  │
│         (Core Models and Entities)           │
└───────────────────┬──────────────────────────┘
                    │
┌───────────────────▼──────────────────────────┐
│             Infrastructure Layer             │
│    (Storage, Rendering, External Services)   │
└──────────────────────────────────────────────┘
```

### 2.1 Key Design Principles

1. **Interface-Based Design**: All components interact through well-defined interfaces, allowing for easy replacement of implementations
2. **Dependency Inversion**: Higher-level modules depend on abstractions, not concrete implementations
3. **Single Responsibility**: Each component has one primary responsibility
4. **Testability**: Components are designed to be easily tested in isolation
5. **Immutability**: Prefer immutable data structures where possible
6. **Simplicity**: Favor simple, straightforward solutions over complex ones
7. **File-First Approach**: Direct mapping between filesystem structure and application model
8. **Self-Contained Data**: All review state stored directly in card frontmatter

## 3. Core Domain Model

The domain layer contains the core business entities and logic, with file paths as natural identifiers.

### 3.1 Card

```go
// Card represents a flashcard with question, answer, and metadata
type Card struct {
    FilePath       string    // Path to the card file (serves as identifier)
    Title          string    // Card title
    Question       string    // Question text (supports Markdown)
    Answer         string    // Answer text (supports Markdown)
    Tags           []string  // Tags for categorization
    Created        time.Time // Creation timestamp
    LastReviewed   time.Time // Last review timestamp - from frontmatter
    ReviewInterval int       // Current interval in days - from frontmatter
    Difficulty     int       // Difficulty rating (0-5) - from frontmatter
    RawContent     string    // Raw markdown content including frontmatter
    Frontmatter    map[string]interface{} // Parsed frontmatter
}
```

### 3.2 Deck

```go
// Deck represents a collection of cards
type Deck struct {
    Path       string // Directory path (serves as identifier)
    Name       string // Directory name for display (derived from path)
    ParentPath string // Path to parent directory or empty for root
}
```

### 3.3 Review Session

```go
// ReviewSession represents an active review of cards
type ReviewSession struct {
    DeckPath    string    // Path of the deck being reviewed
    StartTime   time.Time // Time the session started
    CardPaths   []string  // File paths of cards in the session
    CurrentCard int       // Index of current card
    Completed   []bool    // Tracks which cards are completed
    Ratings     []int     // Rating given to each card
}
```

## 4. Service Interfaces

### 4.1 CardService

```go
// CardService manages operations on individual cards
type CardService interface {
    // Card read operations (no creation/editing in-app)
    GetCard(cardPath string) (Card, error)

    // Review operations
    ReviewCard(cardPath string, rating int) error
    IsDue(cardPath string) bool
    GetDueDate(cardPath string) time.Time
}
```

### 4.2 DeckService

```go
// DeckService manages operations on decks and card collections
type DeckService interface {
    // Deck read operations
    GetDeck(deckPath string) (Deck, error)

    // Deck hierarchy operations
    GetSubdecks(deckPath string) ([]Deck, error)
    GetParentDeck(deckPath string) (Deck, error)

    // Card collection operations
    GetCards(deckPath string) ([]Card, error)
    GetDueCards(deckPath string) ([]Card, error)
    GetCardStats(deckPath string) (map[string]int, error)
}
```

### 4.3 ReviewService

```go
// ReviewService manages the review process
type ReviewService interface {
    // Session management
    StartSession(deckPath string) (ReviewSession, error)
    GetSession() (ReviewSession, error)
    EndSession() (ReviewSessionSummary, error)

    // Card review operations
    GetNextCard() (Card, error)
    SubmitRating(rating int) error
    GetSessionStats() (map[string]interface{}, error)
}

// ReviewSessionSummary contains statistics about a completed review session
type ReviewSessionSummary struct {
    DeckPath      string
    Duration      time.Duration
    CardsReviewed int
    AverageRating float64
    NewCards      int
    ReviewedCards int
}
```

### 4.4 StorageService

```go
// StorageService handles persistence of cards and decks
type StorageService interface {
    // Initialization and cleanup
    Initialize(rootDir string) error
    Close() error

    // Card operations
    LoadCard(filePath string) (Card, error)
    UpdateCardMetadata(card Card) error // Updates frontmatter for review state
    ListCardPaths(deckPath string) ([]string, error)

    // Frontmatter operations
    ParseFrontmatter(content []byte) (map[string]interface{}, []byte, error)
    UpdateFrontmatter(content []byte, updates map[string]interface{}) ([]byte, error)

    // Deck operations
    LoadDeck(dirPath string) (Deck, error)
    ListDeckPaths(parentPath string) ([]string, error)

    // Query operations
    FindCardsByTag(tag string) ([]Card, error)
    SearchCards(query string) ([]Card, error)
}
```

### 4.5 RenderService

```go
// RenderService handles rendering of content for display
type RenderService interface {
    // Markdown rendering
    RenderMarkdown(content string) (string, error)
    RenderMarkdownWithTheme(content string, theme string) (string, error)

    // Code syntax highlighting
    GetAvailableCodeThemes() []string
    SetCodeTheme(theme string)
    EnableLineNumbers(enabled bool)

    // UI styling
    StyleHeading(text string, level int) string
    StyleInfo(text string) string
    StyleWarning(text string) string
    StyleError(text string) string
}
```

### 4.6 ConfigService

```go
// ConfigService manages application configuration
type ConfigService interface {
    // Configuration management
    GetConfig() (Config, error)
    SetConfig(config Config) error

    // Individual settings
    GetString(key string, defaultValue string) string
    GetInt(key string, defaultValue int) int
    GetBool(key string, defaultValue bool) bool
    GetFloat(key string, defaultValue float64) float64

    // Storage
    SaveConfig() error
    ResetToDefaults() error
}

// Config holds application configuration
type Config struct {
    CardsDir           string  // Root directory for cards
    Theme              string  // UI theme
    CodeTheme          string  // Code syntax highlighting theme
    EasyBonus          float64 // SM-2 algorithm easy bonus multiplier
    IntervalModifier   float64 // SM-2 algorithm interval modifier
    NewCardsPerDay     int     // Limit on new cards per day
    MaxInterval        int     // Maximum review interval in days
    ShowLineNumbers    bool    // Whether to show line numbers in code blocks
}
```

## 5. Service Implementations

### 5.1 FileSystemStorage

The `FileSystemStorage` implements the `StorageService` interface using the local filesystem:

+ Cards are stored as Markdown files with YAML frontmatter
+ Decks are represented by directories
+ The directory structure mirrors the deck hierarchy
+ Each file's path serves as its natural identifier
+ Review state is stored directly in card frontmatter

```go
// FileSystemStorage implements StorageService using local filesystem
type FileSystemStorage struct {
    rootDir      string
    cardCache    map[string]Card    // Path -> Card
    deckCache    map[string]Deck    // Path -> Deck
    frontmatter  *yaml.Unmarshaler  // YAML parser for frontmatter
}
```

#### Frontmatter Handling

```go
// Example implementation of UpdateCardMetadata
func (fs *FileSystemStorage) UpdateCardMetadata(card Card) error {
    // Read the current file content
    content, err := os.ReadFile(card.FilePath)
    if err != nil {
        return err
    }

    // Prepare updated frontmatter
    updates := map[string]interface{}{
        "last_reviewed": card.LastReviewed.Format("2006-01-02"),
        "review_interval": card.ReviewInterval,
        "difficulty": card.Difficulty,
    }

    // Update the frontmatter in the content
    newContent, err := fs.UpdateFrontmatter(content, updates)
    if err != nil {
        return err
    }

    // Write the updated content back to the file
    return os.WriteFile(card.FilePath, newContent, 0644)
}
```

### 5.2 SM2Algorithm

Implements the spaced repetition algorithm as a separate component that can be injected into the `CardService`:

```go
// SM2Algorithm implements the SuperMemo-2 spaced repetition algorithm
type SM2Algorithm struct {
    EasyBonus        float64
    IntervalModifier float64
    MaxInterval      int
}

// SM2Algorithm methods
func (sm2 *SM2Algorithm) CalculateNextInterval(card Card, rating int) int {
    // Implementation of the SM-2 algorithm
}

func (sm2 *SM2Algorithm) IsDue(card Card) bool {
    // Check if a card is due for review
}
```

### 5.3 DefaultCardService

```go
// DefaultCardService implements CardService interface
type DefaultCardService struct {
    storage   StorageService
    algorithm *SM2Algorithm
}
```

### 5.4 DefaultDeckService

```go
// DefaultDeckService implements DeckService interface
type DefaultDeckService struct {
    storage StorageService
    cardSvc CardService
}
```

### 5.5 DefaultReviewService

```go
// DefaultReviewService implements ReviewService interface
type DefaultReviewService struct {
    storage   StorageService
    cardSvc   CardService
    deckSvc   DeckService
    algorithm *SM2Algorithm
    session   *ReviewSession
}
```

### 5.6 MarkdownRenderer

```go
// MarkdownRenderer implements RenderService using Goldmark with Chroma highlighting
type MarkdownRenderer struct {
    width          int
    codeTheme      string
    showLineNumbers bool
    styles         map[string]lipgloss.Style
}

// Implementation of RenderMarkdown with syntax highlighting
func (r *MarkdownRenderer) RenderMarkdown(content string) (string, error) {
    // Set up Goldmark with Chroma highlighting
    md := goldmark.New(
        goldmark.WithExtensions(
            extension.GFM,
            highlighting.NewHighlighting(
                highlighting.WithStyle(r.codeTheme),
                highlighting.WithFormatOptions(
                    chromahtml.WithLineNumbers(r.showLineNumbers),
                ),
            ),
        ),
    )

    var buf bytes.Buffer
    if err := md.Convert([]byte(content), &buf); err != nil {
        return "", err
    }

    return buf.String(), nil
}
```

### 5.7 YAMLConfig

```go
// YAMLConfig implements ConfigService using YAML files
type YAMLConfig struct {
    configPath string
    config     Config
}
```

## 6. UI Design

### 6.1 TUI (Terminal User Interface)

The TUI is built using the Bubble Tea framework and consists of multiple views:

1. **DeckListView**: Browse and navigate the deck hierarchy
2. **DeckBrowserView**: View details and statistics for a specific deck
3. **ReviewView**: Conduct a review session with flashcards

Each view interacts with the application services through well-defined interfaces, allowing the UI to be tested independently from the service implementations.

### 6.2 CLI (Command Line Interface)

In addition to the TUI, GoCard provides a command-line interface for quick operations:

```sh
gocard [options]

Options:
  -d, --directory        Set cards directory (default: ~/GoCard)
  -c, --config           Specify config file (default: ~/.gocard.yaml)
  --theme                Set UI theme
  --code-theme           Set code syntax highlighting theme
  --help                 Show help
  --version              Show version information
```

## 7. Project Structure

```sh
gocard/
├── cmd/
│   ├── gocard/              # Main application entry point
│   └── utilities/           # Command-line utilities
├── internal/
│   ├── domain/              # Domain models and core logic
│   │   ├── card.go          # Card entity and related types
│   │   ├── deck.go          # Deck entity and related types
│   │   └── review.go        # Review session and algorithm
│   ├── service/             # Service interfaces and implementations
│   │   ├── interfaces/      # Service interfaces
│   │   ├── card/            # Card service implementation
│   │   ├── deck/            # Deck service implementation
│   │   ├── review/          # Review service implementation
│   │   ├── storage/         # Storage service implementation
│   │   ├── render/          # Render service implementation
│   │   └── config/          # Config service implementation
│   ├── ui/                  # User interface
│   │   ├── tui/             # Terminal UI components
│   │   ├── cli/             # Command-line interface
│   │   └── views/           # TUI views
│   └── util/                # Utility functions and helpers
├── pkg/                     # Public packages that can be imported by other projects
│   ├── algorithm/           # Spaced repetition algorithms
│   └── markdown/            # Markdown processing utilities
└── test/                    # Integration tests
```
