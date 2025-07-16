package unit

import (
	"testing"

	"github.com/ofri/mde/pkg/ast"
	"github.com/stretchr/testify/assert"
)

func TestCoordinateTypes(t *testing.T) {
	t.Run("DocumentPos", func(t *testing.T) {
		pos := ast.DocumentPos{Line: 1, Col: 5}
		
		assert.Equal(t, "DocumentPos{Line:1, Col:5}", pos.String())
		assert.True(t, pos.IsValid())
		
		// Test invalid position
		invalidPos := ast.DocumentPos{Line: -1, Col: 5}
		assert.False(t, invalidPos.IsValid())
	})
	
	t.Run("ContentPos", func(t *testing.T) {
		pos := ast.ContentPos{Line: 1, Col: 11}
		
		assert.Equal(t, "ContentPos{Line:1, Col:11}", pos.String())
		assert.True(t, pos.IsValid())
		
		// Test invalid position
		invalidPos := ast.ContentPos{Line: 1, Col: -5}
		assert.False(t, invalidPos.IsValid())
	})
	
	t.Run("ScreenPos", func(t *testing.T) {
		pos := ast.ScreenPos{Row: 2, Col: 15}
		
		assert.Equal(t, "ScreenPos{Row:2, Col:15}", pos.String())
		assert.True(t, pos.IsValid())
		
		// Test invalid position
		invalidPos := ast.ScreenPos{Row: 2, Col: -1}
		assert.False(t, invalidPos.IsValid())
	})
}

func TestCoordinateTransformation(t *testing.T) {
	editor := ast.NewEditorWithContent("line1\nline2\nline3")
	
	t.Run("without line numbers", func(t *testing.T) {
		// Line numbers are disabled by default
		
		docPos := ast.DocumentPos{Line: 1, Col: 2}
		contentPos := editor.TransformDocumentToContent(docPos)
		
		// Without line numbers, content position should equal document position
		assert.Equal(t, 1, contentPos.Line)
		assert.Equal(t, 2, contentPos.Col)
	})
	
	t.Run("with line numbers", func(t *testing.T) {
		editor.ToggleLineNumbers() // Enable line numbers
		
		docPos := ast.DocumentPos{Line: 1, Col: 2}
		contentPos := editor.TransformDocumentToContent(docPos)
		
		// With line numbers, content position should have +6 column offset
		assert.Equal(t, 1, contentPos.Line)
		assert.Equal(t, 8, contentPos.Col) // 2 + 6 for line numbers
		
		editor.ToggleLineNumbers() // Disable for next test
	})
	
	t.Run("with viewport offset", func(t *testing.T) {
		editor.SetViewPort(80, 24)
		
		// This test is skipped because we can't directly modify viewport offset
		// The viewport is managed internally by the editor
		t.Skip("Viewport offset manipulation not supported in current API")
	})
}

func TestCoordinateValidation(t *testing.T) {
	editor := ast.NewEditorWithContent("line1\nline2\nline3")
	
	t.Run("valid document position", func(t *testing.T) {
		pos := ast.DocumentPos{Line: 1, Col: 2}
		err := editor.ValidateDocumentPos(pos)
		assert.NoError(t, err)
	})
	
	t.Run("invalid document position - negative line", func(t *testing.T) {
		pos := ast.DocumentPos{Line: -1, Col: 2}
		err := editor.ValidateDocumentPos(pos)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "negative coordinates")
	})
	
	t.Run("invalid document position - line out of bounds", func(t *testing.T) {
		pos := ast.DocumentPos{Line: 10, Col: 2}
		err := editor.ValidateDocumentPos(pos)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "line 10 >= document line count")
	})
	
	t.Run("invalid document position - column out of bounds", func(t *testing.T) {
		pos := ast.DocumentPos{Line: 1, Col: 100}
		err := editor.ValidateDocumentPos(pos)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "column 100 > line length")
	})
	
	t.Run("valid content position without line numbers", func(t *testing.T) {
		// Line numbers are disabled by default
		editor.SetViewPort(80, 24)
		
		pos := ast.ContentPos{Line: 1, Col: 5}
		err := editor.ValidateContentPos(pos)
		assert.NoError(t, err)
	})
	
	t.Run("valid content position with line numbers", func(t *testing.T) {
		editor.ToggleLineNumbers() // Enable line numbers
		editor.SetViewPort(80, 24)
		
		pos := ast.ContentPos{Line: 1, Col: 11} // 5 + 6 for line numbers
		err := editor.ValidateContentPos(pos)
		assert.NoError(t, err)
		
		editor.ToggleLineNumbers() // Disable for next test
	})
	
	t.Run("invalid content position - missing line number offset", func(t *testing.T) {
		editor.ToggleLineNumbers() // Enable line numbers
		editor.SetViewPort(80, 24)
		
		pos := ast.ContentPos{Line: 1, Col: 3} // < 6 but line numbers enabled
		err := editor.ValidateContentPos(pos)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing line number offset")
		
		editor.ToggleLineNumbers() // Disable for next test
	})
}

