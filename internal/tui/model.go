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
	themepkg "github.com/ofri/mde/pkg/theme"
	"github.com/ofri/mde/internal/plugins/renderers"
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
	
	// Theme
	currentTheme string
}

type EditorMode int

const (
	ModeNormal EditorMode = iota
	ModeFind
	ModeReplace
	ModeGoto
)

func New() *Model {
	m := &Model{
		editor: ast.NewEditor(),
		currentTheme: "dark", // Default theme
	}
	
	// Synchronize with registry default theme
	m.syncThemeWithRegistry()
	
	return m
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
	
	// Apply editor background theme
	registry := plugin.GetRegistry()
	if themePlugin, err := registry.GetDefaultTheme(); err == nil {
		editorBgStyle := themePlugin.GetStyle(themepkg.EditorBackground)
		editorStyle := editorBgStyle.ToLipgloss().Width(m.width).Height(m.height)
		return editorStyle.Render(lipgloss.JoinVertical(lipgloss.Top, content, statusBar, helpBar))
	}
	
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
	
	// BUG FIX: Configure renderer to match editor settings
	// This ensures consistent coordinate system between editor and renderer
	if err := m.configureRenderer(renderer); err != nil {
		// If configuration fails, fall back to simple rendering
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
	
	// Apply editor background theme
	editorStyle := lipgloss.NewStyle().Width(m.width).Height(editorHeight)
	if registry := plugin.GetRegistry(); registry != nil {
		if currentTheme, err := registry.GetDefaultTheme(); err == nil {
			editorBgStyle := currentTheme.GetStyle(themepkg.EditorBackground)
			editorStyle = editorBgStyle.ToLipgloss().Width(m.width).Height(editorHeight)
		}
	}
	
	return editorStyle.Render(result)
}

// renderSimple provides fallback rendering without plugins
func (m *Model) renderSimple(editorHeight int) string {
	// Get raw document lines without line numbers for cursor positioning
	viewport := m.editor.GetViewport()
	rawLines := make([]string, 0, viewport.GetHeight())
	
	for i := 0; i < viewport.GetHeight(); i++ {
		lineNum := viewport.GetTopLine() + i
		if lineNum >= m.editor.GetDocument().LineCount() {
			break
		}
		rawLines = append(rawLines, m.editor.GetDocument().GetLine(lineNum))
	}
	
	// Add cursor to raw lines first
	cursorPos := m.editor.GetCursor().GetBufferPos()
	cursorRow := cursorPos.Line - viewport.GetTopLine()
	cursorCol := cursorPos.Col - viewport.GetLeftColumn()
	
	if cursorRow >= 0 && cursorRow < len(rawLines) && cursorCol >= 0 {
		line := rawLines[cursorRow]
		runes := []rune(line)
		
		if cursorCol < len(runes) {
			// Cursor is within the line - replace character
			runes[cursorCol] = '█'
			rawLines[cursorRow] = string(runes)
		} else if cursorCol == len(runes) {
			// Cursor is at end of line - append cursor
			rawLines[cursorRow] = line + "█"
		}
		// Note: If cursorCol > len(runes), cursor is beyond line end - don't render
	}
	
	// Now add line numbers if enabled
	lines := make([]string, len(rawLines))
	for i, line := range rawLines {
		if m.editor.ShowLineNumbers() {
			lineNum := viewport.GetTopLine() + i
			lineNumStr := m.editor.FormatLineNumber(lineNum + 1)
			lines[i] = lineNumStr + line
		} else {
			lines[i] = line
		}
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
	
	// Apply editor background theme for fallback mode
	editorStyle := lipgloss.NewStyle().Width(m.width).Height(editorHeight)
	if registry := plugin.GetRegistry(); registry != nil {
		if currentTheme, err := registry.GetDefaultTheme(); err == nil {
			editorBgStyle := currentTheme.GetStyle(themepkg.EditorBackground)
			editorStyle = editorBgStyle.ToLipgloss().Width(m.width).Height(editorHeight)
		}
	}
	
	return editorStyle.Render(b.String())
}

// renderPreviewContent renders the markdown content as HTML preview
func (m *Model) renderPreviewContent() string {
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
	
	// BUG FIX: Configure renderer to match editor settings
	// This ensures consistent coordinate system between editor and renderer
	if err := m.configureRenderer(renderer); err != nil {
		// If configuration fails, fall back to simple rendering
		return m.renderSimple(editorHeight)
	}
	
	themePlugin, err := registry.GetDefaultTheme()
	if err != nil {
		// Fallback to simple rendering
		return m.renderSimple(editorHeight)
	}
	
	// Render the document using plugins in preview mode
	ctx := context.Background()
	renderedLines, err := renderer.RenderPreview(ctx, m.editor.GetDocument(), themePlugin)
	if err != nil {
		// Fallback to simple rendering
		return m.renderSimple(editorHeight)
	}
	
	// Apply viewport
	viewport := m.editor.GetViewport()
	startLine := viewport.GetTopLine()
	endLine := startLine + editorHeight
	
	if startLine >= len(renderedLines) {
		startLine = len(renderedLines) - 1
		if startLine < 0 {
			startLine = 0
		}
	}
	
	if endLine > len(renderedLines) {
		endLine = len(renderedLines)
	}
	
	// Get visible lines
	var visibleLines []plugin.RenderedLine
	if startLine < len(renderedLines) {
		visibleLines = renderedLines[startLine:endLine]
	}
	
	// Pad to fill editor height
	for len(visibleLines) < editorHeight {
		visibleLines = append(visibleLines, plugin.RenderedLine{Content: "", Styles: nil})
	}
	
	// Trim to exact height
	if len(visibleLines) > editorHeight {
		visibleLines = visibleLines[:editorHeight]
	}
	
	// Use the renderer's RenderToString method to apply theme styles
	var content string
	if terminalRenderer, ok := renderer.(*renderers.TerminalRenderer); ok {
		content = terminalRenderer.RenderToString(visibleLines, themePlugin)
	} else {
		// Fallback to plain text
		lines := make([]string, len(visibleLines))
		for i, line := range visibleLines {
			lines[i] = line.Content
		}
		content = strings.Join(lines, "\n")
	}
	
	// Apply editor background theme
	editorStyle := lipgloss.NewStyle().Width(m.width).Height(editorHeight)
	if themePlugin != nil {
		editorBgStyle := themePlugin.GetStyle(themepkg.EditorBackground)
		editorStyle = editorBgStyle.ToLipgloss().Width(m.width).Height(editorHeight)
	}
	
	return editorStyle.Render(content)
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
	viewport := m.editor.GetViewport()
	
	// Calculate visible range with buffer for context
	bufferSize := 100 // Process 100 lines above and below for context
	startLine := viewport.GetTopLine() - bufferSize
	endLine := viewport.GetTopLine() + viewport.GetHeight() + bufferSize
	
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
	viewport := m.editor.GetViewport()
	
	// Parse visible lines plus a buffer for smooth scrolling
	bufferSize := 50 // Parse 50 lines above and below visible area
	startLine := viewport.GetTopLine() - bufferSize
	endLine := viewport.GetTopLine() + viewport.GetHeight() + bufferSize
	
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
func (m *Model) renderLinesWithCursor(renderedLines []plugin.RenderedLine, themePlugin themepkg.Theme, renderer plugin.RendererPlugin) string {
	// Use the renderer's RenderToStringWithCursor method to apply cursor during styling
	if terminalRenderer, ok := renderer.(*renderers.TerminalRenderer); ok {
		// NOTE: Renderer is already configured in the main rendering path
		// so we don't need to configure it again here
		
		// NEW COORDINATE SYSTEM: Use ScreenPos from CursorManager
		screenPos, err := m.editor.GetCursor().GetScreenPos()
		if err != nil {
			// Cursor not visible, render without cursor
			lines := make([]string, len(renderedLines))
			for i, line := range renderedLines {
				lines[i] = line.Content
			}
			return strings.Join(lines, "\n")
		}
		
		return terminalRenderer.RenderToStringWithCursor(renderedLines, themePlugin, screenPos.Row, screenPos.Col)
	} else {
		// Fallback to plain text without cursor
		lines := make([]string, len(renderedLines))
		for i, line := range renderedLines {
			lines[i] = line.Content
		}
		return strings.Join(lines, "\n")
	}
}

// configureRenderer synchronizes the renderer configuration with the editor's settings.
// 
// BUG FIX: This function addresses the cursor positioning bug where the renderer's
// line numbers configuration becomes desynchronized from the editor's line numbers setting.
// 
// ROOT CAUSE: The TUI gets a fresh renderer from the registry but never configures it
// to match the editor's settings, causing coordinate transformation failures.
//
// SOLUTION: Ensure renderer.config.ShowLineNumbers matches editor.ShowLineNumbers()
// before using the renderer for cursor positioning.
func (m *Model) configureRenderer(renderer plugin.RendererPlugin) error {
	// Synchronize renderer configuration with editor settings
	config := map[string]interface{}{
		"showLineNumbers":  m.editor.ShowLineNumbers(),
		"lineNumberWidth": m.editor.GetLineNumberWidth(),
	}
	
	// Configure the renderer to match editor settings
	return renderer.Configure(config)
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
	
	pos := m.editor.GetCursor().GetBufferPos()
	position := fmt.Sprintf("Ln %d, Col %d", pos.Line+1, pos.Col+1)
	
	gap := m.width - lipgloss.Width(status) - lipgloss.Width(position)
	if gap < 1 {
		gap = 1
	}
	
	// Get theme style for status bar
	registry := plugin.GetRegistry()
	statusBarStyle := lipgloss.NewStyle().Width(m.width)
	if themePlugin, err := registry.GetDefaultTheme(); err == nil {
		themeStyle := themePlugin.GetStyle(themepkg.UIStatusBar)
		statusBarStyle = themeStyle.ToLipgloss().Width(m.width)
	}
	
	statusBar := statusBarStyle.Render(status + strings.Repeat(" ", gap) + position)
	
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
		help = "^O Open  ^S Save  ^Q Quit  ^Z Undo  ^Y Redo  ^C Copy  ^V Paste  ^X Cut  ^A Select All  ^L Line Numbers  ^F Find  ^H Replace  ^G Goto  ^P Preview  ^T Theme"
	}
	
	// Get theme style for help bar
	registry := plugin.GetRegistry()
	helpBarStyle := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center)
	if themePlugin, err := registry.GetDefaultTheme(); err == nil {
		themeStyle := themePlugin.GetStyle(themepkg.UIHelpBar)
		helpBarStyle = themeStyle.ToLipgloss().Width(m.width).Align(lipgloss.Center)
	}
	
	helpBar := helpBarStyle.Render(help)
	
	return helpBar
}

// screenToBufferSafe converts screen coordinates to buffer coordinates with safe bounds checking.
// Handles editor area bounds and uses document validation for final position safety.
// USAGE: bufferPos := m.screenToBufferSafe(mouseRow, mouseCol)
func (m *Model) screenToBufferSafe(row, col int) ast.BufferPos {
	// Check if position is within editor area (exclude status and help bars)
	editorHeight := m.height - 2
	if row >= editorHeight {
		row = editorHeight - 1
	}
	if row < 0 {
		row = 0
	}
	
	// Use viewport's safe transformation
	screenPos := ast.ScreenPos{Row: row, Col: col}
	bufferPos := m.editor.GetViewport().ScreenToBuffer(screenPos)
	
	// Apply document bounds validation using existing ValidatePosition
	return m.editor.GetDocument().ValidatePosition(bufferPos)
}