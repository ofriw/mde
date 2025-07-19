// Package integration contains end-to-end cursor behavior tests.
//
// CRITICAL REGRESSION TESTS: These tests catch the "ghost line" cursor bug
// through full TUI integration testing without external frameworks.
//
// AI AGENT GUARDRAILS:
// - CAUTION: TUI rendering changes require explicit user approval
// - VERIFY: Integration test updates needed for intentional behavior changes
// - VALIDATE: Visual output modifications require user confirmation
//
// SECURITY & ROBUSTNESS:
// - Tests use direct TUI model testing (no external dependencies)
// - Input validation: Controlled test content prevents injection
// - Output validation: Multiple assertions verify correct cursor behavior
package integration

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/ofri/mde/internal/plugins"
	"github.com/ofri/mde/internal/tui"
	"github.com/ofri/mde/pkg/plugin"
	"github.com/ofri/mde/test/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTUICursor_InitialPositionGhostLineBug validates full TUI cursor rendering behavior.
//
// CRITICAL BUG REPRODUCTION: This test reproduces the exact "ghost line" bug
// reported where cursor appears at end of line instead of first character.
//
// DESIRED BEHAVIOR:
// - Cursor should appear ON the first character of loaded content
// - Content length should not be extended by cursor rendering
// - Cursor should be visible in TUI output
// - No "Hello World█" pattern should appear (ghost line indicator)
//
// INTEGRATION SCOPE:
// - Full TUI model rendering pipeline
// - Both with and without line numbers
// - Real content loading and display
//
// AI AGENT GUARDRAILS:
// - CAUTION: TUI pipeline changes require explicit user approval
// - VERIFY: These tests must pass after any TUI modifications
// - VALIDATE: Behavior changes need user confirmation and test updates
func TestTUICursor_InitialPositionGhostLineBug(t *testing.T) {
	// CRITICAL TEST: This test reproduces the exact bug reported
	// When opening a file, cursor should be at first character, not at end with ghost line
	
	// Initialize plugins once for the entire test
	err := plugins.InitializePlugins()
	require.NoError(t, err, "Should initialize plugins successfully")
	
	t.Run("cursor at start without ghost line", func(t *testing.T) {
		model := tui.New()
		// Set proper dimensions for rendering
		testutils.SetModelSize(model, 80, 24)
		// Load content that should show cursor at position (0,0) on first character
		testutils.LoadContentIntoModel(model, "Hello World\nSecond Line")
		
		// Ensure line numbers are disabled for this test
		editor := model.GetEditor()
		if editor.ShowLineNumbers() {
			editor.ToggleLineNumbers()
		}
		
		// Reset renderer configuration to ensure clean state
		registry := plugin.GetRegistry()
		if renderer, err := registry.GetDefaultRenderer(); err == nil {
			renderer.Configure(map[string]interface{}{
				"showLineNumbers": false,
			})
		}
		
		// Test initial cursor position directly
		pos := editor.GetCursor().GetBufferPos()
		
		// Verify cursor is at (0,0)
		assert.Equal(t, 0, pos.Line, "Cursor should be at line 0")
		assert.Equal(t, 0, pos.Col, "Cursor should be at column 0")
		
		// Test the View() method to get rendered output
		output := model.View()
		
		// CRITICAL: This test will FAIL with current bug
		// The cursor should be ON the 'H' character, not appended at end
		
		// Check that content is present (cursor replaces first character)
		assert.Contains(t, output, "ello World", "Content should be present with cursor replacing first character")
		
		// Check that there's no ghost line artifact
		// Ghost line bug: cursor appears at end of line instead of on first character
		assert.NotContains(t, output, "Hello World█", "GHOST LINE BUG: Cursor should not be appended at end of line")
		
		// The cursor should be visible somewhere in the output
		assert.Contains(t, output, "█", "Cursor should be visible in output")
		
		// Additional check: if we find the cursor, it should be at the beginning of content
		lines := strings.Split(output, "\n")
		if len(lines) > 0 {
			firstLine := strings.TrimRight(testutils.StripAnsiEscapes(lines[0]), " ")
			// First line should be exactly "Hello World" length with cursor replacing first char
			// It should NOT be "Hello World█" (cursor appended)
			// Use rune count to handle Unicode cursor character
			assert.True(t, utf8.RuneCountInString(firstLine) <= 11, "First line should not be longer than original content due to cursor")
		}
	})
	
	t.Run("cursor at start with line numbers", func(t *testing.T) {
		model := tui.New()
		// Set proper dimensions for rendering
		testutils.SetModelSize(model, 80, 24)
		testutils.LoadContentIntoModel(model, "Hello World\nSecond Line")
		
		// Line numbers are now on by default, so no need to toggle
		
		// Ensure renderer configuration is synchronized with editor
		registry := plugin.GetRegistry()
		if renderer, err := registry.GetDefaultRenderer(); err == nil {
			renderer.Configure(map[string]interface{}{
				"showLineNumbers": true,
			})
		}
		
		// Test initial cursor position
		editor := model.GetEditor()
		pos := editor.GetCursor().GetBufferPos()
		assert.Equal(t, 0, pos.Line, "Cursor should be at line 0")
		assert.Equal(t, 0, pos.Col, "Cursor should be at column 0")
		
		// Test the View() method to get rendered output
		output := model.View()
		
		// Check that content is present (cursor replaces first character)
		assert.Contains(t, output, "ello World", "Content should be present with line numbers and cursor replacing first character")
		
		// Check that line numbers are present
		assert.Contains(t, output, "1", "Line numbers should be present")
		
		// CRITICAL: Even with line numbers, cursor should be on first character of content
		// not appended at end of line
		assert.NotContains(t, output, "Hello World█", "GHOST LINE BUG: Cursor should not be appended at end with line numbers")
		
		// The cursor should be visible
		assert.Contains(t, output, "█", "Cursor should be visible with line numbers")
		
		// Additional check: content length with line numbers
		lines := strings.Split(output, "\n")
		if len(lines) > 0 {
			firstLine := strings.TrimRight(testutils.StripAnsiEscapes(lines[0]), " ")
			// With line numbers, format should be "   1 │ Hello World" with cursor on H
			// NOT "   1 │ Hello World█" (cursor appended)
			// Use rune count to handle Unicode cursor character
			expectedMaxLength := 7 + 11 // line number prefix (7 runes) + content length (11 runes)
			assert.True(t, utf8.RuneCountInString(firstLine) <= expectedMaxLength, "First line with line numbers should not be extended by cursor")
		}
	})
}

