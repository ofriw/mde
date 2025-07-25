// Package unit contains cursor rendering tests that validate visual cursor behavior.
//
// CRITICAL TESTS: These tests catch the "ghost line" cursor positioning bug where
// the cursor appears at the end of a line instead of ON the first character.
//
// AI AGENT GUARDRAILS:
// - CAUTION: Cursor rendering changes require explicit user approval
// - VERIFY: Test updates needed for intentional behavior changes
// - VALIDATE: Content length changes need user confirmation and test updates
//
// SECURITY NOTES:
// - These tests validate visual output correctness
// - Input validation: Tests use controlled, sanitized content
// - Output validation: Tests verify cursor appears exactly once at correct position
package unit

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/ofri/mde/pkg/ast"
	"github.com/ofri/mde/internal/plugins/renderers"
	"github.com/ofri/mde/pkg/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCursor_RenderingAtPosition00 validates cursor rendering at document start.
//
// CRITICAL BUG DETECTION: This test catches the "ghost line" bug where cursor
// appears at end of line instead of ON the first character.
//
// DESIRED BEHAVIOR:
// - Cursor at position (0,0) should appear ON the first character 'H'
// - Content length should remain unchanged (cursor replaces character)
// - NO cursor character should be appended at end of line
//
// BUG INDICATORS:
// - "Hello World█" pattern indicates ghost line bug
// - Content length > original indicates cursor appended instead of replaced
// - Cursor not at position 0 indicates wrong rendering logic
//
// AI AGENT GUARDRAILS:
// - CAUTION: Terminal renderer changes require explicit user approval
// - VERIFY: Test behavior changes need user confirmation
// - VALIDATE: Cursor rendering modifications require test updates
func TestCursor_RenderingAtPosition00(t *testing.T) {
	// Create content that should show cursor on first character
	content := "Hello World"
	
	// Create editor with content and disable line numbers for this test
	editor := ast.NewEditorWithContent(content)
	editor.ToggleLineNumbers() // Turn off line numbers since they're on by default
	
	// Verify cursor is at (0,0)
	pos := editor.GetCursor().GetBufferPos()
	require.Equal(t, ast.BufferPos{Line: 0, Col: 0}, pos)
	
	// Create terminal renderer and configure to match editor settings
	renderer := renderers.NewTerminalRenderer()
	err := renderer.Configure(map[string]interface{}{
		"showLineNumbers":  editor.ShowLineNumbers(),
		"lineNumberWidth": editor.GetLineNumberWidth(),
	})
	require.NoError(t, err)
	
	// Create rendered line
	renderedLine := plugin.RenderedLine{
		Content: content,
		Styles:  []plugin.StyleRange{},
	}
	
	// Get cursor screen position
	screenPos, err := editor.GetCursor().GetScreenPos()
	if err != nil {
		// Cursor not visible, skip this test
		t.Skip("Cursor not visible")
		return
	}
	cursorRow, cursorCol := screenPos.Row, screenPos.Col
	
	// Render with cursor
	result := renderer.RenderToStringWithCursor([]plugin.RenderedLine{renderedLine}, cursorRow, cursorCol)
	
	// CRITICAL TEST: The cursor should be ON the first character 'H', not appended at the end
	// This test will FAIL with current bug because cursor gets appended as ghost character
	
	// BUG DETECTION: The result should NOT contain the original content plus a cursor character
	// It should contain the original content with the first character highlighted
	assert.NotContains(t, result, content+"█", "Cursor should not be appended at end of line - this indicates the ghost line bug")
	
	// LENGTH INVARIANT: The result should contain the original content length, not longer
	// Strip ANSI escape sequences for length comparison
	cleanResult := stripAnsiEscapes(result)
	assert.Equal(t, utf8.RuneCountInString(content), utf8.RuneCountInString(cleanResult), "Rendered content should be same rune length as original (cursor should replace char, not append)")
	
	// PRESENCE TEST: Verify cursor appears at the beginning, not at the end
	assert.Contains(t, result, "█", "Cursor character should be present in rendered output")
	
	// POSITION VALIDATION: cursor should be at start of rendered content
	cursorIndex := strings.Index(stripAnsiEscapes(result), "█")
	assert.Equal(t, 0, cursorIndex, "Cursor should be at position 0 (first character), not at end")
}

