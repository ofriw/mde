// Package ast defines the coordinate system for the MDE editor.
//
// QUICK REFERENCE:
//
// WHAT: Single BufferPos type for all operations
// WHY: Eliminates transformation bugs and synchronization issues  
// HOW: BufferPos → Viewport → ScreenPos (unidirectional)
//
// USAGE PATTERNS:
//   pos := BufferPos{Line: 0, Col: 5}           // Document position
//   screenPos, err := viewport.BufferToScreen(pos) // Transform to screen
//   if err == ErrPositionNotVisible { ... }     // Handle invisible positions
//
// COMMON OPERATIONS:
//   ✅ cursor.SetBufferPos(BufferPos{Line: 10, Col: 0})
//   ✅ screenPos, err := viewport.BufferToScreen(cursor.GetBufferPos())
//   ❌ screenPos := ScreenPos{Row: 10, Col: 0} // Wrong - use viewport
//
// CONSTRAINTS:
//   1. Only BufferPos is authoritative - never create ScreenPos directly
//   2. Always validate: validator.ValidateBufferPos(pos)
//   3. Handle ErrPositionNotVisible when converting to screen coordinates
//   4. Viewport is immutable - create new instances for changes
package ast

import (
	"fmt"
)

// BufferPos represents a position in the document buffer.
// AUTHORITATIVE: This is the single source of truth for all positions.
// USAGE: pos := BufferPos{Line: 0, Col: 5} // Line 1, character 6
type BufferPos struct {
	Line int // 0-indexed line number in document
	Col  int // 0-indexed column position in line
}

// String returns a human-readable representation of the buffer position.
func (p BufferPos) String() string {
	return fmt.Sprintf("BufferPos{Line:%d, Col:%d}", p.Line, p.Col)
}

// IsValid returns true if the position has non-negative coordinates.
func (p BufferPos) IsValid() bool {
	return p.Line >= 0 && p.Col >= 0
}

// ScreenPos represents a position on the terminal screen.
// DERIVED: Always computed from BufferPos via viewport.BufferToScreen()
// NEVER create directly - use viewport transformation
type ScreenPos struct {
	Row int // 0-indexed row on terminal screen
	Col int // 0-indexed column on terminal screen
}

// String returns a human-readable representation of the screen position.
func (p ScreenPos) String() string {
	return fmt.Sprintf("ScreenPos{Row:%d, Col:%d}", p.Row, p.Col)
}

// IsValid returns true if the position has non-negative coordinates.
func (p ScreenPos) IsValid() bool {
	return p.Row >= 0 && p.Col >= 0
}

// CoordinateError represents an error in coordinate validation or transformation.
type CoordinateError struct {
	Type     string // "buffer" or "screen"
	Position string // String representation of the position
	Reason   string // Human-readable reason for the error
}

// Error returns the error message.
func (e CoordinateError) Error() string {
	return fmt.Sprintf("coordinate error in %s position %s: %s", e.Type, e.Position, e.Reason)
}

// NewBufferCoordinateError creates a coordinate error for buffer positions.
func NewBufferCoordinateError(pos BufferPos, reason string) CoordinateError {
	return CoordinateError{
		Type:     "buffer",
		Position: pos.String(),
		Reason:   reason,
	}
}

// NewScreenCoordinateError creates a coordinate error for screen positions.
func NewScreenCoordinateError(pos ScreenPos, reason string) CoordinateError {
	return CoordinateError{
		Type:     "screen",
		Position: pos.String(),
		Reason:   reason,
	}
}

// ErrPositionNotVisible is returned when a buffer position is not visible in the viewport.
// COMMON: Check for this error when converting BufferPos to ScreenPos
// USAGE: if err == ErrPositionNotVisible { /* cursor off-screen */ }
var ErrPositionNotVisible = CoordinateError{
	Type:     "buffer",
	Position: "",
	Reason:   "position not visible in current viewport",
}

// PositionValidator validates buffer positions against document bounds.
type PositionValidator interface {
	// ValidateBufferPos checks if a buffer position is within document bounds.
	// Returns error with clear message if position is invalid.
	ValidateBufferPos(pos BufferPos) error
}