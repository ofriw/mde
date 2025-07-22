package tui

import (
	"strconv"
	"strings"
	"unicode"
	
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/ofri/mde/pkg/terminal"
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
		
		// Update editor viewport with content height (terminal height - UI chrome)
		if m.editor != nil {
			m.editor.SetViewPort(msg.Width, m.GetContentHeight())
		}
		
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKeyInput(msg)
		
	case tea.KeyboardEnhancementsMsg:
		return m, nil
		
	case tea.MouseClickMsg:
		return m.handleMouseClick(msg)
		
	case tea.MouseReleaseMsg:
		return m.handleMouseRelease(msg)
		
	case tea.MouseMotionMsg:
		return m.handleMouseMotion(msg)
		
	case tea.MouseWheelMsg:
		return m.handleMouseWheel(msg)
		
	case fileLoadedMsg, fileSavedMsg, fileOpenPromptMsg:
		return m.handleFileMsg(msg)
	}

	return m, nil
}

func (m *Model) handleKeyInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// Handle modal states first
	if m.mode != ModeNormal {
		return m.handleModalKeyInput(msg)
	}
	
	// Handle Alt+Arrow keys for word movement
	if left, right := terminal.IsWordMovement(msg); left || right {
		if left {
			m.editor.MoveCursorWordLeft()
		} else {
			m.editor.MoveCursorWordRight()
		}
		return m, nil
	}
	
	switch msg.String() {
	case "ctrl+c":
		if m.editor.GetCursor().HasSelection() {
			m.editor.Copy()
			m.showMessage("Copied")
		} else {
			return m, tea.Quit
		}

	case "ctrl+q":
		// Check if file has unsaved changes
		if m.editor.GetDocument().IsModified() {
			m.mode = ModeSavePrompt
			m.savePromptContext = "quit"
			return m, nil
		}
		return m, tea.Quit

	case "ctrl+s":
		return m, m.saveFile()

	case "ctrl+o":
		return m, m.openFile()

	case "ctrl+v":
		m.editor.Paste()

	case "ctrl+x":
		if m.editor.GetCursor().HasSelection() {
			m.editor.Cut()
			m.showMessage("Cut")
		}

	case "up":
		m.editor.MoveCursorUp()

	case "down":
		m.editor.MoveCursorDown()

	case "left":
		m.editor.MoveCursorLeft()

	case "right":
		m.editor.MoveCursorRight()

	case "shift+up":
		if !m.editor.GetCursor().HasSelection() {
			m.editor.GetCursor().StartSelection()
		}
		m.editor.MoveCursorUp()
		m.editor.GetCursor().ExtendSelection()

	case "shift+down":
		if !m.editor.GetCursor().HasSelection() {
			m.editor.GetCursor().StartSelection()
		}
		m.editor.MoveCursorDown()
		m.editor.GetCursor().ExtendSelection()

	case "shift+left":
		if !m.editor.GetCursor().HasSelection() {
			m.editor.GetCursor().StartSelection()
		}
		m.editor.MoveCursorLeft()
		m.editor.GetCursor().ExtendSelection()

	case "shift+right":
		if !m.editor.GetCursor().HasSelection() {
			m.editor.GetCursor().StartSelection()
		}
		m.editor.MoveCursorRight()
		m.editor.GetCursor().ExtendSelection()

	case "ctrl+a":
		// Select all
		m.editor.GetCursor().StartSelection()
		m.editor.MoveCursorToDocumentStart()
		m.editor.GetCursor().ExtendSelection()
		m.editor.MoveCursorToDocumentEnd()
		m.editor.GetCursor().ExtendSelection()

	case "escape":
		// Clear selection
		m.editor.GetCursor().ClearSelection()

	case "ctrl+l":
		// Toggle line numbers
		m.editor.ToggleLineNumbers()
		if m.editor.ShowLineNumbers() {
			m.showMessage("Line numbers enabled")
		} else {
			m.showMessage("Line numbers disabled")
		}
		
	case "ctrl+f":
		// Enter find mode
		m.mode = ModeFind
		m.input = ""
		m.caseSensitive = false
		
	case "ctrl+h":
		// Enter replace mode
		m.mode = ModeReplace
		m.input = ""
		m.replaceText = ""
		m.caseSensitive = false
		
	case "ctrl+g":
		// Enter goto mode
		m.mode = ModeGoto
		m.input = ""
		
	case "ctrl+p":
		// Toggle preview mode
		m.previewMode = !m.previewMode
		if m.previewMode {
			m.showMessage("Preview mode enabled")
		} else {
			m.showMessage("Preview mode disabled")
		}
		
	case "home":
		m.editor.MoveCursorToLineStart()

	case "end":
		m.editor.MoveCursorToLineEnd()

	case "backspace":
		m.editor.DeleteText(1)

	case "delete":
		pos := m.editor.GetCursor().GetBufferPos()
		m.editor.MoveCursorRight()
		m.editor.DeleteText(1)
		m.editor.GetCursor().SetBufferPos(pos)

	case "enter":
		m.editor.InsertText("\n")

	case "space":
		m.editor.InsertText(" ")

	case "tab":
		m.editor.InsertText("\t")

	default:
		// Handle regular character input
		if isPrintableCharacter(msg.String()) {
			m.editor.InsertText(msg.String())
		}
	}

	return m, nil
}

