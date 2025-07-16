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

// RenderPreview renders a preview of the document with markdown formatting
func (r *TerminalRenderer) RenderPreview(ctx context.Context, doc *ast.Document, themePlugin theme.Theme) ([]plugin.RenderedLine, error) {
	// Get the raw text content
	markdownText := doc.GetText()
	
	// Parse markdown and render with formatting
	lines := strings.Split(markdownText, "\n")
	renderedLines := make([]plugin.RenderedLine, 0, len(lines))
	
	for _, line := range lines {
		renderedLine := r.renderMarkdownLine(line, themePlugin)
		renderedLines = append(renderedLines, renderedLine)
	}
	
	return renderedLines, nil
}

// renderMarkdownLine renders a single line with markdown formatting
func (r *TerminalRenderer) renderMarkdownLine(line string, themePlugin theme.Theme) plugin.RenderedLine {
	trimmedLine := strings.TrimSpace(line)
	
	// Handle different markdown elements
	if strings.HasPrefix(trimmedLine, "# ") {
		// H1 heading
		text := strings.TrimPrefix(trimmedLine, "# ")
		headingStyle := themePlugin.GetStyle(theme.MarkdownHeading1)
		return plugin.RenderedLine{
			Content: "# " + text,
			Styles: []plugin.StyleRange{
				{Start: 0, End: len("# " + text), Style: headingStyle},
			},
		}
	} else if strings.HasPrefix(trimmedLine, "## ") {
		// H2 heading
		text := strings.TrimPrefix(trimmedLine, "## ")
		headingStyle := themePlugin.GetStyle(theme.MarkdownHeading2)
		return plugin.RenderedLine{
			Content: "## " + text,
			Styles: []plugin.StyleRange{
				{Start: 0, End: len("## " + text), Style: headingStyle},
			},
		}
	} else if strings.HasPrefix(trimmedLine, "### ") {
		// H3 heading
		text := strings.TrimPrefix(trimmedLine, "### ")
		headingStyle := themePlugin.GetStyle(theme.MarkdownHeading3)
		return plugin.RenderedLine{
			Content: "### " + text,
			Styles: []plugin.StyleRange{
				{Start: 0, End: len("### " + text), Style: headingStyle},
			},
		}
	} else if strings.HasPrefix(trimmedLine, "> ") {
		// Blockquote
		text := strings.TrimPrefix(trimmedLine, "> ")
		quoteStyle := themePlugin.GetStyle(theme.MarkdownQuote)
		return plugin.RenderedLine{
			Content: "  " + text,
			Styles: []plugin.StyleRange{
				{Start: 0, End: len("  " + text), Style: quoteStyle},
			},
		}
	} else if strings.HasPrefix(trimmedLine, "- ") || strings.HasPrefix(trimmedLine, "* ") {
		// Bullet list
		text := trimmedLine[2:]
		listStyle := themePlugin.GetStyle(theme.MarkdownList)
		return plugin.RenderedLine{
			Content: "  • " + text,
			Styles: []plugin.StyleRange{
				{Start: 0, End: len("  • " + text), Style: listStyle},
			},
		}
	} else if strings.HasPrefix(trimmedLine, "```") {
		// Code block delimiter
		codeStyle := themePlugin.GetStyle(theme.MarkdownCodeBlock)
		return plugin.RenderedLine{
			Content: line,
			Styles: []plugin.StyleRange{
				{Start: 0, End: len(line), Style: codeStyle},
			},
		}
	} else if len(trimmedLine) > 0 && (trimmedLine[0] >= '0' && trimmedLine[0] <= '9') && strings.Contains(trimmedLine, ". ") {
		// Numbered list (simple detection)
		parts := strings.SplitN(trimmedLine, ". ", 2)
		if len(parts) == 2 {
			listStyle := themePlugin.GetStyle(theme.MarkdownList)
			return plugin.RenderedLine{
				Content: "  " + parts[0] + ". " + parts[1],
				Styles: []plugin.StyleRange{
					{Start: 0, End: len("  " + parts[0] + ". " + parts[1]), Style: listStyle},
				},
			}
		}
	}
	
	// Handle inline formatting for regular text
	return r.renderInlineFormatting(line, themePlugin)
}