func TestTUICursor_BasicMovement(t *testing.T) {
	// Initialize plugins for TUI rendering
	err := plugins.InitializePlugins()
	require.NoError(t, err, "Should initialize plugins successfully")
	
	// Create a test model with content
	model := tui.New()
	testutils.SetModelSize(model, 80, 24)
	testutils.LoadContentIntoModel(model, "line1\nline2\nline3")
	
	// Test basic cursor movement using the model directly
	editor := model.GetEditor()
	// Ensure line numbers are disabled for this test
	if editor.ShowLineNumbers() {
		editor.ToggleLineNumbers()
	}
	
	// Reset renderer configuration to ensure clean state
	registry := plugin.GetRegistry()
	if renderer, err := registry.GetDefaultRenderer(); err == nil {
		renderer.Configure(map[string]interface{}{
			"showLineNumbers": false,
		})
	}
	
	cursor := editor.GetCursor()
	
	// Test cursor movement
	editor.MoveCursorRight()
	editor.MoveCursorRight()
	editor.MoveCursorDown()
	
	// Verify cursor position
	pos := cursor.GetBufferPos()
	assert.Equal(t, 1, pos.Line, "Cursor should be on line 1 after moving down")
	assert.Equal(t, 2, pos.Col, "Cursor should be at column 2 after moving right twice")
	
	// Test the view output
	output := model.View()
	assert.Contains(t, output, "line1")
	assert.Contains(t, output, "li█e2") // Cursor should replace the 'n' at position (1,2)
	assert.Contains(t, output, "line3")
}