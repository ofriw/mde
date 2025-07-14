package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
}

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

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	// Update editor viewport
	m.editor.SetViewPort(m.width, m.height-2)
	m.editor.AdjustViewPort()
	
	// Render with syntax highlighting
	content := m.renderEditorContent()
	
	statusBar := m.renderStatusBar()
	helpBar := m.renderHelpBar()
	
	return lipgloss.JoinVertical(lipgloss.Top, content, statusBar, helpBar)
}

// renderEditorContent renders the editor content with syntax highlighting
func (m *Model) renderEditorContent() string {
	editorHeight := m.height - 2
	
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
	
	// Apply syntax highlighting line by line
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
			runes[cursorCol] = '█'
		} else {
			runes = append(runes, []rune(strings.Repeat(" ", cursorCol-len(runes)))...)
			runes = append(runes, '█')
		}
		
		lines[cursorRow] = string(runes)
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
	help := "^O Open  ^S Save  ^Q Quit  ^Z Undo  ^Y Redo  ^C Copy  ^V Paste  ^X Cut  ^A Select All  ^L Line Numbers"
	
	helpBar := lipgloss.NewStyle().
		Background(lipgloss.Color("237")).
		Foreground(lipgloss.Color("249")).
		Width(m.width).
		Align(lipgloss.Center).
		Render(help)
	
	return helpBar
}