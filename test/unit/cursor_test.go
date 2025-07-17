package unit

import (
	"io/ioutil"
	"testing"

	"github.com/ofri/mde/pkg/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create temporary files for testing
func createTempFile(t *testing.T, content string) string {
	tmpFile, err := ioutil.TempFile("", "cursor_test_*.txt")
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
	cursor := editor.GetCursor()
	
	// Test screen position calculation
	cursor.SetBufferPos(ast.BufferPos{Line: 0, Col: 0})
	screenPos, err := cursor.GetScreenPos()
	
	// Should be visible at origin
	assert.NoError(t, err)
	assert.Equal(t, 0, screenPos.Row)
	assert.Equal(t, 0, screenPos.Col)
}