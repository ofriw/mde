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
	case tea.KeyCtrlC:
		if m.editor.GetCursor().HasSelection() {
			m.editor.Copy()
			m.showMessage("Copied")
		} else {
			return m, tea.Quit
		}

	case tea.KeyCtrlQ:
		return m, tea.Quit

	case tea.KeyCtrlS:
		return m, m.saveFile()

	case tea.KeyCtrlO:
		return m, m.openFile()

	case tea.KeyCtrlZ:
		if m.editor.CanUndo() {
			m.editor.Undo()
			m.showMessage("Undone")
		}

	case tea.KeyCtrlY:
		if m.editor.CanRedo() {
			m.editor.Redo()
			m.showMessage("Redone")
		}

	case tea.KeyCtrlV:
		m.editor.Paste()

	case tea.KeyCtrlX:
		if m.editor.GetCursor().HasSelection() {
			m.editor.Cut()
			m.showMessage("Cut")
		}

	case tea.KeyUp:
		m.editor.GetCursor().MoveUp()

	case tea.KeyDown:
		m.editor.GetCursor().MoveDown()

	case tea.KeyLeft:
		m.editor.GetCursor().MoveLeft()

	case tea.KeyRight:
		m.editor.GetCursor().MoveRight()

	case tea.KeyShiftUp:
		if !m.editor.GetCursor().HasSelection() {
			m.editor.GetCursor().StartSelection()
		}
		m.editor.GetCursor().MoveUp()
		m.editor.GetCursor().ExtendSelection()

	case tea.KeyShiftDown:
		if !m.editor.GetCursor().HasSelection() {
			m.editor.GetCursor().StartSelection()
		}
		m.editor.GetCursor().MoveDown()
		m.editor.GetCursor().ExtendSelection()

	case tea.KeyShiftLeft:
		if !m.editor.GetCursor().HasSelection() {
			m.editor.GetCursor().StartSelection()
		}
		m.editor.GetCursor().MoveLeft()
		m.editor.GetCursor().ExtendSelection()

	case tea.KeyShiftRight:
		if !m.editor.GetCursor().HasSelection() {
			m.editor.GetCursor().StartSelection()
		}
		m.editor.GetCursor().MoveRight()
		m.editor.GetCursor().ExtendSelection()

	case tea.KeyCtrlA:
		// Select all
		m.editor.GetCursor().StartSelection()
		m.editor.GetCursor().MoveToDocumentStart()
		m.editor.GetCursor().ExtendSelection()
		m.editor.GetCursor().MoveToDocumentEnd()
		m.editor.GetCursor().ExtendSelection()

	case tea.KeyEscape:
		// Clear selection
		m.editor.GetCursor().ClearSelection()

	case tea.KeyCtrlL:
		// Toggle line numbers
		m.editor.ToggleLineNumbers()
		if m.editor.ShowLineNumbers() {
			m.showMessage("Line numbers enabled")
		} else {
			m.showMessage("Line numbers disabled")
		}

	case tea.KeyHome:
		m.editor.GetCursor().MoveToLineStart()

	case tea.KeyEnd:
		m.editor.GetCursor().MoveToLineEnd()

	case tea.KeyBackspace:
		m.editor.DeleteText(1)

	case tea.KeyDelete:
		pos := m.editor.GetCursor().GetPosition()
		m.editor.GetCursor().MoveRight()
		m.editor.DeleteText(1)
		m.editor.GetCursor().SetPosition(pos)

	case tea.KeyEnter:
		m.editor.InsertText("\n")

	case tea.KeyRunes:
		m.editor.InsertText(msg.String())
	}

	return m, nil
}

func (m *Model) showMessage(msg string) {
	m.message = msg
	m.messageTimer = 60 // Show for ~1 second at 60fps
}