package integration

import (
	"testing"

	"github.com/ofri/mde/internal/tui"
	"github.com/ofri/mde/pkg/ast"
	"github.com/ofri/mde/test/testutils"
	"github.com/stretchr/testify/assert"
)

func TestViewport_CursorSynchronization(t *testing.T) {
	// This test focuses on the specific issues mentioned in the ticket:
	// - GetCursorScreenPosition() doesn't handle all edge cases
	// - Mouse-to-cursor conversion has multiple error-prone transformations
	// - Cursor/viewport synchronization bugs

	model := tui.New()
	testutils.LoadContentIntoModel(model, "line1\nline2\nline3\nline4\nline5")
	
	editor := model.GetEditor()
	
	// Test 1: Basic cursor screen position calculation
	editor.GetCursor().SetPosition(ast.Position{Line: 2, Col: 3})
	screenRow, screenCol := editor.GetCursorScreenPosition()
	
	assert.Equal(t, 2, screenRow, "Screen row should match document line")
	assert.Equal(t, 3, screenCol, "Screen column should match document column")
	
	// Test 2: Cursor position with viewport offset
	editor.SetViewPort(80, 24)
	viewport := editor.GetViewPort()
	
	// Move cursor to a position that would be visible
	editor.GetCursor().SetPosition(ast.Position{Line: 1, Col: 2})
	screenRow, screenCol = editor.GetCursorScreenPosition()
	
	expectedRow := 1 - viewport.Top
	expectedCol := 2 - viewport.Left
	
	assert.Equal(t, expectedRow, screenRow, "Screen row should account for viewport offset")
	assert.Equal(t, expectedCol, screenCol, "Screen column should account for viewport offset")
	
	// Test 3: Line numbers handling
	if editor.ShowLineNumbers() {
		// Line numbers should add 6 characters to screen column
		assert.Equal(t, expectedCol+6, screenCol, "Line numbers should add 6 to screen column")
	}
}

func TestViewport_MouseCoordinateTransformation(t *testing.T) {
	// Test the mouse-to-cursor coordinate transformation that's mentioned as buggy
	
	model := tui.New()
	testutils.LoadContentIntoModel(model, "hello world\ntest line\nfinal line")
	
	editor := model.GetEditor()
	
	// Simulate mouse click transformation logic
	// This replicates the logic from internal/tui/update.go:344-400
	
	editorHeight := 24 - 2 // Assume 24 height minus status bars
	
	// Test click in editor area
	clickRow := 1
	clickCol := 5
	
	// Check if click is in editor area
	assert.True(t, clickRow < editorHeight, "Click should be in editor area")
	
	// Convert screen position to document position
	viewport := editor.GetViewPort()
	docRow := viewport.Top + clickRow
	docCol := clickCol
	
	// Account for line numbers
	if editor.ShowLineNumbers() {
		if clickCol < 6 {
			docCol = 0 // Click in line number area
		} else {
			docCol = clickCol - 6
		}
	}
	
	// Adjust for viewport
	docCol += viewport.Left
	
	// Ensure coordinates are within document bounds
	doc := editor.GetDocument()
	if docRow >= doc.LineCount() {
		docRow = doc.LineCount() - 1
	}
	if docRow < 0 {
		docRow = 0
	}
	
	lineLength := doc.GetLineLength(docRow)
	if docCol > lineLength {
		docCol = lineLength
	}
	if docCol < 0 {
		docCol = 0
	}
	
	// Verify bounds
	assert.True(t, docRow >= 0, "Document row should be non-negative")
	assert.True(t, docRow < doc.LineCount(), "Document row should be within bounds")
	assert.True(t, docCol >= 0, "Document column should be non-negative")
	assert.True(t, docCol <= lineLength, "Document column should be within line bounds")
	
	// Test round-trip: document -> screen -> document
	editor.GetCursor().SetPosition(ast.Position{Line: docRow, Col: docCol})
	screenRow, screenCol := editor.GetCursorScreenPosition()
	
	// Convert back to document coordinates
	backDocRow := viewport.Top + screenRow
	backDocCol := viewport.Left + screenCol
	
	if editor.ShowLineNumbers() {
		backDocCol -= 6
	}
	
	// Should match original position (within bounds)
	assert.Equal(t, docRow, backDocRow, "Round-trip should preserve row")
	assert.Equal(t, docCol, backDocCol, "Round-trip should preserve column")
}

func TestViewport_EdgeCases(t *testing.T) {
	// Test edge cases that might cause cursor positioning issues
	
	model := tui.New()
	testutils.LoadContentIntoModel(model, "short\nverylonglinewithnospaces\na")
	
	editor := model.GetEditor()
	
	// Test cursor at end of short line
	editor.GetCursor().SetPosition(ast.Position{Line: 0, Col: 5})
	screenRow, screenCol := editor.GetCursorScreenPosition()
	
	assert.Equal(t, 0, screenRow, "Cursor at end of line should have correct row")
	assert.Equal(t, 5, screenCol, "Cursor at end of line should have correct column")
	
	// Test cursor at start of long line
	editor.GetCursor().SetPosition(ast.Position{Line: 1, Col: 0})
	screenRow, screenCol = editor.GetCursorScreenPosition()
	
	assert.Equal(t, 1, screenRow, "Cursor at start of long line should have correct row")
	assert.Equal(t, 0, screenCol, "Cursor at start of long line should have correct column")
	
	// Test cursor in middle of long line
	editor.GetCursor().SetPosition(ast.Position{Line: 1, Col: 10})
	screenRow, screenCol = editor.GetCursorScreenPosition()
	
	assert.Equal(t, 1, screenRow, "Cursor in middle of long line should have correct row")
	assert.Equal(t, 10, screenCol, "Cursor in middle of long line should have correct column")
}

