package tui

import (
	"strconv"
	"strings"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ofri/mde/pkg/plugin"
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
		m.editor.MoveCursorUp()

	case tea.KeyDown:
		m.editor.MoveCursorDown()

	case tea.KeyLeft:
		m.editor.MoveCursorLeft()

	case tea.KeyRight:
		m.editor.MoveCursorRight()

	case tea.KeyShiftUp:
		if !m.editor.GetCursor().HasSelection() {
			m.editor.GetCursor().StartSelection()
		}
		m.editor.MoveCursorUp()
		m.editor.GetCursor().ExtendSelection()

	case tea.KeyShiftDown:
		if !m.editor.GetCursor().HasSelection() {
			m.editor.GetCursor().StartSelection()
		}
		m.editor.MoveCursorDown()
		m.editor.GetCursor().ExtendSelection()

	case tea.KeyShiftLeft:
		if !m.editor.GetCursor().HasSelection() {
			m.editor.GetCursor().StartSelection()
		}
		m.editor.MoveCursorLeft()
		m.editor.GetCursor().ExtendSelection()

	case tea.KeyShiftRight:
		if !m.editor.GetCursor().HasSelection() {
			m.editor.GetCursor().StartSelection()
		}
		m.editor.MoveCursorRight()
		m.editor.GetCursor().ExtendSelection()

	case tea.KeyCtrlA:
		// Select all
		m.editor.GetCursor().StartSelection()
		m.editor.MoveCursorToDocumentStart()
		m.editor.GetCursor().ExtendSelection()
		m.editor.MoveCursorToDocumentEnd()
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
		
	case tea.KeyCtrlT:
		// Toggle theme
		return m.toggleTheme()

	case tea.KeyHome:
		m.editor.MoveCursorToLineStart()

	case tea.KeyEnd:
		m.editor.MoveCursorToLineEnd()

	case tea.KeyBackspace:
		m.editor.DeleteText(1)

	case tea.KeyDelete:
		pos := m.editor.GetCursor().GetBufferPos()
		m.editor.MoveCursorRight()
		m.editor.DeleteText(1)
		m.editor.GetCursor().SetBufferPos(pos)

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
		m.editor.GetCursor().SetBufferPos(*pos)
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
			m.editor.MoveCursorUp()
		}
		
	case tea.MouseButtonWheelDown:
		// Scroll down
		for i := 0; i < 3; i++ {
			m.editor.MoveCursorDown()
		}
	}
	
	return m, nil
}

func (m *Model) handleMouseClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Use safe coordinate transformation
	bufferPos := m.screenToBufferSafe(msg.Y, msg.X)
	
	// Clear any existing selection and move cursor
	m.editor.GetCursor().ClearSelection()
	m.editor.GetCursor().SetBufferPos(bufferPos)
	
	return m, nil
}

func (m *Model) handleMouseDrag(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Use safe coordinate transformation
	bufferPos := m.screenToBufferSafe(msg.Y, msg.X)
	
	// Start selection if not already started
	if !m.editor.GetCursor().HasSelection() {
		m.editor.GetCursor().StartSelection()
	}
	
	// Move cursor to dragged position and extend selection
	m.editor.GetCursor().SetBufferPos(bufferPos)
	m.editor.GetCursor().ExtendSelection()
	
	return m, nil
}

// toggleTheme switches between light and dark themes
func (m *Model) toggleTheme() (tea.Model, tea.Cmd) {
	registry := plugin.GetRegistry()
	
	// Toggle between light and dark themes
	if m.currentTheme == "dark" {
		m.currentTheme = "light"
	} else {
		m.currentTheme = "dark"
	}
	
	// Set the new theme as default
	if err := registry.SetDefaultTheme(m.currentTheme); err != nil {
		m.showMessage("Error switching theme: " + err.Error())
		return m, nil
	}
	
	m.showMessage("Theme switched to " + m.currentTheme)
	return m, nil
}

// syncThemeWithRegistry synchronizes the model's currentTheme with the registry default
func (m *Model) syncThemeWithRegistry() {
	registry := plugin.GetRegistry()
	if defaultTheme, err := registry.GetDefaultTheme(); err == nil {
		m.currentTheme = defaultTheme.Name()
	}
}