// renderInlineFormatting handles bold, italic, code, and links
func (r *TerminalRenderer) renderInlineFormatting(line string, themePlugin theme.Theme) plugin.RenderedLine {
	content := line
	styles := []plugin.StyleRange{}
	
	// Handle **bold**
	for {
		// Simple regex replacement for bold
		if start := strings.Index(content, "**"); start != -1 {
			if end := strings.Index(content[start+2:], "**"); end != -1 {
				end += start + 2
				boldStyle := themePlugin.GetStyle(theme.MarkdownBold)
				styles = append(styles, plugin.StyleRange{
					Start: start,
					End:   end + 2,
					Style: boldStyle,
				})
				// Remove the ** markers for display
				content = content[:start] + content[start+2:end] + content[end+2:]
				// Adjust the end position after removal
				styles[len(styles)-1].End = end - 2
			} else {
				break
			}
		} else {
			break
		}
	}
	
	// Handle *italic*
	for {
		if start := strings.Index(content, "*"); start != -1 {
			if end := strings.Index(content[start+1:], "*"); end != -1 {
				end += start + 1
				// Make sure it's not part of a bold (already processed)
				validItalic := true
				for _, style := range styles {
					if start >= style.Start && end <= style.End {
						validItalic = false
						break
					}
				}
				if validItalic {
					italicStyle := themePlugin.GetStyle(theme.MarkdownItalic)
					styles = append(styles, plugin.StyleRange{
						Start: start,
						End:   end + 1,
						Style: italicStyle,
					})
					// Remove the * markers for display
					content = content[:start] + content[start+1:end] + content[end+1:]
					// Adjust the end position after removal
					styles[len(styles)-1].End = end - 1
				} else {
					break
				}
			} else {
				break
			}
		} else {
			break
		}
	}
	
	// Handle `code`
	for {
		if start := strings.Index(content, "`"); start != -1 {
			if end := strings.Index(content[start+1:], "`"); end != -1 {
				end += start + 1
				codeStyle := themePlugin.GetStyle(theme.MarkdownCode)
				styles = append(styles, plugin.StyleRange{
					Start: start,
					End:   end + 1,
					Style: codeStyle,
				})
				// Remove the ` markers for display
				content = content[:start] + content[start+1:end] + content[end+1:]
				// Adjust the end position after removal
				styles[len(styles)-1].End = end - 1
			} else {
				break
			}
		} else {
			break
		}
	}
	
	// If no styles applied, use normal text style
	if len(styles) == 0 {
		textStyle := themePlugin.GetStyle(theme.TextNormal)
		styles = append(styles, plugin.StyleRange{
			Start: 0,
			End:   len(content),
			Style: textStyle,
		})
	}
	
	return plugin.RenderedLine{
		Content: content,
		Styles:  styles,
	}
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
		// Markdown-specific tokens
		case ast.TokenHeading:
			elementType = theme.MarkdownHeading
		case ast.TokenBold:
			elementType = theme.MarkdownBold
		case ast.TokenItalic:
			elementType = theme.MarkdownItalic
		case ast.TokenCode:
			elementType = theme.MarkdownCode
		case ast.TokenCodeBlock:
			elementType = theme.MarkdownCodeBlock
		case ast.TokenLink:
			elementType = theme.MarkdownLink
		case ast.TokenLinkText:
			elementType = theme.MarkdownLinkText
		case ast.TokenLinkURL:
			elementType = theme.MarkdownLinkURL
		case ast.TokenImage:
			elementType = theme.MarkdownImage
		case ast.TokenQuote:
			elementType = theme.MarkdownQuote
		case ast.TokenList:
			elementType = theme.MarkdownList
		case ast.TokenDelimiter:
			elementType = theme.MarkdownDelimiter
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

// RenderToStringWithCursor renders lines with cursor positioning
// cursorRow, cursorCol are in content coordinates (ContentPos)
func (r *TerminalRenderer) RenderToStringWithCursor(lines []plugin.RenderedLine, themePlugin theme.Theme, cursorRow, cursorCol int) string {
	var result strings.Builder
	
	for i, line := range lines {
		// Add line number if enabled
		if r.config.ShowLineNumbers {
			lineNumStyle := themePlugin.GetStyle(theme.EditorLineNumber)
			lineNumStr := fmt.Sprintf("%4d │ ", i+1)
			styledLineNum := lineNumStyle.ToLipgloss().Render(lineNumStr)
			result.WriteString(styledLineNum)
		}
		
		// Render the line content with styles, including cursor if on this line
		if i == cursorRow {
			content := r.renderLineWithStylesAndCursor(line, themePlugin, cursorCol)
			result.WriteString(content)
		} else {
			content := r.renderLineWithStyles(line, themePlugin)
			result.WriteString(content)
		}
		
		// Add newline except for last line
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}
	
	return result.String()
}

// renderLineWithStylesAndCursor applies styles to a line and adds cursor at specified position
func (r *TerminalRenderer) renderLineWithStylesAndCursor(line plugin.RenderedLine, themePlugin theme.Theme, cursorCol int) string {
	// COORDINATE SYSTEM: cursorCol is in content coordinates (ContentPos)
	// This means it already includes line number offset if line numbers are enabled
	// We need to convert back to line-content coordinates for rendering
	adjustedCursorCol := cursorCol
	if r.config.ShowLineNumbers {
		// Subtract 6 for line number prefix "%4d │ " to get position within line content
		adjustedCursorCol = cursorCol - 6
	}
	
	// Ensure cursor is within bounds
	if adjustedCursorCol < 0 {
		adjustedCursorCol = 0
	}
	
	runes := []rune(line.Content)
	
	// If cursor is beyond the line, place it at the end
	if adjustedCursorCol > len(runes) {
		adjustedCursorCol = len(runes)
	}
	
	// Create a new style range for the cursor
	cursorStyle := themePlugin.GetStyle(theme.EditorCursor)
	
	// If cursor is at end of line, append cursor character
	if adjustedCursorCol == len(runes) {
		// Render the line normally, then append cursor
		normalContent := r.renderLineWithStyles(line, themePlugin)
		cursorChar := cursorStyle.ToLipgloss().Render("█")
		return normalContent + cursorChar
	}
	
	// Replace the character at cursor position with cursor character
	runes[adjustedCursorCol] = '█'
	
	// Create new rendered line with cursor character
	lineWithCursor := plugin.RenderedLine{
		Content: string(runes),
		Styles:  line.Styles, // Keep existing styles
	}
	
	return r.renderLineWithStyles(lineWithCursor, themePlugin)
}

// renderLineWithStyles applies styles to a line
func (r *TerminalRenderer) renderLineWithStyles(line plugin.RenderedLine, themePlugin theme.Theme) string {
	if len(line.Styles) == 0 {
		// Apply default text styling if no styles are provided
		defaultStyle := themePlugin.GetStyle(theme.TextNormal)
		return defaultStyle.ToLipgloss().Render(line.Content)
	}
	
	// Get default style for unstyled text
	defaultStyle := themePlugin.GetStyle(theme.TextNormal)
	
	// Sort styles by start position to handle overlapping styles properly
	// For now, we'll process them in the order they appear
	
	var result strings.Builder
	runes := []rune(line.Content)
	lastEnd := 0
	
	for _, styleRange := range line.Styles {
		// Add unstyled text before this style with default styling
		if styleRange.Start > lastEnd {
			unstyledText := string(runes[lastEnd:styleRange.Start])
			styledUnstyledText := defaultStyle.ToLipgloss().Render(unstyledText)
			result.WriteString(styledUnstyledText)
		}
		
		// Apply the style - ensure bounds are valid
		if styleRange.Start >= 0 && styleRange.End <= len(runes) && styleRange.Start < styleRange.End {
			text := string(runes[styleRange.Start:styleRange.End])
			styledText := styleRange.Style.ToLipgloss().Render(text)
			result.WriteString(styledText)
			lastEnd = styleRange.End
		}
	}
	
	// Add any remaining unstyled text with default styling
	if lastEnd < len(runes) {
		remainingText := string(runes[lastEnd:])
		styledRemainingText := defaultStyle.ToLipgloss().Render(remainingText)
		result.WriteString(styledRemainingText)
	}
	
	return result.String()
}