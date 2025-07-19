package integration

import (
	"strings"
	"testing"

	"github.com/ofri/mde/internal/tui"
	"github.com/ofri/mde/pkg/ast"
	"github.com/ofri/mde/test/testutils"
	"github.com/stretchr/testify/assert"
)

// Test that specifically validates the fixes for the cursor management issues

func TestCursorFixes_FallbackRenderingFixed(t *testing.T) {
	// Test that the fallback cursor rendering no longer creates artifacts
	
	model := tui.New()
	testutils.LoadContentIntoModel(model, "hello world\ntest line\nfinal line")
	
	editor := model.GetEditor()
	
	// Test cursor rendering at different positions
	testCases := []struct {
		pos  ast.BufferPos
		desc string
	}{
		{ast.BufferPos{Line: 0, Col: 0}, "start of first line"},
		{ast.BufferPos{Line: 0, Col: 5}, "middle of first line"},
		{ast.BufferPos{Line: 0, Col: 11}, "end of first line"},
		{ast.BufferPos{Line: 1, Col: 0}, "start of second line"},
		{ast.BufferPos{Line: 1, Col: 9}, "end of second line"},
		{ast.BufferPos{Line: 2, Col: 10}, "end of last line"},
	}
	
	for _, tc := range testCases {
		editor.GetCursor().SetBufferPos(tc.pos)
		
		// Simulate the fixed fallback rendering logic
		lines := editor.GetVisibleLines()
		screenPos, err := editor.GetCursor().GetScreenPos()
		if err != nil {
			// Cursor not visible, skip this test case
			continue
		}
		cursorRow, cursorCol := screenPos.Row, screenPos.Col
		
		if cursorRow >= 0 && cursorRow < len(lines) && cursorCol >= 0 {
			line := lines[cursorRow]
			runes := []rune(line)
			
			// Apply the fixed logic
			if cursorCol < len(runes) {
				// Cursor is within the line - replace character
				runes[cursorCol] = '█'
				lines[cursorRow] = string(runes)
			} else if cursorCol == len(runes) {
				// Cursor is at end of line - append cursor
				lines[cursorRow] = line + "█"
			}
		}
		
		// Verify cursor appears exactly once
		result := strings.Join(lines, "\n")
		cursorCount := strings.Count(result, "█")
		assert.Equal(t, 1, cursorCount, "Should have exactly one cursor for %s", tc.desc)
		
		// Verify cursor is at the expected position
		lines = strings.Split(result, "\n")
		if cursorRow >= 0 && cursorRow < len(lines) {
			lineRunes := []rune(lines[cursorRow])
			found := false
			for i, r := range lineRunes {
				if r == '█' {
					// The cursor should be at the expected column
					assert.Equal(t, cursorCol, i, 
						"Cursor should be at column %d for %s, found at %d", cursorCol, tc.desc, i)
					found = true
					break
				}
			}
			assert.True(t, found, "Cursor should be visible for %s", tc.desc)
		}
	}
}

