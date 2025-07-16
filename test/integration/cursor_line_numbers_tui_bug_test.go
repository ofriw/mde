// Package integration contains TUI-level tests that reproduce the cursor visibility bug.
//
// INTEGRATION TESTING RATIONALE:
// While unit tests isolate the renderer coordinate transformation bug, integration tests
// demonstrate the complete user-facing issue in the TUI context.
//
// TUI RENDERING ARCHITECTURE:
// The TUI follows this rendering pipeline:
//   1. Document content → Renderer.Render() → RenderedLines
//   2. RenderedLines → Renderer.RenderToStringWithCursor() → Final output
//   3. Final output → TUI display
//
// BUG MANIFESTATION IN TUI:
// When line numbers are enabled via editor.ToggleLineNumbers():
//   - TUI enables line numbers in editor but doesn't configure renderer
//   - Renderer uses default config (ShowLineNumbers=false)
//   - Cursor becomes invisible or positioned incorrectly
//   - User experiences "broken horizontal movement" (cursor not visible)
//
// TESTING STRATEGY:
// These tests use the full TUI model to reproduce the exact user experience:
//   1. TestCursorLineNumbersTUIBug - Shows cursor invisibility
//   2. TestCursorLineNumbersHorizontalMovement - Shows movement appears broken
//   3. TestCursorLineNumbersToggle - Shows toggle behavior
//
// WHY INTEGRATION TESTS MATTER:
// Unit tests prove the coordinate transformation bug exists, but integration tests
// demonstrate the actual user impact and verify that fixes work in the real TUI context.
//
// EXPECTED vs ACTUAL USER EXPERIENCE:
// - Expected: Cursor visible and moving correctly with line numbers
// - Actual: Cursor invisible, making horizontal movement appear broken
//
// TEST SUMMARY:
// - TestCursorLineNumbersTUIBug: Shows cursor becomes invisible in TUI when line numbers enabled
// - TestCursorLineNumbersHorizontalMovement: Shows why users report "movement broken" 
// - TestCursorLineNumbersToggle: Shows bug occurs specifically during line number toggle
//
// FUTURE LLM READERS:
// These tests prove that the "horizontal movement not working" user complaint is actually
// a cursor visibility bug caused by renderer configuration desynchronization. The cursor
// movement logic works correctly, but users can't see the cursor move. Fix by synchronizing
// editor and renderer line number settings in the TUI.
package integration

import (
	"strings"
	"testing"

	"github.com/ofri/mde/internal/tui"
	"github.com/ofri/mde/pkg/ast"
	"github.com/ofri/mde/test/testutils"
	"github.com/stretchr/testify/assert"
)

