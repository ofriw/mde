package tui

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/ofri/mde/pkg/ast"
	"github.com/ofri/mde/pkg/plugin"
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
	
	// Save prompt context
	savePromptContext string
	
	// Preview mode
	previewMode  bool
	
	// Mouse state tracking
	mouseStartPos *ast.BufferPos // Starting position for drag selection
	isDragging    bool            // Whether we're currently dragging
}

type EditorMode int

const (
	ModeNormal EditorMode = iota
	ModeFind
	ModeReplace
	ModeGoto
	ModeSavePrompt
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
	return tea.RequestKeyReleases
}

// GetContentHeight returns the available height for editor content.
// Terminal height minus UI chrome (status bar + help bar = 2 lines).
func (m *Model) GetContentHeight() int {
	const uiChromeHeight = 2 // status bar (1) + help bar (1)
	contentHeight := m.height - uiChromeHeight
	if contentHeight < 1 {
		contentHeight = 1 // minimum height
	}
	return contentHeight
}

func (m *Model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	// Viewport is now updated only when window size changes
	
	// Render content based on mode
	var content string
	if m.previewMode {
		content = m.renderPreviewContent()
	} else {
		content = m.renderEditorContent()
	}
	
	statusBar := m.renderStatusBar()
	helpBar := m.renderHelpBar()
	
	// No background styling - use terminal's default
	editorStyle := lipgloss.NewStyle().Width(m.width).Height(m.height)
	return editorStyle.Render(lipgloss.JoinVertical(lipgloss.Top, content, statusBar, helpBar))
}

// renderEditorContent renders the editor content with syntax highlighting
// IMPORTANT: This uses the internal plugin system for modularization.
// Plugins are compiled into the binary and cannot fail at runtime unless there's a programming error.
// Any plugin errors indicate a bug and should crash with explicit errors.
func (m *Model) renderEditorContent() string {
	editorHeight := m.GetContentHeight()
	
	// Get renderer plugin - must exist as it's compiled into the binary
	registry := plugin.GetRegistry()
	renderer, err := registry.GetDefaultRenderer()
	if err != nil {
		panic(fmt.Sprintf("FATAL: Failed to get default renderer plugin: %v\nThis is a programming error - renderer plugin must be registered at startup", err))
	}
	
	// Configure renderer to match editor settings
	if err := m.configureRenderer(renderer); err != nil {
		panic(fmt.Sprintf("FATAL: Failed to configure renderer: %v\nThis is a programming error - renderer configuration should never fail", err))
	}
	
	// Create render context with viewport information
	// This ensures we only render what's visible, fixing scrolling issues
	// and improving performance for large documents
	renderCtx := &plugin.RenderContext{
		Document:        m.editor.GetDocument(),
		Viewport:        m.editor.GetViewport(),
		ShowLineNumbers: m.editor.ShowLineNumbers(),
	}
	
	// Render only the visible portion of the document
	// This is the key fix for scrolling - we now respect the viewport boundaries
	ctx := context.Background()
	renderedLines, err := renderer.RenderVisible(ctx, renderCtx)
	if err != nil {
		panic(fmt.Sprintf("FATAL: Renderer failed to render visible content: %v\nThis is a programming error - internal renderer should never fail", err))
	}
	
	// Convert rendered lines to string and add cursor
	content := m.renderLinesWithCursor(renderedLines, renderer)
	
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
	
	// No background styling - use terminal's default
	editorStyle := lipgloss.NewStyle().Width(m.width).Height(editorHeight)
	return editorStyle.Render(result)
}


