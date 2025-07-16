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
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/ofri/mde/pkg/ast"
	"github.com/ofri/mde/internal/plugins/renderers"
	"github.com/ofri/mde/internal/plugins/themes"
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
	
	// Create editor with content
	editor := ast.NewEditorWithContent(content)
	
	// Verify cursor is at (0,0)
	pos := editor.GetCursor().GetPosition()
	require.Equal(t, ast.Position{Line: 0, Col: 0}, pos)
	
	// Create terminal renderer
	renderer := renderers.NewTerminalRenderer()
	
	// Create theme
	theme := &themes.DarkTheme{}
	
	// Create rendered line
	renderedLine := plugin.RenderedLine{
		Content: content,
		Styles:  []plugin.StyleRange{},
	}
	
	// Get cursor screen position
	cursorRow, cursorCol := editor.GetCursorScreenPosition()
	
	// Render with cursor
	result := renderer.RenderToStringWithCursor([]plugin.RenderedLine{renderedLine}, theme, cursorRow, cursorCol)
	
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
	
	// Create editor with content and enable line numbers
	editor := ast.NewEditorWithContent(content)
	editor.ToggleLineNumbers()
	
	// Verify cursor is at (0,0)
	pos := editor.GetCursor().GetPosition()
	require.Equal(t, ast.Position{Line: 0, Col: 0}, pos)
	
	// Create terminal renderer with line numbers
	renderer := renderers.NewTerminalRenderer()
	err := renderer.Configure(map[string]interface{}{
		"showLineNumbers": true,
	})
	require.NoError(t, err)
	
	// Create theme
	theme := &themes.DarkTheme{}
	
	// Create rendered line
	renderedLine := plugin.RenderedLine{
		Content: content,
		Styles:  []plugin.StyleRange{},
	}
	
	// Get cursor screen position (should account for line numbers)
	cursorRow, cursorCol := editor.GetCursorScreenPosition()
	
	// Render with cursor
	result := renderer.RenderToStringWithCursor([]plugin.RenderedLine{renderedLine}, theme, cursorRow, cursorCol)
	
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
	
	lineNumPrefixLen := 7 // "   1 │ " (7 runes)
	expectedCursorPos := lineNumPrefixLen // Should be right after line number prefix
	assert.Equal(t, expectedCursorPos, cursorRuneIndex, "Cursor should be at first character position after line number prefix")
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
	// Create editor with empty line
	editor := ast.NewEditorWithContent("")
	
	// Verify cursor is at (0,0) on empty line
	pos := editor.GetCursor().GetPosition()
	require.Equal(t, ast.Position{Line: 0, Col: 0}, pos, "Cursor should be at (0,0) on empty line")
	
	// Create renderer and theme
	renderer := renderers.NewTerminalRenderer()
	theme := themes.NewDarkTheme()
	
	// Render empty document
	doc := editor.GetDocument()
	renderedLines, err := renderer.Render(context.Background(), doc, theme)
	require.NoError(t, err)
	
	// Get cursor position
	contentPos := editor.GetCursorContentPosition()
	
	// Render with cursor
	result := renderer.RenderToStringWithCursor(renderedLines, theme, contentPos.Line, contentPos.Col)
	
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
	
	// Move cursor to end of line
	editor.GetCursor().SetPosition(ast.Position{Line: 0, Col: 5}) // After 'o'
	
	// Verify cursor position
	pos := editor.GetCursor().GetPosition()
	require.Equal(t, ast.Position{Line: 0, Col: 5}, pos, "Cursor should be at end of line")
	
	// Create renderer and theme
	renderer := renderers.NewTerminalRenderer()
	theme := themes.NewDarkTheme()
	
	// Render document
	doc := editor.GetDocument()
	renderedLines, err := renderer.Render(context.Background(), doc, theme)
	require.NoError(t, err)
	
	// Get cursor position
	contentPos := editor.GetCursorContentPosition()
	
	// Render with cursor
	result := renderer.RenderToStringWithCursor(renderedLines, theme, contentPos.Line, contentPos.Col)
	
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
	// Create editor with empty line and line numbers
	editor := ast.NewEditorWithContent("")
	editor.ToggleLineNumbers()
	
	// Verify cursor is at (0,0)
	pos := editor.GetCursor().GetPosition()
	require.Equal(t, ast.Position{Line: 0, Col: 0}, pos, "Cursor should be at (0,0)")
	
	// Create renderer with line numbers
	renderer := renderers.NewTerminalRenderer()
	err := renderer.Configure(map[string]interface{}{
		"showLineNumbers": true,
	})
	require.NoError(t, err)
	
	theme := themes.NewDarkTheme()
	
	// Render document
	doc := editor.GetDocument()
	renderedLines, err := renderer.Render(context.Background(), doc, theme)
	require.NoError(t, err)
	
	// Get cursor position (should include line number offset)
	contentPos := editor.GetCursorContentPosition()
	
	// Render with cursor
	result := renderer.RenderToStringWithCursor(renderedLines, theme, contentPos.Line, contentPos.Col)
	
	// Verify cursor visible after line number prefix
	cleanResult := stripAnsiEscapes(result)
	assert.Contains(t, result, "█", "Cursor should be visible with line numbers")
	assert.Contains(t, cleanResult, "1", "Line number should be present")
	assert.Contains(t, cleanResult, "│", "Line number separator should be present")
	
	// Verify expected format: "   1 │ █"
	expectedPrefix := "   1 │ "
	assert.True(t, strings.HasPrefix(cleanResult, expectedPrefix), "Should have line number prefix")
	assert.Equal(t, len(expectedPrefix), strings.Index(cleanResult, "█"), "Cursor should be immediately after line number prefix")
}