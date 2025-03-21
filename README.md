# GoCard: A File-Based Spaced Repetition System

[![Go Report Card](https://goreportcard.com/badge/github.com/DavidMiserak/GoCard)](https://goreportcard.com/report/github.com/DavidMiserak/GoCard)
[![Build Status](https://github.com/DavidMiserak/GoCard/workflows/Go/badge.svg)](https://github.com/DavidMiserak/GoCard/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

![GoCard Logo](assets/gocard-logo.webp)

GoCard is a lightweight, file-based spaced repetition system (SRS)
built in Go. It uses plain Markdown files organized in directories as
its data source, making it perfect for developers who prefer working
with text files and version control.

## Features

- **File-Based Storage**: All flashcards are stored as Markdown files in regular directories
- **Git-Friendly**: Easily track changes, collaborate, and back up your knowledge base
- **Terminal Interface**: Clean, distraction-free TUI for focused learning
- **Markdown Support**: Full Markdown rendering with code syntax highlighting
- **Cross-Platform**: Works on Linux, macOS, and Windows
- **Spaced Repetition Algorithm**: Implements the SM-2 algorithm for efficient learning
- **Code-Focused**: Special features for programming-related cards:
  - Syntax highlighting for 50+ languages
  - Side-by-side diff view for comparing code
- **Session Statistics**: Track your learning progress with detailed review stats

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

GoCard follows a standard Go project layout:

```sh
github.com/DavidMiserak/GoCard/
├── cmd/gocard/          # Main application entry point
├── internal/            # Private implementation packages
│   ├── algorithm/       # Spaced repetition algorithms (SM-2)
│   ├── card/            # Core card data models
│   ├── storage/         # File-based card storage
│   └── ui/              # Terminal user interface
├── assets/              # Application resources
└── docs/                # Documentation
```

This package organization provides:
- Clean separation of concerns
- Better testability of individual components
- Easier maintenance and extensibility
- Adherence to Go best practices

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
tags: algorithms, techniques, arrays
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

## File Format

Cards are stored as markdown files with a YAML frontmatter section for metadata:

```markdown
---
tags: tag1, tag2, tag3
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

## Directory Structure

Organize your cards however you want! The directory structure becomes the deck structure:

```sh
~/gocard/
├── algorithms/
│   ├── sorting/
│   │   ├── quicksort.md
│   │   └── mergesort.md
│   └── searching/
│       ├── binary-search.md
│       └── depth-first-search.md
├── go-programming/
│   ├── concurrency/
│   │   ├── goroutines.md
│   │   └── channels.md
│   └── interfaces.md
└── vocabulary/
    ├── spanish.md
    └── german.md
```

## Spaced Repetition System

GoCard implements the SM-2 algorithm for spaced repetition, similar to
Anki. After reviewing a card, you rate how well you remembered it on a
scale of 0-5:

- **0-2**: Difficult, short interval
- **3**: Correct, but required effort
- **4-5**: Easy, longer interval

The review intervals are calculated based on your performance and
stored in the markdown file's frontmatter.

## Terminal Interface

GoCard provides a clean, minimalist terminal interface optimized for focused learning:

- **Distraction-Free**: Simple design that lets you focus on learning
- **Markdown Rendering**: Beautiful rendering of card content with syntax highlighting
- **Keyboard-Driven**: Efficient workflow with intuitive keyboard shortcuts
- **Progress Tracking**: Monitor your review session progress
- **Session Statistics**: Summary view after completing a review session

## Keyboard Shortcuts

| Key     | Action              |
|---------|---------------------|
| `Space` | Show answer         |
| `0-5`   | Rate card difficulty|
| `?`     | Toggle help         |
| `q`     | Quit                |

Additional shortcuts planned for future versions:
- `e` - Edit current card
- `n` - Create new card
- `d` - Delete current card
- `t` - Add/edit tags
- `s` - Search cards

## Review Process

The review process follows a simple flow:

1. Cards due for review will be loaded automatically
2. For each card:
   - The question is shown first
   - Press `Space` to reveal the answer
   - Rate your recall from 0-5:
     - `0-2`: Difficult/incorrect (short interval)
     - `3`: Correct but required effort (moderate interval increase)
     - `4-5`: Easy (longer interval increase)
3. After reviewing all due cards, a summary is displayed showing:
   - Number of cards reviewed
   - Current card statistics (new, young, mature)
   - Next scheduled review date

## Configuration

Configuration is stored in `~/.gocard.yaml`:

```yaml
default_cards_dir: ~/gocard
theme: "auto"  # auto, light, dark
highlight_theme: "monokai"  # code highlighting theme
spaced_repetition:
  easy_bonus: 1.3
  interval_modifier: 1.0
  new_cards_per_day: 20
```

## Development

### Running Tests

```bash
go test ./...
```

### Linting and Formatting

```bash
# Format code
go fmt ./...

# Run linters
golangci-lint run
```

### Setting Up Pre-commit Hooks

```bash
pre-commit install
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgements

- Inspired by Anki and SuperMemo
- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) terminal UI framework
- Markdown rendering via [Goldmark](https://github.com/yuin/goldmark) and [Glamour](https://github.com/charmbracelet/glamour)
- Terminal styling with [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- Code syntax highlighting via [Chroma](https://github.com/alecthomas/chroma)