// renderPreviewContent renders the markdown content in preview mode
// Uses the internal plugin system for consistent rendering
func (m *Model) renderPreviewContent() string {
	editorHeight := m.GetContentHeight()
	
	// Get renderer plugin - must exist as it's compiled into the binary
	registry := plugin.GetRegistry()
	renderer, err := registry.GetDefaultRenderer()
	if err != nil {
		panic(fmt.Sprintf("FATAL: Failed to get default renderer plugin: %v\nThis is a programming error - renderer plugin must be registered at startup", err))
	}
	
	// Configure renderer to match editor settings
	if err := m.configureRenderer(renderer); err != nil {
		panic(fmt.Sprintf("FATAL: Failed to configure renderer: %v\nThis is a programming error - renderer configuration should never fail", err))
	}
	
	// Create render context for preview mode
	// Preview mode doesn't show line numbers but still respects viewport boundaries
	renderCtx := &plugin.RenderContext{
		Document:        m.editor.GetDocument(),
		Viewport:        m.editor.GetViewport(),
		ShowLineNumbers: false, // Preview mode never shows line numbers
	}
	
	// Render only the visible portion of the document in preview mode
	// This fixes scrolling in preview mode and improves performance
	ctx := context.Background()
	renderedLines, err := renderer.RenderPreviewVisible(ctx, renderCtx)
	if err != nil {
		panic(fmt.Sprintf("FATAL: Renderer failed to render preview content: %v\nThis is a programming error - internal renderer should never fail", err))
	}
	
	// The renderer now returns only visible lines, so we can use them directly
	// Pad to fill editor height if needed
	for len(renderedLines) < editorHeight {
		renderedLines = append(renderedLines, plugin.RenderedLine{Content: "", Styles: nil})
	}
	
	// Trim to exact height (safety measure)
	if len(renderedLines) > editorHeight {
		renderedLines = renderedLines[:editorHeight]
	}
	
	// Use the renderer's RenderToString method to convert to terminal output
	// The renderer MUST be a TerminalRenderer as it's the only implementation
	terminalRenderer, ok := renderer.(*renderers.TerminalRenderer)
	if !ok {
		panic(fmt.Sprintf("FATAL: Renderer is not a TerminalRenderer: got %T\nThis is a programming error - only TerminalRenderer is supported", renderer))
	}
	content := terminalRenderer.RenderToString(renderedLines)
	
	// No background styling - use terminal's default
	editorStyle := lipgloss.NewStyle().Width(m.width).Height(editorHeight)
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
		panic(fmt.Sprintf("FATAL: Failed to convert markdown to HTML: %v\nThis is a programming error - goldmark should never fail", err))
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
		panic(fmt.Sprintf("FATAL: Failed to get default parser plugin: %v\nThis is a programming error - parser plugin must be registered at startup", err))
	}
	
	ctx := context.Background()
	_, err = parser.Parse(ctx, m.editor.GetDocument().GetText())
	if err != nil {
		panic(fmt.Sprintf("FATAL: Parser failed to parse document: %v\nThis is a programming error - internal parser should never fail", err))
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
			panic(fmt.Sprintf("FATAL: Parser failed to get syntax highlighting for line %d: %v\nThis is a programming error - internal parser should never fail", i, err))
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
			panic(fmt.Sprintf("FATAL: Parser failed to get syntax highlighting for line %d: %v\nThis is a programming error - internal parser should never fail", i, err))
		}
		doc.SetLineTokens(i, tokens)
	}
}

// renderLinesWithCursor converts rendered lines to display string with cursor
func (m *Model) renderLinesWithCursor(renderedLines []plugin.RenderedLine, renderer plugin.RendererPlugin) string {
	// The renderer MUST be a TerminalRenderer as it's the only implementation
	terminalRenderer, ok := renderer.(*renderers.TerminalRenderer)
	if !ok {
		panic(fmt.Sprintf("FATAL: Renderer is not a TerminalRenderer: got %T\nThis is a programming error - only TerminalRenderer is supported", renderer))
	}
	
	// Get cursor position and viewport for calculation
	cursorPos := m.editor.GetCursor().GetBufferPos()
	viewport := m.editor.GetViewport()
	
	// Check if cursor is within the visible viewport
	// This is critical because we only render visible lines now
	if cursorPos.Line < viewport.GetTopLine() || 
	   cursorPos.Line >= viewport.GetTopLine() + viewport.GetHeight() {
		// Cursor is outside the visible area - render without cursor
		lines := make([]string, len(renderedLines))
		for i, line := range renderedLines {
			lines[i] = line.Content
		}
		return strings.Join(lines, "\n")
	}
	
	// Calculate cursor position relative to the rendered lines
	// Since we only render visible lines, the cursor row is relative to viewport top
	cursorRow := cursorPos.Line - viewport.GetTopLine()
	cursorCol := cursorPos.Col
	
	// The renderer has already applied horizontal scrolling and line numbers
	// So we need to use the screen position's column which accounts for these
	screenPos, err := m.editor.GetCursor().GetScreenPos()
	if err == nil {
		// Use the properly calculated screen column that accounts for line numbers
		cursorCol = screenPos.Col
	}
	
	// Render with cursor at the viewport-relative position
	return terminalRenderer.RenderToStringWithCursor(renderedLines, cursorRow, cursorCol)
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
	
	// Status bar style - use inverse colors (terminal will handle this)
	statusBarStyle := lipgloss.NewStyle().
		Reverse(true).
		Width(m.width)
	
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
	case ModeSavePrompt:
		filename := m.editor.GetDocument().GetFilename()
		help = fmt.Sprintf("Save changes to %s? (y/n/c)", filename)
	default:
		help = "^O Open  ^S Save  ^Q Quit  ^C Copy  ^V Paste  ^X Cut  ^A Select All  ^L Line Numbers  ^F Find  ^H Replace  ^G Goto  ^P Preview"
	}
	
	// Help bar style - use reverse for background like status bar
	helpBarStyle := lipgloss.NewStyle().
		Reverse(true).
		Width(m.width).
		Align(lipgloss.Center)
	
	helpBar := helpBarStyle.Render(help)
	
	return helpBar
}

// screenToBufferSafe converts screen coordinates to buffer coordinates with safe bounds checking.
// Handles editor area bounds and uses document validation for final position safety.
// USAGE: bufferPos := m.screenToBufferSafe(mouseRow, mouseCol)
func (m *Model) screenToBufferSafe(row, col int) ast.BufferPos {
	// Check if position is within editor area (exclude status and help bars)
	editorHeight := m.GetContentHeight()
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