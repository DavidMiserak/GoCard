# GoCard: A File-Based Spaced Repetition System

[![Go Report Card](https://goreportcard.com/badge/github.com/DavidMiserak/GoCard)](https://goreportcard.com/report/github.com/DavidMiserak/GoCard)
[![Build Status](https://github.com/DavidMiserak/GoCard/workflows/Go/badge.svg)](https://github.com/DavidMiserak/GoCard/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

![GoCard Logo](assets/gocard-logo.webp)

GoCard is a lightweight, file-based spaced repetition system (SRS) built in Go.
It uses plain Markdown files organized in directories as its data source, making it perfect for developers who prefer working with text files and version control.

## Features

- **File-Based Storage**: All flashcards are stored as Markdown files in regular directories
- **Git-Friendly**: Easily track changes, collaborate, and back up your knowledge base
- **Terminal Interface**: Clean, distraction-free TUI for focused learning
- **Rich Markdown Support**: Full Markdown rendering with syntax highlighting
- **Spaced Repetition Algorithm**: Enhanced SM-2 algorithm implementation
- **Interactive Statistics**: Comprehensive views of your learning progress
- **Cross-Platform**: Works on Linux, macOS, and Windows

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
go build -o gocard ./cmd/gocard
```

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
created: 2025-04-02
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
gocard -dir ~/GoCard
```

## Project Structure

GoCard follows a standard Go project layout with a focus on modularity and clean separation of concerns:

```sh
github.com/DavidMiserak/GoCard/
├── cmd/gocard/                # Main application entry point
├── internal/                  # Private implementation packages
│   ├── data/                  # Data handling and storage
│   │   ├── dummy_store.go     # Sample data for demo mode
│   │   ├── markdown_parser.go # Markdown parsing for cards
│   │   ├── markdown_writer.go # Writing cards back to markdown
│   │   └── store.go           # Main data store functionality
│   ├── model/                 # Data models
│   │   ├── card.go            # Card model
│   │   └── deck.go            # Deck model
│   ├── srs/                   # Spaced repetition algorithm
│   │   └── algorithm.go       # SM-2 implementation
│   └── ui/                    # Terminal user interface
│       ├── browse_decks.go    # Deck browsing screen
│       ├── main_menu.go       # Main menu screen
│       ├── markdown_renderer.go # Markdown rendering
│       ├── stats_screen.go    # Statistics screens
│       ├── study_screen.go    # Card study interface
│       └── styles.go          # UI styling
├── assets/                    # Application resources
└── docs/                      # Documentation
```

## Command-Line Options

GoCard supports the following command-line options:

```sh
Usage: gocard [options]

Options:
-dir        Directory containing flashcard decks (default: ~/GoCard)
```

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

## Key Features

### Spaced Repetition

GoCard implements an enhanced version of the SuperMemo-2 (SM-2) algorithm for optimal learning efficiency:

- **Adaptive Intervals**: Review intervals automatically adjust based on your performance
- **Five-Point Rating Scale**:
  - 1: Blackout (complete failure)
  - 2: Wrong (significant difficulty)
  - 3: Hard (correct with difficulty)
  - 4: Good (correct with some effort)
  - 5: Easy (correct with no effort)
- **Smart Scheduling**: Cards are prioritized based on your learning history

### Rich Statistics

GoCard provides comprehensive statistics to help you track your learning progress:

- **Summary View**: Overall stats including retention rate and daily progress
- **Deck Review**: Deck-specific metrics and rating distribution
- **Review Forecast**: Visual representation of upcoming reviews

### Terminal UI

The clean, distraction-free terminal interface includes:

- **Deck Browser**: Navigate and manage your deck collection
- **Study Interface**: Focus on one card at a time with markdown rendering
- **Statistics Screens**: Interactive visualizations of your progress

## Keyboard Shortcuts

| Key                | Action                   |
|--------------------|--------------------------|
| `Space`            | Show answer              |
| `1-5`              | Rate card difficulty     |
| `↑/k`              | Move up/scroll up        |
| `↓/j`              | Move down/scroll down    |
| `Enter`            | Select/confirm           |
| `Tab`              | Switch tab (in statistics)|
| `b`                | Back to previous screen  |
| `q`                | Quit                     |

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

Before submitting PR:

1. Ensure tests pass: `go test ./...`
2. Format your code: `go fmt ./...`
3. Follow [conventional commits](https://www.conventionalcommits.org/) for commit messages

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgements

- Inspired by Anki and SuperMemo
- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) terminal UI framework
- Markdown rendering via [Goldmark](https://github.com/yuin/goldmark) and [Glamour](https://github.com/charmbracelet/glamour)
- Terminal styling with [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- Code syntax highlighting via [Chroma](https://github.com/alecthomas/chroma)
