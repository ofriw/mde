package tui

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/ofri/mde/pkg/ast"
	"github.com/ofri/mde/pkg/plugin"
	"github.com/ofri/mde/pkg/theme"
)

type Model struct {
	editor       *ast.Editor
	width        int
	height       int
	message      string
	messageTimer int
	err          error
	
	// Modal states
	mode         EditorMode
	input        string
	replaceText  string
	caseSensitive bool
	
	// Preview mode
	previewMode  bool
}

type EditorMode int

const (
	ModeNormal EditorMode = iota
	ModeFind
	ModeReplace
	ModeGoto
)

func New() *Model {
	return &Model{
		editor: ast.NewEditor(),
	}
}

func (m *Model) SetFilename(filename string) {
	err := m.editor.LoadFile(filename)
	if err != nil {
		m.err = err
		return
	}
	
	// Parse the document for syntax highlighting
	m.parseDocument()
}

// Public methods for testing
func (m *Model) GetEditor() *ast.Editor {
	return m.editor
}

func (m *Model) IsPreviewMode() bool {
	return m.previewMode
}

func (m *Model) TogglePreviewMode() {
	m.previewMode = !m.previewMode
}

func (m *Model) ConvertMarkdownToHTML(markdownText string) string {
	return m.convertMarkdownToHTML(markdownText)
}

func (m *Model) ConvertHTMLToTerminalText(htmlContent string) string {
	return m.convertHTMLToTerminalText(htmlContent)
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	// Update editor viewport
	viewportHeight := m.height - 2
	if viewportHeight < 1 {
		viewportHeight = 1
	}
	m.editor.SetViewPort(m.width, viewportHeight)
	m.editor.AdjustViewPort()
	
	// Render content based on mode
	var content string
	if m.previewMode {
		content = m.renderPreviewContent()
	} else {
		content = m.renderEditorContent()
	}
	
	statusBar := m.renderStatusBar()
	helpBar := m.renderHelpBar()
	
	return lipgloss.JoinVertical(lipgloss.Top, content, statusBar, helpBar)
}

// renderEditorContent renders the editor content with syntax highlighting
func (m *Model) renderEditorContent() string {
	editorHeight := m.height - 2
	if editorHeight < 1 {
		editorHeight = 1
	}
	
	// Try to get renderer and theme plugins
	registry := plugin.GetRegistry()
	renderer, err := registry.GetDefaultRenderer()
	if err != nil {
		// Fallback to simple rendering
		return m.renderSimple(editorHeight)
	}
	
	theme, err := registry.GetDefaultTheme()
	if err != nil {
		// Fallback to simple rendering
		return m.renderSimple(editorHeight)
	}
	
	// Render the document using plugins
	ctx := context.Background()
	renderedLines, err := renderer.Render(ctx, m.editor.GetDocument(), theme)
	if err != nil {
		// Fallback to simple rendering
		return m.renderSimple(editorHeight)
	}
	
	// Convert rendered lines to string and add cursor
	content := m.renderLinesWithCursor(renderedLines, theme, renderer)
	
	// Pad to fill editor height
	lines := strings.Split(content, "\n")
	for len(lines) < editorHeight {
		lines = append(lines, "")
	}
	
	// Trim to exact height
	if len(lines) > editorHeight {
		lines = lines[:editorHeight]
	}
	
	result := strings.Join(lines, "\n")
	
	return lipgloss.NewStyle().
		Width(m.width).
		Height(editorHeight).
		Render(result)
}

// renderSimple provides fallback rendering without plugins
func (m *Model) renderSimple(editorHeight int) string {
	// Get visible lines
	lines := m.editor.GetVisibleLines()
	
	// Add cursor
	cursorRow, cursorCol := m.editor.GetCursorScreenPosition()
	if cursorRow >= 0 && cursorRow < len(lines) && cursorCol >= 0 {
		line := lines[cursorRow]
		runes := []rune(line)
		
		if cursorCol < len(runes) {
			runes[cursorCol] = '█'
		} else {
			runes = append(runes, []rune(strings.Repeat(" ", cursorCol-len(runes)))...)
			runes = append(runes, '█')
		}
		
		lines[cursorRow] = string(runes)
	}
	
	// Join lines and pad
	var b strings.Builder
	for i, line := range lines {
		b.WriteString(line)
		if i < len(lines)-1 {
			b.WriteString("\n")
		}
	}
	
	// Pad remaining lines
	for i := len(lines); i < editorHeight; i++ {
		b.WriteString("\n")
	}
	
	return lipgloss.NewStyle().
		Width(m.width).
		Height(editorHeight).
		Render(b.String())
}

