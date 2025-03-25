# GoCard: A File-Based Spaced Repetition System

[![Go Report Card](https://goreportcard.com/badge/github.com/DavidMiserak/GoCard)](https://goreportcard.com/report/github.com/DavidMiserak/GoCard)
[![Build Status](https://github.com/DavidMiserak/GoCard/workflows/Go/badge.svg)](https://github.com/DavidMiserak/GoCard/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

![GoCard Logo](assets/gocard-logo.webp)

GoCard is a lightweight, file-based spaced repetition system (SRS) built in Go.
It uses plain Markdown files organized in directories as its data source, making it perfect for developers who prefer working with text files and version control.

> **Project Status**: Version 0.2.0 - Enhanced markdown support, terminal UI improvements, card editing interface with auto-save, tutorial mode for first-time users, and more.

## Features

- **File-Based Storage**: All flashcards are stored as Markdown files in regular directories
- **Git-Friendly**: Easily track changes, collaborate, and back up your knowledge base
- **Terminal Interface**: Clean, distraction-free TUI for focused learning
- **Rich Markdown Support**: Full Markdown rendering with Goldmark parser
- **Code-Focused**: Enhanced features for programming-related cards:
  - Syntax highlighting for 50+ languages with customizable themes
  - Side-by-side diff view for comparing code
- **Card Editor**: Built-in editor with auto-save and markdown preview
- **Cross-Platform**: Works on Linux, macOS, and Windows
- **Spaced Repetition Algorithm**: Enhanced SM-2 algorithm implementation
- **Real-time File Watching**: Changes to card files are automatically detected and loaded
- **Hierarchical Deck Navigation**: Browse your decks with vim and emacs keybindings
- **Tutorial Mode**: Interactive guidance for first-time users
- **Session Statistics**: Detailed summaries of your learning progress

## Installation

### From Binary Release

Download the latest binary for your platform from the
[releases page](https://github.com/DavidMiserak/GoCard/releases).

### Using Go Install

```bash
go install github.com/DavidMiserak/GoCard/cmd/gocard@latest
```

### Building from Source

```bash
git clone https://github.com/DavidMiserak/GoCard.git
cd GoCard
go build -o GoCard ./cmd/gocard
```

## Project Structure

GoCard follows a standard Go project layout with a focus on modularity and clean separation of concerns:

```sh
github.com/DavidMiserak/GoCard/
├── cmd/gocard/                      # Main application entry point
├── internal/                        # Private implementation packages
│   ├── algorithm/                   # Spaced repetition algorithms (SM-2)
│   ├── card/                        # Core card data models
│   ├── deck/                        # Deck management and organization
│   ├── storage/                     # File-based card storage
│   │   ├── card_store.go            # CardStore core struct
│   │   ├── card_ops.go              # Card operations
│   │   ├── deck_ops.go              # Deck operations
│   │   ├── review.go                # Review functionality
│   │   ├── stats.go                 # Statistics and metrics
│   │   ├── search.go                # Search functionality
│   │   ├── io/                      # File system operations
│   │   │   ├── file_ops.go          # File handling utilities
│   │   │   ├── logger.go            # Logging system
│   │   │   └── watcher.go           # File change monitoring
│   │   └── parser/                  # Content parsing
│   │       ├── markdown.go          # Markdown parsing
│   │       └── formatter.go         # Markdown formatting
│   └── ui/                          # Terminal user interface
│       ├── input/                   # Input handling and key bindings
│       ├── render/                  # Rendering utilities
│       └── views/                   # Screen components
│           ├── deck_browser_view.go # Deck content browser
│           ├── deck_list_view.go    # Hierarchical deck navigation
│           ├── review_view.go       # Card review interface
│           └── views.go             # View interfaces and common code
├── assets/                          # Application resources
└── docs/                            # Documentation
```

## Command-Line Options

GoCard supports the following command-line options:

```sh
Usage: gocard [options] [cards_directory]

Options:
-tui        Use terminal UI mode (default: true)
-example    Run in example mode with sample cards
-verbose    Enable detailed logging (useful for debugging)
-config     Path to configuration file (default: ~/.gocard.yaml)
-version    Show version information
-h, -help   Show help information
```

- The optional `cards_directory` parameter specifies the directory for your flashcards (default: ~/GoCard)
- The `-example` flag creates sample decks and cards to demonstrate GoCard functionality
- The `-verbose` flag enables detailed logs of file operations and debugging information
- The `-config` flag allows you to specify a custom configuration file path

## Quick Start

1. Create a directory for your flashcards:

```bash
mkdir -p ~/GoCard/programming
```

2. Create your first card as a markdown file:

```bash
touch ~/GoCard/programming/two-pointer-technique.md
```

3. Edit the file with the following markdown structure:

```markdown
---
tags: [algorithms, techniques, arrays]
created: 2023-04-15
review_interval: 0
---

# Two-Pointer Technique

## Question

What is the two-pointer technique in algorithms and when should it be used?

## Answer

The two-pointer technique uses two pointers to iterate through a data structure simultaneously.

It's particularly useful for:
- Sorted array operations
- Finding pairs with certain conditions
- String manipulation (palindromes)
- Linked list cycle detection

Example (Two Sum in sorted array):
```python
def two_sum(nums, target):
  left, right = 0, len(nums) - 1
  while left < right:
    current_sum = nums[left] + nums[right]
    if current_sum == target:
      return [left, right]
    elif current_sum < target:
      left += 1
      else:
      right -= 1
    return [-1, -1]  # No solution
```

4. Launch GoCard and point it to your directory:

```bash
gocard ~/GoCard
```

## Tutorial Mode

GoCard includes an interactive tutorial mode for first-time users. When you first run the application, you'll be guided through:

- Basic GoCard concepts and organization
- File-based storage structure
- Deck navigation and review process
- Card creation and editing
- Keyboard shortcuts and navigation

The tutorial includes sample cards that demonstrate effective use of various features.

## Card Creation and Editing

GoCard offers a built-in card editor with:

- Title and tags input fields
- Question and answer editing with multi-line support
- Markdown preview mode (toggle with `Ctrl+p`)
- Auto-save functionality to prevent data loss
- Syntax highlighting preview for code blocks

To access the editor:

- Press `Ctrl+n` to create a new card
- Press `Ctrl+e` to edit an existing card

## File Format

Cards are stored as markdown files with a YAML frontmatter section for metadata:

```markdown
---
tags: [tag1, tag2, tag3]
created: YYYY-MM-DD
last_reviewed: YYYY-MM-DD
review_interval: N
difficulty: 0-5
---

# Card Title

## Question

Your question goes here. This can be multiline and include any markdown.

## Answer

Your answer goes here. This can include:
- Lists
- Code blocks
- Images
- Tables
- And any other markdown formatting
```

You can edit these files directly with any text editor, and GoCard will automatically detect changes.

## Keyboard Shortcuts

| Key                  | Action                     |
|----------------------|----------------------------|
| `Space`              | Show answer                |
| `0-5`                | Rate card difficulty       |
| `Ctrl+o`             | Change deck                |
| `Ctrl+alt+n`         | Create new deck            |
| `Ctrl+n`             | Create new card            |
| `Ctrl+e`             | Edit current card          |
| `Ctrl+p`             | Toggle preview mode        |
| `Ctrl+s`             | Save card                  |
| `Ctrl+shift+s`       | Save and exit              |
| `Ctrl+f`             | Search cards               |
| `Ctrl+h` or `F1`     | Toggle help                |
| `Ctrl+q`             | Quit                       |
| **Navigation Keys**  |                            |
| `↑/k`                | Move up in lists           |
| `↓/j`                | Move down in lists         |
| `Enter`              | Select/move forward        |
| `Esc`                | Go back                    |
| `Ctrl+x d`           | Delete current card        |
| `Ctrl+t`             | Add/edit tags              |
| `F2`                 | Rename deck                |
| `Ctrl+alt+d`         | Delete deck                |
| `Ctrl+m`             | Move cards between decks   |

## Configuration

Configuration is stored in `~/.gocard.yaml`:

```yaml
# GoCard Configuration File
# This file controls the behavior of GoCard

# Cards directory (default: ~/GoCard)
cards_dir: "~/GoCard"

# Logging settings
logging:
  # Log level: debug, info, warn, error
  level: "info"
  # Enable file logging
  file_enabled: false
  # Log file path
  file_path: "~/.gocard.log"

  # UI settings
ui:
  # Theme: auto, light, dark
  theme: "auto"
  # Code highlighting theme
  highlight_theme: "monokai"
  # Show line numbers in code blocks
  show_line_numbers: true

  # Spaced repetition settings
spaced_repetition:
  # Bonus for easy cards (higher = longer intervals)
  easy_bonus: 1.3
  # Global interval modifier (higher = longer intervals)
  interval_modifier: 1.0
  # Maximum interval in days
  max_interval: 365
  # Number of new cards per day
  new_cards_per_day: 20
```

## Development

### Setting Up Your Development Environment

```bash
# Clone the repository
git clone https://github.com/DavidMiserak/GoCard.git
cd GoCard

# Install dependencies
go mod download

# Set up pre-commit hooks
make pre-commit-setup
```

### Running Tests

```bash
go test ./...
# or
make test
```

### Linting and Formatting

```bash
# Format code
go fmt ./...
# or
make format

# Run linters
golangci-lint run
```

### Building

```bash
# Build for your current platform
go build -o GoCard ./cmd/gocard
# or
make GoCard

# Clean build artifacts
make clean
```

### CI/CD

GoCard uses GitHub Actions for continuous integration:

- Automated testing on multiple platforms
- Code linting with golangci-lint
- Automated builds for Linux, macOS, and Windows
- Release automation when tagging versions

## Roadmap

The project is under active development with several features planned:

1. **Search and Filter Functionality**: Advanced search capabilities for your card collection
2. **Import/Export Compatibility**: Interoperability with other SRS systems like Anki
3. **Customizable Styling and Themes**: Personalize your learning environment
4. **Code Testing Integration**: Run and test code snippets directly from cards
5. **Configuration File Management**: More customization options

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

Before submitting PR:

1. Ensure tests pass: `make test`
2. Format your code: `make format`
3. Run pre-commit hooks: `pre-commit run --all-files`
4. Follow [conventional commits](https://www.conventionalcommits.org/) for commit messages

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgements

- Inspired by Anki and SuperMemo
- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) terminal UI framework
- Markdown rendering via [Goldmark](https://github.com/yuin/goldmark) and [Glamour](https://github.com/charmbracelet/glamour)
- Terminal styling with [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- Code syntax highlighting via [Chroma](https://github.com/alecthomas/chroma)
