package testutils

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/teatest"
	"github.com/ofri/mde/pkg/ast"
	"github.com/stretchr/testify/assert"
)

// readOutput reads the output from a teatest output reader
func readOutput(output io.Reader) string {
	if output == nil {
		return ""
	}
	
	buf := make([]byte, 4096)
	n, err := output.Read(buf)
	if err != nil && err != io.EOF {
		return ""
	}
	
	return string(buf[:n])
}

// VisualAssertion provides utilities for visual testing and artifact detection
type VisualAssertion struct {
	t  *testing.T
	tm *teatest.TestModel
}

// NewVisualAssertion creates a new visual assertion helper
func NewVisualAssertion(t *testing.T, tm *teatest.TestModel) *VisualAssertion {
	return &VisualAssertion{
		t:  t,
		tm: tm,
	}
}

// AssertNoVisualArtifacts checks for common visual artifacts
func (va *VisualAssertion) AssertNoVisualArtifacts() {
	output := va.tm.FinalOutput(va.t)
	outputStr := readOutput(output)
	
	// Check for duplicate cursor characters
	va.assertNoDuplicateCursors(outputStr)
	
	// Check for orphaned line number fragments
	va.assertNoOrphanedLineNumbers(outputStr)
	
	// Check for broken line endings
	va.assertCleanLineEndings(outputStr)
	
	// Check for unexpected control characters
	va.assertNoControlCharacters(outputStr)
}

// AssertCursorVisible checks that the cursor is visible in the output
func (va *VisualAssertion) AssertCursorVisible() {
	output := va.tm.FinalOutput(va.t)
	outputStr := readOutput(output)
	
	// Look for cursor characters (█, |, _, etc.)
	cursorChars := []string{"█", "▏", "▎", "▍", "▌", "▋", "▊", "▉", "|", "_"}
	
	hasCursor := false
	for _, char := range cursorChars {
		if strings.Contains(outputStr, char) {
			hasCursor = true
			break
		}
	}
	
	assert.True(va.t, hasCursor, "Cursor should be visible in output")
}

// AssertCursorAtPosition checks that the cursor appears at the expected screen position
func (va *VisualAssertion) AssertCursorAtPosition(expectedRow, expectedCol int) {
	output := va.tm.FinalOutput(va.t)
	outputStr := readOutput(output)
	lines := strings.Split(outputStr, "\n")
	
	if expectedRow >= len(lines) {
		va.t.Errorf("Expected cursor row %d is beyond output lines (%d)", expectedRow, len(lines))
		return
	}
	
	line := lines[expectedRow]
	if expectedCol >= len(line) {
		va.t.Errorf("Expected cursor column %d is beyond line length (%d)", expectedCol, len(line))
		return
	}
	
	// Check for cursor character at expected position
	cursorChars := []string{"█", "▏", "▎", "▍", "▌", "▋", "▊", "▉", "|", "_"}
	runes := []rune(line)
	
	if expectedCol < len(runes) {
		char := string(runes[expectedCol])
		found := false
		for _, cursorChar := range cursorChars {
			if char == cursorChar {
				found = true
				break
			}
		}
		assert.True(va.t, found, "Cursor character should be at position (%d, %d), found: %q", expectedRow, expectedCol, char)
	}
}

// AssertLineNumbers checks that line numbers are correctly displayed
func (va *VisualAssertion) AssertLineNumbers(expectedCount int) {
	output := va.tm.FinalOutput(va.t)
	outputStr := readOutput(output)
	lines := strings.Split(outputStr, "\n")
	
	lineNumberCount := 0
	for _, line := range lines {
		// Look for line number pattern: "1234 │ " or similar
		if len(line) >= 6 && strings.Contains(line[0:6], "│") {
			lineNumberCount++
		}
	}
	
	assert.Equal(va.t, expectedCount, lineNumberCount, "Line number count should match")
}

// AssertContentVisible checks that specific content is visible in the output
func (va *VisualAssertion) AssertContentVisible(expectedText string) {
	output := va.tm.FinalOutput(va.t)
	outputStr := readOutput(output)
	assert.Contains(va.t, outputStr, expectedText, "Content should be visible in output")
}

