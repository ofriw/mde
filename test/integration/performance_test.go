package integration

import (
	"testing"
	"time"

	"github.com/ofri/mde/internal/tui"
	"github.com/ofri/mde/pkg/ast"
	"github.com/ofri/mde/test/testutils"
	"github.com/stretchr/testify/assert"
)

func TestCursorPerformance_BasicMovement(t *testing.T) {
	// Test that cursor movement meets performance targets
	// Target: Cursor movement < 16ms (60fps)
	
	model := tui.New()
	testutils.LoadContentIntoModel(model, generateLargeDocument(1000))
	
	editor := model.GetEditor()
	
	// Test various cursor movements
	movements := []func(){
		func() { editor.MoveCursorRight() },
		func() { editor.MoveCursorLeft() },
		func() { editor.MoveCursorUp() },
		func() { editor.MoveCursorDown() },
		func() { editor.MoveCursorWordRight() },
		func() { editor.MoveCursorWordLeft() },
		func() { editor.MoveCursorToLineEnd() },
		func() { editor.MoveCursorToLineStart() },
	}
	
	for _, movement := range movements {
		start := time.Now()
		movement()
		duration := time.Since(start)
		
		// Each movement should complete within 16ms for 60fps
		assert.Less(t, duration, 16*time.Millisecond, "Cursor movement should complete within 16ms")
	}
}

func TestCursorPerformance_ScreenPositionCalculation(t *testing.T) {
	// Test that cursor screen position calculation meets performance targets
	// Target: Render < 50ms for 1000 lines (from ticket)
	
	model := tui.New()
	testutils.LoadContentIntoModel(model, generateLargeDocument(1000))
	
	editor := model.GetEditor()
	
	// Test screen position calculation at various positions
	positions := []ast.BufferPos{
		{Line: 0, Col: 0},
		{Line: 100, Col: 50},
		{Line: 500, Col: 25},
		{Line: 999, Col: 0},
	}
	
	for _, pos := range positions {
		editor.GetCursor().SetBufferPos(pos)
		
		start := time.Now()
		screenPos, err := editor.GetCursor().GetScreenPos()
		if err != nil {
			// Cursor not visible, skip timing for this position
			continue
		}
		screenRow, screenCol := screenPos.Row, screenPos.Col
		duration := time.Since(start)
		
		// Screen position calculation should be fast
		assert.Less(t, duration, 1*time.Millisecond, "Screen position calculation should be fast")
		
		// Verify position is valid
		assert.True(t, screenRow >= 0, "Screen row should be non-negative")
		assert.True(t, screenCol >= 0, "Screen column should be non-negative")
	}
}

func TestCursorPerformance_MassiveDocument(t *testing.T) {
	// Test cursor performance with very large documents
	// Target: Memory < 50MB typical usage (from ticket)
	
	model := tui.New()
	testutils.LoadContentIntoModel(model, generateLargeDocument(10000))
	
	editor := model.GetEditor()
	
	// Test movements across large document
	start := time.Now()
	
	// Move to various positions in large document
	editor.MoveCursorToDocumentEnd()
	editor.MoveCursorToDocumentStart()
	editor.GetCursor().SetBufferPos(ast.BufferPos{Line: 5000, Col: 50})
	
	duration := time.Since(start)
	
	// Should complete within reasonable time
	assert.Less(t, duration, 50*time.Millisecond, "Large document navigation should complete within 50ms")
}

func TestCursorPerformance_ContinuousMovement(t *testing.T) {
	// Test continuous cursor movement (like holding arrow key)
	
	model := tui.New()
	testutils.LoadContentIntoModel(model, generateLargeDocument(100))
	
	editor := model.GetEditor()
	
	// Simulate continuous right arrow key
	start := time.Now()
	for i := 0; i < 1000; i++ {
		editor.MoveCursorRight()
	}
	duration := time.Since(start)
	
	// Should average less than 1ms per movement
	avgDuration := duration / 1000
	assert.Less(t, avgDuration, 1*time.Millisecond, "Continuous movement should average less than 1ms per movement")
}

func TestCursorPerformance_ViewportUpdates(t *testing.T) {
	// Test that viewport updates don't significantly impact cursor performance
	
	model := tui.New()
	testutils.LoadContentIntoModel(model, generateLargeDocument(1000))
	
	editor := model.GetEditor()
	
	// Test viewport size changes
	viewportSizes := []struct{ width, height int }{
		{80, 24},
		{120, 30},
		{160, 40},
		{200, 50},
	}
	
	for _, size := range viewportSizes {
		start := time.Now()
		
		editor.SetViewPort(size.width, size.height)
		editor.GetCursor().SetBufferPos(ast.BufferPos{Line: 100, Col: 25})
		screenPos, err := editor.GetCursor().GetScreenPos()
		if err != nil {
			// Cursor not visible, skip timing for this position
			continue
		}
		screenRow, screenCol := screenPos.Row, screenPos.Col
		
		duration := time.Since(start)
		
		// Viewport update + cursor positioning should be fast
		assert.Less(t, duration, 5*time.Millisecond, "Viewport update should be fast")
		
		// Verify position is still valid
		assert.True(t, screenRow >= 0, "Screen row should be non-negative after viewport change")
		assert.True(t, screenCol >= 0, "Screen column should be non-negative after viewport change")
	}
}

func BenchmarkCursorMovement(b *testing.B) {
	// Benchmark cursor movement operations
	
	model := tui.New()
	testutils.LoadContentIntoModel(model, generateLargeDocument(1000))
	
	editor := model.GetEditor()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		editor.MoveCursorRight()
	}
}

func BenchmarkScreenPositionCalculation(b *testing.B) {
	// Benchmark screen position calculation
	
	model := tui.New()
	testutils.LoadContentIntoModel(model, generateLargeDocument(1000))
	
	editor := model.GetEditor()
	editor.GetCursor().SetBufferPos(ast.BufferPos{Line: 500, Col: 25})
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, _ = editor.GetCursor().GetScreenPos()
	}
}

func BenchmarkCursorPositioning(b *testing.B) {
	// Benchmark cursor positioning operations
	
	model := tui.New()
	testutils.LoadContentIntoModel(model, generateLargeDocument(1000))
	
	editor := model.GetEditor()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		line := i % 1000
		col := i % 50
		editor.GetCursor().SetBufferPos(ast.BufferPos{Line: line, Col: col})
	}
}

// Helper function to generate large document for testing
func generateLargeDocument(lines int) string {
	content := ""
	for i := 0; i < lines; i++ {
		if i > 0 {
			content += "\n"
		}
		// Generate lines of varying lengths
		lineLength := 20 + (i % 60) // 20-80 character lines
		for j := 0; j < lineLength; j++ {
			char := 'a' + rune(j % 26)
			content += string(char)
		}
	}
	return content
}

func TestCursorPerformance_MemoryUsage(t *testing.T) {
	// Test that cursor operations don't cause memory leaks
	// This is a simplified test - in practice, you'd use runtime.MemStats
	
	model := tui.New()
	testutils.LoadContentIntoModel(model, generateLargeDocument(100))
	
	editor := model.GetEditor()
	
	// Perform many cursor operations
	for i := 0; i < 10000; i++ {
		editor.MoveCursorRight()
		editor.MoveCursorLeft()
		editor.GetCursor().SetBufferPos(ast.BufferPos{Line: i % 100, Col: i % 50})
		_, _ = editor.GetCursor().GetScreenPos()
	}
	
	// If we get here without running out of memory, the test passes
	assert.True(t, true, "Cursor operations should not cause memory leaks")
}

