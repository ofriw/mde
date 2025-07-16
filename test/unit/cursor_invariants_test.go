package unit

import (
	"math/rand"
	"testing"

	"github.com/ofri/mde/pkg/ast"
	"github.com/stretchr/testify/assert"
)

// Property-based tests for cursor invariants

func TestCursor_Invariant_PositionAlwaysValid(t *testing.T) {
	// Property: cursor position should always be within document bounds
	for i := 0; i < 100; i++ {
		doc := generateRandomDocument()
		cursor := ast.NewCursor(doc)
		
		// Apply random movements
		for j := 0; j < 20; j++ {
			applyRandomMovement(cursor)
			
			pos := cursor.GetPosition()
			assert.True(t, pos.Line >= 0, "Line should be non-negative")
			assert.True(t, pos.Line < doc.LineCount(), "Line should be within document bounds")
			assert.True(t, pos.Col >= 0, "Column should be non-negative")
			assert.True(t, pos.Col <= doc.GetLineLength(pos.Line), "Column should be within line bounds")
		}
	}
}

func TestCursor_Invariant_DesiredColumnConsistency(t *testing.T) {
	// Property: desired column should be maintained across vertical movements
	for i := 0; i < 50; i++ {
		doc := generateRandomDocument()
		cursor := ast.NewCursor(doc)
		
		// Set cursor to a random position
		randomPos := generateRandomPosition(doc)
		cursor.SetPosition(randomPos)
		
		// Move up then down (or vice versa) should maintain desired column
		originalPos := cursor.GetPosition()
		
		// Try moving up then down
		cursor.MoveUp()
		cursor.MoveDown()
		
		// If we're on the same line, column should be preserved
		if cursor.GetPosition().Line == originalPos.Line {
			assert.Equal(t, originalPos.Col, cursor.GetPosition().Col, 
				"Desired column should be maintained")
		}
	}
}

func TestCursor_Invariant_MovementReversibility(t *testing.T) {
	// Property: opposite movements should cancel each other out when possible
	for i := 0; i < 50; i++ {
		doc := generateRandomDocument()
		cursor := ast.NewCursor(doc)
		
		// Set cursor to a random position
		randomPos := generateRandomPosition(doc)
		cursor.SetPosition(randomPos)
		originalPos := cursor.GetPosition()
		
		// Test left-right reversibility
		cursor.MoveRight()
		rightPos := cursor.GetPosition()
		cursor.MoveLeft()
		
		// If we could move right and left, we should be back at original
		if rightPos != originalPos {
			assert.Equal(t, originalPos, cursor.GetPosition(), 
				"Left-right movement should be reversible")
		}
		
		// Test up-down reversibility
		cursor.SetPosition(originalPos)
		cursor.MoveDown()
		downPos := cursor.GetPosition()
		cursor.MoveUp()
		
		// If we could move down and up, we should be back at original
		if downPos != originalPos {
			assert.Equal(t, originalPos, cursor.GetPosition(), 
				"Up-down movement should be reversible")
		}
	}
}

func TestCursor_Invariant_BoundaryBehavior(t *testing.T) {
	// Property: movements at boundaries should be safe and consistent
	for i := 0; i < 50; i++ {
		doc := generateRandomDocument()
		cursor := ast.NewCursor(doc)
		
		// Test document start boundary
		cursor.MoveToDocumentStart()
		pos := cursor.GetPosition()
		assert.Equal(t, ast.Position{Line: 0, Col: 0}, pos)
		
		// Moving left/up from start should not change position
		cursor.MoveLeft()
		assert.Equal(t, pos, cursor.GetPosition())
		cursor.MoveUp()
		assert.Equal(t, pos, cursor.GetPosition())
		
		// Test document end boundary
		cursor.MoveToDocumentEnd()
		pos = cursor.GetPosition()
		assert.Equal(t, doc.LineCount()-1, pos.Line)
		assert.Equal(t, doc.GetLineLength(pos.Line), pos.Col)
		
		// Moving right/down from end should not change position
		cursor.MoveRight()
		assert.Equal(t, pos, cursor.GetPosition())
		cursor.MoveDown()
		assert.Equal(t, pos, cursor.GetPosition())
	}
}

func TestCursor_Invariant_SelectionConsistency(t *testing.T) {
	// Property: selection should always be consistent with cursor position
	for i := 0; i < 50; i++ {
		doc := generateRandomDocument()
		cursor := ast.NewCursor(doc)
		
		// Start selection at random position
		startPos := generateRandomPosition(doc)
		cursor.SetPosition(startPos)
		cursor.StartSelection()
		
		// Move cursor and extend selection
		for j := 0; j < 10; j++ {
			applyRandomMovement(cursor)
			cursor.ExtendSelection()
			
			selection := cursor.GetSelection()
			assert.NotNil(t, selection, "Selection should exist")
			assert.Equal(t, startPos, selection.Start, "Selection start should not change")
			assert.Equal(t, cursor.GetPosition(), selection.End, "Selection end should match cursor")
		}
		
		// Clear selection
		cursor.ClearSelection()
		assert.False(t, cursor.HasSelection())
		assert.Nil(t, cursor.GetSelection())
	}
}