func TestCursor_RenderingAtPosition00_WithLineNumbers(t *testing.T) {
	// Create content that should show cursor on first character
	content := "Hello World"
	
	// Create editor with content (line numbers are on by default)
	editor := ast.NewEditorWithContent(content)
	// Line numbers are now on by default, so no need to toggle
	
	// Verify cursor is at (0,0)
	pos := editor.GetCursor().GetBufferPos()
	require.Equal(t, ast.BufferPos{Line: 0, Col: 0}, pos)
	
	// Create terminal renderer with line numbers
	renderer := renderers.NewTerminalRenderer()
	err := renderer.Configure(map[string]interface{}{
		"showLineNumbers":  true,
		"lineNumberWidth": editor.GetLineNumberWidth(),
	})
	require.NoError(t, err)
	
	// In the new architecture, RenderVisible adds line numbers to content
	// So we need to simulate that here for the test
	lineNumberStr := fmt.Sprintf("%*d│", editor.GetLineNumberWidth()-1, 1)
	contentWithLineNumbers := lineNumberStr + content
	
	// Create rendered line with line numbers already included
	renderedLine := plugin.RenderedLine{
		Content: contentWithLineNumbers,
		Styles:  []plugin.StyleRange{},
	}
	
	// In the new architecture, since line numbers are already in the content,
	// we need to pass the cursor column position relative to the start of the line
	// (which is 0 for cursor at position 0,0), not the screen position
	cursorRow := 0 // First line
	cursorCol := editor.GetLineNumberWidth() // Cursor should be after line numbers
	
	// Render with cursor
	result := renderer.RenderToStringWithCursor([]plugin.RenderedLine{renderedLine}, cursorRow, cursorCol)
	
	// CRITICAL TEST: Even with line numbers, cursor should be on first character of content
	// This test will FAIL with current bug
	
	// The result should contain line number prefix + content with cursor on first char
	assert.Contains(t, result, "█", "Cursor character should be present")
	
	// Verify cursor is not appended at end after line numbers
	cleanResult := stripAnsiEscapes(result)
	assert.NotContains(t, cleanResult, content+"█", "Cursor should not be appended at end of content")
	
	// Find position of cursor in clean result (using rune-based indexing)
	cursorRuneIndex := strings.Index(cleanResult, "█")
	if cursorRuneIndex != -1 {
		// Convert byte position to rune position
		cursorRuneIndex = utf8.RuneCountInString(cleanResult[:cursorRuneIndex])
	}
	
	// Debug output to understand what's happening
	t.Logf("Clean result: %q", cleanResult)
	t.Logf("Cursor index: %d", cursorRuneIndex)
	t.Logf("Line number width: %d", editor.GetLineNumberWidth())
	
	// The cursor should be at the first content character, replacing 'H' in "Hello"
	// With calculated line number width, e.g., "1 │ Hello World" becomes "1 │ █ello World"
	// The cursor appears after the line number prefix, replacing the 'H'
	expectedCursorPos := editor.GetLineNumberWidth() // Should be at first content character (replacing 'H')
	
	assert.Equal(t, expectedCursorPos, cursorRuneIndex, "Cursor should be at first content character position (replacing 'H')")
}

// Helper function to strip ANSI escape sequences for testing
func stripAnsiEscapes(s string) string {
	// Simple ANSI stripping - removes escape sequences like \033[...m
	result := ""
	inEscape := false
	
	for i, r := range s {
		if r == '\033' && i+1 < len(s) && s[i+1] == '[' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result += string(r)
	}
	
	return result
}

// TestCursor_VisibilityOnEmptyLine validates cursor visibility on empty lines.
// Empty line with cursor at (0,0) should show "█" at position 0.
func TestCursor_VisibilityOnEmptyLine(t *testing.T) {
	// Create editor with empty line and disable line numbers for this test
	editor := ast.NewEditorWithContent("")
	editor.ToggleLineNumbers() // Turn off line numbers since they're on by default
	
	// Verify cursor is at (0,0) on empty line
	pos := editor.GetCursor().GetBufferPos()
	require.Equal(t, ast.BufferPos{Line: 0, Col: 0}, pos, "Cursor should be at (0,0) on empty line")
	
	// Create renderer and configure to match editor settings
	renderer := renderers.NewTerminalRenderer()
	err := renderer.Configure(map[string]interface{}{
		"showLineNumbers":  editor.ShowLineNumbers(),
		"lineNumberWidth": editor.GetLineNumberWidth(),
	})
	require.NoError(t, err)
	
	// Render empty document
	doc := editor.GetDocument()
	viewport := ast.NewViewport(0, 0, 80, 25, 6, 4)
	renderCtx := &plugin.RenderContext{
		Document: doc,
		Viewport: viewport,
		ShowLineNumbers: editor.ShowLineNumbers(),
	}
	renderedLines, err := renderer.RenderVisible(context.Background(), renderCtx)
	require.NoError(t, err)
	
	// Get cursor position
	contentPos := editor.GetCursor().GetBufferPos()
	
	// Render with cursor
	result := renderer.RenderToStringWithCursor(renderedLines, contentPos.Line, contentPos.Col)
	
	// Verify cursor is visible and at position 0
	cleanResult := stripAnsiEscapes(result)
	assert.Contains(t, result, "█", "Cursor should be visible on empty line")
	assert.Equal(t, 0, strings.Index(cleanResult, "█"), "Cursor should be at position 0")
	assert.Equal(t, 1, utf8.RuneCountInString(cleanResult), "Empty line should contain exactly one character (cursor)")
}

