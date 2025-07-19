package plugin

import (
	"context"
	"github.com/ofri/mde/pkg/ast"
	"github.com/charmbracelet/lipgloss"
)

// RendererPlugin defines the interface for document renderers
type RendererPlugin interface {
	// Name returns the plugin name
	Name() string
	
	// Render renders the document
	Render(ctx context.Context, doc *ast.Document) ([]RenderedLine, error)
	
	// RenderPreview renders a preview of the document
	RenderPreview(ctx context.Context, doc *ast.Document) ([]RenderedLine, error)
	
	// RenderLine renders a single line with syntax highlighting
	RenderLine(ctx context.Context, line string, tokens []ast.Token) (RenderedLine, error)
	
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
	Style Style
}

// Style represents basic styling information using ANSI colors
type Style struct {
	// Foreground color (ANSI color code 0-15)
	Foreground string
	
	// Background color (ANSI color code 0-15)
	Background string
	
	// Bold text
	Bold bool
	
	// Italic text
	Italic bool
	
	// Underline text
	Underline bool
	
	// Strikethrough text
	Strikethrough bool
}

// ToLipgloss converts a Style to a lipgloss.Style
func (s Style) ToLipgloss() lipgloss.Style {
	style := lipgloss.NewStyle()
	
	if s.Foreground != "" {
		style = style.Foreground(lipgloss.Color(s.Foreground))
	}
	
	if s.Background != "" {
		style = style.Background(lipgloss.Color(s.Background))
	}
	
	if s.Bold {
		style = style.Bold(true)
	}
	
	if s.Italic {
		style = style.Italic(true)
	}
	
	if s.Underline {
		style = style.Underline(true)
	}
	
	if s.Strikethrough {
		style = style.Strikethrough(true)
	}
	
	return style
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