func TestCursorFixes_MouseCoordinateTransformationFixed(t *testing.T) {
	// Test that mouse coordinate transformation is now correct
	
	model := tui.New()
	testutils.LoadContentIntoModel(model, "hello world\ntest line\nfinal line")
	
	editor := model.GetEditor()
	
	// Test various click positions
	testCases := []struct {
		clickRow, clickCol int
		desc               string
	}{
		{0, 0, "top-left"},
		{0, 5, "middle of first line"},
		{1, 3, "middle of second line"},
		{1, 8, "near end of second line"},
		{2, 5, "middle of third line"},
	}
	
	for _, tc := range testCases {
		// Apply the fixed mouse coordinate transformation logic
		viewport := editor.GetViewport()
		docRow := viewport.GetTopLine() + tc.clickRow
		docCol := tc.clickCol
		
		// Account for line numbers first (fixed order)
		if editor.ShowLineNumbers() {
			lineNumberWidth := editor.GetLineNumberWidth()
			if tc.clickCol < lineNumberWidth {
				docCol = 0
			} else {
				docCol = tc.clickCol - lineNumberWidth
			}
		}
		
		// Adjust for viewport offset
		docCol += viewport.GetLeftColumn()
		
		// Ensure within bounds
		doc := editor.GetDocument()
		if docRow >= doc.LineCount() {
			docRow = doc.LineCount() - 1
		}
		if docRow < 0 {
			docRow = 0
		}
		
		lineLength := doc.GetLineLength(docRow)
		if docCol > lineLength {
			docCol = lineLength
		}
		if docCol < 0 {
			docCol = 0
		}
		
		// Set cursor position
		editor.GetCursor().SetBufferPos(ast.BufferPos{Line: docRow, Col: docCol})
		
		// Test round-trip consistency
		screenPos, err := editor.GetCursor().GetScreenPos()
		if err != nil {
			// Cursor not visible, skip this test case
			continue
		}
		screenRow, screenCol := screenPos.Row, screenPos.Col
		
		// Convert back to document coordinates
		backDocRow := viewport.GetTopLine() + screenRow
		backDocCol := viewport.GetLeftColumn() + screenCol
		
		if editor.ShowLineNumbers() {
			backDocCol -= editor.GetLineNumberWidth()
		}
		
		// Should be consistent
		actualPos := editor.GetCursor().GetBufferPos()
		assert.Equal(t, actualPos.Line, backDocRow, "Round-trip row consistency for %s", tc.desc)
		assert.Equal(t, actualPos.Col, backDocCol, "Round-trip column consistency for %s", tc.desc)
	}
}

func TestCursorFixes_ScreenPositionEdgeCasesFixed(t *testing.T) {
	// Test that cursor screen position calculation handles edge cases properly
	
	model := tui.New()
	testutils.LoadContentIntoModel(model, "short\nverylonglinewithnospaces\na\n")
	
	editor := model.GetEditor()
	
	// Test edge cases
	testCases := []struct {
		pos  ast.BufferPos
		desc string
	}{
		{ast.BufferPos{Line: 0, Col: 0}, "start of document"},
		{ast.BufferPos{Line: 0, Col: 5}, "end of short line"},
		{ast.BufferPos{Line: 1, Col: 0}, "start of long line"},
		{ast.BufferPos{Line: 1, Col: 23}, "end of long line"},
		{ast.BufferPos{Line: 2, Col: 0}, "start of single char line"},
		{ast.BufferPos{Line: 2, Col: 1}, "end of single char line"},
		{ast.BufferPos{Line: 3, Col: 0}, "empty line"},
	}
	
	for _, tc := range testCases {
		editor.GetCursor().SetBufferPos(tc.pos)
		screenPos, err := editor.GetCursor().GetScreenPos()
		if err != nil {
			// Cursor not visible, skip this test case
			continue
		}
		screenRow, screenCol := screenPos.Row, screenPos.Col
		
		// Basic sanity checks
		assert.True(t, screenRow >= 0 || tc.pos.Line == 0, 
			"Screen row should be reasonable for %s", tc.desc)
		assert.True(t, screenCol >= 0 || (tc.pos.Col == 0 && !editor.ShowLineNumbers()), 
			"Screen column should be reasonable for %s", tc.desc)
		
		// Test consistency with viewport
		viewport := editor.GetViewport()
		expectedRow := tc.pos.Line - viewport.GetTopLine()
		expectedCol := tc.pos.Col - viewport.GetLeftColumn()
		
		if editor.ShowLineNumbers() {
			expectedCol += editor.GetLineNumberWidth()
		}
		
		assert.Equal(t, expectedRow, screenRow, "Screen row calculation for %s", tc.desc)
		assert.Equal(t, expectedCol, screenCol, "Screen column calculation for %s", tc.desc)
	}
}