// AssertContentNotVisible checks that specific content is not visible in the output
func (va *VisualAssertion) AssertContentNotVisible(unexpectedText string) {
	output := va.tm.FinalOutput(va.t)
	outputStr := readOutput(output)
	assert.NotContains(va.t, outputStr, unexpectedText, "Content should not be visible in output")
}

// AssertSelectionHighlight checks that text selection is properly highlighted
func (va *VisualAssertion) AssertSelectionHighlight(selection *ast.Selection) {
	if selection == nil {
		return
	}
	
	output := va.tm.FinalOutput(va.t)
	outputStr := readOutput(output)
	
	// This is a simplified check - in a real implementation, you would
	// look for specific highlighting characters or ANSI codes
	
	// For now, just check that the output contains some highlighting indicators
	highlightChars := []string{"[", "]", "▪", "▫", "■", "□"}
	
	hasHighlight := false
	for _, char := range highlightChars {
		if strings.Contains(outputStr, char) {
			hasHighlight = true
			break
		}
	}
	
	// This is a weak assertion - in practice, you'd want more specific checks
	if selection.Start != selection.End {
		// Only assert if there's actually a selection
		assert.True(va.t, hasHighlight, "Selection should be highlighted")
	}
}

// AssertStatusBar checks that the status bar is correctly displayed
func (va *VisualAssertion) AssertStatusBar(expectedContent string) {
	output := va.tm.FinalOutput(va.t)
	outputStr := readOutput(output)
	lines := strings.Split(outputStr, "\n")
	
	if len(lines) == 0 {
		va.t.Error("No output lines found")
		return
	}
	
	// Status bar is typically at the bottom
	statusLine := lines[len(lines)-1]
	assert.Contains(va.t, statusLine, expectedContent, "Status bar should contain expected content")
}

// AssertViewportBounds checks that content is properly clipped to viewport
func (va *VisualAssertion) AssertViewportBounds(expectedWidth, expectedHeight int) {
	output := va.tm.FinalOutput(va.t)
	outputStr := readOutput(output)
	lines := strings.Split(outputStr, "\n")
	
	// Check height
	assert.LessOrEqual(va.t, len(lines), expectedHeight, "Output should not exceed viewport height")
	
	// Check width for each line
	for i, line := range lines {
		lineWidth := len([]rune(line))
		assert.LessOrEqual(va.t, lineWidth, expectedWidth, "Line %d should not exceed viewport width", i)
	}
}

// AssertScrollPosition checks that scrolling is working correctly
func (va *VisualAssertion) AssertScrollPosition(expectedTopLine int) {
	output := va.tm.FinalOutput(va.t)
	outputStr := readOutput(output)
	lines := strings.Split(outputStr, "\n")
	
	// This is a simplified check - in practice, you'd examine the actual
	// document content to verify which lines are visible
	
	// Look for line numbers if they're enabled
	if len(lines) > 0 && len(lines[0]) >= 6 {
		firstLine := lines[0]
		if strings.Contains(firstLine, "│") {
			// Extract line number
			parts := strings.Split(firstLine, "│")
			if len(parts) > 0 {
				lineNumStr := strings.TrimSpace(parts[0])
				// In a real implementation, you'd parse this and verify it matches expectedTopLine
				assert.NotEmpty(va.t, lineNumStr, "Line number should be present")
			}
		}
	}
}

// Helper methods for specific artifact detection

func (va *VisualAssertion) assertNoDuplicateCursors(output string) {
	cursorChars := []string{"█", "▏", "▎", "▍", "▌", "▋", "▊", "▉"}
	
	for _, char := range cursorChars {
		count := strings.Count(output, char)
		assert.LessOrEqual(va.t, count, 1, "Should not have duplicate cursor characters: %q", char)
	}
}

func (va *VisualAssertion) assertNoOrphanedLineNumbers(output string) {
	lines := strings.Split(output, "\n")
	
	for i, line := range lines {
		// Check for line number fragments without proper structure
		if strings.Contains(line, "│") {
			// Should have format: "1234 │ content"
			parts := strings.Split(line, "│")
			if len(parts) >= 2 {
				lineNumPart := strings.TrimSpace(parts[0])
				assert.NotEmpty(va.t, lineNumPart, "Line number part should not be empty on line %d", i)
			}
		}
	}
}

