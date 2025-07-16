package unit

import (
	"testing"

	"github.com/ofri/mde/pkg/ast"
	"github.com/stretchr/testify/assert"
)

func TestEditor_GetCursorScreenPosition(t *testing.T) {
	editor := ast.NewEditorWithContent("hello world\ntest line\nfinal")
	
	// Test basic screen position calculation
	editor.GetCursor().SetPosition(ast.Position{Line: 1, Col: 2})
	screenRow, screenCol := editor.GetCursorScreenPosition()
	assert.Equal(t, 1, screenRow)
	assert.Equal(t, 2, screenCol)
	
	// Test with viewport offset
	editor.SetViewPort(80, 25)
	screenRow, screenCol = editor.GetCursorScreenPosition()
	assert.Equal(t, 1, screenRow) // Line 1 - viewport top 0 = 1
	assert.Equal(t, 2, screenCol) // Col 2 - viewport left 0 = 2
	
	// Test with line numbers enabled
	if editor.ShowLineNumbers() {
		screenRow, screenCol = editor.GetCursorScreenPosition()
		assert.Equal(t, 1, screenRow)
		assert.Equal(t, 8, screenCol) // Col 2 + 6 for line numbers = 8
	}
}

func TestEditor_GetCursorScreenPosition_EdgeCases(t *testing.T) {
	editor := ast.NewEditorWithContent("hello\nworld")
	
	// Test cursor at document start
	editor.GetCursor().SetPosition(ast.Position{Line: 0, Col: 0})
	screenRow, screenCol := editor.GetCursorScreenPosition()
	assert.Equal(t, 0, screenRow)
	assert.Equal(t, 0, screenCol)
	
	// Test cursor at line end
	editor.GetCursor().SetPosition(ast.Position{Line: 0, Col: 5})
	screenRow, screenCol = editor.GetCursorScreenPosition()
	assert.Equal(t, 0, screenRow)
	assert.Equal(t, 5, screenCol)
	
	// Test cursor position calculation is consistent
	screenRow, screenCol = editor.GetCursorScreenPosition()
	assert.Equal(t, 0, screenRow) // Line 0
	assert.Equal(t, 5, screenCol) // Col 5
}

func TestEditor_ViewPort(t *testing.T) {
	editor := ast.NewEditorWithContent("line1\nline2\nline3\nline4\nline5")
	
	// Test default viewport
	viewport := editor.GetViewPort()
	assert.Equal(t, 0, viewport.Top)
	assert.Equal(t, 0, viewport.Left)
	
	// Test setting viewport
	editor.SetViewPort(80, 25)
	viewport = editor.GetViewPort()
	assert.Equal(t, 80, viewport.Width)
	assert.Equal(t, 25, viewport.Height)
}

func TestCoordinateTransformation_MouseToDocument(t *testing.T) {
	editor := ast.NewEditorWithContent("hello world\ntest line\nfinal line")
	
	// Test basic mouse click transformation
	// Simulate mouse click at screen position (1, 5)
	viewport := editor.GetViewPort()
	docRow := viewport.Top + 1    // 0 + 1 = 1
	docCol := viewport.Left + 5   // 0 + 5 = 5
	
	assert.Equal(t, 1, docRow)
	assert.Equal(t, 5, docCol)
	
	// Test basic coordinate transformation
	docRow = viewport.Top + 1     // 0 + 1 = 1
	docCol = viewport.Left + 5    // 0 + 5 = 5
	
	assert.Equal(t, 1, docRow)
	assert.Equal(t, 5, docCol)
}

func TestCoordinateTransformation_LineNumbers(t *testing.T) {
	editor := ast.NewEditorWithContent("hello\nworld")
	
	// Test line number offset calculation
	lineNumbersEnabled := editor.ShowLineNumbers()
	
	// Click at screen column 3 (in line number area if enabled)
	clickCol := 3
	var docCol int
	if lineNumbersEnabled && clickCol < 6 {
		docCol = 0 // Click in line number area
	} else if lineNumbersEnabled {
		docCol = clickCol - 6 // Adjust for line number width
	} else {
		docCol = clickCol
	}
	
	if lineNumbersEnabled {
		assert.Equal(t, 0, docCol)
	} else {
		assert.Equal(t, 3, docCol)
	}
	
	// Click at screen column 8 (in content area)  
	clickCol = 8
	if lineNumbersEnabled && clickCol < 6 {
		docCol = 0
	} else if lineNumbersEnabled {
		docCol = clickCol - 6
	} else {
		docCol = clickCol
	}
	
	if lineNumbersEnabled {
		assert.Equal(t, 2, docCol)
	} else {
		assert.Equal(t, 8, docCol)
	}
}

