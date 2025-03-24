# Changelog

All notable changes to GoCard will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2025-03-23

### Added

- Improved Markdown rendering using Goldmark parser
- Code syntax highlighting for 50+ programming languages
- Enhanced deck organization and navigation
- Real-time file watching with fsnotify
- Terminal UI improvements with better keyboard shortcuts
- Card creation and editing interface with auto-save
- Preview mode for card editing
- Enhanced SM-2 algorithm implementation
- Deck statistics and review session summaries
- Tutorial mode for first-time users

### Fixed

- Corrected file path handling in deck operations
- Fixed memory leaks in file watching system
- Improved error handling throughout the application
- Resolved concurrent file operation issues

### Development

- Improved Go test coverage
- Enhanced developer documentation
- Added more code examples
- Setup cross-platform build pipeline

## [0.1.0] - 2025-03-22

### Added

- File-based storage for flashcards using Markdown files
- SM-2 spaced repetition algorithm implementation
- Real-time file watching with fsnotify
- Deck organization through directory structure
- Goldmark for Markdown processing
- Basic card navigation and review interface
- Markdown rendering with code syntax highlighting
- Terminal user interface with bubbles/bubbletea
- Support for keyboard shortcuts and vim-style navigation
- Example cards and tutorial for first-time users

### Development

- Added GitHub Actions workflows for testing and building
- Set up cross-platform build support (Linux, macOS, Windows)
- Added code coverage reporting
- Implemented pre-commit hooks configuration
- Created comprehensive project structure