func (m *Model) showMessage(msg string) {
	m.message = msg
	m.messageTimer = 60 // Show for ~1 second at 60fps
}

// isPrintableCharacter checks if the input represents a single printable character
func isPrintableCharacter(s string) bool {
	if len(s) != 1 {
		return false
	}
	r := rune(s[0])
	return unicode.IsPrint(r) && !unicode.IsControl(r)
}
func (m *Model) handleModalKeyInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "escape":
		// Exit modal mode
		m.mode = ModeNormal
		m.input = ""
		m.replaceText = ""
		m.savePromptContext = ""
		return m, nil
		
	case "enter":
		switch m.mode {
		case ModeFind:
			return m.handleFind()
		case ModeReplace:
			return m.handleReplace()
		case ModeGoto:
			return m.handleGoto()
		}
		return m, nil
		
	case "backspace":
		// Remove last character from input
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
		return m, nil
		
	case "space":
		// Add space to input
		m.input += " "
		return m, nil
		
	default:
		// Handle save prompt responses and regular character input
		if m.mode == ModeSavePrompt {
			return m.handleSavePrompt(msg.String())
		}
		// Add character to input for other modes
		if isPrintableCharacter(msg.String()) {
			m.input += msg.String()
		}
		return m, nil
	}
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

func (m *Model) handleSavePrompt(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "y", "Y":
		// Save and execute context action
		if err := m.editor.SaveFile(""); err != nil {
			m.showMessage("Error saving file: " + err.Error())
			m.mode = ModeNormal
			m.savePromptContext = ""
			return m, nil
		}
		m.showMessage("File saved")
		m.mode = ModeNormal
		
		// Execute context action
		context := m.savePromptContext
		m.savePromptContext = ""
		
		if context == "quit" {
			return m, tea.Quit
		}
		
		return m, nil
		
	case "n", "N":
		// Don't save, execute context action
		m.mode = ModeNormal
		
		// Execute context action
		context := m.savePromptContext
		m.savePromptContext = ""
		
		if context == "quit" {
			return m, tea.Quit
		}
		
		return m, nil
		
	case "c", "C":
		// Cancel, return to editor
		m.mode = ModeNormal
		m.savePromptContext = ""
		return m, nil
	}
	
	return m, nil
}


func (m *Model) handleMouseClick(msg tea.MouseClickMsg) (tea.Model, tea.Cmd) {
	// Only handle mouse events in normal mode
	if m.mode != ModeNormal {
		return m, nil
	}
	
	mouse := msg.Mouse()
	
	// Only handle left button clicks
	if mouse.Button != tea.MouseLeft {
		return m, nil
	}
	
	// Position cursor at click location
	bufferPos := m.screenToBufferSafe(mouse.Y, mouse.X)
	
	// Clear any existing selection and move cursor
	m.editor.GetCursor().ClearSelection()
	m.editor.GetCursor().SetBufferPos(bufferPos)
	
	// Track for potential drag
	m.mouseStartPos = &bufferPos
	m.isDragging = false
	
	return m, nil
}

func (m *Model) handleMouseRelease(msg tea.MouseReleaseMsg) (tea.Model, tea.Cmd) {
	// Only handle mouse events in normal mode
	if m.mode != ModeNormal {
		return m, nil
	}
	
	// End drag selection
	m.isDragging = false
	m.mouseStartPos = nil
	return m, nil
}

func (m *Model) handleMouseMotion(msg tea.MouseMotionMsg) (tea.Model, tea.Cmd) {
	// Only handle mouse events in normal mode
	if m.mode != ModeNormal {
		return m, nil
	}
	
	if m.mouseStartPos == nil {
		return m, nil
	}
	
	mouse := msg.Mouse()
	
	// Convert screen coordinates to buffer position
	bufferPos := m.screenToBufferSafe(mouse.Y, mouse.X)
	
	if !m.isDragging {
		// Start selection on first motion
		m.editor.GetCursor().StartSelection()
		m.isDragging = true
	}
	
	// Update cursor and extend selection
	m.editor.GetCursor().SetBufferPos(bufferPos)
	m.editor.GetCursor().ExtendSelection()
	
	return m, nil
}

func (m *Model) handleMouseWheel(msg tea.MouseWheelMsg) (tea.Model, tea.Cmd) {
	// Only handle mouse events in normal mode
	if m.mode != ModeNormal {
		return m, nil
	}
	
	const scrollAmount = 3 // Standard amount - matches vim and bubbles/viewport
	
	mouse := msg.Mouse()
	
	switch mouse.Button {
	case tea.MouseWheelUp:
		// Scroll viewport up 3 lines (standard) - content moves down
		m.editor.ScrollViewportUp(scrollAmount)
		
	case tea.MouseWheelDown:
		// Scroll viewport down 3 lines (standard) - content moves up
		m.editor.ScrollViewportDown(scrollAmount)
		
	case tea.MouseWheelLeft:
		// Horizontal scroll left - typically 2 columns is standard
		m.editor.ScrollViewportLeft(2)
		
	case tea.MouseWheelRight:
		// Horizontal scroll right - typically 2 columns is standard
		m.editor.ScrollViewportRight(2)
	}
	
	return m, nil
}

