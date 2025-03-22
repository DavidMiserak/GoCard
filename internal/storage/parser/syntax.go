// File: internal/storage/parser/syntax.go
package parser

import (
	"bytes"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	ghtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// SyntaxHighlightingConfig contains options for syntax highlighting
type SyntaxHighlightingConfig struct {
	Theme            string // Chroma theme name
	ShowLineNumbers  bool   // Whether to show line numbers
	HighlightedStyle string // CSS class for highlighted lines
	LineNumbersStyle string // CSS class for line numbers
	WrapLongLines    bool   // Whether to wrap long lines
	TabWidth         int    // Number of spaces for tabs
	DefaultLang      string // Default language if none specified
}

// DefaultSyntaxConfig returns the default syntax highlighting configuration
func DefaultSyntaxConfig() *SyntaxHighlightingConfig {
	return &SyntaxHighlightingConfig{
		Theme:            "monokai",
		ShowLineNumbers:  true,
		HighlightedStyle: "highlighted",
		LineNumbersStyle: "line-numbers",
		WrapLongLines:    true,
		TabWidth:         4,
		DefaultLang:      "text",
	}
}

// NewMarkdownParser creates a new Goldmark parser with syntax highlighting support
func NewMarkdownParser(config *SyntaxHighlightingConfig) goldmark.Markdown {
	if config == nil {
		config = DefaultSyntaxConfig()
	}

	// Create a custom renderer for code blocks
	codeBlockRenderer := NewCodeBlockRenderer(config)

	// Create the markdown parser with extensions
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM, // GitHub Flavored Markdown
			extension.Footnote,
			extension.Typographer,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(), // Allow HTML attributes in markdown
		),
		goldmark.WithRendererOptions(
			ghtml.WithUnsafe(), // Allow HTML in markdown
			renderer.WithNodeRenderers(
				util.Prioritized(codeBlockRenderer, 100),
			),
		),
	)
}

// CodeBlockRenderer is a custom renderer for code blocks using Chroma
type CodeBlockRenderer struct {
	config *SyntaxHighlightingConfig
}

// NewCodeBlockRenderer creates a new CodeBlockRenderer
func NewCodeBlockRenderer(config *SyntaxHighlightingConfig) *CodeBlockRenderer {
	return &CodeBlockRenderer{
		config: config,
	}
}

// RegisterFuncs registers this renderer for code blocks
func (r *CodeBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
	reg.Register(ast.KindCodeBlock, r.renderCodeBlock)
}

// renderFencedCodeBlock renders a fenced code block with syntax highlighting
func (r *CodeBlockRenderer) renderFencedCodeBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	node := n.(*ast.FencedCodeBlock)

	// Get the language from the code block
	lang := string(node.Language(source))
	if lang == "" {
		lang = r.config.DefaultLang
	}

	// Get the code content
	var code bytes.Buffer
	content := node.Lines()
	for i := 0; i < content.Len(); i++ {
		line := content.At(i)
		code.Write(line.Value(source))
	}

	// Highlight the code
	highlighted, err := r.highlightCode(code.String(), lang)
	if err != nil {
		// Fallback to plain text if highlighting fails
		highlighted = escapeHTML(code.String())
		if _, err := w.WriteString("<pre><code>"); err != nil {
			return ast.WalkStop, err
		}
		if _, err := w.WriteString(highlighted); err != nil {
			return ast.WalkStop, err
		}
		if _, err := w.WriteString("</code></pre>"); err != nil {
			return ast.WalkStop, err
		}

	} else {
		// Write the highlighted code
		if _, err := w.WriteString(highlighted); err != nil {
			return ast.WalkStop, err
		}
	}

	return ast.WalkSkipChildren, nil
}

// renderCodeBlock renders a standard code block with syntax highlighting
func (r *CodeBlockRenderer) renderCodeBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	node := n.(*ast.CodeBlock)

	// Get the code content
	var code bytes.Buffer
	content := node.Lines()
	for i := 0; i < content.Len(); i++ {
		line := content.At(i)
		code.Write(line.Value(source))
	}

	// Highlight the code (using default language)
	highlighted, err := r.highlightCode(code.String(), r.config.DefaultLang)
	if err != nil {
		// Fallback to plain text if highlighting fails
		highlighted = escapeHTML(code.String())
		if _, err := w.WriteString("<pre><code>"); err != nil {
			return ast.WalkStop, err
		}
		if _, err := w.WriteString(highlighted); err != nil {
			return ast.WalkStop, err
		}
		if _, err := w.WriteString("</code></pre>"); err != nil {
			return ast.WalkStop, err
		}

	} else {
		// Write the highlighted code
		if _, err := w.WriteString(highlighted); err != nil {
			return ast.WalkStop, err
		}
	}

	return ast.WalkSkipChildren, nil
}

// highlightCode highlights code using Chroma
func (r *CodeBlockRenderer) highlightCode(code, language string) (string, error) {
	// Get the lexer for the language
	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// Process the code to create tokens
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return "", err
	}

	// Get the style and formatter
	style := styles.Get(r.config.Theme)
	if style == nil {
		style = styles.Fallback
	}

	// Create HTML formatter with options
	var options []html.Option
	options = append(options, html.WithClasses(false))

	if r.config.ShowLineNumbers {
		options = append(options, html.WithLineNumbers(true))
	}

	// Create formatter with the collected options
	formatter := html.New(options...)

	// Format the code
	var buf bytes.Buffer
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// escapeHTML escapes HTML characters in a string
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// RenderMarkdownWithHighlighting renders markdown content to HTML with syntax highlighting
func RenderMarkdownWithHighlighting(content string, config *SyntaxHighlightingConfig) (string, error) {
	md := NewMarkdownParser(config)
	var buf bytes.Buffer
	if err := md.Convert([]byte(content), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