// renderPreviewContent renders the markdown content as HTML preview
func (m *Model) renderPreviewContent() string {
	editorHeight := m.height - 2
	if editorHeight < 1 {
		editorHeight = 1
	}
	
	// Get markdown content from editor
	markdownText := m.editor.GetDocument().GetText()
	
	// For large documents, optimize by only rendering visible portion
	var htmlContent string
	if len(markdownText) > 50000 { // Optimize for documents > 50KB
		htmlContent = m.convertMarkdownToHTMLLazy(markdownText)
	} else {
		htmlContent = m.convertMarkdownToHTML(markdownText)
	}
	
	// Convert HTML to terminal-friendly text
	terminalText := m.convertHTMLToTerminalText(htmlContent)
	
	// Split into lines and handle viewport
	lines := strings.Split(terminalText, "\n")
	
	// Apply viewport
	viewport := m.editor.GetViewPort()
	startLine := viewport.Top
	endLine := startLine + editorHeight
	
	if startLine >= len(lines) {
		startLine = len(lines) - 1
		if startLine < 0 {
			startLine = 0
		}
	}
	
	if endLine > len(lines) {
		endLine = len(lines)
	}
	
	// Get visible lines
	var visibleLines []string
	if startLine < len(lines) {
		visibleLines = lines[startLine:endLine]
	}
	
	// Pad to fill editor height
	for len(visibleLines) < editorHeight {
		visibleLines = append(visibleLines, "")
	}
	
	// Trim to exact height
	if len(visibleLines) > editorHeight {
		visibleLines = visibleLines[:editorHeight]
	}
	
	result := strings.Join(visibleLines, "\n")
	
	return lipgloss.NewStyle().
		Width(m.width).
		Height(editorHeight).
		Render(result)
}

// convertMarkdownToHTML converts markdown text to HTML using goldmark
func (m *Model) convertMarkdownToHTML(markdownText string) string {
	var buf bytes.Buffer
	
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
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)
	
	if err := md.Convert([]byte(markdownText), &buf); err != nil {
		// If conversion fails, return original text
		return markdownText
	}
	
	return buf.String()
}

// convertMarkdownToHTMLLazy converts only the visible portion of markdown for large documents
func (m *Model) convertMarkdownToHTMLLazy(markdownText string) string {
	lines := strings.Split(markdownText, "\n")
	viewport := m.editor.GetViewPort()
	
	// Calculate visible range with buffer for context
	bufferSize := 100 // Process 100 lines above and below for context
	startLine := viewport.Top - bufferSize
	endLine := viewport.Top + viewport.Height + bufferSize
	
	if startLine < 0 {
		startLine = 0
	}
	if endLine > len(lines) {
		endLine = len(lines)
	}
	
	// Extract visible portion
	visibleLines := lines[startLine:endLine]
	visibleText := strings.Join(visibleLines, "\n")
	
	// Convert the visible portion to HTML
	return m.convertMarkdownToHTML(visibleText)
}

