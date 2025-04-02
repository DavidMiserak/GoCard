# GoCard: Product Requirements Document

## 1. Executive Summary

GoCard is a lightweight, file-based spaced repetition system (SRS)
built in Go. It addresses the needs of developers and technical users
who prefer working with plain text files and version control systems
for managing their knowledge. GoCard provides a distraction-free
terminal interface that combines powerful learning algorithms with the
simplicity of Markdown files.

## 2. Problem Statement

Existing spaced repetition and flashcard applications typically store
data in proprietary formats or databases, making it difficult to:

- Version control flashcard content
- Collaborate with others on knowledge bases
- Back up content using standard tools
- Edit content with preferred text editors
- Maintain ownership of learning data

Additionally, many applications include distracting elements that
detract from the learning experience.

## 3. Goals and Objectives

- Create a spaced repetition system that uses plain Markdown files as its primary data source
- Implement an enhanced version of the SuperMemo-2 (SM-2) algorithm for optimal learning efficiency
- Provide a clean, keyboard-driven terminal interface for distraction-free studying
- Deliver comprehensive statistics to help users track their learning progress
- Ensure cross-platform compatibility across Linux, MacOS, and Windows
- Support rich Markdown rendering with syntax highlighting for code examples

## 4. Target Audience

GoCard targets:

- Software developers and technical professionals
- Users who prefer terminal-based applications
- People who want to maintain version control of their learning materials
- Learners who value data portability and ownership
- Users who appreciate minimalist, distraction-free interfaces

## 5. User Stories

1. As a developer, I want to store my flashcards as Markdown files so I can version control them with Git.
2. As a learner, I want the system to schedule my reviews optimally so I can maximize retention with minimal time investment.
3. As a user, I want a distraction-free interface so I can focus solely on learning.
4. As a programmer, I want proper syntax highlighting for code snippets so I can study programming concepts effectively.
5. As a power user, I want keyboard shortcuts for all actions so I can navigate efficiently.
6. As a student, I want comprehensive statistics so I can track my learning progress.
7. As a knowledge worker, I want to organize cards into decks so I can separate different subject areas.

## 6. Feature Requirements

### 6.1 Core Functionality

#### File-Based Storage

- **Must have:** Store all flashcards as Markdown files in standard directories
- **Must have:** Support standard Markdown formatting in questions and answers
- **Must have:** Include YAML front-matter for card metadata (tags, review dates, intervals)

#### Spaced Repetition Algorithm

- **Must have:** Implement an enhanced SM-2 algorithm
- **Must have:** Support a five-point rating scale (1-Blackout to 5-Easy)
- **Must have:** Dynamically adjust intervals based on performance
- **Must have:** Track review history and success rates

#### Card Organization

- **Must have:** Support organization via directories (as decks)
- **Must have:** Allow tagging of cards via front-matter
- **Should have:** Support for filtering cards by tags

### 6.2 User Interface

#### Terminal Interface

- **Must have:** Clean, distraction-free TUI for focused learning
- **Must have:** Full keyboard navigation
- **Must have:** Rich Markdown rendering with syntax highlighting
- **Must have:** Progress indicators for study sessions
- **Must have:** Visual distinction between question and answer views

#### Study Screen

- **Must have:** Clear separation between question and answer
- **Must have:** Rating buttons (1-5) for card difficulty
- **Must have:** Card count and progress indicators
- **Should have:** Scrolling support for long answers

#### Statistics Views

- **Must have:** Summary statistics (total cards, retention rate)
- **Must have:** Deck-specific metrics
- **Must have:** Review forecast visualization
- **Must have:** Study history visualization

## 7. Technical Requirements

### 7.1 Performance

- Application startup time under 1 second
- Smooth performance with collections of 10,000+ cards
- Responsive UI with no perceptible lag during interactions

### 7.2 Dependencies

- Go programming language
- Bubble Tea framework for terminal UI
- Goldmark and Glamour for Markdown rendering
- Lip Gloss for terminal styling
- Chroma for code syntax highlighting

### 7.3 Cross-Platform Compatibility

- Must work consistently on Linux, MacOS, and Windows
- Must handle path differences between operating systems
- Must work with various terminal emulators

## 8. User Interface Design

### 8.1 Navigation Structure

1. Main Menu

   - Study
   - Browse Decks
   - Statistics
   - Quit

2. Browse Decks Screen

   - List of decks with metadata (card count, due cards, last studied)
   - Pagination for large collections
   - Options to study selected deck

3. Study Screen

   - Question display
   - Answer reveal on key-press
   - Rating interface (1-5)
   - Progress indicator

4. Statistics Screen

   - Tab navigation between views:
     - Summary (overall statistics)
     - Deck Review (deck-specific metrics)
     - Review Forecast (upcoming reviews)

### 8.2 Keyboard Shortcuts

- Space: Show answer
- 1-5: Rate card difficulty
- ↑/k: Move up/scroll up
- ↓/j: Move down/scroll down
- Enter: Select/confirm
- Tab: Switch tab (in statistics)
- b: Back to previous screen
- q: Quit

## 9. Implementation Constraints

- Terminal-based interface only (no GUI)
- File system for data storage
- Must work with standard terminal dimensions
- Must handle Unicode characters correctly
- Must preserve card formatting during reading/writing

## 10. Success Metrics

- User retention and daily usage patterns
- Average retention rate improvement over time
- Speed of card review (cards per minute)
- Growth in user flashcard collections
- User feedback on terminal interface usability

## 11. Future Considerations

- Synchronization between devices
- Web interface for mobile access
- Improved data visualization for statistics
- Integration with external learning resources
- Enhanced collaboration features
- Audio support for language learning
- Image support in cards
