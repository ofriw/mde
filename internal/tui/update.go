package tui

import (
	"strconv"
	"strings"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ofri/mde/pkg/ast"
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
		
	case tea.MouseMsg:
		return m.handleMouseEvent(msg)
		
	case fileLoadedMsg, fileSavedMsg, fileOpenPromptMsg:
		return m.handleFileMsg(msg)
	}

	return m, nil
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle modal states first
	if m.mode != ModeNormal {
		return m.handleModalKeyPress(msg)
	}
	
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
		
	case tea.KeyCtrlF:
		// Enter find mode
		m.mode = ModeFind
		m.input = ""
		m.caseSensitive = false
		
	case tea.KeyCtrlH:
		// Enter replace mode
		m.mode = ModeReplace
		m.input = ""
		m.replaceText = ""
		m.caseSensitive = false
		
	case tea.KeyCtrlG:
		// Enter goto mode
		m.mode = ModeGoto
		m.input = ""
		
	case tea.KeyCtrlP:
		// Toggle preview mode
		m.previewMode = !m.previewMode
		if m.previewMode {
			m.showMessage("Preview mode enabled")
		} else {
			m.showMessage("Preview mode disabled")
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

	case tea.KeySpace:
		m.editor.InsertText(" ")

	case tea.KeyTab:
		m.editor.InsertText("\t")

	case tea.KeyRunes:
		m.editor.InsertText(msg.String())
	}

	return m, nil
}

func (m *Model) showMessage(msg string) {
	m.message = msg
	m.messageTimer = 60 // Show for ~1 second at 60fps
}
func (m *Model) handleModalKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		// Exit modal mode
		m.mode = ModeNormal
		m.input = ""
		m.replaceText = ""
		return m, nil
		
	case tea.KeyEnter:
		switch m.mode {
		case ModeFind:
			return m.handleFind()
		case ModeReplace:
			return m.handleReplace()
		case ModeGoto:
			return m.handleGoto()
		}
		return m, nil
		
	case tea.KeyBackspace:
		// Remove last character from input
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
		return m, nil
		
	case tea.KeySpace:
		// Add space to input
		m.input += " "
		return m, nil
		
	case tea.KeyRunes:
		// Add character to input
		m.input += msg.String()
		return m, nil
	}
	
	return m, nil
}

func (m *Model) handleFind() (tea.Model, tea.Cmd) {
	if m.input == "" {
		m.showMessage("Nothing to search for")
		m.mode = ModeNormal
		return m, nil
	}
	
	pos := m.editor.FindText(m.input, m.caseSensitive)
	if pos == nil {
		m.showMessage("Not found: " + m.input)
	} else {
		m.editor.GetCursor().SetPosition(*pos)
		m.showMessage("Found: " + m.input)
	}
	
	m.mode = ModeNormal
	m.input = ""
	return m, nil
}

func (m *Model) handleReplace() (tea.Model, tea.Cmd) {
	if m.input == "" {
		m.showMessage("Nothing to replace")
		m.mode = ModeNormal
		return m, nil
	}
	
	success := m.editor.ReplaceText(m.input, m.replaceText, m.caseSensitive)
	if success {
		m.showMessage("Replaced: " + m.input + " with: " + m.replaceText)
	} else {
		m.showMessage("No match found at cursor")
	}
	
	m.mode = ModeNormal
	m.input = ""
	m.replaceText = ""
	return m, nil
}

func (m *Model) handleGoto() (tea.Model, tea.Cmd) {
	if m.input == "" {
		m.showMessage("Enter a line number")
		m.mode = ModeNormal
		return m, nil
	}
	
	lineNum, err := strconv.Atoi(strings.TrimSpace(m.input))
	if err != nil {
		m.showMessage("Invalid line number: " + m.input)
	} else {
		m.editor.GotoLine(lineNum)
		m.showMessage("Jumped to line " + m.input)
	}
	
	m.mode = ModeNormal
	m.input = ""
	return m, nil
}

