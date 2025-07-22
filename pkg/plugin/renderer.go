package plugin

import (
	"context"
	"github.com/ofri/mde/pkg/ast"
	"github.com/charmbracelet/lipgloss"
)

// RenderContext provides all necessary context for viewport-aware rendering.
// This structure encapsulates the document, viewport, and display settings,
// allowing renderers to efficiently render only the visible portion of a document.
//
// DESIGN RATIONALE:
// Previously, renderers would process entire documents and the TUI would discard
// non-visible content. This was inefficient for large files and broke scrolling
// because the viewport position was ignored during rendering.
//
// With RenderContext, renderers receive explicit viewport boundaries and can
// optimize by only processing visible lines. This fixes scrolling issues and
// improves performance for large documents.
type RenderContext struct {
	// Document is the source document to render
	Document *ast.Document
	
	// Viewport defines the visible region of the document
	// This includes TopLine, LeftColumn, Width, and Height
	Viewport *ast.Viewport
	
	// ShowLineNumbers indicates whether to render line number prefixes
	// When true, renderers should add line numbers and account for their width
	// in horizontal scrolling calculations
	ShowLineNumbers bool
}

// RendererPlugin defines the interface for document renderers
type RendererPlugin interface {
	// Name returns the plugin name
	Name() string
	
	// RenderVisible renders only the visible portion of the document defined by the viewport.
	// This is the primary rendering method that should be used for editor content.
	//
	// IMPLEMENTATION NOTES:
	// - Renderers MUST respect the viewport boundaries (TopLine to TopLine+Height)
	// - Line numbers should be added if ShowLineNumbers is true in the context
	// - Horizontal scrolling (LeftColumn) should be applied after line numbers
	// - The returned slice should contain exactly the visible lines, no more, no less
	RenderVisible(ctx context.Context, renderCtx *RenderContext) ([]RenderedLine, error)
	
	// RenderPreviewVisible renders the visible portion of the document in preview mode.
	// This method applies markdown formatting and other preview-specific transformations
	// while still respecting the viewport boundaries.
	//
	// IMPLEMENTATION NOTES:
	// - Preview mode typically doesn't show line numbers
	// - Markdown formatting should be applied (headers, bold, italic, etc.)
	// - Viewport boundaries must still be respected for performance
	RenderPreviewVisible(ctx context.Context, renderCtx *RenderContext) ([]RenderedLine, error)
	
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