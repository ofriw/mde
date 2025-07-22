package unit

import (
	"os"
	"testing"

	"github.com/ofri/mde/pkg/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create temporary files for testing
func createTempFile(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "cursor_test_*.txt")
	require.NoError(t, err)
	
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	
	err = tmpFile.Close()
	require.NoError(t, err)
	
	return tmpFile.Name()
}

func TestCursor_NewEditor(t *testing.T) {
	editor := ast.NewEditor()
	cursor := editor.GetCursor()
	
	assert.Equal(t, ast.BufferPos{Line: 0, Col: 0}, cursor.GetBufferPos())
	assert.False(t, cursor.HasSelection())
	assert.Nil(t, cursor.GetSelection())
}

func TestCursor_SetPosition(t *testing.T) {
	editor := ast.NewEditorWithContent("line1\nline2\nline3")
	cursor := editor.GetCursor()
	
	// Test valid position
	cursor.SetBufferPos(ast.BufferPos{Line: 1, Col: 2})
	assert.Equal(t, ast.BufferPos{Line: 1, Col: 2}, cursor.GetBufferPos())
	
	// Test position validation - should clamp to document bounds
	cursor.SetBufferPos(ast.BufferPos{Line: 10, Col: 20})
	pos := cursor.GetBufferPos()
	doc := editor.GetDocument()
	assert.True(t, pos.Line < doc.LineCount())
	assert.True(t, pos.Col <= doc.GetLineLength(pos.Line))
}

func TestCursor_BasicMovement(t *testing.T) {
	editor := ast.NewEditorWithContent("hello\nworld")
	cursor := editor.GetCursor()
	
	// Test right movement
	cursor.SetBufferPos(ast.BufferPos{Line: 0, Col: 2})
	editor.MoveCursorRight()
	assert.Equal(t, ast.BufferPos{Line: 0, Col: 3}, cursor.GetBufferPos())
	
	// Test left movement
	editor.MoveCursorLeft()
	assert.Equal(t, ast.BufferPos{Line: 0, Col: 2}, cursor.GetBufferPos())
	
	// Test down movement
	editor.MoveCursorDown()
	assert.Equal(t, ast.BufferPos{Line: 1, Col: 2}, cursor.GetBufferPos())
	
	// Test up movement
	editor.MoveCursorUp()
	assert.Equal(t, ast.BufferPos{Line: 0, Col: 2}, cursor.GetBufferPos())
}

func TestCursor_LineMovement(t *testing.T) {
	editor := ast.NewEditorWithContent("hello\nworld")
	cursor := editor.GetCursor()
	
	// Test move to line start
	cursor.SetBufferPos(ast.BufferPos{Line: 0, Col: 3})
	editor.MoveCursorToLineStart()
	assert.Equal(t, ast.BufferPos{Line: 0, Col: 0}, cursor.GetBufferPos())
	
	// Test move to line end
	editor.MoveCursorToLineEnd()
	assert.Equal(t, ast.BufferPos{Line: 0, Col: 5}, cursor.GetBufferPos())
}

func TestCursor_DocumentMovement(t *testing.T) {
	editor := ast.NewEditorWithContent("hello\nworld\ntest")
	cursor := editor.GetCursor()
	
	// Test move to document start
	cursor.SetBufferPos(ast.BufferPos{Line: 2, Col: 3})
	editor.MoveCursorToDocumentStart()
	assert.Equal(t, ast.BufferPos{Line: 0, Col: 0}, cursor.GetBufferPos())
	
	// Test move to document end
	editor.MoveCursorToDocumentEnd()
	pos := cursor.GetBufferPos()
	doc := editor.GetDocument()
	assert.Equal(t, doc.LineCount()-1, pos.Line)
	assert.Equal(t, doc.GetLineLength(pos.Line), pos.Col)
}

func TestCursor_Selection(t *testing.T) {
	editor := ast.NewEditorWithContent("hello\nworld")
	cursor := editor.GetCursor()
	
	// Test selection creation
	assert.False(t, cursor.HasSelection())
	cursor.StartSelection()
	assert.True(t, cursor.HasSelection())
	
	// Test selection extension
	cursor.SetBufferPos(ast.BufferPos{Line: 0, Col: 5})
	cursor.ExtendSelection()
	
	selection := cursor.GetSelection()
	assert.NotNil(t, selection)
	assert.Equal(t, ast.BufferPos{Line: 0, Col: 0}, selection.Start)
	assert.Equal(t, ast.BufferPos{Line: 0, Col: 5}, selection.End)
	
	// Test clear selection
	cursor.ClearSelection()
	assert.False(t, cursor.HasSelection())
	assert.Nil(t, cursor.GetSelection())
}

func TestCursor_ScreenPosition(t *testing.T) {
	editor := ast.NewEditorWithContent("hello\nworld")
	editor.ToggleLineNumbers() // Turn off line numbers since they're on by default
	cursor := editor.GetCursor()
	
	// Test screen position calculation
	cursor.SetBufferPos(ast.BufferPos{Line: 0, Col: 0})
	screenPos, err := cursor.GetScreenPos()
	
	// Should be visible at origin
	assert.NoError(t, err)
	assert.Equal(t, 0, screenPos.Row)
	assert.Equal(t, 0, screenPos.Col)
}