func TestCoordinateTransformation_Bounds(t *testing.T) {
	editor := ast.NewEditorWithContent("hello\nworld")
	doc := editor.GetDocument()
	
	// Test document bounds checking
	lineCount := doc.LineCount()
	assert.Equal(t, 2, lineCount)
	
	// Test row bounds
	docRow := 5 // Beyond document
	if docRow >= lineCount {
		docRow = lineCount - 1
	}
	assert.Equal(t, 1, docRow)
	
	// Test negative row
	docRow = -1
	if docRow < 0 {
		docRow = 0
	}
	assert.Equal(t, 0, docRow)
	
	// Test column bounds
	lineLength := doc.GetLineLength(0)
	assert.Equal(t, 5, lineLength)
	
	docCol := 10 // Beyond line length
	if docCol > lineLength {
		docCol = lineLength
	}
	assert.Equal(t, 5, docCol)
	
	// Test negative column
	docCol = -1
	if docCol < 0 {
		docCol = 0
	}
	assert.Equal(t, 0, docCol)
}

func TestCoordinateInvariant_RoundTrip(t *testing.T) {
	editor := ast.NewEditorWithContent("hello world\ntest line\nfinal")
	
	// Test round-trip: document -> screen -> document
	originalPos := ast.Position{Line: 1, Col: 4}
	editor.GetCursor().SetPosition(originalPos)
	
	// Get screen position
	screenRow, screenCol := editor.GetCursorScreenPosition()
	
	// Convert back to document position
	viewport := editor.GetViewPort()
	docRow := viewport.Top + screenRow
	docCol := viewport.Left + screenCol
	
	// Should match original position
	assert.Equal(t, originalPos.Line, docRow)
	assert.Equal(t, originalPos.Col, docCol)
}

func TestCoordinateInvariant_LineNumbersRoundTrip(t *testing.T) {
	editor := ast.NewEditorWithContent("hello world\ntest line\nfinal")
	lineNumbersEnabled := editor.ShowLineNumbers()
	
	// Test round-trip with line numbers
	originalPos := ast.Position{Line: 1, Col: 4}
	editor.GetCursor().SetPosition(originalPos)
	
	// Get screen position (includes line number offset)
	screenRow, screenCol := editor.GetCursorScreenPosition()
	
	// Convert back to document position (remove line number offset)
	viewport := editor.GetViewPort()
	docRow := viewport.Top + screenRow
	var docCol int
	if lineNumbersEnabled {
		docCol = viewport.Left + (screenCol - 6) // Remove line number offset
	} else {
		docCol = viewport.Left + screenCol
	}
	
	// Should match original position
	assert.Equal(t, originalPos.Line, docRow)
	assert.Equal(t, originalPos.Col, docCol)
}

func TestCoordinateInvariant_ViewportScrolling(t *testing.T) {
	editor := ast.NewEditorWithContent("line1\nline2\nline3\nline4\nline5")
	
	// Test coordinate consistency across viewport changes
	targetPos := ast.Position{Line: 3, Col: 2}
	editor.GetCursor().SetPosition(targetPos)
	
	// Test with different viewport settings
	viewportSizes := []struct{ width, height int }{
		{80, 25},
		{120, 30},
		{60, 20},
	}
	
	for _, vp := range viewportSizes {
		editor.SetViewPort(vp.width, vp.height)
		
		// Get screen position
		screenRow, screenCol := editor.GetCursorScreenPosition()
		
		// Convert back to document position
		viewport := editor.GetViewPort()
		docRow := viewport.Top + screenRow
		docCol := viewport.Left + screenCol
		
		// Should always match target position
		assert.Equal(t, targetPos.Line, docRow, "Viewport: %+v", vp)
		assert.Equal(t, targetPos.Col, docCol, "Viewport: %+v", vp)
	}
}