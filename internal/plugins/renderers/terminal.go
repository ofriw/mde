package renderers

import (
	"context"
	"fmt"
	"strings"
	"github.com/ofri/mde/pkg/ast"
	"github.com/ofri/mde/pkg/plugin"
	"github.com/ofri/mde/pkg/theme"
)

// TerminalRenderer implements the RendererPlugin interface for terminal output
type TerminalRenderer struct {
	config plugin.RendererConfig
}

// NewTerminalRenderer creates a new terminal renderer
func NewTerminalRenderer() *TerminalRenderer {
	return &TerminalRenderer{
		config: plugin.RendererConfig{
			MaxWidth:        80,
			TabWidth:        4,
			ShowLineNumbers: false,
			PreviewMode:     false,
			Options:         make(map[string]interface{}),
		},
	}
}

// Name returns the plugin name
func (r *TerminalRenderer) Name() string {
	return "terminal"
}

// Configure configures the renderer with options
func (r *TerminalRenderer) Configure(options map[string]interface{}) error {
	if maxWidth, ok := options["maxWidth"].(int); ok {
		r.config.MaxWidth = maxWidth
	}
	
	if tabWidth, ok := options["tabWidth"].(int); ok {
		r.config.TabWidth = tabWidth
	}
	
	if showLineNumbers, ok := options["showLineNumbers"].(bool); ok {
		r.config.ShowLineNumbers = showLineNumbers
	}
	
	if previewMode, ok := options["previewMode"].(bool); ok {
		r.config.PreviewMode = previewMode
	}
	
	// Store custom options
	for key, value := range options {
		r.config.Options[key] = value
	}
	
	return nil
}

// Render renders the document with the given theme
func (r *TerminalRenderer) Render(ctx context.Context, doc *ast.Document, themePlugin theme.Theme) ([]plugin.RenderedLine, error) {
	lines := make([]plugin.RenderedLine, 0, doc.LineCount())
	
	for i := 0; i < doc.LineCount(); i++ {
		line := doc.GetLine(i)
		
		// For now, render as plain text with basic styling
		renderedLine, err := r.renderTextLine(line, themePlugin)
		if err != nil {
			return nil, fmt.Errorf("failed to render line %d: %w", i, err)
		}
		
		lines = append(lines, renderedLine)
	}
	
	return lines, nil
}

// RenderPreview renders a preview of the document
func (r *TerminalRenderer) RenderPreview(ctx context.Context, doc *ast.Document, themePlugin theme.Theme) ([]plugin.RenderedLine, error) {
	// For now, preview mode is the same as regular rendering
	// In the future, this could render markdown with formatted output
	return r.Render(ctx, doc, themePlugin)
}

// RenderLine renders a single line with syntax highlighting
func (r *TerminalRenderer) RenderLine(ctx context.Context, line string, tokens []ast.Token, themePlugin theme.Theme) (plugin.RenderedLine, error) {
	if len(tokens) == 0 {
		// No syntax highlighting, render as plain text
		return r.renderTextLine(line, themePlugin)
	}
	
	// Apply syntax highlighting
	content := line
	styles := make([]plugin.StyleRange, 0, len(tokens))
	
	for _, token := range tokens {
		var elementType theme.ElementType
		switch token.Kind() {
		case ast.TokenKeyword:
			elementType = theme.SyntaxKeyword
		case ast.TokenString:
			elementType = theme.SyntaxString
		case ast.TokenComment:
			elementType = theme.SyntaxComment
		case ast.TokenNumber:
			elementType = theme.SyntaxNumber
		default:
			elementType = theme.TextNormal
		}
		
		style := themePlugin.GetStyle(elementType)
		styles = append(styles, plugin.StyleRange{
			Start: token.Start(),
			End:   token.End(),
			Style: style,
		})
	}
	
	return plugin.RenderedLine{
		Content: content,
		Styles:  styles,
		Metadata: map[string]interface{}{
			"syntax_highlighted": true,
		},
	}, nil
}

// renderTextLine renders a plain text line with basic styling
func (r *TerminalRenderer) renderTextLine(line string, themePlugin theme.Theme) (plugin.RenderedLine, error) {
	// Apply tab expansion
	content := r.expandTabs(line)
	
	// Apply basic text styling
	textStyle := themePlugin.GetStyle(theme.TextNormal)
	
	styles := []plugin.StyleRange{
		{
			Start: 0,
			End:   len(content),
			Style: textStyle,
		},
	}
	
	return plugin.RenderedLine{
		Content: content,
		Styles:  styles,
		Metadata: map[string]interface{}{
			"plain_text": true,
		},
	}, nil
}

// expandTabs expands tabs to spaces
func (r *TerminalRenderer) expandTabs(line string) string {
	if r.config.TabWidth <= 0 {
		return line
	}
	
	var result strings.Builder
	col := 0
	
	for _, ch := range line {
		if ch == '\t' {
			// Calculate spaces needed to reach next tab stop
			spaces := r.config.TabWidth - (col % r.config.TabWidth)
			result.WriteString(strings.Repeat(" ", spaces))
			col += spaces
		} else {
			result.WriteRune(ch)
			col++
		}
	}
	
	return result.String()
}

// RenderToString converts rendered lines to terminal output
func (r *TerminalRenderer) RenderToString(lines []plugin.RenderedLine, themePlugin theme.Theme) string {
	var result strings.Builder
	
	for i, line := range lines {
		// Add line number if enabled
		if r.config.ShowLineNumbers {
			lineNumStyle := themePlugin.GetStyle(theme.EditorLineNumber)
			lineNumStr := fmt.Sprintf("%4d │ ", i+1)
			styledLineNum := lineNumStyle.ToLipgloss().Render(lineNumStr)
			result.WriteString(styledLineNum)
		}
		
		// Render the line content with styles
		content := r.renderLineWithStyles(line, themePlugin)
		result.WriteString(content)
		
		// Add newline except for last line
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}
	
	return result.String()
}

// renderLineWithStyles applies styles to a line
func (r *TerminalRenderer) renderLineWithStyles(line plugin.RenderedLine, themePlugin theme.Theme) string {
	if len(line.Styles) == 0 {
		return line.Content
	}
	
	// Sort styles by start position
	// (In a real implementation, you'd want to handle overlapping styles)
	
	var result strings.Builder
	runes := []rune(line.Content)
	lastEnd := 0
	
	for _, styleRange := range line.Styles {
		// Add unstyled text before this style
		if styleRange.Start > lastEnd {
			result.WriteString(string(runes[lastEnd:styleRange.Start]))
		}
		
		// Apply the style
		if styleRange.End <= len(runes) {
			text := string(runes[styleRange.Start:styleRange.End])
			styledText := styleRange.Style.ToLipgloss().Render(text)
			result.WriteString(styledText)
			lastEnd = styleRange.End
		}
	}
	
	// Add any remaining unstyled text
	if lastEnd < len(runes) {
		result.WriteString(string(runes[lastEnd:]))
	}
	
	return result.String()
}