func TestCursorCoordinateMethods(t *testing.T) {
	editor := ast.NewEditorWithContent("line1\nline2\nline3")
	
	t.Run("GetCursorDocumentPosition", func(t *testing.T) {
		// Move cursor to position (1, 3) in document
		editor.GetCursor().SetPosition(ast.Position{Line: 1, Col: 3})
		
		docPos := editor.GetCursorDocumentPosition()
		assert.Equal(t, 1, docPos.Line)
		assert.Equal(t, 3, docPos.Col)
	})
	
	t.Run("GetCursorContentPosition without line numbers", func(t *testing.T) {
		// Line numbers are disabled by default
		editor.GetCursor().SetPosition(ast.Position{Line: 1, Col: 3})
		
		contentPos := editor.GetCursorContentPosition()
		assert.Equal(t, 1, contentPos.Line)
		assert.Equal(t, 3, contentPos.Col)
	})
	
	t.Run("GetCursorContentPosition with line numbers", func(t *testing.T) {
		editor.ToggleLineNumbers() // Enable line numbers
		editor.GetCursor().SetPosition(ast.Position{Line: 1, Col: 3})
		
		contentPos := editor.GetCursorContentPosition()
		assert.Equal(t, 1, contentPos.Line)
		assert.Equal(t, 9, contentPos.Col) // 3 + 6 for line numbers
		
		editor.ToggleLineNumbers() // Disable for next test
	})
}

func TestViewportInfo(t *testing.T) {
	editor := ast.NewEditor()
	editor.SetViewPort(80, 24)
	editor.ToggleLineNumbers() // Enable line numbers
	
	info := editor.GetViewportInfo()
	
	assert.Equal(t, 0, info.Top)
	assert.Equal(t, 0, info.Left)
	assert.Equal(t, 80, info.Width)
	assert.Equal(t, 24, info.Height)
	assert.True(t, info.LineNumbers)
	
	expectedString := "ViewportInfo{Top:0, Left:0, Width:80, Height:24, LineNumbers:true}"
	assert.Equal(t, expectedString, info.String())
}

func TestCoordinateErrors(t *testing.T) {
	t.Run("DocumentCoordinateError", func(t *testing.T) {
		pos := ast.DocumentPos{Line: -1, Col: 5}
		err := ast.NewDocumentCoordinateError(pos, "test error")
		
		assert.Equal(t, "coordinate error in document position DocumentPos{Line:-1, Col:5}: test error", err.Error())
	})
	
	t.Run("ContentCoordinateError", func(t *testing.T) {
		pos := ast.ContentPos{Line: 1, Col: 3}
		err := ast.NewContentCoordinateError(pos, "missing offset")
		
		assert.Equal(t, "coordinate error in content position ContentPos{Line:1, Col:3}: missing offset", err.Error())
	})
}

func TestCoordinateTransformationConsistency(t *testing.T) {
	editor := ast.NewEditorWithContent("line1\nline2\nline3")
	
	t.Run("transformation consistency", func(t *testing.T) {
		// Test various document positions
		testCases := []ast.DocumentPos{
			{Line: 0, Col: 0},
			{Line: 0, Col: 5},
			{Line: 1, Col: 2},
			{Line: 2, Col: 0},
		}
		
		for _, useLineNumbers := range []bool{false, true} {
			// Set line numbers state
			currentLineNumbers := editor.ShowLineNumbers()
			if useLineNumbers != currentLineNumbers {
				editor.ToggleLineNumbers()
			}
			
			for _, docPos := range testCases {
				// Skip invalid positions
				if err := editor.ValidateDocumentPos(docPos); err != nil {
					continue
				}
				
				// Transform document → content
				contentPos := editor.TransformDocumentToContent(docPos)
				
				// Validate content position
				err := editor.ValidateContentPos(contentPos)
				assert.NoError(t, err, "Content position should be valid for docPos=%s, lineNumbers=%t", docPos, useLineNumbers)
				
				// Verify line number offset
				if useLineNumbers {
					assert.True(t, contentPos.Col >= 6, "Content position with line numbers should have col >= 6")
				}
			}
		}
	})
}