// convertHTMLToTerminalText converts HTML to terminal-friendly text
func (m *Model) convertHTMLToTerminalText(htmlContent string) string {
	// Simple HTML to text conversion for terminal display
	// This is a basic implementation - could be enhanced with proper HTML parsing
	
	result := htmlContent
	
	// Remove HTML tags (basic implementation)
	result = strings.ReplaceAll(result, "<p>", "")
	result = strings.ReplaceAll(result, "</p>", "\n")
	result = strings.ReplaceAll(result, "<br>", "\n")
	result = strings.ReplaceAll(result, "<br/>", "\n")
	result = strings.ReplaceAll(result, "<br />", "\n")
	
	// Handle headings (with and without IDs)
	for i := 1; i <= 6; i++ {
		h := fmt.Sprintf("h%d", i)
		// Handle headings with IDs using regex-like replacement
		idPattern := "<" + h + " id=\""
		endPattern := "\">"
		if strings.Contains(result, idPattern) {
			// Find all occurrences and replace them
			for strings.Contains(result, idPattern) {
				start := strings.Index(result, idPattern)
				if start != -1 {
					end := strings.Index(result[start:], endPattern)
					if end != -1 {
						// Replace the entire ID portion
						before := result[:start]
						after := result[start+end+len(endPattern):]
						result = before + strings.Repeat("#", i) + " " + after
					}
				}
			}
		}
		// Handle simple headings
		result = strings.ReplaceAll(result, "<"+h+">", strings.Repeat("#", i)+" ")
		result = strings.ReplaceAll(result, "</"+h+">", "\n\n")
	}
	
	// Handle emphasis
	result = strings.ReplaceAll(result, "<strong>", "**")
	result = strings.ReplaceAll(result, "</strong>", "**")
	result = strings.ReplaceAll(result, "<em>", "*")
	result = strings.ReplaceAll(result, "</em>", "*")
	result = strings.ReplaceAll(result, "<b>", "**")
	result = strings.ReplaceAll(result, "</b>", "**")
	result = strings.ReplaceAll(result, "<i>", "*")
	result = strings.ReplaceAll(result, "</i>", "*")
	
	// Handle code
	result = strings.ReplaceAll(result, "<code>", "`")
	result = strings.ReplaceAll(result, "</code>", "`")
	result = strings.ReplaceAll(result, "<pre>", "```\n")
	result = strings.ReplaceAll(result, "</pre>", "\n```\n")
	
	// Handle lists
	result = strings.ReplaceAll(result, "<ul>", "")
	result = strings.ReplaceAll(result, "</ul>", "\n")
	result = strings.ReplaceAll(result, "<ol>", "")
	result = strings.ReplaceAll(result, "</ol>", "\n")
	result = strings.ReplaceAll(result, "<li>", "• ")
	result = strings.ReplaceAll(result, "</li>", "\n")
	
	// Handle blockquotes
	result = strings.ReplaceAll(result, "<blockquote>", "> ")
	result = strings.ReplaceAll(result, "</blockquote>", "\n")
	
	// Handle links - show as [text](url)
	// This is a basic implementation - more robust regex could be used
	result = strings.ReplaceAll(result, "<a href=\"", "[")
	result = strings.ReplaceAll(result, "\">", "](")
	result = strings.ReplaceAll(result, "</a>", ")")
	
	// Clean up extra whitespace
	lines := strings.Split(result, "\n")
	var cleanLines []string
	for _, line := range lines {
		cleanLines = append(cleanLines, strings.TrimSpace(line))
	}
	
	return strings.Join(cleanLines, "\n")
}

// parseDocument parses the current document content for syntax highlighting
func (m *Model) parseDocument() {
	registry := plugin.GetRegistry()
	parser, err := registry.GetDefaultParser()
	if err != nil {
		// No parser available, skip parsing
		return
	}
	
	ctx := context.Background()
	_, err = parser.Parse(ctx, m.editor.GetDocument().GetText())
	if err != nil {
		// Parsing failed, continue without syntax highlighting
		return
	}
	
	// For large documents, only parse visible lines for performance
	doc := m.editor.GetDocument()
	lineCount := doc.LineCount()
	
	// Optimize for large documents (> 1000 lines)
	if lineCount > 1000 {
		m.parseVisibleLines(parser, ctx)
	} else {
		// Parse all lines for smaller documents
		m.parseAllLines(parser, ctx)
	}
}

// parseAllLines parses all lines in the document
func (m *Model) parseAllLines(parser plugin.ParserPlugin, ctx context.Context) {
	doc := m.editor.GetDocument()
	for i := 0; i < doc.LineCount(); i++ {
		line := doc.GetLine(i)
		tokens, err := parser.GetSyntaxHighlighting(ctx, line)
		if err != nil {
			// Skip this line if parsing fails
			continue
		}
		doc.SetLineTokens(i, tokens)
	}
}

