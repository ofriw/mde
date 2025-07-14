package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.messageTimer > 0 {
		m.messageTimer--
		if m.messageTimer == 0 {
			m.message = ""
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.adjustViewport()
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
		
	case fileLoadedMsg, fileSavedMsg, fileOpenPromptMsg:
		return m.handleFileMsg(msg)
	}

	return m, nil
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyCtrlQ:
		return m, tea.Quit

	case tea.KeyCtrlS:
		return m, m.saveFile()

	case tea.KeyCtrlO:
		return m, m.openFile()

	case tea.KeyUp:
		m.moveCursorUp()

	case tea.KeyDown:
		m.moveCursorDown()

	case tea.KeyLeft:
		m.moveCursorLeft()

	case tea.KeyRight:
		m.moveCursorRight()

	case tea.KeyHome:
		m.cursor.col = 0

	case tea.KeyEnd:
		if m.cursor.row < len(m.content) {
			m.cursor.col = len(m.content[m.cursor.row])
		}

	case tea.KeyBackspace:
		m.handleBackspace()

	case tea.KeyDelete:
		m.handleDelete()

	case tea.KeyEnter:
		m.handleEnter()

	case tea.KeyRunes:
		m.insertText(msg.String())
	}

	m.adjustViewport()
	return m, nil
}

func (m *Model) moveCursorUp() {
	if m.cursor.row > 0 {
		m.cursor.row--
		if m.cursor.col > len(m.content[m.cursor.row]) {
			m.cursor.col = len(m.content[m.cursor.row])
		}
	}
}

func (m *Model) moveCursorDown() {
	if m.cursor.row < len(m.content)-1 {
		m.cursor.row++
		if m.cursor.col > len(m.content[m.cursor.row]) {
			m.cursor.col = len(m.content[m.cursor.row])
		}
	}
}

func (m *Model) moveCursorLeft() {
	if m.cursor.col > 0 {
		m.cursor.col--
	} else if m.cursor.row > 0 {
		m.cursor.row--
		m.cursor.col = len(m.content[m.cursor.row])
	}
}

func (m *Model) moveCursorRight() {
	if m.cursor.row < len(m.content) && m.cursor.col < len(m.content[m.cursor.row]) {
		m.cursor.col++
	} else if m.cursor.row < len(m.content)-1 {
		m.cursor.row++
		m.cursor.col = 0
	}
}

func (m *Model) handleBackspace() {
	if m.cursor.col > 0 {
		line := m.content[m.cursor.row]
		m.content[m.cursor.row] = line[:m.cursor.col-1] + line[m.cursor.col:]
		m.cursor.col--
		m.modified = true
	} else if m.cursor.row > 0 {
		prevLine := m.content[m.cursor.row-1]
		currLine := m.content[m.cursor.row]
		m.cursor.col = len(prevLine)
		m.content[m.cursor.row-1] = prevLine + currLine
		m.content = append(m.content[:m.cursor.row], m.content[m.cursor.row+1:]...)
		m.cursor.row--
		m.modified = true
	}
}

func (m *Model) handleDelete() {
	if m.cursor.row < len(m.content) && m.cursor.col < len(m.content[m.cursor.row]) {
		line := m.content[m.cursor.row]
		m.content[m.cursor.row] = line[:m.cursor.col] + line[m.cursor.col+1:]
		m.modified = true
	} else if m.cursor.row < len(m.content)-1 {
		currLine := m.content[m.cursor.row]
		nextLine := m.content[m.cursor.row+1]
		m.content[m.cursor.row] = currLine + nextLine
		m.content = append(m.content[:m.cursor.row+1], m.content[m.cursor.row+2:]...)
		m.modified = true
	}
}

func (m *Model) handleEnter() {
	if m.cursor.row >= len(m.content) {
		return
	}
	
	line := m.content[m.cursor.row]
	before := line[:m.cursor.col]
	after := line[m.cursor.col:]
	
	m.content[m.cursor.row] = before
	newContent := make([]string, 0, len(m.content)+1)
	newContent = append(newContent, m.content[:m.cursor.row+1]...)
	newContent = append(newContent, after)
	newContent = append(newContent, m.content[m.cursor.row+1:]...)
	m.content = newContent
	
	m.cursor.row++
	m.cursor.col = 0
	m.modified = true
}

func (m *Model) insertText(text string) {
	if m.cursor.row >= len(m.content) {
		return
	}
	
	line := m.content[m.cursor.row]
	before := line[:m.cursor.col]
	after := line[m.cursor.col:]
	
	m.content[m.cursor.row] = before + text + after
	m.cursor.col += len(text)
	m.modified = true
}

func (m *Model) adjustViewport() {
	editorHeight := m.height - 2
	
	if m.cursor.row < m.offset {
		m.offset = m.cursor.row
	} else if m.cursor.row >= m.offset+editorHeight {
		m.offset = m.cursor.row - editorHeight + 1
	}
}