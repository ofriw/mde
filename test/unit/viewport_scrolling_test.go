package unit

import (
	"testing"
	
	"github.com/ofri/mde/pkg/ast"
	"github.com/stretchr/testify/assert"
)

func TestViewportScrolling(t *testing.T) {
	t.Run("scroll viewport without moving cursor", func(t *testing.T) {
		// Create editor with content
		content := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10"
		editor := ast.NewEditorWithContent(content)
		editor.SetViewPort(80, 5) // Small viewport to test scrolling
		
		// Position cursor at line 3
		editor.MoveCursorDown()
		editor.MoveCursorDown()
		initialCursorPos := editor.GetCursorBufferPosition()
		assert.Equal(t, ast.BufferPos{Line: 2, Col: 0}, initialCursorPos)
		
		// Get initial viewport position
		initialViewport := editor.GetViewport()
		assert.Equal(t, 0, initialViewport.GetTopLine())
		
		// Scroll viewport down
		editor.ScrollViewportDown(3)
		
		// Check viewport moved
		scrolledViewport := editor.GetViewport()
		assert.Equal(t, 3, scrolledViewport.GetTopLine())
		
		// Check cursor stayed at same buffer position
		cursorAfterScroll := editor.GetCursorBufferPosition()
		assert.Equal(t, initialCursorPos, cursorAfterScroll)
		
		// Scroll viewport up
		editor.ScrollViewportUp(2)
		
		// Check viewport moved up
		finalViewport := editor.GetViewport()
		assert.Equal(t, 1, finalViewport.GetTopLine())
		
		// Cursor should still be at same position
		finalCursorPos := editor.GetCursorBufferPosition()
		assert.Equal(t, initialCursorPos, finalCursorPos)
	})
	
	t.Run("viewport scrolling limits", func(t *testing.T) {
		content := "Line 1\nLine 2\nLine 3"
		editor := ast.NewEditorWithContent(content)
		editor.SetViewPort(80, 5)
		
		// Try to scroll up past beginning
		editor.ScrollViewportUp(10)
		viewport := editor.GetViewport()
		assert.Equal(t, 0, viewport.GetTopLine(), "Should not scroll past beginning")
		
		// Try to scroll down past end
		editor.ScrollViewportDown(10)
		viewport = editor.GetViewport()
		assert.Equal(t, 2, viewport.GetTopLine(), "Should not scroll past last line")
	})
	
	t.Run("horizontal viewport scrolling", func(t *testing.T) {
		content := "This is a very long line that requires horizontal scrolling to see the entire content"
		editor := ast.NewEditorWithContent(content)
		editor.SetViewPort(40, 5) // Narrow viewport
		
		// Get initial viewport
		initialViewport := editor.GetViewport()
		assert.Equal(t, 0, initialViewport.GetLeftColumn())
		
		// Scroll right
		editor.ScrollViewportRight(10)
		scrolledViewport := editor.GetViewport()
		assert.Equal(t, 10, scrolledViewport.GetLeftColumn())
		
		// Scroll left
		editor.ScrollViewportLeft(5)
		finalViewport := editor.GetViewport()
		assert.Equal(t, 5, finalViewport.GetLeftColumn())
		
		// Try to scroll left past beginning
		editor.ScrollViewportLeft(20)
		viewport := editor.GetViewport()
		assert.Equal(t, 0, viewport.GetLeftColumn())
	})
}