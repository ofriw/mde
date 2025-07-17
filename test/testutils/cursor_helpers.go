package testutils

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/ofri/mde/internal/tui"
	"github.com/ofri/mde/pkg/ast"
	"github.com/stretchr/testify/assert"
)

// CursorTestHelper provides utilities for cursor testing in TUI context
type CursorTestHelper struct {
	t      *testing.T
	model  *tui.Model
	width  int
	height int
}

// NewCursorTestHelper creates a new cursor test helper
func NewCursorTestHelper(t *testing.T, content string, width, height int) *CursorTestHelper {
	model := tui.New()
	if content != "" {
		LoadContentIntoModel(model, content)
	}
	SetModelSize(model, width, height)
	
	return &CursorTestHelper{
		t:      t,
		model:  model,
		width:  width,
		height: height,
	}
}

// GetCursorPosition returns the current cursor position in the document
func (h *CursorTestHelper) GetCursorPosition() ast.BufferPos {
	return h.model.GetEditor().GetCursor().GetBufferPos()
}

// SetCursorPosition sets the cursor to a specific position
func (h *CursorTestHelper) SetCursorPosition(pos ast.BufferPos) {
	h.model.GetEditor().GetCursor().SetBufferPos(pos)
}

// GetScreenPosition returns the screen position of the cursor
func (h *CursorTestHelper) GetScreenPosition() (int, int) {
	screenPos, err := h.model.GetEditor().GetCursor().GetScreenPos()
	if err != nil {
		return -1, -1  // Cursor not visible
	}
	return screenPos.Row, screenPos.Col
}

// GetViewPort returns the current viewport
func (h *CursorTestHelper) GetViewPort() *ast.Viewport {
	return h.model.GetEditor().GetViewport()
}

// SendKey sends a key press to the model
func (h *CursorTestHelper) SendKey(key string) {
	var msg tea.Msg
	switch key {
	case "left":
		msg = tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		msg = tea.KeyMsg{Type: tea.KeyRight}
	case "up":
		msg = tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		msg = tea.KeyMsg{Type: tea.KeyDown}
	case "home":
		msg = tea.KeyMsg{Type: tea.KeyHome}
	case "end":
		msg = tea.KeyMsg{Type: tea.KeyEnd}
	case "ctrl+home":
		msg = tea.KeyMsg{Type: tea.KeyCtrlHome}
	case "ctrl+end":
		msg = tea.KeyMsg{Type: tea.KeyCtrlEnd}
	case "ctrl+left":
		msg = tea.KeyMsg{Type: tea.KeyCtrlLeft}
	case "ctrl+right":
		msg = tea.KeyMsg{Type: tea.KeyCtrlRight}
	default:
		if len(key) == 1 {
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		} else {
			h.t.Fatalf("Unknown key: %s", key)
		}
	}
	
	h.model.Update(msg)
}

// SendMouseClick sends a mouse click event to the model
func (h *CursorTestHelper) SendMouseClick(x, y int) {
	msg := tea.MouseMsg{
		Type: tea.MouseLeft,
		X:    x,
		Y:    y,
	}
	h.model.Update(msg)
}

// SendMouseDrag sends a mouse drag event to the model
func (h *CursorTestHelper) SendMouseDrag(x, y int) {
	msg := tea.MouseMsg{
		Type: tea.MouseLeft,
		X:    x,
		Y:    y,
	}
	h.model.Update(msg)
}

// AssertCursorPosition asserts that the cursor is at the expected position
func (h *CursorTestHelper) AssertCursorPosition(expected ast.BufferPos) {
	actual := h.GetCursorPosition()
	assert.Equal(h.t, expected, actual, "Cursor position mismatch")
}

// AssertScreenPosition asserts that the cursor screen position is as expected
func (h *CursorTestHelper) AssertScreenPosition(expectedRow, expectedCol int) {
	actualRow, actualCol := h.GetScreenPosition()
	assert.Equal(h.t, expectedRow, actualRow, "Cursor screen row mismatch")
	assert.Equal(h.t, expectedCol, actualCol, "Cursor screen column mismatch")
}

