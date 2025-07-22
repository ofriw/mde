package unit

import (
	"context"
	"strings"
	"testing"
	
	"github.com/ofri/mde/internal/plugins/renderers"
	"github.com/ofri/mde/pkg/ast"
	"github.com/ofri/mde/pkg/plugin"
	"github.com/stretchr/testify/assert"
)

// TestNoDuplicateLineNumbers verifies that line numbers are not duplicated
// when rendering with cursor. This tests the fix for the bug where line numbers
// appeared twice in the viewport area containing the cursor.
func TestNoDuplicateLineNumbers(t *testing.T) {
	// Create a document with multiple lines
	doc := ast.NewDocument("Line 1\nLine 2\nLine 3\nLine 4\nLine 5")
	
	// Create viewport showing lines 2-4 (indices 1-3)
	viewport := ast.NewViewport(1, 0, 80, 3, 4, 4)
	
	// Create and configure renderer
	renderer := renderers.NewTerminalRenderer()
	err := renderer.Configure(map[string]interface{}{
		"showLineNumbers": true,
		"lineNumberWidth": 4,
	})
	assert.NoError(t, err)
	
	// Create render context
	renderCtx := &plugin.RenderContext{
		Document:        doc,
		Viewport:        viewport,
		ShowLineNumbers: true,
	}
	
	// Render visible lines
	lines, err := renderer.RenderVisible(context.Background(), renderCtx)
	assert.NoError(t, err)
	assert.Len(t, lines, 3, "Should render 3 visible lines")
	
	// Verify line numbers are in the content
	assert.Contains(t, lines[0].Content, "2│", "First visible line should have line number 2")
	assert.Contains(t, lines[1].Content, "3│", "Second visible line should have line number 3")
	assert.Contains(t, lines[2].Content, "4│", "Third visible line should have line number 4")
	
	// Render with cursor on second visible line (line 3 in document)
	cursorRow := 1  // Second line in rendered output
	cursorCol := 4  // After line number "  3│"
	
	output := renderer.RenderToStringWithCursor(lines, cursorRow, cursorCol)
	
	// Count occurrences of each line number
	line2Count := strings.Count(output, "2│")
	line3Count := strings.Count(output, "3│")
	line4Count := strings.Count(output, "4│")
	
	// Each line number should appear exactly once
	assert.Equal(t, 1, line2Count, "Line number 2 should appear exactly once")
	assert.Equal(t, 1, line3Count, "Line number 3 should appear exactly once")
	assert.Equal(t, 1, line4Count, "Line number 4 should appear exactly once")
	
	// Verify cursor is present
	assert.Contains(t, output, "█", "Cursor should be present in output")
	
	// Verify the cursor is on line 3 after the line number
	outputLines := strings.Split(output, "\n")
	assert.Contains(t, outputLines[1], "3│█", "Cursor should be on line 3 after line number")
}