func TestCursor_Invariant_SelectionTextAccuracy(t *testing.T) {
	// Property: selection text should match actual document content
	for i := 0; i < 30; i++ {
		doc := generateRandomDocument()
		cursor := ast.NewCursor(doc)
		
		// Create random selection
		pos1 := generateRandomPosition(doc)
		pos2 := generateRandomPosition(doc)
		
		cursor.SetPosition(pos1)
		cursor.StartSelection()
		cursor.SetPosition(pos2)
		cursor.ExtendSelection()
		
		selectionText := cursor.GetSelectionText()
		
		// Verify selection text matches expected content
		if cursor.HasSelection() {
			selection := cursor.GetSelection()
			start := selection.Start
			end := selection.End
			
			// Normalize order
			if start.Line > end.Line || (start.Line == end.Line && start.Col > end.Col) {
				start, end = end, start
			}
			
			// Build expected text
			expectedText := buildExpectedSelectionText(doc, start, end)
			assert.Equal(t, expectedText, selectionText, 
				"Selection text should match document content")
		}
	}
}

func TestCursor_Invariant_ScreenPositionConsistency(t *testing.T) {
	// Property: screen position calculation should be consistent
	for i := 0; i < 50; i++ {
		doc := generateRandomDocument()
		editor := ast.NewEditorWithContent(doc.GetText())
		
		// Test with random viewport settings
		width := 80 + rand.Intn(40)
		height := 25 + rand.Intn(10)
		editor.SetViewPort(width, height)
		
		// Set cursor to random position
		pos := generateRandomPosition(doc)
		editor.GetCursor().SetPosition(pos)
		
		// Get screen position
		screenRow, screenCol := editor.GetCursorScreenPosition()
		
		// Screen position should be consistent with viewport
		viewport := editor.GetViewPort()
		expectedScreenRow := pos.Line - viewport.Top
		expectedScreenCol := pos.Col - viewport.Left
		
		if editor.ShowLineNumbers() {
			expectedScreenCol += 6
		}
		
		assert.Equal(t, expectedScreenRow, screenRow, "Screen row should match calculation")
		assert.Equal(t, expectedScreenCol, screenCol, "Screen column should match calculation")
	}
}

// Helper functions for property-based testing

func generateRandomDocument() *ast.Document {
	// Generate 1-10 lines with random content
	lineCount := 1 + rand.Intn(10)
	var lines []string
	
	for i := 0; i < lineCount; i++ {
		lineLength := rand.Intn(20) // 0-19 characters
		line := generateRandomString(lineLength)
		lines = append(lines, line)
	}
	
	content := ""
	for i, line := range lines {
		if i > 0 {
			content += "\n"
		}
		content += line
	}
	
	if content == "" {
		return ast.NewEmptyDocument()
	}
	return ast.NewDocument(content)
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 "
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func generateRandomPosition(doc *ast.Document) ast.Position {
	line := rand.Intn(doc.LineCount())
	col := rand.Intn(doc.GetLineLength(line) + 1) // +1 to allow end of line
	return ast.Position{Line: line, Col: col}
}

func applyRandomMovement(cursor *ast.Cursor) {
	movements := []func(){
		cursor.MoveLeft,
		cursor.MoveRight,
		cursor.MoveUp,
		cursor.MoveDown,
		cursor.MoveWordLeft,
		cursor.MoveWordRight,
		cursor.MoveToLineStart,
		cursor.MoveToLineEnd,
	}
	
	movement := movements[rand.Intn(len(movements))]
	movement()
}

func buildExpectedSelectionText(doc *ast.Document, start, end ast.Position) string {
	if start.Line == end.Line {
		// Single line selection
		line := doc.GetLine(start.Line)
		runes := []rune(line)
		if start.Col < len(runes) && end.Col <= len(runes) {
			return string(runes[start.Col:end.Col])
		}
		return ""
	}
	
	// Multi-line selection
	var result []string
	
	// First line
	firstLine := doc.GetLine(start.Line)
	firstRunes := []rune(firstLine)
	if start.Col < len(firstRunes) {
		result = append(result, string(firstRunes[start.Col:]))
	}
	
	// Middle lines
	for i := start.Line + 1; i < end.Line; i++ {
		result = append(result, doc.GetLine(i))
	}
	
	// Last line
	lastLine := doc.GetLine(end.Line)
	lastRunes := []rune(lastLine)
	if end.Col <= len(lastRunes) {
		result = append(result, string(lastRunes[:end.Col]))
	}
	
	if len(result) == 0 {
		return ""
	}
	
	// Join with newlines
	finalResult := result[0]
	for i := 1; i < len(result); i++ {
		finalResult += "\n" + result[i]
	}
	
	return finalResult
}

func TestCursor_Invariant_UnicodeHandling(t *testing.T) {
	// Property: cursor should handle Unicode characters correctly
	unicodeTexts := []string{
		"こんにちは世界",
		"Hello 🌍 World",
		"Ελληνικά",
		"العربية",
		"🚀🌟💫",
	}
	
	for _, text := range unicodeTexts {
		doc := ast.NewDocument(text)
		cursor := ast.NewCursor(doc)
		
		// Test movement through Unicode text
		for i := 0; i < len([]rune(text)); i++ {
			pos := cursor.GetPosition()
			assert.True(t, pos.Line >= 0)
			assert.True(t, pos.Col >= 0)
			assert.True(t, pos.Col <= len([]rune(text)))
			
			cursor.MoveRight()
		}
		
		// Should be at end
		pos := cursor.GetPosition()
		assert.Equal(t, 0, pos.Line)
		assert.Equal(t, len([]rune(text)), pos.Col)
	}
}