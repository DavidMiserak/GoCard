# GoCard v0.1.0 Release Notes

We're excited to announce the first release of GoCard, a lightweight, file-based spaced repetition system built in Go. This initial release provides the core functionality while establishing a solid foundation for future development.

## Features

### Core Functionality

- **File-Based Storage**: All flashcards are stored as Markdown files in regular directories, making them Git-friendly and easy to edit with your favorite text editor
- **Spaced Repetition**: Implementation of the SM-2 algorithm (similar to Anki) for efficient learning
- **Directory-Based Deck Organization**: Directories represent decks, giving you a natural way to organize your knowledge
- **Real-Time File Watching**: Changes to card files are automatically detected and loaded

### User Interface

- **Terminal Interface**: Clean, distraction-free TUI for focused learning
- **Markdown Rendering**: Beautiful rendering of card content with syntax highlighting
- **Keyboard-Driven**: Efficient workflow with intuitive keyboard shortcuts including vim/emacs-style navigation
- **Getting Started Experience**: Tutorial and example cards for first-time users

### Developer Features

- **Markdown Support**: Full Markdown rendering with code syntax highlighting
- **Cross-Platform**: Works on Linux, macOS, and Windows

## Installation

### Download Binary

Download the binary for your platform from the releases page.

### Building from Source

```bash
git clone https://github.com/DavidMiserak/GoCard.git
cd GoCard
go build -o GoCard ./cmd/gocard
```

### Using Go Install

```bash
go install github.com/DavidMiserak/GoCard/cmd/gocard@latest
```

## Getting Started

1. Launch GoCard:
```bash
gocard
```

2. By default, GoCard will create a directory at `~/GoCard` to store your cards.

3. Follow the tutorial that appears on first run to learn the basics.

4. Start creating and reviewing your own cards!

## Feedback and Contributions

This is an early release, and we welcome your feedback and contributions. Please file issues for bugs or feature requests on our GitHub repository.

## Coming Soon

We're actively working on more features:
- Search and filter functionality
- Import/export compatibility with other SRS systems
- Customizable styling and themes
- Code testing integration
- Additional configuration options

Thank you for trying GoCard!
