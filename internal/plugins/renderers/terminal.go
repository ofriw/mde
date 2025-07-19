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

// Render renders the document
func (r *TerminalRenderer) Render(ctx context.Context, doc *ast.Document) ([]plugin.RenderedLine, error) {
	lines := make([]plugin.RenderedLine, 0, doc.LineCount())
	
	for i := 0; i < doc.LineCount(); i++ {
		line := doc.GetLine(i)
		
		// For now, render as plain text with basic styling
		renderedLine, err := r.renderTextLine(line)
		if err != nil {
			return nil, fmt.Errorf("failed to render line %d: %w", i, err)
		}
		
		lines = append(lines, renderedLine)
	}
	
	return lines, nil
}

// RenderPreview renders a preview of the document with markdown formatting
func (r *TerminalRenderer) RenderPreview(ctx context.Context, doc *ast.Document) ([]plugin.RenderedLine, error) {
	// Get the raw text content
	markdownText := doc.GetText()
	
	// Parse markdown and render with formatting
	lines := strings.Split(markdownText, "\n")
	renderedLines := make([]plugin.RenderedLine, 0, len(lines))
	
	for _, line := range lines {
		renderedLine := r.renderMarkdownLine(line)
		renderedLines = append(renderedLines, renderedLine)
	}
	
	return renderedLines, nil
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

// RenderToString converts rendered lines to terminal output
func (r *TerminalRenderer) RenderToString(lines []plugin.RenderedLine) string {
	var result strings.Builder
	
	for i, line := range lines {
		// Add line number if enabled
		if r.config.ShowLineNumbers {
			lineNumStr := r.formatLineNumber(i+1, len(lines))
			lineNumStyle := plugin.Style{Foreground: getAccessibleColor(ColorGray)}
			styledLineNum := lineNumStyle.ToLipgloss().Render(lineNumStr)
			result.WriteString(styledLineNum)
		}
		
		// Render the line content with styles
		content := r.renderLineWithStyles(line)
		result.WriteString(content)
		
		// Add newline except for last line
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}
	
	return result.String()
}

// RenderToStringWithCursor renders lines with cursor positioning
// cursorRow, cursorCol are in screen coordinates (ScreenPos)
func (r *TerminalRenderer) RenderToStringWithCursor(lines []plugin.RenderedLine, cursorRow, cursorCol int) string {
	var result strings.Builder
	
	for i, line := range lines {
		// Add line number if enabled
		if r.config.ShowLineNumbers {
			lineNumStr := r.formatLineNumber(i+1, len(lines))
			lineNumStyle := plugin.Style{Foreground: getAccessibleColor(ColorGray)}
			styledLineNum := lineNumStyle.ToLipgloss().Render(lineNumStr)
			result.WriteString(styledLineNum)
		}
		
		// Render the line content with styles, including cursor if on this line
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
	// ScreenPos coordinates are absolute screen positions, but we need position within line content
	// Subtract line number offset to get position within the line content
	adjustedCursorCol := cursorCol
	if r.config.ShowLineNumbers {
		adjustedCursorCol -= r.config.LineNumberWidth
	}
	
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