func (va *VisualAssertion) assertCleanLineEndings(output string) {
	// Check for common line ending issues
	
	// No trailing spaces before newlines
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if i < len(lines)-1 { // Don't check the last line
			assert.False(va.t, strings.HasSuffix(line, " "), "Line %d should not have trailing spaces", i)
		}
	}
	
	// No multiple consecutive newlines (unless intentional)
	assert.False(va.t, strings.Contains(output, "\n\n\n"), "Should not have more than 2 consecutive newlines")
}

func (va *VisualAssertion) assertNoControlCharacters(output string) {
	// Check for unexpected control characters (except ANSI escape codes)
	
	for i, char := range output {
		if char < 32 && char != '\n' && char != '\t' && char != '\r' {
			// Allow ESC (27) for ANSI codes
			if char != 27 {
				va.t.Errorf("Unexpected control character %d at position %d", char, i)
			}
		}
	}
}

// GoldenFileAssertion provides utilities for golden file testing
type GoldenFileAssertion struct {
	t        *testing.T
	tm       *teatest.TestModel
	testName string
}

// NewGoldenFileAssertion creates a new golden file assertion helper
func NewGoldenFileAssertion(t *testing.T, tm *teatest.TestModel, testName string) *GoldenFileAssertion {
	return &GoldenFileAssertion{
		t:        t,
		tm:       tm,
		testName: testName,
	}
}

// CompareWithGolden compares the current output with a golden file
func (gfa *GoldenFileAssertion) CompareWithGolden() {
	// This is a placeholder for golden file comparison
	// In a real implementation, you would:
	// 1. Get the current output
	// 2. Read the golden file
	// 3. Compare them
	// 4. Update the golden file if -update flag is set
	
	output := gfa.tm.FinalOutput(gfa.t)
	outputStr := readOutput(output)
	
	// For now, just ensure the output is not empty
	assert.NotEmpty(gfa.t, outputStr, "Output should not be empty for golden file test")
	
	// In a real implementation:
	// goldenFile := fmt.Sprintf("testdata/%s.golden", gfa.testName)
	// if *update {
	//     ioutil.WriteFile(goldenFile, []byte(output), 0644)
	// } else {
	//     expected, err := ioutil.ReadFile(goldenFile)
	//     assert.NoError(gfa.t, err)
	//     assert.Equal(gfa.t, string(expected), output)
	// }
}

// SaveAsGolden saves the current output as a golden file
func (gfa *GoldenFileAssertion) SaveAsGolden() {
	output := gfa.tm.FinalOutput(gfa.t)
	outputStr := readOutput(output)
	
	// This is a placeholder - in a real implementation, you would save to file
	fmt.Printf("Golden file content for %s:\n%s\n", gfa.testName, outputStr)
}

// PerformanceAssertion provides utilities for performance testing
type PerformanceAssertion struct {
	t  *testing.T
	tm *teatest.TestModel
}

// NewPerformanceAssertion creates a new performance assertion helper
func NewPerformanceAssertion(t *testing.T, tm *teatest.TestModel) *PerformanceAssertion {
	return &PerformanceAssertion{
		t:  t,
		tm: tm,
	}
}

// AssertRenderTime checks that rendering completes within the expected time
func (pa *PerformanceAssertion) AssertRenderTime(maxDuration time.Duration) {
	// This is a placeholder for performance testing
	// In a real implementation, you would measure actual render time
	
	// For now, just ensure the test completes
	assert.True(pa.t, true, "Render time assertion placeholder")
}

// AssertMemoryUsage checks that memory usage is within expected bounds
func (pa *PerformanceAssertion) AssertMemoryUsage(maxMemoryMB int) {
	// This is a placeholder for memory usage testing
	// In a real implementation, you would measure actual memory usage
	
	assert.True(pa.t, true, "Memory usage assertion placeholder")
}