// AssertCursorVisible asserts that the cursor is visible within the viewport
func (h *CursorTestHelper) AssertCursorVisible() {
	pos := h.GetCursorPosition()
	viewport := h.GetViewPort()
	
	assert.True(h.t, pos.Line >= viewport.GetTopLine(), "Cursor line should be >= viewport top")
	assert.True(h.t, pos.Line < viewport.GetTopLine()+viewport.GetHeight(), "Cursor line should be < viewport bottom")
	assert.True(h.t, pos.Col >= viewport.GetLeftColumn(), "Cursor column should be >= viewport left")
	assert.True(h.t, pos.Col < viewport.GetLeftColumn()+viewport.GetWidth(), "Cursor column should be < viewport right")
}

// RunTUITest runs a TUI test with teatest framework
func (h *CursorTestHelper) RunTUITest(testFunc func(tm *teatest.TestModel)) {
	tm := teatest.NewTestModel(h.t, h.model, teatest.WithInitialTermSize(h.width, h.height))
	testFunc(tm)
}

// GetDocumentContent returns the current document content
func (h *CursorTestHelper) GetDocumentContent() string {
	return h.model.GetEditor().GetDocument().GetText()
}

// GetVisibleContent returns the currently visible content
func (h *CursorTestHelper) GetVisibleContent() []string {
	return h.model.GetEditor().GetVisibleLines()
}

// GetSelection returns the current selection
func (h *CursorTestHelper) GetSelection() *ast.Selection {
	return h.model.GetEditor().GetCursor().GetSelection()
}

// HasSelection returns true if there is an active selection
func (h *CursorTestHelper) HasSelection() bool {
	return h.model.GetEditor().GetCursor().HasSelection()
}

// GetSelectionText returns the selected text
func (h *CursorTestHelper) GetSelectionText() string {
	return h.model.GetEditor().GetSelectionText()
}

// SimulateTyping simulates typing a string character by character
func (h *CursorTestHelper) SimulateTyping(text string) {
	for _, char := range text {
		h.SendKey(string(char))
	}
}

// SimulateResize simulates a terminal resize event
func (h *CursorTestHelper) SimulateResize(width, height int) {
	h.width = width
	h.height = height
	SetModelSize(h.model, width, height)
}

// FormatCursorState returns a formatted string describing the cursor state
func (h *CursorTestHelper) FormatCursorState() string {
	pos := h.GetCursorPosition()
	screenRow, screenCol := h.GetScreenPosition()
	viewport := h.GetViewPort()
	
	return fmt.Sprintf(
		"Cursor: doc(%d,%d) screen(%d,%d) viewport(top=%d,left=%d,w=%d,h=%d)",
		pos.Line, pos.Col, screenRow, screenCol,
		viewport.GetTopLine(), viewport.GetLeftColumn(), viewport.GetWidth(), viewport.GetHeight(),
	)
}

// AssertNoCursorArtifacts checks that there are no visual artifacts in the rendered output
func (h *CursorTestHelper) AssertNoCursorArtifacts() {
	// This is a placeholder - in a real implementation, you would check the
	// rendered output for duplicate cursor characters or other artifacts
	// For now, we'll just ensure the cursor position is valid
	pos := h.GetCursorPosition()
	doc := h.model.GetEditor().GetDocument()
	
	assert.True(h.t, pos.Line >= 0, "Cursor line should be non-negative")
	assert.True(h.t, pos.Line < doc.LineCount(), "Cursor line should be within document")
	assert.True(h.t, pos.Col >= 0, "Cursor column should be non-negative")
	assert.True(h.t, pos.Col <= doc.GetLineLength(pos.Line), "Cursor column should be within line")
}

// WaitForStableState waits for the cursor to reach a stable state
func (h *CursorTestHelper) WaitForStableState() {
	// In a real implementation, this might wait for animations to complete
	// or for the cursor to stop moving. For now, it's a no-op.
}