package plugin

import (
	"context"
	"github.com/ofri/mde/pkg/ast"
	"github.com/ofri/mde/pkg/theme"
)

// RendererPlugin defines the interface for document renderers
type RendererPlugin interface {
	// Name returns the plugin name
	Name() string
	
	// Render renders the document with the given theme
	Render(ctx context.Context, doc *ast.Document, theme theme.Theme) ([]RenderedLine, error)
	
	// RenderPreview renders a preview of the document
	RenderPreview(ctx context.Context, doc *ast.Document, theme theme.Theme) ([]RenderedLine, error)
	
	// RenderLine renders a single line with syntax highlighting
	RenderLine(ctx context.Context, line string, tokens []ast.Token, theme theme.Theme) (RenderedLine, error)
	
	// Configure configures the renderer with options
	Configure(options map[string]interface{}) error
}

// RenderedLine represents a rendered line with styling information
type RenderedLine struct {
	// Content is the rendered text content
	Content string
	
	// Styles contains styling information for different ranges
	Styles []StyleRange
	
	// Metadata contains additional information about the line
	Metadata map[string]interface{}
}

// StyleRange represents a range of characters with specific styling
type StyleRange struct {
	// Start position in the line
	Start int
	
	// End position in the line
	End int
	
	// Style to apply
	Style theme.Style
}

// RendererConfig holds configuration for renderers
type RendererConfig struct {
	// Maximum line width for wrapping
	MaxWidth int
	
	// Tab width for rendering
	TabWidth int
	
	// Show line numbers
	ShowLineNumbers bool
	
	// Width of line number prefix (calculated dynamically)
	LineNumberWidth int
	
	// Preview mode settings
	PreviewMode bool
	
	// Custom renderer options
	Options map[string]interface{}
}