// Package unit contains tests that reproduce the cursor positioning bug when line numbers are enabled.
//
// ARCHITECTURAL CONTEXT:
// The MDE editor uses a coordinate transformation chain:
//   DocumentPos (raw document coordinates) → ContentPos (includes line number offset) → ScreenPos (terminal coordinates)
//
// The bug occurs because:
//   1. Editor tracks line numbers state via editor.ShowLineNumbers()
//   2. Renderer has separate line numbers config via renderer.config.ShowLineNumbers
//   3. These two settings can become desynchronized
//   4. When desynchronized, cursor coordinate transformation fails
//
// COORDINATE SYSTEM EXPLANATION:
// - DocumentPos: Position in raw document content (e.g., Line=0, Col=0 for first character)
// - ContentPos: Position including line number offset (e.g., Line=0, Col=6 when line numbers enabled)
// - The renderer uses ContentPos but must know if line numbers are enabled to transform correctly
//
// BUG MANIFESTATION:
// When editor.ShowLineNumbers()=true but renderer.config.ShowLineNumbers=false:
//   - Editor calculates ContentPos with line number offset (Col=6)
//   - Renderer doesn't subtract line number offset during rendering
//   - Cursor appears at wrong position or becomes invisible
//
// TESTING STRATEGY:
// These tests isolate the renderer coordinate transformation logic to demonstrate
// the exact mechanism of the bug without TUI complexity.
//
// TEST SUMMARY:
// - TestCursorLineNumbersBug: Reproduces the coordinate transformation bug by showing
//   that desynchronized editor/renderer configurations cause cursor misplacement
// - TestCursorLineNumbersFixed: Proves the bug is fixable by synchronizing configurations
//
// FUTURE LLM READERS:
// These tests provide a clear diagnosis of the cursor positioning bug. The root cause
// is configuration desynchronization between editor.ShowLineNumbers() and 
// renderer.config.ShowLineNumbers. The fix is to ensure TUI synchronizes these settings.
package unit

