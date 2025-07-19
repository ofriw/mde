package parsers

import (
	"context"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldmarkText "github.com/yuin/goldmark/text"

	mdeAST "github.com/ofri/mde/pkg/ast"
	"github.com/ofri/mde/pkg/plugin"
)

// CommonMarkParser implements the ParserPlugin interface using goldmark
type CommonMarkParser struct {
	name     string
	goldmark goldmark.Markdown
	config   *plugin.ParserConfig
}

// NewCommonMarkParser creates a new CommonMark parser
func NewCommonMarkParser() *CommonMarkParser {
	// Create goldmark instance with extensions
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.Linkify,
			extension.TaskList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	return &CommonMarkParser{
		name:     "commonmark",
		goldmark: md,
		config: &plugin.ParserConfig{
			Extensions:         []string{"gfm", "table", "strikethrough", "linkify", "tasklist"},
			SyntaxHighlighting: true,
			Options:            make(map[string]interface{}),
		},
	}
}

// Name returns the plugin name
func (p *CommonMarkParser) Name() string {
	return p.name
}

// Parse parses markdown text into an AST
func (p *CommonMarkParser) Parse(ctx context.Context, text string) (*mdeAST.Document, error) {
	// Parse with goldmark for validation (full AST conversion comes later)
	source := []byte(text)
	reader := goldmarkText.NewReader(source)
	
	_ = p.goldmark.Parser().Parse(reader)
	
	// Convert goldmark AST to our document model
	doc := mdeAST.NewDocument(text)
	
	// Apply syntax highlighting to each line
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		tokens, err := p.GetSyntaxHighlighting(ctx, line)
		if err != nil {
			// Log error but continue with empty tokens
			tokens = []mdeAST.Token{}
		}
		
		// Set tokens for the line
		if i < doc.LineCount() {
			doc.SetLineTokens(i, tokens)
		}
	}
	
	return doc, nil
}


// GetSyntaxHighlighting returns syntax highlighting tokens for a line
func (p *CommonMarkParser) GetSyntaxHighlighting(ctx context.Context, line string) ([]mdeAST.Token, error) {
	var tokens []mdeAST.Token
	
	// Parse the line to identify markdown elements
	tokens = append(tokens, p.parseHeadings(line)...)
	tokens = append(tokens, p.parseBoldItalic(line)...)
	tokens = append(tokens, p.parseCode(line)...)
	tokens = append(tokens, p.parseLinks(line)...)
	tokens = append(tokens, p.parseQuotes(line)...)
	tokens = append(tokens, p.parseLists(line)...)
	
	// Sort tokens by start position
	for i := 0; i < len(tokens)-1; i++ {
		for j := i + 1; j < len(tokens); j++ {
			if tokens[i].Start() > tokens[j].Start() {
				tokens[i], tokens[j] = tokens[j], tokens[i]
			}
		}
	}
	
	return tokens, nil
}

// Configure configures the parser with options
func (p *CommonMarkParser) Configure(options map[string]interface{}) error {
	if options == nil {
		return nil
	}
	
	// Update config
	if extensions, ok := options["extensions"].([]string); ok {
		p.config.Extensions = extensions
	}
	
	if syntaxHighlighting, ok := options["syntax_highlighting"].(bool); ok {
		p.config.SyntaxHighlighting = syntaxHighlighting
	}
	
	for key, value := range options {
		p.config.Options[key] = value
	}
	
	return nil
}


// Syntax highlighting helper methods

func (p *CommonMarkParser) parseHeadings(line string) []mdeAST.Token {
	var tokens []mdeAST.Token
	re := regexp.MustCompile(`^(#{1,6})\s*(.*)`)
	matches := re.FindStringSubmatch(line)
	
	if len(matches) > 0 {
		hashLen := len(matches[1])
		tokens = append(tokens, mdeAST.NewToken(0, hashLen, mdeAST.TokenDelimiter))
		
		if len(matches[2]) > 0 {
			start := hashLen
			if strings.HasPrefix(line[hashLen:], " ") {
				start++ // Skip space after #
			}
			tokens = append(tokens, mdeAST.NewToken(start, len(line), mdeAST.TokenHeading))
		}
	}
	
	return tokens
}

func (p *CommonMarkParser) parseBoldItalic(line string) []mdeAST.Token {
	var tokens []mdeAST.Token
	
	// Bold (**text** or __text__)
	boldRe := regexp.MustCompile(`\*\*(.*?)\*\*|__(.*?)__`)
	for _, match := range boldRe.FindAllStringSubmatchIndex(line, -1) {
		if len(match) >= 4 {
			// Mark the entire bold text
			tokens = append(tokens, mdeAST.NewToken(match[0], match[1], mdeAST.TokenBold))
		}
	}
	
	// Italic (*text* or _text_) - avoid conflicts with bold
	italicRe := regexp.MustCompile(`(?:\*([^*]+?)\*)|(?:_([^_]+?)_)`)
	for _, match := range italicRe.FindAllStringSubmatchIndex(line, -1) {
		if len(match) >= 4 {
			// Skip if this is part of bold (already handled)
			if p.isInsideBold(line, match[0]) {
				continue
			}
			
			// Mark the entire italic text
			tokens = append(tokens, mdeAST.NewToken(match[0], match[1], mdeAST.TokenItalic))
		}
	}
	
	return tokens
}