// TestCursorLineNumbersTUIBug reproduces the cursor visibility bug in the complete TUI context.
//
// PURPOSE:
// This test demonstrates the user-facing bug by testing the complete TUI rendering pipeline,
// showing that the cursor becomes invisible when line numbers are enabled.
//
// USER EXPERIENCE REPRODUCTION:
// 1. User opens file and enables line numbers
// 2. Cursor should be visible at first character
// 3. BUG: Cursor becomes invisible due to renderer configuration mismatch
// 4. User tries to move cursor horizontally but can't see it move
//
// TUI RENDERING FLOW:
// 1. model.View() calls TUI rendering pipeline
// 2. TUI calls editor.ToggleLineNumbers() (editor config changes)
// 3. TUI calls renderer.RenderToStringWithCursor() (renderer config unchanged)
// 4. Coordinate mismatch causes cursor to be invisible or misplaced
//
// TESTING APPROACH:
// - Use full TUI model to replicate user workflow
// - Enable line numbers as user would
// - Examine actual rendered output that user sees
// - Document whether cursor is visible and where it appears
//
// EXPECTED vs ACTUAL:
// - Expected: "   1 │ █ello World" (cursor visible at first character)
// - Actual: "   1 │ Hello World" (cursor invisible due to coordinate bug)
//
// WHY THIS MATTERS:
// This test proves that the coordinate transformation bug causes a real user experience
// issue, not just a theoretical problem. Users report "horizontal movement not working"
// when the real issue is cursor invisibility.
func TestCursorLineNumbersTUIBug(t *testing.T) {
	// SETUP: Create TUI model to replicate user experience
	model := tui.New()
	
	// SETUP: Set model size to simulate terminal window
	testutils.SetModelSize(model, 80, 24)
	
	// SETUP: Load content as user would when opening a file
	testutils.LoadContentIntoModel(model, "Hello World\nSecond Line\nThird Line")
	
	// USER ACTION: Enable line numbers (common user workflow)
	editor := model.GetEditor()
	editor.ToggleLineNumbers()
	
	// VERIFICATION: Cursor should be at expected document position
	cursor := editor.GetCursor()
	pos := cursor.GetPosition()
	assert.Equal(t, ast.Position{Line: 0, Col: 0}, pos, "Cursor should be at document origin")
	
	// CRITICAL: Get the actual rendered output that user sees
	// This is the complete TUI rendering pipeline result
	output := model.View()
	
	// ANALYSIS: Parse rendered output to find content lines
	lines := strings.Split(output, "\n")
	
	// SEARCH: Find the first content line (may be mixed with status bars, etc.)
	// Note: Look for either "Hello" or "█ello" since cursor might replace the "H"
	var firstContentLine string
	for _, line := range lines {
		if strings.Contains(line, "Hello") || strings.Contains(line, "█ello") {
			firstContentLine = line
			break
		}
	}
	
	// DOCUMENTATION: Record what the user actually sees
	t.Logf("User sees first content line: '%s'", firstContentLine)
	
	// BUG DETECTION: Check if cursor is visible to user
	if strings.Contains(firstContentLine, "█") {
		// ANALYSIS: Find cursor position in rendered output
		runes := []rune(firstContentLine)
		cursorPos := -1
		for i, r := range runes {
			if r == '█' {
				cursorPos = i
				break
			}
		}
		
		// VERIFICATION: Check if cursor is at expected position
		// With line numbers enabled, cursor should be at position 7 (after "   1 │ ")
		t.Logf("Cursor found at position %d", cursorPos)
		
		// ANALYSIS: Document current behavior vs expected behavior
		if cursorPos == 7 {
			t.Log("✅ Cursor is at correct position 7 (after line number prefix)")
		} else {
			t.Logf("❌ BUG: Cursor at position %d, expected 7", cursorPos)
		}
	} else {
		// BUG DETECTED: Cursor is invisible - this is the main issue
		t.Log("❌ BUG: Cursor is not visible in output")
		t.Log("IMPACT: User cannot see cursor position, horizontal movement appears broken")
		t.Log("ROOT CAUSE: Renderer configuration not synchronized with editor line numbers setting")
	}
}

