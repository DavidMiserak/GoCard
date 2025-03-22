# GoCard Examples

This directory contains example flashcards demonstrating various features of GoCard.

## Directory Structure

- **programming/** - Programming language specific cards
  - **go/** - Go programming language
  - **python/** - Python programming language
- **algorithms/** - Algorithm concepts and implementations
- **concepts/** - Computer science concepts
- **math/** - Mathematical concepts
- **language-learning/** - Natural language learning examples
  - **vocabulary/** - Vocabulary examples
- **gocard-features/** - Cards showcasing GoCard features

## Using These Examples

Copy these examples to your GoCard directory to try them out:

```bash
cp -r examples/* ~/GoCard/
```

Or point GoCard directly to this directory:

```bash
gocard ./examples
```

## Creating Your Own Cards

Use these examples as templates for creating your own cards. Each card is a Markdown file with YAML frontmatter for metadata.

Basic structure:

```markdown
---
tags: tag1, tag2
created: YYYY-MM-DD
last_reviewed: YYYY-MM-DD
review_interval: 0
difficulty: 0
---

# Card Title

## Question

Your question goes here?

## Answer

Your answer goes here.
```
