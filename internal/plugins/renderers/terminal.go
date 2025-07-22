package renderers

import (
	"context"
	"fmt"
	"strings"
	"github.com/ofri/mde/pkg/ast"
	"github.com/ofri/mde/pkg/plugin"
)

// ANSI color codes for terminal-inherited theming
const (
	// Basic colors (0-7)
	ColorBlack   = "0"
	ColorRed     = "1"
	ColorGreen   = "2"
	ColorYellow  = "3"
	ColorBlue    = "4"
	ColorMagenta = "5"
	ColorCyan    = "6"
	ColorWhite   = "7"
	
	// Bright colors (8-15)
	ColorBrightBlack   = "8"  // Gray
	ColorBrightRed     = "9"
	ColorBrightGreen   = "10"
	ColorBrightYellow  = "11"
	ColorBrightBlue    = "12"
	ColorBrightMagenta = "13"
	ColorBrightCyan    = "14"
	ColorBrightWhite   = "15"
	
	// Aliases for common uses
	ColorGray    = ColorBrightBlack
	ColorDefault = ""  // Use terminal's default color
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
			ShowLineNumbers: true,
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
	
	if lineNumberWidth, ok := options["lineNumberWidth"].(int); ok {
		r.config.LineNumberWidth = lineNumberWidth
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


// RenderVisible renders only visible lines from the viewport.
//
// CRITICAL: This method ADDS LINE NUMBERS to content. It's the only place
// in the rendering pipeline that does this.
//
// RESPONSIBILITIES:
// - Extract visible lines based on viewport (startLine to endLine)
// - Add line numbers if ShowLineNumbers is true
// - Apply horizontal scrolling while preserving line numbers
//
// FOR LLM: After this method, RenderedLine.Content includes line numbers.
func (r *TerminalRenderer) RenderVisible(ctx context.Context, renderCtx *plugin.RenderContext) ([]plugin.RenderedLine, error) {
	viewport := renderCtx.Viewport
	doc := renderCtx.Document
	
	// Calculate the range of lines to render based on viewport
	startLine := viewport.GetTopLine()
	endLine := startLine + viewport.GetHeight()
	
	// Ensure we don't go beyond document bounds
	if startLine < 0 {
		startLine = 0
	}
	if endLine > doc.LineCount() {
		endLine = doc.LineCount()
	}
	
	// Handle edge case: viewport starts beyond document end
	if startLine >= doc.LineCount() {
		return []plugin.RenderedLine{}, nil
	}
	
	// Pre-allocate slice for visible lines
	lines := make([]plugin.RenderedLine, 0, endLine-startLine)
	
	// Process only the visible lines
	for i := startLine; i < endLine; i++ {
		lineContent := doc.GetLine(i)
		
		// Add line numbers if enabled
		if renderCtx.ShowLineNumbers {
			// Format line number with proper width and separator
			// Use same format as editor: "%Nd │ " (includes space after │)
			lineNumStr := fmt.Sprintf("%*d│ ", viewport.GetLineNumberWidth()-2, i+1)
			lineContent = lineNumStr + lineContent
		}
		
		// Apply horizontal scrolling
		if viewport.GetLeftColumn() > 0 {
			lineContent = r.applyHorizontalScroll(lineContent, viewport.GetLeftColumn(), renderCtx.ShowLineNumbers, viewport.GetLineNumberWidth())
		}
		
		// Render the line with syntax highlighting (future enhancement)
		renderedLine, err := r.renderTextLine(lineContent)
		if err != nil {
			return nil, fmt.Errorf("failed to render line %d: %w", i, err)
		}
		
		lines = append(lines, renderedLine)
	}
	
	return lines, nil
}

// RenderPreviewVisible implements viewport-aware rendering for preview mode.
// This method renders markdown with formatting while respecting viewport boundaries.
//
// PREVIEW MODE DIFFERENCES:
// - No line numbers are shown
// - Markdown formatting is applied (headers, emphasis, etc.)
// - Content is rendered for display, not editing
func (r *TerminalRenderer) RenderPreviewVisible(ctx context.Context, renderCtx *plugin.RenderContext) ([]plugin.RenderedLine, error) {
	viewport := renderCtx.Viewport
	doc := renderCtx.Document
	
	// For preview mode, we need to work with the full markdown text
	// to properly parse markdown elements that might span multiple lines
	markdownText := doc.GetText()
	allLines := strings.Split(markdownText, "\n")
	
	// Calculate visible range
	startLine := viewport.GetTopLine()
	endLine := startLine + viewport.GetHeight()
	
	// Ensure bounds are valid
	if startLine < 0 {
		startLine = 0
	}
	if endLine > len(allLines) {
		endLine = len(allLines)
	}
	if startLine >= len(allLines) {
		return []plugin.RenderedLine{}, nil
	}
	
	// Extract visible lines
	visibleLines := allLines[startLine:endLine]
	renderedLines := make([]plugin.RenderedLine, 0, len(visibleLines))
	
	// Render each visible line with markdown formatting
	for _, line := range visibleLines {
		renderedLine := r.renderMarkdownLine(line)
		
		// Apply horizontal scrolling to preview content
		if viewport.GetLeftColumn() > 0 && len(renderedLine.Content) > viewport.GetLeftColumn() {
			// For preview mode, we simply trim from the left
			renderedLine.Content = renderedLine.Content[viewport.GetLeftColumn():]
			
			// Adjust style ranges for horizontal scroll
			for j := range renderedLine.Styles {
				renderedLine.Styles[j].Start -= viewport.GetLeftColumn()
				renderedLine.Styles[j].End -= viewport.GetLeftColumn()
				
				// Clamp to visible range
				if renderedLine.Styles[j].Start < 0 {
					renderedLine.Styles[j].Start = 0
				}
				if renderedLine.Styles[j].End < 0 {
					renderedLine.Styles[j].End = 0
				}
			}
		}
		
		renderedLines = append(renderedLines, renderedLine)
	}
	
	return renderedLines, nil
}

// applyHorizontalScroll applies horizontal scrolling to a line while preserving line numbers.
// This is a helper method that handles the complexity of scrolling content while keeping
// line numbers visible.
//
// ALGORITHM:
// 1. If line numbers are shown, preserve them during horizontal scroll
// 2. Apply the scroll offset only to the content portion
// 3. Handle edge cases where content is shorter than scroll offset
func (r *TerminalRenderer) applyHorizontalScroll(line string, leftColumn int, hasLineNumbers bool, lineNumberWidth int) string {
	if !hasLineNumbers {
		// Simple case: no line numbers, just trim from left
		if len(line) > leftColumn {
			return line[leftColumn:]
		}
		return ""
	}
	
	// Complex case: preserve line numbers while scrolling content
	if len(line) <= lineNumberWidth {
		// Line only contains line number (or less), return as-is
		return line
	}
	
	// Split line number and content
	lineNumPart := line[:lineNumberWidth]
	contentPart := line[lineNumberWidth:]
	
	// Apply scroll to content portion only
	if len(contentPart) > leftColumn {
		return lineNumPart + contentPart[leftColumn:]
	}
	
	// Content is entirely scrolled off, but keep line number
	return lineNumPart
}

// renderMarkdownLine renders a single line with markdown formatting
func (r *TerminalRenderer) renderMarkdownLine(line string) plugin.RenderedLine {
	trimmedLine := strings.TrimSpace(line)
	
	// Handle different markdown elements
	if strings.HasPrefix(trimmedLine, "# ") {
		// H1 heading - bright red and bold
		return plugin.RenderedLine{
			Content: line,
			Styles: []plugin.StyleRange{
				{Start: 0, End: len(line), Style: plugin.Style{Foreground: ColorBrightRed, Bold: true}},
			},
		}
	} else if strings.HasPrefix(trimmedLine, "## ") {
		// H2 heading - bright green and bold
		return plugin.RenderedLine{
			Content: line,
			Styles: []plugin.StyleRange{
				{Start: 0, End: len(line), Style: plugin.Style{Foreground: ColorBrightGreen, Bold: true}},
			},
		}
	} else if strings.HasPrefix(trimmedLine, "### ") {
		// H3 heading - bright yellow and bold
		return plugin.RenderedLine{
			Content: line,
			Styles: []plugin.StyleRange{
				{Start: 0, End: len(line), Style: plugin.Style{Foreground: ColorBrightYellow, Bold: true}},
			},
		}
	} else if strings.HasPrefix(trimmedLine, "#### ") {
		// H4 heading - bright blue and bold
		return plugin.RenderedLine{
			Content: line,
			Styles: []plugin.StyleRange{
				{Start: 0, End: len(line), Style: plugin.Style{Foreground: ColorBrightBlue, Bold: true}},
			},
		}
	} else if strings.HasPrefix(trimmedLine, "##### ") {
		// H5 heading - bright magenta and bold
		return plugin.RenderedLine{
			Content: line,
			Styles: []plugin.StyleRange{
				{Start: 0, End: len(line), Style: plugin.Style{Foreground: ColorBrightMagenta, Bold: true}},
			},
		}
	} else if strings.HasPrefix(trimmedLine, "###### ") {
		// H6 heading - bright cyan and bold
		return plugin.RenderedLine{
			Content: line,
			Styles: []plugin.StyleRange{
				{Start: 0, End: len(line), Style: plugin.Style{Foreground: ColorBrightCyan, Bold: true}},
			},
		}
	} else if strings.HasPrefix(trimmedLine, "> ") {
		// Blockquote - gray/bright black
		text := strings.TrimPrefix(trimmedLine, "> ")
		return plugin.RenderedLine{
			Content: "  " + text,
			Styles: []plugin.StyleRange{
				{Start: 0, End: len("  " + text), Style: plugin.Style{Foreground: getAccessibleColor(ColorGray)}},
			},
		}
	} else if strings.HasPrefix(trimmedLine, "- ") || strings.HasPrefix(trimmedLine, "* ") {
		// Bullet list
		text := trimmedLine[2:]
		return plugin.RenderedLine{
			Content: "  • " + text,
			Styles: []plugin.StyleRange{
				{Start: 0, End: 3, Style: plugin.Style{Foreground: ColorYellow}}, // Bullet
			},
		}
	} else if strings.HasPrefix(trimmedLine, "```") {
		// Code block delimiter - cyan
		return plugin.RenderedLine{
			Content: line,
			Styles: []plugin.StyleRange{
				{Start: 0, End: len(line), Style: plugin.Style{Foreground: ColorCyan}},
			},
		}
	} else if len(trimmedLine) > 0 && (trimmedLine[0] >= '0' && trimmedLine[0] <= '9') && strings.Contains(trimmedLine, ". ") {
		// Numbered list (simple detection)
		parts := strings.SplitN(trimmedLine, ". ", 2)
		if len(parts) == 2 {
			return plugin.RenderedLine{
				Content: "  " + parts[0] + ". " + parts[1],
				Styles: []plugin.StyleRange{
					{Start: 2, End: 2 + len(parts[0]) + 1, Style: plugin.Style{Foreground: ColorYellow}}, // Number
				},
			}
		}
	}
	
	// Handle inline formatting for regular text
	return r.renderInlineFormatting(line)
}

// renderInlineFormatting handles bold, italic, code, and links
func (r *TerminalRenderer) renderInlineFormatting(line string) plugin.RenderedLine {
	content := line
	styles := []plugin.StyleRange{}
	
	// For simplicity, we'll apply basic styling without complex parsing
	// In a real implementation, this would properly parse markdown
	
	// If no styles applied, return as-is with default style
	if len(styles) == 0 {
		return plugin.RenderedLine{
			Content: content,
			Styles:  []plugin.StyleRange{},
		}
	}
	
	return plugin.RenderedLine{
		Content: content,
		Styles:  styles,
	}
}

// RenderLine renders a single line with syntax highlighting
func (r *TerminalRenderer) RenderLine(ctx context.Context, line string, tokens []ast.Token) (plugin.RenderedLine, error) {
	if len(tokens) == 0 {
		// No syntax highlighting, render as plain text
		return r.renderTextLine(line)
	}
	
	// Apply syntax highlighting
	content := line
	styles := make([]plugin.StyleRange, 0, len(tokens))
	
	for _, token := range tokens {
		var style plugin.Style
		switch token.Kind() {
		case ast.TokenKeyword:
			style = plugin.Style{Foreground: getAccessibleColor(ColorMagenta)}
		case ast.TokenString:
			style = plugin.Style{Foreground: getAccessibleColor(ColorGreen)}
		case ast.TokenComment:
			style = plugin.Style{Foreground: getAccessibleColor(ColorGray)}
		case ast.TokenNumber:
			style = plugin.Style{Foreground: getAccessibleColor(ColorYellow)}
		// Markdown-specific tokens
		case ast.TokenHeading:
			style = plugin.Style{Foreground: ColorBrightRed, Bold: true}
		case ast.TokenBold:
			style = plugin.Style{Bold: true}
		case ast.TokenItalic:
			style = plugin.Style{Italic: true}
		case ast.TokenCode:
			style = plugin.Style{Foreground: ColorCyan}
		case ast.TokenCodeBlock:
			style = plugin.Style{Foreground: ColorCyan}
		case ast.TokenLink:
			style = plugin.Style{Foreground: getAccessibleColor(ColorBlue), Underline: true}
		case ast.TokenLinkText:
			style = plugin.Style{Foreground: getAccessibleColor(ColorBlue)}
		case ast.TokenLinkURL:
			style = plugin.Style{Foreground: getAccessibleColor(ColorGray)}
		case ast.TokenImage:
			style = plugin.Style{Foreground: ColorMagenta}
		case ast.TokenQuote:
			style = plugin.Style{Foreground: getAccessibleColor(ColorGray)}
		case ast.TokenList:
			style = plugin.Style{Foreground: ColorYellow}
		case ast.TokenDelimiter:
			style = plugin.Style{Foreground: getAccessibleColor(ColorGray)}
		default:
			// No special styling
			continue
		}
		
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
func (r *TerminalRenderer) renderTextLine(line string) (plugin.RenderedLine, error) {
	// Apply tab expansion
	content := r.expandTabs(line)
	
	// No special styling for plain text
	return plugin.RenderedLine{
		Content: content,
		Styles:  []plugin.StyleRange{},
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

// RenderToString converts rendered lines to terminal output with proper styling.
//
// CRITICAL: This method does NOT add line numbers. Line numbers are already
// included in RenderedLine.Content by RenderVisible/RenderPreviewVisible.
//
// RENDERING PIPELINE:
// 1. RenderVisible/RenderPreviewVisible: Adds line numbers, applies viewport
// 2. RenderToString/RenderToStringWithCursor: Applies styling and cursor
//
// FOR LLM: RenderedLine.Content already contains line numbers if enabled.
func (r *TerminalRenderer) RenderToString(lines []plugin.RenderedLine) string {
	var result strings.Builder
	
	for i, line := range lines {
		// Render the line content with styles
		// NOTE: line.Content already includes line numbers if ShowLineNumbers is true
		content := r.renderLineWithStyles(line)
		result.WriteString(content)
		
		// Add newline except for last line
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}
	
	return result.String()
}

// RenderToStringWithCursor renders lines with cursor at the specified position.
//
// CRITICAL: Like RenderToString, this method does NOT add line numbers.
// Line numbers are already in RenderedLine.Content.
//
// PARAMETERS:
// - cursorRow: Line index in 'lines' array (0-based)
// - cursorCol: Column position within that line's Content string
//
// FOR LLM: cursorCol is already adjusted for line numbers by the caller.
func (r *TerminalRenderer) RenderToStringWithCursor(lines []plugin.RenderedLine, cursorRow, cursorCol int) string {
	var result strings.Builder
	
	for i, line := range lines {
		// Render the line content with styles, including cursor if on this line
		// NOTE: line.Content already includes line numbers if ShowLineNumbers is true
		if i == cursorRow {
			content := r.renderLineWithStylesAndCursor(line, cursorCol)
			result.WriteString(content)
		} else {
			content := r.renderLineWithStyles(line)
			result.WriteString(content)
		}
		
		// Add newline except for last line
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}
	
	return result.String()
}

// renderLineWithStylesAndCursor applies styles to a line and adds cursor at specified position.
//
// CURSOR POSITIONING:
// - cursorCol is in ScreenPos coordinates (already includes line number offset)
// - The viewport transformation handles line number offset calculation
// - End-of-line: extend line with space, replace with cursor → "Hello█"
// - Within line: replace existing character with cursor → "He█lo"
// - Empty line: extend with space, replace with cursor → "█"
func (r *TerminalRenderer) renderLineWithStylesAndCursor(line plugin.RenderedLine, cursorCol int) string {
	// CRITICAL ARCHITECTURAL NOTE:
	// Line numbers are already included in line.Content by RenderVisible.
	// The cursorCol parameter is the position within line.Content where the cursor
	// should be placed. No adjustment for line numbers is needed here.
	
	// Use cursorCol directly - it's already the correct position within line.Content
	adjustedCursorCol := cursorCol
	
	// Bounds checking
	if adjustedCursorCol < 0 {
		adjustedCursorCol = 0
	}
	
	runes := []rune(line.Content)
	
	// Clamp cursor to end of line
	if adjustedCursorCol > len(runes) {
		adjustedCursorCol = len(runes)
	}
	
	// Extend line with spaces to cursor position if needed
	if adjustedCursorCol >= len(runes) {
		// Extend line with spaces to cursor position
		spaceCount := adjustedCursorCol + 1 - len(runes)
		runes = append(runes, []rune(strings.Repeat(" ", spaceCount))...)
	}
	
	// Replace character at cursor position with cursor block
	runes[adjustedCursorCol] = '█'
	
	// Create new rendered line with cursor
	lineWithCursor := plugin.RenderedLine{
		Content: string(runes),
		Styles:  line.Styles,
	}
	
	return r.renderLineWithStyles(lineWithCursor)
}

// formatLineNumber formats a line number using the appropriate width
func (r *TerminalRenderer) formatLineNumber(lineNum, totalLines int) string {
	if !r.config.ShowLineNumbers {
		return ""
	}
	
	// Calculate digits needed for the total number of lines
	digits := len(fmt.Sprintf("%d", totalLines))
	
	// Create format string: "%Nd │ " where N is the digit count
	formatStr := fmt.Sprintf("%%%dd │ ", digits)
	
	return fmt.Sprintf(formatStr, lineNum)
}

// renderLineWithStyles applies styles to a line
func (r *TerminalRenderer) renderLineWithStyles(line plugin.RenderedLine) string {
	if len(line.Styles) == 0 {
		// No styles, return content as-is
		return line.Content
	}
	
	// Sort styles by start position to handle overlapping styles properly
	// For now, we'll process them in the order they appear
	
	var result strings.Builder
	runes := []rune(line.Content)
	lastEnd := 0
	
	for _, styleRange := range line.Styles {
		// Add unstyled text before this style
		if styleRange.Start > lastEnd {
			result.WriteString(string(runes[lastEnd:styleRange.Start]))
		}
		
		// Apply the style - ensure bounds are valid
		if styleRange.Start >= 0 && styleRange.End <= len(runes) && styleRange.Start < styleRange.End {
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