func (p *CommonMarkParser) parseCode(line string) []mdeAST.Token {
	var tokens []mdeAST.Token
	
	// Inline code (`code`)
	codeRe := regexp.MustCompile("`([^`]+)`")
	for _, match := range codeRe.FindAllStringSubmatchIndex(line, -1) {
		if len(match) >= 4 {
			tokens = append(tokens, mdeAST.NewToken(match[0], match[1], mdeAST.TokenCode))
		}
	}
	
	// Code block start (```)
	if strings.HasPrefix(strings.TrimSpace(line), "```") {
		tokens = append(tokens, mdeAST.NewToken(0, len(line), mdeAST.TokenCodeBlock))
	}
	
	return tokens
}

func (p *CommonMarkParser) parseLinks(line string) []mdeAST.Token {
	var tokens []mdeAST.Token
	
	// Links [text](url)
	linkRe := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	for _, match := range linkRe.FindAllStringSubmatchIndex(line, -1) {
		if len(match) >= 6 {
			// Link text
			tokens = append(tokens, mdeAST.NewToken(match[2], match[3], mdeAST.TokenLinkText))
			// Link URL
			tokens = append(tokens, mdeAST.NewToken(match[4], match[5], mdeAST.TokenLinkURL))
			// Delimiters
			tokens = append(tokens, mdeAST.NewToken(match[0], match[2], mdeAST.TokenDelimiter))
			tokens = append(tokens, mdeAST.NewToken(match[3], match[4], mdeAST.TokenDelimiter))
			tokens = append(tokens, mdeAST.NewToken(match[5], match[1], mdeAST.TokenDelimiter))
		}
	}
	
	// Images ![alt](url)
	imageRe := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	for _, match := range imageRe.FindAllStringSubmatchIndex(line, -1) {
		if len(match) >= 6 {
			tokens = append(tokens, mdeAST.NewToken(match[0], match[1], mdeAST.TokenImage))
		}
	}
	
	return tokens
}

func (p *CommonMarkParser) parseQuotes(line string) []mdeAST.Token {
	var tokens []mdeAST.Token
	
	if strings.HasPrefix(strings.TrimSpace(line), ">") {
		// Find the first >
		index := strings.Index(line, ">")
		if index >= 0 {
			tokens = append(tokens, mdeAST.NewToken(index, index+1, mdeAST.TokenDelimiter))
			if index+1 < len(line) {
				tokens = append(tokens, mdeAST.NewToken(index+1, len(line), mdeAST.TokenQuote))
			}
		}
	}
	
	return tokens
}

func (p *CommonMarkParser) parseLists(line string) []mdeAST.Token {
	var tokens []mdeAST.Token
	
	trimmed := strings.TrimSpace(line)
	
	// Unordered lists (-, *, +)
	if len(trimmed) > 0 && (trimmed[0] == '-' || trimmed[0] == '*' || trimmed[0] == '+') {
		index := strings.Index(line, string(trimmed[0]))
		if index >= 0 {
			tokens = append(tokens, mdeAST.NewToken(index, index+1, mdeAST.TokenDelimiter))
			if index+1 < len(line) {
				tokens = append(tokens, mdeAST.NewToken(index+1, len(line), mdeAST.TokenList))
			}
		}
	}
	
	// Ordered lists (1., 2., etc.)
	orderedRe := regexp.MustCompile(`^\s*(\d+\.)\s*(.*)`)
	matches := orderedRe.FindStringSubmatch(line)
	if len(matches) > 0 {
		numLen := len(matches[1])
		start := strings.Index(line, matches[1])
		if start >= 0 {
			tokens = append(tokens, mdeAST.NewToken(start, start+numLen, mdeAST.TokenDelimiter))
			if start+numLen < len(line) {
				tokens = append(tokens, mdeAST.NewToken(start+numLen, len(line), mdeAST.TokenList))
			}
		}
	}
	
	return tokens
}

func (p *CommonMarkParser) isInsideBold(line string, pos int) bool {
	// Check if position is inside bold markup
	boldRe := regexp.MustCompile(`\*\*(.*?)\*\*|__(.*?)__`)
	for _, match := range boldRe.FindAllStringSubmatchIndex(line, -1) {
		if len(match) >= 2 && pos >= match[0] && pos < match[1] {
			return true
		}
	}
	return false
}