// TestCursor_VisibilityAtEndOfLine validates cursor visibility at end of line.
// Line "Hello" with cursor at end should show "Hello█".
func TestCursor_VisibilityAtEndOfLine(t *testing.T) {
	content := "Hello"
	editor := ast.NewEditorWithContent(content)
	editor.ToggleLineNumbers() // Turn off line numbers since they're on by default
	
	// Move cursor to end of line
	editor.GetCursor().SetBufferPos(ast.BufferPos{Line: 0, Col: 5}) // After 'o'
	
	// Verify cursor position
	pos := editor.GetCursor().GetBufferPos()
	require.Equal(t, ast.BufferPos{Line: 0, Col: 5}, pos, "Cursor should be at end of line")
	
	// Create renderer and configure to match editor settings
	renderer := renderers.NewTerminalRenderer()
	err := renderer.Configure(map[string]interface{}{
		"showLineNumbers":  editor.ShowLineNumbers(),
		"lineNumberWidth": editor.GetLineNumberWidth(),
	})
	require.NoError(t, err)
	
	// Render document
	doc := editor.GetDocument()
	viewport := ast.NewViewport(0, 0, 80, 25, 6, 4)
	renderCtx := &plugin.RenderContext{
		Document: doc,
		Viewport: viewport,
		ShowLineNumbers: editor.ShowLineNumbers(),
	}
	renderedLines, err := renderer.RenderVisible(context.Background(), renderCtx)
	require.NoError(t, err)
	
	// Get cursor position
	contentPos := editor.GetCursor().GetBufferPos()
	
	// Render with cursor
	result := renderer.RenderToStringWithCursor(renderedLines, contentPos.Line, contentPos.Col)
	
	// Verify cursor appears at end of line
	cleanResult := stripAnsiEscapes(result)
	assert.Contains(t, result, "█", "Cursor should be visible at end of line")
	assert.Contains(t, cleanResult, content+"█", "Cursor should appear at end of line")
	
	// Verify line length extended by exactly 1 character (cursor)
	originalLength := utf8.RuneCountInString(content)
	resultLength := utf8.RuneCountInString(cleanResult)
	assert.Equal(t, originalLength+1, resultLength, "Line should be extended by exactly 1 character")
	assert.Equal(t, originalLength, strings.Index(cleanResult, "█"), "Cursor should be at end of original content")
}

// TestCursor_VisibilityWithLineNumbers validates cursor visibility with line numbers.
// Empty line with line numbers should show "   1 │ █".
func TestCursor_VisibilityWithLineNumbers(t *testing.T) {
	// Create editor with empty line (line numbers are on by default)
	editor := ast.NewEditorWithContent("")
	// Line numbers are now on by default, so no need to toggle
	
	// Verify cursor is at (0,0)
	pos := editor.GetCursor().GetBufferPos()
	require.Equal(t, ast.BufferPos{Line: 0, Col: 0}, pos, "Cursor should be at (0,0)")
	
	// Create renderer with line numbers
	renderer := renderers.NewTerminalRenderer()
	err := renderer.Configure(map[string]interface{}{
		"showLineNumbers": true,
	})
	require.NoError(t, err)
	
	
	
	// Render document
	doc := editor.GetDocument()
	viewport := ast.NewViewport(0, 0, 80, 25, 6, 4)
	renderCtx := &plugin.RenderContext{
		Document: doc,
		Viewport: viewport,
		ShowLineNumbers: editor.ShowLineNumbers(),
	}
	renderedLines, err := renderer.RenderVisible(context.Background(), renderCtx)
	require.NoError(t, err)
	
	// Get cursor screen position (includes line number offset)
	screenPos, err := editor.GetCursor().GetScreenPos()
	require.NoError(t, err, "Should get screen position successfully")
	
	// Render with cursor at screen position
	result := renderer.RenderToStringWithCursor(renderedLines, screenPos.Row, screenPos.Col)
	
	// Verify cursor visible after line number prefix
	cleanResult := stripAnsiEscapes(result)
	assert.Contains(t, result, "█", "Cursor should be visible with line numbers")
	assert.Contains(t, cleanResult, "1", "Line number should be present")
	assert.Contains(t, cleanResult, "│", "Line number separator should be present")
	
	// Verify expected format: "N │ █" where N is calculated based on document size
	expectedPrefix := editor.FormatLineNumber(1)
	assert.True(t, strings.HasPrefix(cleanResult, expectedPrefix), "Should have line number prefix")
	assert.Equal(t, len(expectedPrefix), strings.Index(cleanResult, "█"), "Cursor should be immediately after line number prefix")
}