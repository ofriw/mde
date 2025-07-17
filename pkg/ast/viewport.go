// Package ast defines the immutable viewport system for coordinate transformations.
//
// VIEWPORT QUICK REFERENCE:
//
// WHAT: Transforms BufferPos to ScreenPos for terminal display
// WHY: Handles scrolling and line number offset in one place
// HOW: viewport.BufferToScreen(bufferPos) → screenPos, error
//
// COMMON OPERATIONS:
//   ✅ screenPos, err := viewport.BufferToScreen(bufferPos)
//   ✅ newViewport := viewport.WithTopLine(10) // Scroll to line 10
//   ✅ if err == ErrPositionNotVisible { /* handle off-screen */ }
//   ❌ viewport.topLine = 10 // Wrong - viewport is immutable
//
// TRANSFORMATION FORMULA:
//   screenRow = bufferPos.Line - viewport.topLine
//   screenCol = bufferPos.Col - viewport.leftColumn + viewport.lineNumberWidth
package ast

import (
	"fmt"
)

// Viewport represents the visible area of the document and provides
// coordinate transformation from BufferPos to ScreenPos.
// IMMUTABLE: Create new instances for changes to prevent sync issues
type Viewport struct {
	topLine         int  // First visible document line (0-indexed)
	leftColumn      int  // First visible document column (0-indexed)
	width           int  // Viewport width in characters
	height          int  // Viewport height in lines
	lineNumberWidth int  // Width of line number prefix (0 or 6)
	tabWidth        int  // Tab width in spaces
}

// NewViewport creates a new immutable viewport with the given parameters.
func NewViewport(topLine, leftColumn, width, height, lineNumberWidth, tabWidth int) *Viewport {
	return &Viewport{
		topLine:         topLine,
		leftColumn:      leftColumn,
		width:           width,
		height:          height,
		lineNumberWidth: lineNumberWidth,
		tabWidth:        tabWidth,
	}
}

// BufferToScreen converts a buffer position to a screen position.
// RETURNS: ScreenPos if visible, ErrPositionNotVisible if off-screen
// USAGE: screenPos, err := viewport.BufferToScreen(bufferPos)
func (v *Viewport) BufferToScreen(pos BufferPos) (ScreenPos, error) {
	// Check if position is visible
	if !v.isVisible(pos) {
		return ScreenPos{}, ErrPositionNotVisible
	}
	
	// Transform coordinates
	screenRow := pos.Line - v.topLine
	screenCol := pos.Col - v.leftColumn + v.lineNumberWidth
	
	return ScreenPos{Row: screenRow, Col: screenCol}, nil
}

// isVisible checks if a buffer position is visible in the current viewport.
func (v *Viewport) isVisible(pos BufferPos) bool {
	// Check vertical bounds
	if pos.Line < v.topLine || pos.Line >= v.topLine+v.height {
		return false
	}
	
	// Check horizontal bounds (considering line number width)
	visibleLeft := v.leftColumn
	visibleRight := v.leftColumn + v.width - v.lineNumberWidth
	
	if pos.Col < visibleLeft || pos.Col >= visibleRight {
		return false
	}
	
	return true
}

// GetTopLine returns the first visible document line.
func (v *Viewport) GetTopLine() int {
	return v.topLine
}

// GetLeftColumn returns the first visible document column.
func (v *Viewport) GetLeftColumn() int {
	return v.leftColumn
}

// GetWidth returns the viewport width in characters.
func (v *Viewport) GetWidth() int {
	return v.width
}

// GetHeight returns the viewport height in lines.
func (v *Viewport) GetHeight() int {
	return v.height
}

// GetLineNumberWidth returns the width of the line number prefix.
func (v *Viewport) GetLineNumberWidth() int {
	return v.lineNumberWidth
}

// GetTabWidth returns the tab width in spaces.
func (v *Viewport) GetTabWidth() int {
	return v.tabWidth
}

// String returns a human-readable representation of the viewport.
func (v *Viewport) String() string {
	return fmt.Sprintf("Viewport{TopLine:%d, LeftColumn:%d, Width:%d, Height:%d, LineNumberWidth:%d, TabWidth:%d}",
		v.topLine, v.leftColumn, v.width, v.height, v.lineNumberWidth, v.tabWidth)
}

// WithTopLine creates a new viewport with updated top line.
func (v *Viewport) WithTopLine(topLine int) *Viewport {
	return &Viewport{
		topLine:         topLine,
		leftColumn:      v.leftColumn,
		width:           v.width,
		height:          v.height,
		lineNumberWidth: v.lineNumberWidth,
		tabWidth:        v.tabWidth,
	}
}

// WithLeftColumn creates a new viewport with updated left column.
func (v *Viewport) WithLeftColumn(leftColumn int) *Viewport {
	return &Viewport{
		topLine:         v.topLine,
		leftColumn:      leftColumn,
		width:           v.width,
		height:          v.height,
		lineNumberWidth: v.lineNumberWidth,
		tabWidth:        v.tabWidth,
	}
}

// WithDimensions creates a new viewport with updated dimensions.
func (v *Viewport) WithDimensions(width, height int) *Viewport {
	return &Viewport{
		topLine:         v.topLine,
		leftColumn:      v.leftColumn,
		width:           width,
		height:          height,
		lineNumberWidth: v.lineNumberWidth,
		tabWidth:        v.tabWidth,
	}
}

// ScreenToBuffer converts a screen position to a buffer position.
// SAFE: Always returns a valid BufferPos, handling line number offsets correctly
// USAGE: bufferPos := viewport.ScreenToBuffer(ScreenPos{Row: 5, Col: 10})
func (v *Viewport) ScreenToBuffer(pos ScreenPos) BufferPos {
	// Convert screen coordinates to buffer coordinates
	bufferLine := pos.Row + v.topLine
	bufferCol := pos.Col - v.lineNumberWidth + v.leftColumn
	
	// Ensure minimum bounds (negative coordinates become 0)
	if bufferLine < 0 {
		bufferLine = 0
	}
	if bufferCol < 0 {
		bufferCol = 0
	}
	
	return BufferPos{Line: bufferLine, Col: bufferCol}
}