# GoCard v0.2.0 Release Notes

We're excited to announce GoCard v0.2.0, building on our initial release with significant improvements to the core functionality and user experience.

## What's New

### Enhanced Markdown Support

- **Goldmark Integration**: Completely reworked markdown processing using Goldmark for more reliable and feature-rich rendering
- **Syntax Highlighting**: Support for over 50 programming languages with customizable themes
- **Rich Content**: Better rendering of tables, lists, and code blocks

### Improved User Interface

- **Card Editor**: Full-featured card creation and editing interface with markdown preview
- **Auto-Save**: Automatic saving of changes to prevent data loss
- **Tutorial Mode**: Interactive tutorial for first-time users
- **Enhanced Navigation**: Vim and Emacs-style keyboard shortcuts for efficient navigation
- **Deck Browser**: Improved deck browsing with statistics and breadcrumb navigation

### File System Integration

- **Real-time Watching**: Robust file system monitoring with fsnotify for seamless external editing
- **Directory Organization**: Enhanced deck management through directory structure
- **File Operations**: Better handling of file moves, renames, and deletions

### Learning Features

- **Enhanced SM-2 Algorithm**: Improved spaced repetition scheduling
- **Review Sessions**: Comprehensive review session interface with statistics
- **Learning Insights**: Better tracking of your learning progress

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

## Upgrading from v0.1.0

This release is fully compatible with card files created in v0.1.0. No migration steps are needed - simply install the new version and run it with your existing card directory.

## Feedback and Contributions

We welcome your feedback and contributions! Please file issues for bugs or feature requests on our GitHub repository.

## Coming Soon

We're actively working on more features:

- Search and filter functionality
- Import/export compatibility with other SRS systems
- Customizable styling and themes
- Code testing integration
- Additional configuration options

Thank you for using GoCard!