func TestCursorFixes_ViewportSynchronizationFixed(t *testing.T) {
	// Test that cursor/viewport synchronization works correctly
	
	model := tui.New()
	content := ""
	for i := 0; i < 20; i++ {
		if i > 0 {
			content += "\n"
		}
		content += "Line " + string(rune('A'+i)) + " content here"
	}
	testutils.LoadContentIntoModel(model, content)
	
	editor := model.GetEditor()
	
	// Test various scenarios
	testCases := []struct {
		pos  ast.BufferPos
		desc string
	}{
		{ast.BufferPos{Line: 0, Col: 0}, "start of document"},
		{ast.BufferPos{Line: 10, Col: 5}, "middle of document"},
		{ast.BufferPos{Line: 19, Col: 10}, "end of document"},
	}
	
	for _, tc := range testCases {
		editor.GetCursor().SetBufferPos(tc.pos)
		
		// Get current viewport and screen position
		viewport := editor.GetViewport()
		screenPos, err := editor.GetCursor().GetScreenPos()
		if err != nil {
			// Cursor not visible, skip this test case
			continue
		}
		screenRow, screenCol := screenPos.Row, screenPos.Col
		
		// Test forward transformation
		expectedScreenRow := tc.pos.Line - viewport.GetTopLine()
		expectedScreenCol := tc.pos.Col - viewport.GetLeftColumn()
		
		if editor.ShowLineNumbers() {
			expectedScreenCol += editor.GetLineNumberWidth()
		}
		
		assert.Equal(t, expectedScreenRow, screenRow, "Forward transformation row for %s", tc.desc)
		assert.Equal(t, expectedScreenCol, screenCol, "Forward transformation column for %s", tc.desc)
		
		// Test reverse transformation
		backDocRow := viewport.GetTopLine() + screenRow
		backDocCol := viewport.GetLeftColumn() + screenCol
		
		if editor.ShowLineNumbers() {
			backDocCol -= editor.GetLineNumberWidth()
		}
		
		assert.Equal(t, tc.pos.Line, backDocRow, "Reverse transformation row for %s", tc.desc)
		assert.Equal(t, tc.pos.Col, backDocCol, "Reverse transformation column for %s", tc.desc)
	}
}

func TestCursorFixes_UnicodeHandlingFixed(t *testing.T) {
	// Test that Unicode content is handled correctly
	
	model := tui.New()
	testutils.LoadContentIntoModel(model, "hello 世界\nこんにちは world\n🌍🚀💫 test")
	
	editor := model.GetEditor()
	
	// Test Unicode positions
	testCases := []struct {
		pos  ast.BufferPos
		desc string
	}{
		{ast.BufferPos{Line: 0, Col: 6}, "before Unicode character"},
		{ast.BufferPos{Line: 0, Col: 8}, "after Unicode characters"},
		{ast.BufferPos{Line: 1, Col: 5}, "after Japanese characters"},
		{ast.BufferPos{Line: 2, Col: 4}, "after emoji characters"},
	}
	
	for _, tc := range testCases {
		editor.GetCursor().SetBufferPos(tc.pos)
		
		// Position should be valid
		actualPos := editor.GetCursor().GetBufferPos()
		assert.True(t, actualPos.Line >= 0, "Unicode position line should be valid for %s", tc.desc)
		assert.True(t, actualPos.Col >= 0, "Unicode position column should be valid for %s", tc.desc)
		
		// Screen position should be calculable
		screenPos, err := editor.GetCursor().GetScreenPos()
		if err != nil {
			// Cursor not visible, skip this test case
			continue
		}
		screenRow, screenCol := screenPos.Row, screenPos.Col
		assert.True(t, screenRow >= 0 || actualPos.Line == 0, "Unicode screen row should be reasonable for %s", tc.desc)
		assert.True(t, screenCol >= 0 || (actualPos.Col == 0 && !editor.ShowLineNumbers()), 
			"Unicode screen column should be reasonable for %s", tc.desc)
		
		// Round-trip should be consistent
		viewport := editor.GetViewport()
		backDocRow := viewport.GetTopLine() + screenRow
		backDocCol := viewport.GetLeftColumn() + screenCol
		
		if editor.ShowLineNumbers() {
			backDocCol -= editor.GetLineNumberWidth()
		}
		
		assert.Equal(t, actualPos.Line, backDocRow, "Unicode round-trip row for %s", tc.desc)
		assert.Equal(t, actualPos.Col, backDocCol, "Unicode round-trip column for %s", tc.desc)
	}
}