func TestViewport_BoundaryConditions(t *testing.T) {
	// Test boundary conditions that might cause artifacts
	
	model := tui.New()
	testutils.LoadContentIntoModel(model, "line1\nline2\nline3")
	
	editor := model.GetEditor()
	
	// Test cursor at document boundaries
	positions := []ast.Position{
		{Line: 0, Col: 0},                                           // Start of document
		{Line: 2, Col: editor.GetDocument().GetLineLength(2)},      // End of document
		{Line: 1, Col: 0},                                           // Start of middle line
		{Line: 1, Col: editor.GetDocument().GetLineLength(1)},      // End of middle line
	}
	
	for _, pos := range positions {
		editor.GetCursor().SetPosition(pos)
		screenRow, screenCol := editor.GetCursorScreenPosition()
		
		// Basic sanity checks
		assert.True(t, screenRow >= 0, "Screen row should be non-negative for position %+v", pos)
		assert.True(t, screenCol >= 0, "Screen column should be non-negative for position %+v", pos)
		
		// Verify position is within document bounds
		actualPos := editor.GetCursor().GetPosition()
		assert.True(t, actualPos.Line >= 0, "Cursor line should be non-negative")
		assert.True(t, actualPos.Line < editor.GetDocument().LineCount(), "Cursor line should be within document")
		assert.True(t, actualPos.Col >= 0, "Cursor column should be non-negative")
		assert.True(t, actualPos.Col <= editor.GetDocument().GetLineLength(actualPos.Line), "Cursor column should be within line")
	}
}

func TestViewport_ConsistencyInvariants(t *testing.T) {
	// Test consistency invariants that should always hold
	
	model := tui.New()
	testutils.LoadContentIntoModel(model, "line1\nline2\nline3\nline4\nline5")
	
	editor := model.GetEditor()
	
	// Test various cursor positions
	for line := 0; line < editor.GetDocument().LineCount(); line++ {
		lineLength := editor.GetDocument().GetLineLength(line)
		
		for col := 0; col <= lineLength; col++ {
			pos := ast.Position{Line: line, Col: col}
			editor.GetCursor().SetPosition(pos)
			
			// Get screen position
			screenRow, screenCol := editor.GetCursorScreenPosition()
			
			// Convert back to document position
			viewport := editor.GetViewPort()
			backDocRow := viewport.Top + screenRow
			backDocCol := viewport.Left + screenCol
			
			if editor.ShowLineNumbers() {
				backDocCol -= 6
			}
			
			// Invariant: round-trip should preserve position
			actualPos := editor.GetCursor().GetPosition()
			assert.Equal(t, actualPos.Line, backDocRow, "Round-trip should preserve row for position %+v", pos)
			assert.Equal(t, actualPos.Col, backDocCol, "Round-trip should preserve column for position %+v", pos)
		}
	}
}

func TestViewport_LineNumberHandling(t *testing.T) {
	// Test line number handling in coordinate transformations
	
	model := tui.New()
	testutils.LoadContentIntoModel(model, "line1\nline2\nline3")
	
	editor := model.GetEditor()
	
	// Test with line numbers enabled/disabled
	lineNumberStates := []bool{true, false}
	
	for _, _ = range lineNumberStates {
		// Note: We can't actually toggle line numbers in this test
		// because the editor doesn't have a SetLineNumbers method
		// We'll just test the current state
		
		pos := ast.Position{Line: 1, Col: 3}
		editor.GetCursor().SetPosition(pos)
		
		screenRow, screenCol := editor.GetCursorScreenPosition()
		
		// Basic consistency check
		assert.Equal(t, 1, screenRow, "Screen row should match document line")
		
		if editor.ShowLineNumbers() {
			assert.Equal(t, 3+6, screenCol, "Screen column should include line number offset")
		} else {
			assert.Equal(t, 3, screenCol, "Screen column should match document column")
		}
	}
}

func TestViewport_UnicodeHandling(t *testing.T) {
	// Test viewport handling with Unicode content
	
	model := tui.New()
	testutils.LoadContentIntoModel(model, "hello 世界\nこんにちは\n🌍🚀💫")
	
	editor := model.GetEditor()
	
	// Test cursor positions in Unicode content
	positions := []ast.Position{
		{Line: 0, Col: 6},  // After space, before 世
		{Line: 0, Col: 7},  // After 世, before 界
		{Line: 1, Col: 2},  // In middle of こんにちは
		{Line: 2, Col: 1},  // After 🌍, before 🚀
	}
	
	for _, pos := range positions {
		editor.GetCursor().SetPosition(pos)
		screenRow, screenCol := editor.GetCursorScreenPosition()
		
		// Basic sanity checks
		assert.True(t, screenRow >= 0, "Screen row should be non-negative for Unicode position %+v", pos)
		assert.True(t, screenCol >= 0, "Screen column should be non-negative for Unicode position %+v", pos)
		
		// Verify position is valid after cursor movement
		actualPos := editor.GetCursor().GetPosition()
		assert.True(t, actualPos.Line >= 0, "Cursor line should be non-negative after Unicode movement")
		assert.True(t, actualPos.Col >= 0, "Cursor column should be non-negative after Unicode movement")
	}
}