import (
	"context"
	"strings"
	"testing"

	"github.com/ofri/mde/pkg/ast"
	"github.com/ofri/mde/internal/plugins/renderers"
	"github.com/ofri/mde/internal/plugins/themes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCursorLineNumbersBug reproduces the cursor positioning bug through isolated renderer testing.
//
// PURPOSE:
// This test demonstrates the exact mechanism of the bug by testing the renderer's coordinate
// transformation logic in isolation, without TUI complexity.
//
// BUG REPRODUCTION STRATEGY:
// 1. Create editor with line numbers enabled (editor.ShowLineNumbers() = true)
// 2. Create renderer with default config (renderer.config.ShowLineNumbers = false)
// 3. Show that this desynchronization causes incorrect cursor positioning
//
// COORDINATE FLOW ANALYSIS:
// When line numbers are enabled in editor:
//   - Document position: (0,0) [first character]
//   - Content position: (0,6) [includes "   1 │ " prefix]
//   - Renderer receives ContentPos(0,6) but thinks line numbers are disabled
//   - Result: Cursor appears at character 6 instead of character 0
//
// WHY THIS HAPPENS:
// The renderer's RenderToStringWithCursor() method uses this logic:
//   if r.config.ShowLineNumbers {
//       adjustedCursorCol = cursorCol - 6  // Subtract line number offset
//   }
// But when r.config.ShowLineNumbers is false, no adjustment happens, so the cursor
// appears at the raw ContentPos column (6) instead of the intended position (0).
//
// EXPECTED vs ACTUAL:
// - Expected: Cursor at first character ("█ello World")
// - Actual: Cursor at position 6 ("Hello █orld")
//
// TESTING RATIONALE:
// This test isolates the renderer bug from TUI complexity to make the root cause clear.
// It shows that the issue is purely a coordinate transformation problem.
func TestCursorLineNumbersBug(t *testing.T) {
	// SETUP: Create editor with content to test cursor positioning
	editor := ast.NewEditorWithContent("Hello World\nSecond Line")
	
	// CRITICAL: Enable line numbers in editor but NOT in renderer
	// This desynchronization is the root cause of the bug
	editor.ToggleLineNumbers()
	
	// SETUP: Create renderer with default config
	// renderer.config.ShowLineNumbers will be false (default)
	renderer := renderers.NewTerminalRenderer()
	theme := themes.NewDarkTheme()
	
	// VERIFICATION: Cursor should be at document origin
	cursor := editor.GetCursor()
	pos := cursor.GetPosition()
	assert.Equal(t, ast.Position{Line: 0, Col: 0}, pos, "Cursor should be at document position (0,0)")
	
	// KEY INSIGHT: Content position includes line number offset
	// When line numbers are enabled, ContentPos.Col = DocumentPos.Col + 6
	// This is correct behavior - the editor properly calculates the offset
	contentPos := editor.GetCursorContentPosition()
	assert.Equal(t, ast.ContentPos{Line: 0, Col: 6}, contentPos, "Content position should include line number offset")
	
	// SETUP: Render document content (this step is always correct)
	doc := editor.GetDocument()
	renderedLines, err := renderer.Render(context.Background(), doc, theme)
	require.NoError(t, err)
	
	// BUG REPRODUCTION: Pass ContentPos to renderer that doesn't know about line numbers
	// The renderer receives ContentPos{Line: 0, Col: 6} but doesn't know to subtract the offset
	// because renderer.config.ShowLineNumbers = false
	result := renderer.RenderToStringWithCursor(renderedLines, theme, contentPos.Line, contentPos.Col)
	
	// ANALYSIS: Find where cursor actually appears
	lines := strings.Split(result, "\n")
	firstLine := lines[0]
	
	// MEASUREMENT: Calculate actual cursor position in rendered line
	runes := []rune(firstLine)
	cursorPos := -1
	for i, r := range runes {
		if r == '█' {
			cursorPos = i
			break
		}
	}
	
	// BUG DEMONSTRATION: Cursor appears at position 6 (wrong) instead of 0 (correct)
	// This happens because renderer doesn't subtract the line number offset from ContentPos
	assert.Equal(t, 6, cursorPos, "BUG: Cursor should be at position 0 but appears at position %d", cursorPos)
	assert.Equal(t, "Hello █orld", firstLine, "BUG: Cursor appears in wrong position")
	
	// DOCUMENTATION: Record the buggy behavior for future reference
	t.Logf("BUG REPRODUCED: Cursor at position %d in line '%s'", cursorPos, firstLine)
	t.Logf("ROOT CAUSE: editor.ShowLineNumbers()=%v, renderer.config.ShowLineNumbers=%v", 
		editor.ShowLineNumbers(), false)
}

// TestCursorLineNumbersFixed demonstrates the correct solution to the cursor positioning bug.
//
// PURPOSE:
// This test shows how to fix the bug by synchronizing the renderer configuration
// with the editor's line numbers setting.
//
// SOLUTION STRATEGY:
// 1. Create editor with line numbers enabled (editor.ShowLineNumbers() = true)
// 2. Create renderer and configure it to match (renderer.config.ShowLineNumbers = true)
// 3. Demonstrate that synchronized configuration produces correct cursor positioning
//
// WHY THIS WORKS:
// When both editor and renderer have line numbers enabled:
//   - Editor calculates ContentPos(0,6) for DocumentPos(0,0)
//   - Renderer sees ContentPos(0,6) and knows to subtract line number offset
//   - adjustedCursorCol = contentPos.Col - 6 = 6 - 6 = 0
//   - Cursor appears at correct position (first character)
//
// COORDINATE FLOW (FIXED):
// 1. Document position: (0,0) [first character]
// 2. Content position: (0,6) [includes "   1 │ " prefix]
// 3. Renderer adjustment: 6 - 6 = 0 [subtracts line number offset]
// 4. Final position: 0 [cursor at first character]
//
// IMPLEMENTATION DETAILS:
// The fix requires calling renderer.Configure() with the editor's line numbers setting.
// This ensures the renderer's coordinate transformation logic matches the editor's
// coordinate calculation logic.
//
// TESTING RATIONALE:
// This test proves that the bug is fixable through proper configuration synchronization
// without requiring changes to the core rendering or cursor logic.
func TestCursorLineNumbersFixed(t *testing.T) {
	// SETUP: Create editor with content to test cursor positioning
	editor := ast.NewEditorWithContent("Hello World\nSecond Line")
	
	// SETUP: Enable line numbers in editor
	editor.ToggleLineNumbers()
	
	// SETUP: Create renderer with default config
	renderer := renderers.NewTerminalRenderer()
	theme := themes.NewDarkTheme()
	
	// SOLUTION: Configure renderer to match editor's line numbers setting
	// This synchronizes the two configuration states that caused the bug
	err := renderer.Configure(map[string]interface{}{
		"showLineNumbers": editor.ShowLineNumbers(),
	})
	require.NoError(t, err, "Renderer should accept line numbers configuration")
	
	// VERIFICATION: Cursor should be at document origin
	cursor := editor.GetCursor()
	pos := cursor.GetPosition()
	assert.Equal(t, ast.Position{Line: 0, Col: 0}, pos, "Cursor should be at document position (0,0)")
	
	// VERIFICATION: Content position includes line number offset (same as before)
	contentPos := editor.GetCursorContentPosition()
	assert.Equal(t, ast.ContentPos{Line: 0, Col: 6}, contentPos, "Content position should include line number offset")
	
	// SETUP: Render document content (this step is always correct)
	doc := editor.GetDocument()
	renderedLines, err := renderer.Render(context.Background(), doc, theme)
	require.NoError(t, err)
	
	// SOLUTION DEMONSTRATION: Pass ContentPos to properly configured renderer
	// The renderer receives ContentPos{Line: 0, Col: 6} and knows to subtract the offset
	// because renderer.config.ShowLineNumbers = true
	result := renderer.RenderToStringWithCursor(renderedLines, theme, contentPos.Line, contentPos.Col)
	
	// VERIFICATION: Cursor should appear at correct position
	lines := strings.Split(result, "\n")
	firstLine := lines[0]
	
	// ASSERTION: The line should have line number prefix followed by cursor at first character
	// Expected format: "   1 │ █ello World"
	assert.True(t, strings.HasPrefix(firstLine, "   1 │ █ello"), 
		"Expected line to start with '   1 │ █ello' but got: '%s'", firstLine)
	
	// DOCUMENTATION: Record the fixed behavior for future reference
	t.Logf("FIXED: Cursor correctly positioned in line '%s'", firstLine)
	t.Logf("SOLUTION: Synchronized editor.ShowLineNumbers()=%v with renderer.config.ShowLineNumbers=%v", 
		editor.ShowLineNumbers(), true)
}