func (m *Model) handleMouseEvent(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Only handle mouse events in normal mode
	if m.mode != ModeNormal {
		return m, nil
	}
	
	switch msg.Button {
	case tea.MouseButtonLeft:
		if msg.Action == tea.MouseActionPress {
			// Handle left click - position cursor
			return m.handleMouseClick(msg)
		} else if msg.Action == tea.MouseActionMotion {
			// Handle mouse drag for selection
			return m.handleMouseDrag(msg)
		}
		
	case tea.MouseButtonWheelUp:
		// Scroll up
		for i := 0; i < 3; i++ {
			m.editor.GetCursor().MoveUp()
		}
		
	case tea.MouseButtonWheelDown:
		// Scroll down
		for i := 0; i < 3; i++ {
			m.editor.GetCursor().MoveDown()
		}
	}
	
	return m, nil
}

func (m *Model) handleMouseClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Convert screen coordinates to editor coordinates
	clickRow := msg.Y
	clickCol := msg.X
	
	// Check if click is in editor area (not status/help bars)
	editorHeight := m.height - 2
	if clickRow >= editorHeight {
		// Click is in status or help bar, ignore
		return m, nil
	}
	
	// Convert screen position to document position
	viewport := m.editor.GetViewPort()
	docRow := viewport.Top + clickRow
	docCol := clickCol
	
	// Account for line numbers
	if m.editor.ShowLineNumbers() {
		if clickCol < 6 {
			// Click is in line number area, move to start of line
			docCol = 0
		} else {
			docCol = clickCol - 6
		}
	}
	
	// Adjust for viewport
	docCol += viewport.Left
	
	// Ensure coordinates are within document bounds
	if docRow >= m.editor.GetDocument().LineCount() {
		docRow = m.editor.GetDocument().LineCount() - 1
	}
	
	if docRow < 0 {
		docRow = 0
	}
	
	lineLength := len(m.editor.GetDocument().GetLine(docRow))
	if docCol > lineLength {
		docCol = lineLength
	}
	
	if docCol < 0 {
		docCol = 0
	}
	
	// Clear any existing selection
	m.editor.GetCursor().ClearSelection()
	
	// Move cursor to clicked position
	newPos := ast.Position{Line: docRow, Col: docCol}
	m.editor.GetCursor().SetPosition(newPos)
	
	return m, nil
}

func (m *Model) handleMouseDrag(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Convert screen coordinates to editor coordinates
	dragRow := msg.Y
	dragCol := msg.X
	
	// Check if drag is in editor area
	editorHeight := m.height - 2
	if dragRow >= editorHeight {
		return m, nil
	}
	
	// Convert screen position to document position
	viewport := m.editor.GetViewPort()
	docRow := viewport.Top + dragRow
	docCol := dragCol
	
	// Account for line numbers
	if m.editor.ShowLineNumbers() {
		if dragCol < 6 {
			docCol = 0
		} else {
			docCol = dragCol - 6
		}
	}
	
	// Adjust for viewport
	docCol += viewport.Left
	
	// Ensure coordinates are within document bounds
	if docRow >= m.editor.GetDocument().LineCount() {
		docRow = m.editor.GetDocument().LineCount() - 1
	}
	
	if docRow < 0 {
		docRow = 0
	}
	
	lineLength := len(m.editor.GetDocument().GetLine(docRow))
	if docCol > lineLength {
		docCol = lineLength
	}
	
	if docCol < 0 {
		docCol = 0
	}
	
	// Start selection if not already started
	if !m.editor.GetCursor().HasSelection() {
		m.editor.GetCursor().StartSelection()
	}
	
	// Move cursor to dragged position and extend selection
	newPos := ast.Position{Line: docRow, Col: docCol}
	m.editor.GetCursor().SetPosition(newPos)
	m.editor.GetCursor().ExtendSelection()
	
	return m, nil
}