// parseVisibleLines parses only the visible lines and a buffer around them
func (m *Model) parseVisibleLines(parser plugin.ParserPlugin, ctx context.Context) {
	doc := m.editor.GetDocument()
	viewport := m.editor.GetViewPort()
	
	// Parse visible lines plus a buffer for smooth scrolling
	bufferSize := 50 // Parse 50 lines above and below visible area
	startLine := viewport.Top - bufferSize
	endLine := viewport.Top + viewport.Height + bufferSize
	
	// Ensure bounds are valid
	if startLine < 0 {
		startLine = 0
	}
	if endLine > doc.LineCount() {
		endLine = doc.LineCount()
	}
	
	// Parse lines in the visible range
	for i := startLine; i < endLine; i++ {
		line := doc.GetLine(i)
		tokens, err := parser.GetSyntaxHighlighting(ctx, line)
		if err != nil {
			// Skip this line if parsing fails
			continue
		}
		doc.SetLineTokens(i, tokens)
	}
}

// renderLinesWithCursor converts rendered lines to display string with cursor
func (m *Model) renderLinesWithCursor(renderedLines []plugin.RenderedLine, themePlugin theme.Theme, renderer plugin.RendererPlugin) string {
	// Convert to plain text for now
	lines := make([]string, len(renderedLines))
	for i, line := range renderedLines {
		lines[i] = line.Content
	}
	
	// Add cursor
	cursorRow, cursorCol := m.editor.GetCursorScreenPosition()
	if cursorRow >= 0 && cursorRow < len(lines) && cursorCol >= 0 {
		line := lines[cursorRow]
		runes := []rune(line)
		
		if cursorCol < len(runes) {
			// Replace character at cursor position
			runes[cursorCol] = '█'
			lines[cursorRow] = string(runes)
		} else if cursorCol == len(runes) {
			// Cursor is at end of line - append cursor without extra spaces
			lines[cursorRow] = line + "█"
		}
		// If cursor is beyond line end, don't add padding spaces
		// This prevents the cascading effect described in the bug report
	}
	
	return strings.Join(lines, "\n")
}

func (m *Model) renderStatusBar() string {
	filename := m.editor.GetDocument().GetFilename()
	if filename == "" {
		filename = "[No Name]"
	}
	if m.editor.GetDocument().IsModified() {
		filename += " [Modified]"
	}
	
	status := filename
	if m.message != "" {
		status = m.message
	}
	
	pos := m.editor.GetCursor().GetPosition()
	position := fmt.Sprintf("Ln %d, Col %d", pos.Line+1, pos.Col+1)
	
	gap := m.width - lipgloss.Width(status) - lipgloss.Width(position)
	if gap < 1 {
		gap = 1
	}
	
	statusBar := lipgloss.NewStyle().
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("252")).
		Width(m.width).
		Render(status + strings.Repeat(" ", gap) + position)
	
	return statusBar
}

func (m *Model) renderHelpBar() string {
	var help string
	switch m.mode {
	case ModeFind:
		help = "Find: " + m.input + " | Enter: Search | Esc: Cancel"
	case ModeReplace:
		help = "Replace: " + m.input + " with: " + m.replaceText + " | Enter: Replace | Esc: Cancel"
	case ModeGoto:
		help = "Goto line: " + m.input + " | Enter: Go | Esc: Cancel"
	default:
		help = "^O Open  ^S Save  ^Q Quit  ^Z Undo  ^Y Redo  ^C Copy  ^V Paste  ^X Cut  ^A Select All  ^L Line Numbers  ^F Find  ^H Replace  ^G Goto  ^P Preview"
	}
	
	helpBar := lipgloss.NewStyle().
		Background(lipgloss.Color("237")).
		Foreground(lipgloss.Color("249")).
		Width(m.width).
		Align(lipgloss.Center).
		Render(help)
	
	return helpBar
}