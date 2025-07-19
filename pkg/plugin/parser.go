package plugin

import (
	"context"
	"github.com/ofri/mde/pkg/ast"
)

// ParserPlugin defines the interface for markdown parsers
type ParserPlugin interface {
	// Name returns the plugin name
	Name() string
	
	// Parse parses markdown text into an AST
	Parse(ctx context.Context, text string) (*ast.Document, error)
	
	// GetSyntaxHighlighting returns syntax highlighting tokens for a line
	GetSyntaxHighlighting(ctx context.Context, line string) ([]ast.Token, error)
	
	// Configure configures the parser with options
	Configure(options map[string]interface{}) error
}

// ParserConfig holds configuration for parsers
type ParserConfig struct {
	// Extensions to enable (tables, strikethrough, etc.)
	Extensions []string
	
	// Syntax highlighting options
	SyntaxHighlighting bool
	
	// Custom parser options
	Options map[string]interface{}
}