func TestCursor_WordMovement(t *testing.T) {
	content := "hello world test"
	editor := ast.NewEditorWithContent(content)
	cursor := editor.GetCursor()
	
	// Test improved word movement functionality
	t.Run("word right movement", func(t *testing.T) {
		// From start of "hello" to start of "world"
		cursor.SetBufferPos(ast.BufferPos{Line: 0, Col: 0})
		editor.MoveCursorWordRight()
		assert.Equal(t, ast.BufferPos{Line: 0, Col: 6}, cursor.GetBufferPos())
		
		// From start of "world" to start of "test"
		cursor.SetBufferPos(ast.BufferPos{Line: 0, Col: 6})
		editor.MoveCursorWordRight()
		assert.Equal(t, ast.BufferPos{Line: 0, Col: 12}, cursor.GetBufferPos())
		
		// From start of "test" to end of line
		cursor.SetBufferPos(ast.BufferPos{Line: 0, Col: 12})
		editor.MoveCursorWordRight()
		assert.Equal(t, ast.BufferPos{Line: 0, Col: 16}, cursor.GetBufferPos())
	})
	
	t.Run("word left movement", func(t *testing.T) {
		// From end of line to start of "test"
		cursor.SetBufferPos(ast.BufferPos{Line: 0, Col: 16})
		editor.MoveCursorWordLeft()
		assert.Equal(t, ast.BufferPos{Line: 0, Col: 12}, cursor.GetBufferPos())
		
		// From start of "test" to start of "world"
		cursor.SetBufferPos(ast.BufferPos{Line: 0, Col: 12})
		editor.MoveCursorWordLeft()
		assert.Equal(t, ast.BufferPos{Line: 0, Col: 6}, cursor.GetBufferPos())
		
		// From start of "world" to start of "hello"
		cursor.SetBufferPos(ast.BufferPos{Line: 0, Col: 6})
		editor.MoveCursorWordLeft()
		assert.Equal(t, ast.BufferPos{Line: 0, Col: 0}, cursor.GetBufferPos())
	})
	
	t.Run("movement from middle of word", func(t *testing.T) {
		// From middle of "world" left should go to start of "world"
		cursor.SetBufferPos(ast.BufferPos{Line: 0, Col: 8})
		editor.MoveCursorWordLeft()
		assert.Equal(t, ast.BufferPos{Line: 0, Col: 6}, cursor.GetBufferPos())
		
		// From middle of "world" right should go to start of "test"
		cursor.SetBufferPos(ast.BufferPos{Line: 0, Col: 8})
		editor.MoveCursorWordRight()
		assert.Equal(t, ast.BufferPos{Line: 0, Col: 12}, cursor.GetBufferPos())
	})
	
	t.Run("movement from whitespace", func(t *testing.T) {
		// From space between "hello" and "world" - left should go to start of "hello"
		cursor.SetBufferPos(ast.BufferPos{Line: 0, Col: 5})
		editor.MoveCursorWordLeft()
		assert.Equal(t, ast.BufferPos{Line: 0, Col: 0}, cursor.GetBufferPos())
		
		// From space between "world" and "test" - right should go to start of "test"
		cursor.SetBufferPos(ast.BufferPos{Line: 0, Col: 11})
		editor.MoveCursorWordRight()
		assert.Equal(t, ast.BufferPos{Line: 0, Col: 12}, cursor.GetBufferPos())
	})
	
	t.Run("cross-line whitespace handling", func(t *testing.T) {
		// Test with multiple lines and various whitespace
		multiLineContent := "foo\n\n  bar"
		multiLineEditor := ast.NewEditorWithContent(multiLineContent)
		multiLineCursor := multiLineEditor.GetCursor()
		
		// From start, should skip empty line and whitespace to reach "bar"
		multiLineCursor.SetBufferPos(ast.BufferPos{Line: 0, Col: 0})
		multiLineEditor.MoveCursorWordRight()
		assert.Equal(t, ast.BufferPos{Line: 2, Col: 2}, multiLineCursor.GetBufferPos()) // Start of "bar"
		
		// From end of "bar", should go back to start of "foo"
		multiLineCursor.SetBufferPos(ast.BufferPos{Line: 2, Col: 5}) // End of "bar"
		multiLineEditor.MoveCursorWordLeft()
		assert.Equal(t, ast.BufferPos{Line: 2, Col: 2}, multiLineCursor.GetBufferPos()) // Start of "bar"
		
		multiLineEditor.MoveCursorWordLeft()
		assert.Equal(t, ast.BufferPos{Line: 0, Col: 0}, multiLineCursor.GetBufferPos()) // Start of "foo"
	})
}

func TestCursor_WordSelection(t *testing.T) {
	content := "hello world test"
	editor := ast.NewEditorWithContent(content)
	cursor := editor.GetCursor()
	
	t.Run("word selection right", func(t *testing.T) {
		cursor.SetBufferPos(ast.BufferPos{Line: 0, Col: 0}) // Start of "hello"
		
		// Start selection and move word right
		cursor.StartSelection()
		editor.MoveCursorWordRight()
		cursor.ExtendSelection()
		
		assert.True(t, cursor.HasSelection())
		selection := cursor.GetSelection()
		assert.Equal(t, ast.BufferPos{Line: 0, Col: 0}, selection.Start)
		assert.Equal(t, ast.BufferPos{Line: 0, Col: 6}, selection.End) // Start of "world"
	})
	
	t.Run("word selection left", func(t *testing.T) {
		cursor.SetBufferPos(ast.BufferPos{Line: 0, Col: 15}) // Last char of "test"
		cursor.ClearSelection()
		
		// Start selection and move word left
		cursor.StartSelection()
		editor.MoveCursorWordLeft()
		cursor.ExtendSelection()
		
		assert.True(t, cursor.HasSelection())
		selection := cursor.GetSelection()
		// When moving left, the selection Start/End might be flipped compared to movement direction
		assert.Equal(t, ast.BufferPos{Line: 0, Col: 15}, selection.Start) // Original position 
		assert.Equal(t, ast.BufferPos{Line: 0, Col: 12}, selection.End)   // After moving left
	})
}