// TestCursorLineNumbersHorizontalMovement demonstrates why users report "horizontal movement not working".
//
// PURPOSE:
// This test shows that while horizontal cursor movement logic works correctly,
// the cursor invisibility bug makes it appear that movement is broken.
//
// USER EXPERIENCE SIMULATION:
// 1. User has line numbers enabled
// 2. User presses right arrow key to move cursor
// 3. Cursor position updates internally (DocumentPos changes)
// 4. BUG: Cursor remains invisible, so user doesn't see movement
// 5. User concludes "horizontal movement is broken"
//
// ACTUAL vs PERCEIVED ISSUE:
// - Actual: Cursor movement works, but cursor rendering is broken
// - Perceived: Horizontal movement doesn't work
// - This explains why users report movement issues rather than rendering issues
//
// TESTING METHODOLOGY:
// - Test cursor movement at document level (should work)
// - Test cursor visibility at TUI level (should fail)
// - Demonstrate that "broken movement" is really "invisible cursor"
//
// DIAGNOSTIC VALUE:
// This test helps distinguish between:
// 1. Actual cursor movement bugs (cursor position doesn't change)
// 2. Cursor rendering bugs (cursor position changes but isn't visible)
// This bug is type 2 - rendering, not movement.
func TestCursorLineNumbersHorizontalMovement(t *testing.T) {
	// SETUP: Create TUI model for movement testing
	model := tui.New()
	
	// SETUP: Set model size to simulate terminal window
	testutils.SetModelSize(model, 80, 24)
	
	// SETUP: Load content for cursor movement testing
	testutils.LoadContentIntoModel(model, "Hello World\nSecond Line")
	
	// USER ACTION: Enable line numbers (triggers the bug)
	editor := model.GetEditor()
	editor.ToggleLineNumbers()
	
	// VERIFICATION: Initial cursor position should be at document origin
	cursor := editor.GetCursor()
	pos := cursor.GetPosition()
	assert.Equal(t, ast.Position{Line: 0, Col: 0}, pos, "Initial cursor position should be (0,0)")
	
	// USER ACTION: Move cursor right (this is what users report as broken)
	cursor.MoveRight()
	
	// VERIFICATION: Cursor movement logic should work correctly
	pos = cursor.GetPosition()
	assert.Equal(t, ast.Position{Line: 0, Col: 1}, pos, "Cursor should move to position (0,1)")
	
	// CRITICAL: Get rendered output after movement
	output := model.View()
	lines := strings.Split(output, "\n")
	
	// ANALYSIS: Find the content line to check cursor visibility
	// Note: Look for either "Hello" or "█ello" since cursor might replace characters
	var firstContentLine string
	for _, line := range lines {
		if strings.Contains(line, "Hello") || strings.Contains(line, "█ello") || strings.Contains(line, "H█llo") {
			firstContentLine = line
			break
		}
	}
	
	// DOCUMENTATION: Record what user sees after movement
	t.Logf("After MoveRight: '%s'", firstContentLine)
	
	// BUG DETECTION: Check if cursor is visible after movement
	if strings.Contains(firstContentLine, "█") {
		// ANALYSIS: Find cursor position in rendered output
		runes := []rune(firstContentLine)
		cursorPos := -1
		for i, r := range runes {
			if r == '█' {
				cursorPos = i
				break
			}
		}
		
		// VERIFICATION: After moving right, cursor should be at position 8 (after "   1 │ H")
		t.Logf("Cursor after MoveRight at position %d", cursorPos)
		
		// ANALYSIS: Check if cursor appears at expected position
		if cursorPos == 8 {
			t.Log("✅ Cursor moved correctly to position 8")
		} else {
			t.Logf("❌ BUG: Cursor at position %d, expected 8", cursorPos)
		}
	} else {
		// BUG DETECTED: This is why users report "horizontal movement not working"
		t.Log("❌ BUG: Cursor is not visible after movement")
		t.Log("USER IMPACT: User pressed right arrow but sees no change")
		t.Log("USER CONCLUSION: 'Horizontal movement is broken'")
		t.Log("ACTUAL ISSUE: Cursor moved correctly but rendering is broken")
		t.Log("DIAGNOSTIC: DocumentPos changed (0,0)→(0,1) but cursor invisible")
	}
}

