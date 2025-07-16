package unit

import (
	"io/ioutil"
	"os"
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

func TestCursor_NewCursor(t *testing.T) {
	doc := ast.NewEmptyDocument()
	cursor := ast.NewCursor(doc)
	
	assert.Equal(t, ast.Position{Line: 0, Col: 0}, cursor.GetPosition())
	assert.False(t, cursor.HasSelection())
	assert.Nil(t, cursor.GetSelection())
}

func TestCursor_SetPosition(t *testing.T) {
	doc := ast.NewDocument("line1\nline2\nline3")
	cursor := ast.NewCursor(doc)
	
	// Test valid position
	cursor.SetPosition(ast.Position{Line: 1, Col: 2})
	assert.Equal(t, ast.Position{Line: 1, Col: 2}, cursor.GetPosition())
	
	// Test position validation - should clamp to document bounds
	cursor.SetPosition(ast.Position{Line: 10, Col: 20})
	pos := cursor.GetPosition()
	assert.True(t, pos.Line < doc.LineCount())
	assert.True(t, pos.Col <= doc.GetLineLength(pos.Line))
}

func TestCursor_MoveLeft(t *testing.T) {
	doc := ast.NewDocument("hello\nworld")
	cursor := ast.NewCursor(doc)
	
	// Move to middle of first line
	cursor.SetPosition(ast.Position{Line: 0, Col: 3})
	cursor.MoveLeft()
	assert.Equal(t, ast.Position{Line: 0, Col: 2}, cursor.GetPosition())
	
	// Move to start of first line - should not move further
	cursor.SetPosition(ast.Position{Line: 0, Col: 0})
	cursor.MoveLeft()
	assert.Equal(t, ast.Position{Line: 0, Col: 0}, cursor.GetPosition())
	
	// Move from start of second line to end of first line
	cursor.SetPosition(ast.Position{Line: 1, Col: 0})
	cursor.MoveLeft()
	assert.Equal(t, ast.Position{Line: 0, Col: 5}, cursor.GetPosition())
}

func TestCursor_MoveRight(t *testing.T) {
	doc := ast.NewDocument("hello\nworld")
	cursor := ast.NewCursor(doc)
	
	// Move within line
	cursor.SetPosition(ast.Position{Line: 0, Col: 2})
	cursor.MoveRight()
	assert.Equal(t, ast.Position{Line: 0, Col: 3}, cursor.GetPosition())
	
	// Move from end of first line to start of second line
	cursor.SetPosition(ast.Position{Line: 0, Col: 5})
	cursor.MoveRight()
	assert.Equal(t, ast.Position{Line: 1, Col: 0}, cursor.GetPosition())
	
	// Move at end of last line - should not move further
	cursor.SetPosition(ast.Position{Line: 1, Col: 5})
	cursor.MoveRight()
	assert.Equal(t, ast.Position{Line: 1, Col: 5}, cursor.GetPosition())
}

func TestCursor_MoveUp(t *testing.T) {
	doc := ast.NewDocument("hello\nworld\ntest")
	cursor := ast.NewCursor(doc)
	
	// Move up maintaining column
	cursor.SetPosition(ast.Position{Line: 2, Col: 2})
	cursor.MoveUp()
	assert.Equal(t, ast.Position{Line: 1, Col: 2}, cursor.GetPosition())
	
	// Move up to shorter line - should clamp column
	cursor.SetPosition(ast.Position{Line: 1, Col: 4})
	cursor.MoveUp()
	assert.Equal(t, ast.Position{Line: 0, Col: 4}, cursor.GetPosition())
	
	// Move up from first line - should not move
	cursor.SetPosition(ast.Position{Line: 0, Col: 2})
	cursor.MoveUp()
	assert.Equal(t, ast.Position{Line: 0, Col: 2}, cursor.GetPosition())
}

func TestCursor_MoveDown(t *testing.T) {
	doc := ast.NewDocument("hello\nworld\ntest")
	cursor := ast.NewCursor(doc)
	
	// Move down maintaining column
	cursor.SetPosition(ast.Position{Line: 0, Col: 2})
	cursor.MoveDown()
	assert.Equal(t, ast.Position{Line: 1, Col: 2}, cursor.GetPosition())
	
	// Move down from last line - should not move
	cursor.SetPosition(ast.Position{Line: 2, Col: 2})
	cursor.MoveDown()
	assert.Equal(t, ast.Position{Line: 2, Col: 2}, cursor.GetPosition())
}

func TestCursor_MoveToLineStart(t *testing.T) {
	doc := ast.NewDocument("hello\nworld")
	cursor := ast.NewCursor(doc)
	
	cursor.SetPosition(ast.Position{Line: 0, Col: 3})
	cursor.MoveToLineStart()
	assert.Equal(t, ast.Position{Line: 0, Col: 0}, cursor.GetPosition())
}

func TestCursor_MoveToLineEnd(t *testing.T) {
	doc := ast.NewDocument("hello\nworld")
	cursor := ast.NewCursor(doc)
	
	cursor.SetPosition(ast.Position{Line: 0, Col: 2})
	cursor.MoveToLineEnd()
	assert.Equal(t, ast.Position{Line: 0, Col: 5}, cursor.GetPosition())
}

func TestCursor_MoveToDocumentStart(t *testing.T) {
	doc := ast.NewDocument("hello\nworld\ntest")
	cursor := ast.NewCursor(doc)
	
	cursor.SetPosition(ast.Position{Line: 2, Col: 3})
	cursor.MoveToDocumentStart()
	assert.Equal(t, ast.Position{Line: 0, Col: 0}, cursor.GetPosition())
}

func TestCursor_MoveToDocumentEnd(t *testing.T) {
	doc := ast.NewDocument("hello\nworld\ntest")
	cursor := ast.NewCursor(doc)
	
	cursor.SetPosition(ast.Position{Line: 0, Col: 0})
	cursor.MoveToDocumentEnd()
	assert.Equal(t, ast.Position{Line: 2, Col: 4}, cursor.GetPosition())
}

func TestCursor_MoveWordLeft(t *testing.T) {
	doc := ast.NewDocument("hello world\ntest line")
	cursor := ast.NewCursor(doc)
	
	// Move from middle of word to start
	cursor.SetPosition(ast.Position{Line: 0, Col: 8})
	cursor.MoveWordLeft()
	pos := cursor.GetPosition()
	assert.True(t, pos.Col <= 8) // Should move left
	
	// Move from start of line to end of previous line
	cursor.SetPosition(ast.Position{Line: 1, Col: 0})
	cursor.MoveWordLeft()
	assert.Equal(t, ast.Position{Line: 0, Col: 11}, cursor.GetPosition())
}

func TestCursor_MoveWordRight(t *testing.T) {
	doc := ast.NewDocument("hello world\ntest line")
	cursor := ast.NewCursor(doc)
	
	// Move from middle of word to end
	cursor.SetPosition(ast.Position{Line: 0, Col: 2})
	cursor.MoveWordRight()
	pos := cursor.GetPosition()
	assert.True(t, pos.Col >= 2) // Should move right
	
	// Move from end of line to start of next line
	cursor.SetPosition(ast.Position{Line: 0, Col: 11})
	cursor.MoveWordRight()
	assert.Equal(t, ast.Position{Line: 1, Col: 0}, cursor.GetPosition())
}

func TestCursor_Selection(t *testing.T) {
	doc := ast.NewDocument("hello world\ntest line")
	cursor := ast.NewCursor(doc)
	
	// Test starting selection
	cursor.SetPosition(ast.Position{Line: 0, Col: 0})
	cursor.StartSelection()
	assert.True(t, cursor.HasSelection())
	
	selection := cursor.GetSelection()
	require.NotNil(t, selection)
	assert.Equal(t, ast.Position{Line: 0, Col: 0}, selection.Start)
	assert.Equal(t, ast.Position{Line: 0, Col: 0}, selection.End)
	
	// Test extending selection
	cursor.SetPosition(ast.Position{Line: 0, Col: 5})
	cursor.ExtendSelection()
	selection = cursor.GetSelection()
	assert.Equal(t, ast.Position{Line: 0, Col: 0}, selection.Start)
	assert.Equal(t, ast.Position{Line: 0, Col: 5}, selection.End)
	
	// Test clearing selection
	cursor.ClearSelection()
	assert.False(t, cursor.HasSelection())
	assert.Nil(t, cursor.GetSelection())
}

func TestCursor_GetSelectionText(t *testing.T) {
	doc := ast.NewDocument("hello world\ntest line")
	cursor := ast.NewCursor(doc)
	
	// Test no selection
	assert.Equal(t, "", cursor.GetSelectionText())
	
	// Test single line selection
	cursor.SetPosition(ast.Position{Line: 0, Col: 0})
	cursor.StartSelection()
	cursor.SetPosition(ast.Position{Line: 0, Col: 5})
	cursor.ExtendSelection()
	assert.Equal(t, "hello", cursor.GetSelectionText())
	
	// Test multi-line selection
	cursor.SetPosition(ast.Position{Line: 0, Col: 6})
	cursor.StartSelection()
	cursor.SetPosition(ast.Position{Line: 1, Col: 4})
	cursor.ExtendSelection()
	assert.Equal(t, "world\ntest", cursor.GetSelectionText())
}

func TestCursor_EdgeCases(t *testing.T) {
	t.Run("empty document", func(t *testing.T) {
		doc := ast.NewEmptyDocument()
		cursor := ast.NewCursor(doc)
		
		// Should not crash on empty document
		cursor.MoveLeft()
		cursor.MoveRight()
		cursor.MoveUp()
		cursor.MoveDown()
		cursor.MoveToLineStart()
		cursor.MoveToLineEnd()
		cursor.MoveToDocumentStart()
		cursor.MoveToDocumentEnd()
		cursor.MoveWordLeft()
		cursor.MoveWordRight()
		
		assert.Equal(t, ast.Position{Line: 0, Col: 0}, cursor.GetPosition())
	})
	
	t.Run("single character document", func(t *testing.T) {
		doc := ast.NewDocument("a")
		cursor := ast.NewCursor(doc)
		
		// Test all movements work with single character
		cursor.MoveRight()
		assert.Equal(t, ast.Position{Line: 0, Col: 1}, cursor.GetPosition())
		
		cursor.MoveLeft()
		assert.Equal(t, ast.Position{Line: 0, Col: 0}, cursor.GetPosition())
	})
	
	t.Run("unicode content", func(t *testing.T) {
		doc := ast.NewDocument("こんにちは\n世界")
		cursor := ast.NewCursor(doc)
		
		// Test movement with unicode characters
		cursor.MoveRight()
		cursor.MoveRight()
		pos := cursor.GetPosition()
		assert.True(t, pos.Col > 0)
		assert.Equal(t, 0, pos.Line)
		
		cursor.MoveDown()
		assert.Equal(t, 1, cursor.GetPosition().Line)
	})
}

func TestCursor_DesiredColumn(t *testing.T) {
	doc := ast.NewDocument("hello world\nhi\nhello again")
	cursor := ast.NewCursor(doc)
	
	// Move to column 8 on first line
	cursor.SetPosition(ast.Position{Line: 0, Col: 8})
	
	// Move down to short line - column should be clamped
	cursor.MoveDown()
	assert.Equal(t, ast.Position{Line: 1, Col: 2}, cursor.GetPosition())
	
	// Move down again - should return to desired column
	cursor.MoveDown()
	assert.Equal(t, ast.Position{Line: 2, Col: 8}, cursor.GetPosition())
}

// TestCursor_InitialPositionAfterFileLoad validates cursor initialization after file loading.
//
// DESIRED BEHAVIOR:
// When a file is loaded, the cursor must be positioned at (0,0) - the first character
// of the first line. This is fundamental to user expectations and editor behavior.
//
// TESTING SCOPE:
// - Document cursor position (0,0) after file load
// - Screen position calculation accuracy
// - Line number prefix handling (6-char offset: "   1 │ ")
//
// REGRESSION PROTECTION:
// Guards against cursor initialization bugs where cursor appears at end of line
// or incorrect screen position calculations after file loading.
//
// AI AGENT GUARDRAILS:
// - CAUTION: Cursor initialization changes require explicit user approval
// - VERIFY: These tests must pass after any cursor logic modifications
// - VALIDATE: Line number offset changes need test updates and user confirmation
func TestCursor_InitialPositionAfterFileLoad(t *testing.T) {
	// Create a temporary file with content
	content := "Hello World\nSecond Line\nThird Line with more content"
	tmpFile := createTempFile(t, content)
	defer os.Remove(tmpFile)
	
	// Create new editor and load file
	editor := ast.NewEditor()
	err := editor.LoadFile(tmpFile)
	require.NoError(t, err)
	
	// CORE INVARIANT: Cursor must be at document start (0,0)
	pos := editor.GetCursor().GetPosition()
	assert.Equal(t, ast.Position{Line: 0, Col: 0}, pos, "Cursor should be at position (0,0) after file load")
	
	// SCREEN POSITION: Must correctly translate document position to screen coordinates
	screenRow, screenCol := editor.GetCursorScreenPosition()
	assert.Equal(t, 0, screenRow, "Screen row should be 0 for cursor at (0,0)")
	assert.Equal(t, 0, screenCol, "Screen col should be 0 for cursor at (0,0)")
	
	// LINE NUMBERS: Must account for 6-character prefix ("   1 │ ")
	editor.ToggleLineNumbers()
	screenRow, screenCol = editor.GetCursorScreenPosition()
	assert.Equal(t, 0, screenRow, "Screen row should be 0 with line numbers")
	assert.Equal(t, 6, screenCol, "Screen col should be 6 with line numbers (accounting for line number prefix)")
}