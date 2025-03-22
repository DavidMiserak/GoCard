// File: internal/storage/parser/markdown.go (updated)
package parser

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/DavidMiserak/GoCard/internal/card"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"gopkg.in/yaml.v3"
)

// ParseMarkdown parses a markdown file into a Card structure
// Uses Goldmark for proper markdown processing
func ParseMarkdown(content []byte) (*card.Card, error) {
	// Check if the file starts with YAML frontmatter
	if !bytes.HasPrefix(content, []byte("---\n")) {
		return nil, fmt.Errorf("markdown file must start with YAML frontmatter")
	}

	// Split the content into frontmatter and markdown
	parts := bytes.SplitN(content, []byte("---\n"), 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid markdown format")
	}

	frontmatter := parts[1]
	markdownContent := parts[2]

	// Create a temporary struct that matches the YAML structure exactly
	type frontMatterData struct {
		Tags           []string  `yaml:"tags,omitempty"`
		Created        time.Time `yaml:"created,omitempty"`
		LastReviewed   time.Time `yaml:"last_reviewed,omitempty"`
		ReviewInterval int       `yaml:"review_interval"`
		Difficulty     int       `yaml:"difficulty,omitempty"`
	}

	// Parse YAML frontmatter into temporary struct
	var fmData frontMatterData
	if err := yaml.Unmarshal(frontmatter, &fmData); err != nil {
		return nil, fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	// Create and populate the Card struct
	cardObj := &card.Card{
		Tags:           fmData.Tags,
		Created:        fmData.Created,
		LastReviewed:   fmData.LastReviewed,
		ReviewInterval: fmData.ReviewInterval,
		Difficulty:     fmData.Difficulty,
	}

	// Extract title, question, and answer from markdown using regex
	// This preserves the raw markdown content for proper rendering later
	mdStr := string(markdownContent)

	// Title regex: Finds a level 1 heading (# Title)
	titleRegex := regexp.MustCompile(`(?m)^# (.+)$`)
	if match := titleRegex.FindStringSubmatch(mdStr); len(match) > 1 {
		cardObj.Title = strings.TrimSpace(match[1])
	}

	// Question section regex: Finds content between "## Question" and the next heading or end of content
	questionRegex := regexp.MustCompile(`(?ms)^## Question\s*\n(.*?)(?:^## |\z)`)
	if match := questionRegex.FindStringSubmatch(mdStr); len(match) > 1 {
		cardObj.Question = strings.TrimSpace(match[1])
	}

	// Answer section regex: Finds content between "## Answer" and the next heading or end of content
	answerRegex := regexp.MustCompile(`(?ms)^## Answer\s*\n(.*?)(?:^## |\z)`)
	if match := answerRegex.FindStringSubmatch(mdStr); len(match) > 1 {
		cardObj.Answer = strings.TrimSpace(match[1])
	}

	// Validate the extracted markdown content by parsing it with Goldmark
	// This ensures that the markdown is well-formed before we store it
	md := createGoldmarkParser()

	// Parse the question and answer to validate them
	// We don't need the output, we just want to ensure they parse correctly
	_ = md.Parser().Parse(text.NewReader([]byte(cardObj.Question)))
	_ = md.Parser().Parse(text.NewReader([]byte(cardObj.Answer)))

	return cardObj, nil
}

// RenderMarkdown renders markdown content to HTML
func RenderMarkdown(content string) (string, error) {
	// First, attempt to render with syntax highlighting
	// This is the preferred method that handles code blocks with proper highlighting
	config := DefaultSyntaxConfig()
	html, err := RenderMarkdownWithHighlighting(content, config)

	// If there's an error with the syntax highlighting, fall back to the standard renderer
	if err != nil {
		md := createGoldmarkParser()
		var buf bytes.Buffer
		if err := md.Convert([]byte(content), &buf); err != nil {
			return "", fmt.Errorf("failed to render markdown: %w", err)
		}
		return buf.String(), nil
	}

	return html, nil
}

// createGoldmarkParser creates a configured Goldmark parser
func createGoldmarkParser() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,         // GitHub Flavored Markdown
			extension.Footnote,    // Support footnotes
			extension.Typographer, // Smart typography
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(), // Allow HTML in markdown
		),
	)
}