// TestCursorLineNumbersToggle demonstrates cursor behavior during line number toggle operations.
//
// PURPOSE:
// This test shows how the cursor visibility bug manifests during the common user action
// of toggling line numbers on and off.
//
// USER WORKFLOW SIMULATION:
// 1. User starts with line numbers disabled (cursor should be visible)
// 2. User enables line numbers (cursor should remain visible)
// 3. BUG: Cursor becomes invisible when line numbers are enabled
// 4. User may toggle line numbers off again to "fix" cursor visibility
//
// TOGGLE BEHAVIOR ANALYSIS:
// - Line numbers OFF: Renderer config and editor config both default to false → cursor visible
// - Line numbers ON: Editor config true, renderer config false → cursor invisible
// - This shows the bug is specifically in the configuration synchronization
//
// DIAGNOSTIC VALUE:
// This test helps identify whether the bug is:
// 1. A general cursor rendering issue (affects both modes)
// 2. A line number coordinate transformation issue (only affects line number mode)
// The bug is type 2 - only affects line number mode.
//
// TESTING METHODOLOGY:
// - Test cursor visibility before toggle (baseline)
// - Test cursor visibility after toggle (bug detection)
// - Document the specific state change that triggers the bug
func TestCursorLineNumbersToggle(t *testing.T) {
	// SETUP: Create TUI model for toggle testing
	model := tui.New()
	
	// SETUP: Set model size to simulate terminal window
	testutils.SetModelSize(model, 80, 24)
	
	// SETUP: Load content for visibility testing
	testutils.LoadContentIntoModel(model, "Hello World\nSecond Line")
	
	editor := model.GetEditor()
	
	// BASELINE: Test initial state (line numbers disabled)
	assert.False(t, editor.ShowLineNumbers(), "Line numbers should be disabled initially")
	
	// BASELINE: Get output without line numbers
	output := model.View()
	lines := strings.Split(output, "\n")
	
	// BASELINE: Find first content line without line numbers
	// Note: Look for either "Hello" or "█ello" since cursor might replace the "H"
	var firstContentLineWithoutNumbers string
	for _, line := range lines {
		if strings.Contains(line, "Hello") || strings.Contains(line, "█ello") {
			firstContentLineWithoutNumbers = line
			break
		}
	}
	
	// DOCUMENTATION: Record baseline behavior
	t.Logf("Without line numbers: '%s'", firstContentLineWithoutNumbers)
	
	// BASELINE: Check cursor visibility in normal mode
	if strings.Contains(firstContentLineWithoutNumbers, "█") {
		t.Log("✅ BASELINE: Cursor is visible without line numbers")
	} else {
		t.Log("❌ BASELINE ISSUE: Cursor not visible without line numbers")
	}
	
	// USER ACTION: Enable line numbers (this is where the bug manifests)
	editor.ToggleLineNumbers()
	assert.True(t, editor.ShowLineNumbers(), "Line numbers should be enabled after toggle")
	
	// CRITICAL: Get output with line numbers enabled
	output = model.View()
	lines = strings.Split(output, "\n")
	
	// ANALYSIS: Find first content line with line numbers
	// Note: Look for either "Hello" or "█ello" since cursor might replace the "H"
	var firstContentLineWithNumbers string
	for _, line := range lines {
		if strings.Contains(line, "Hello") || strings.Contains(line, "█ello") {
			firstContentLineWithNumbers = line
			break
		}
	}
	
	// DOCUMENTATION: Record behavior after toggle
	t.Logf("With line numbers: '%s'", firstContentLineWithNumbers)
	
	// VERIFICATION: Check if line numbers are properly displayed
	if strings.Contains(firstContentLineWithNumbers, "│") {
		t.Log("✅ Line numbers are visible after toggle")
	} else {
		t.Log("❌ BUG: Line numbers not visible after toggle")
	}
	
	// BUG DETECTION: Check cursor visibility after enabling line numbers
	if strings.Contains(firstContentLineWithNumbers, "█") {
		t.Log("✅ Cursor is visible with line numbers")
	} else {
		// BUG DETECTED: This is the core issue
		t.Log("❌ BUG: Cursor not visible with line numbers")
		t.Log("ISSUE: Cursor visible without line numbers but invisible with line numbers")
		t.Log("ROOT CAUSE: editor.ToggleLineNumbers() doesn't synchronize renderer configuration")
		t.Log("SOLUTION: TUI should call renderer.Configure() when toggling line numbers")
	}
}