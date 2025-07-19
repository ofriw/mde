package integration

import (
	"io/ioutil"
	"os"
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

// Helper function to create temporary files for testing
func createTempFile(t *testing.T, content string) string {
	tmpFile, err := ioutil.TempFile("", "file_opening_test_*.txt")
	require.NoError(t, err)
	
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	
	err = tmpFile.Close()
	require.NoError(t, err)
	
	return tmpFile.Name()
}


func TestFileOpening_RealFileLoadWorkflow(t *testing.T) {
	// CRITICAL TEST: This test reproduces the exact bug by using the real file loading workflow
	// This bypasses the testutils.LoadContentIntoModel helper which masks the bug
	
	// Initialize plugins once for the entire test
	err := plugins.InitializePlugins()
	require.NoError(t, err, "Should initialize plugins successfully")
	
	t.Run("cursor position after real file load", func(t *testing.T) {
		// Create a temporary file
		content := "Hello World\nSecond Line\nThird Line"
		tmpFile := createTempFile(t, content)
		defer os.Remove(tmpFile)
		
		// Create TUI model
		model := tui.New()
		
		// Use the real file loading mechanism (not test helper)
		model.SetFilename(tmpFile)
		
		// This is the critical test - cursor should be at (0,0) after file load
		editor := model.GetEditor()
		pos := editor.GetCursor().GetBufferPos()
		
		// Test the actual cursor position
		assert.Equal(t, 0, pos.Line, "Cursor should be at line 0 after file load")
		assert.Equal(t, 0, pos.Col, "Cursor should be at column 0 after file load")
		
		// Test the screen position calculation (line numbers are on by default)
		screenRow, screenCol := func() (int, int) {
		screenPos, err := editor.GetCursor().GetScreenPos()
		if err != nil { return -1, -1 }
		return screenPos.Row, screenPos.Col
	}()
		assert.Equal(t, 0, screenRow, "Screen row should be 0 for cursor at (0,0)")
		expectedCol := editor.GetLineNumberWidth() // Line numbers are on by default
		assert.Equal(t, expectedCol, screenCol, "Screen col should include line number offset")
		
		// Test with line numbers disabled
		editor.ToggleLineNumbers() // Turn OFF line numbers since they're on by default
		screenRow, screenCol = func() (int, int) {
		screenPos, err := editor.GetCursor().GetScreenPos()
		if err != nil { return -1, -1 }
		return screenPos.Row, screenPos.Col
	}()
		assert.Equal(t, 0, screenRow, "Screen row should be 0 without line numbers")
		assert.Equal(t, 0, screenCol, "Screen col should be 0 without line numbers")
	})
	
	t.Run("cursor rendering after real file load", func(t *testing.T) {
		// Create a temporary file
		content := "Hello World\nSecond Line"
		tmpFile := createTempFile(t, content)
		defer os.Remove(tmpFile)
		
		// Create TUI model and load file
		model := tui.New()
		// Set proper dimensions for rendering
		testutils.SetModelSize(model, 80, 24)
		model.SetFilename(tmpFile)
		
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
		
		// Test TUI rendering with the real file loading workflow
		// Use the View() method directly instead of teatest
		model.Update(nil) // Initialize the model
		output := model.View()
		
		// CRITICAL: This test will FAIL with current bug
		// The cursor should be ON the 'H' character, not appended at end
		
		// Check that content is present (cursor replaces first character)
		assert.Contains(t, output, "ello World", "File content should be present with cursor replacing first character")
		
		// CRITICAL TEST: Check for ghost line artifact
		// This is the exact bug - cursor should not be appended at end of line
		assert.NotContains(t, output, "Hello World█", "GHOST LINE BUG: Cursor should not be appended at end of line")
		
		// The cursor should be visible somewhere in the output
		assert.Contains(t, output, "█", "Cursor should be visible in output after file load")
		
		// Additional verification: check content length
		lines := strings.Split(output, "\n")
		if len(lines) > 0 {
			firstLine := strings.TrimRight(testutils.StripAnsiEscapes(lines[0]), " ")
			// First line should be exactly "Hello World" length with cursor replacing first char
			// It should NOT be "Hello World█" (cursor appended)
			assert.True(t, utf8.RuneCountInString(firstLine) <= 11, "First line should not be longer than original content")
		}
	})
	
	t.Run("cursor rendering with line numbers after real file load", func(t *testing.T) {
		// Create a temporary file
		content := "Hello World\nSecond Line"
		tmpFile := createTempFile(t, content)
		defer os.Remove(tmpFile)
		
		// Create TUI model and load file
		model := tui.New()
		// Set proper dimensions for rendering
		testutils.SetModelSize(model, 80, 24)
		model.SetFilename(tmpFile)
		
		// Enable line numbers
		model.GetEditor().ToggleLineNumbers()
		
		// Ensure renderer configuration is synchronized with editor
		registry := plugin.GetRegistry()
		if renderer, err := registry.GetDefaultRenderer(); err == nil {
			renderer.Configure(map[string]interface{}{
				"showLineNumbers": true,
			})
		}
		
		// Test TUI rendering
		model.Update(nil) // Initialize the model
		output := model.View()
		
		// Check that content is present (cursor replaces first character)
		assert.Contains(t, output, "ello World", "File content should be present with cursor replacing first character")
		
		// Check that line numbers are present
		assert.Contains(t, output, "1", "Line numbers should be present")
		
		// CRITICAL TEST: Check for ghost line artifact with line numbers
		assert.NotContains(t, output, "Hello World█", "GHOST LINE BUG: Cursor should not be appended at end with line numbers")
		
		// The cursor should be visible
		assert.Contains(t, output, "█", "Cursor should be visible with line numbers")
		
		// Additional verification: check content length with line numbers
		lines := strings.Split(output, "\n")
		if len(lines) > 0 {
			firstLine := strings.TrimRight(testutils.StripAnsiEscapes(lines[0]), " ")
			// With line numbers, format should be "   1 │ Hello World" with cursor on H
			// NOT "   1 │ Hello World█" (cursor appended)
			expectedMaxLength := 7 + 11 // line number prefix (7 runes) + content length (11 runes)
			assert.True(t, utf8.RuneCountInString(firstLine) <= expectedMaxLength, "First line with line numbers should not be extended by cursor")
		}
	})
	
	t.Run("empty file cursor position", func(t *testing.T) {
		// Create an empty file
		tmpFile := createTempFile(t, "")
		defer os.Remove(tmpFile)
		
		// Load empty file
		model := tui.New()
		model.SetFilename(tmpFile)
		
		// Test cursor position in empty file
		editor := model.GetEditor()
		pos := editor.GetCursor().GetBufferPos()
		
		assert.Equal(t, 0, pos.Line, "Cursor should be at line 0 in empty file")
		assert.Equal(t, 0, pos.Col, "Cursor should be at column 0 in empty file")
		
		// Test screen position (line numbers are on by default)
		screenRow, screenCol := func() (int, int) {
		screenPos, err := editor.GetCursor().GetScreenPos()
		if err != nil { return -1, -1 }
		return screenPos.Row, screenPos.Col
	}()
		assert.Equal(t, 0, screenRow, "Screen row should be 0 for empty file")
		expectedCol := editor.GetLineNumberWidth() // Line numbers are on by default
		assert.Equal(t, expectedCol, screenCol, "Screen col should include line number offset for empty file")
	})
}

// testutils.ReadOutput reads the output from a teatest output reader
