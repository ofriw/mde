package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ofri/mde/pkg/ast"
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
	}
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
	
	// Get visible lines
	lines := m.editor.GetVisibleLines()
	
	// Add selection highlighting and cursor
	cursorRow, cursorCol := m.editor.GetCursorScreenPosition()
	
	// TODO: Add selection highlighting here
	// For now, just show cursor
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
	
	// Join lines
	var b strings.Builder
	for i, line := range lines {
		b.WriteString(line)
		if i < len(lines)-1 {
			b.WriteString("\n")
		}
	}
	
	// Pad remaining lines
	editorHeight := m.height - 2
	for i := len(lines); i < editorHeight; i++ {
		b.WriteString("\n")
	}
	
	editor := lipgloss.NewStyle().
		Width(m.width).
		Height(editorHeight).
		Render(b.String())
	
	statusBar := m.renderStatusBar()
	helpBar := m.renderHelpBar()
	
	return lipgloss.JoinVertical(lipgloss.Top, editor, statusBar, helpBar)
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