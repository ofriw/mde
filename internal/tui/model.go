package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	content      []string
	cursor       position
	filename     string
	width        int
	height       int
	offset       int
	modified     bool
	message      string
	messageTimer int
	err          error
}

type position struct {
	row int
	col int
}

func New() *Model {
	return &Model{
		content: []string{""},
		cursor:  position{0, 0},
	}
}

func (m *Model) SetFilename(filename string) {
	m.filename = filename
}

func (m *Model) Init() tea.Cmd {
	if m.filename != "" {
		return m.loadFile(m.filename)
	}
	return nil
}

func (m *Model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	var b strings.Builder
	
	editorHeight := m.height - 2
	
	for i := 0; i < editorHeight; i++ {
		lineIdx := i + m.offset
		if lineIdx < len(m.content) {
			line := m.content[lineIdx]
			
			if lineIdx == m.cursor.row {
				if m.cursor.col < len(line) {
					line = line[:m.cursor.col] + "█" + line[m.cursor.col:]
				} else if m.cursor.col == len(line) {
					line = line + "█"
				} else {
					line = line + strings.Repeat(" ", m.cursor.col-len(line)) + "█"
				}
			}
			
			b.WriteString(line)
		}
		
		if i < editorHeight-1 {
			b.WriteString("\n")
		}
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
	filename := m.filename
	if filename == "" {
		filename = "[No Name]"
	}
	if m.modified {
		filename += " [Modified]"
	}
	
	status := filename
	if m.message != "" {
		status = m.message
	}
	
	position := fmt.Sprintf("Ln %d, Col %d", m.cursor.row+1, m.cursor.col+1)
	
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
	help := "^O Open  ^S Save  ^Q Quit"
	
	helpBar := lipgloss.NewStyle().
		Background(lipgloss.Color("237")).
		Foreground(lipgloss.Color("249")).
		Width(m.width).
		Align(lipgloss.Center).
		Render(help)
	
	return helpBar
}