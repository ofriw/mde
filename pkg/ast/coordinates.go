// Package ast defines coordinate types and transformation interfaces for the MDE editor.
//
// COORDINATE SYSTEMS:
//
// This editor uses 3 distinct coordinate systems to prevent confusion:
//
// 1. DocumentPos - Position in the raw document content (0-indexed)
//    - Line: Line number in the document
//    - Col: Character position in the line
//    - Used by: Cursor, Document, Edit operations
//
// 2. ContentPos - Position in rendered content area (includes line number offset)
//    - Line: Line number in the viewport
//    - Col: Character position including line number prefix (if enabled)
//    - Used by: Renderer input, TUI display
//
// 3. ScreenPos - Absolute position on terminal screen
//    - Row: Terminal row (including status bar, etc.)
//    - Col: Terminal column
//    - Used by: Terminal output, mouse events
//
// TRANSFORMATION CHAIN:
// DocumentPos → ContentPos → ScreenPos
//
// CRITICAL INVARIANT:
// Each transformation happens exactly once in exactly one component.
// Editor owns DocumentPos→ContentPos, TUI owns ContentPos→ScreenPos.
package ast

import (
	"fmt"
)

// DocumentPos represents a position in the raw document content.
// This is the authoritative coordinate system for all edit operations.
type DocumentPos struct {
	Line int // 0-indexed line number in document
	Col  int // 0-indexed column position in line
}

// String returns a human-readable representation of the document position.
func (p DocumentPos) String() string {
	return fmt.Sprintf("DocumentPos{Line:%d, Col:%d}", p.Line, p.Col)
}

// IsValid returns true if the position has non-negative coordinates.
func (p DocumentPos) IsValid() bool {
	return p.Line >= 0 && p.Col >= 0
}

// ContentPos represents a position in the rendered content area.
// This includes viewport offset and line number prefix offset.
type ContentPos struct {
	Line int // 0-indexed line number in viewport
	Col  int // 0-indexed column position including line number prefix
}

// String returns a human-readable representation of the content position.
func (p ContentPos) String() string {
	return fmt.Sprintf("ContentPos{Line:%d, Col:%d}", p.Line, p.Col)
}

// IsValid returns true if the position has non-negative coordinates.
func (p ContentPos) IsValid() bool {
	return p.Line >= 0 && p.Col >= 0
}

// ScreenPos represents an absolute position on the terminal screen.
// This includes all UI elements (status bar, help bar, etc.).
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

// ViewportInfo provides debugging information about the current viewport state.
// This is essential for LLM troubleshooting of coordinate transformations.
type ViewportInfo struct {
	Top         int  // First visible document line
	Left        int  // First visible document column
	Width       int  // Viewport width in characters
	Height      int  // Viewport height in lines
	LineNumbers bool // Whether line numbers are enabled
}

// String returns a human-readable representation of the viewport info.
func (v ViewportInfo) String() string {
	return fmt.Sprintf("ViewportInfo{Top:%d, Left:%d, Width:%d, Height:%d, LineNumbers:%t}",
		v.Top, v.Left, v.Width, v.Height, v.LineNumbers)
}

// CoordinateTransformer handles the transformation from document coordinates
// to content coordinates. This transformation includes viewport offset and
// line number offset.
type CoordinateTransformer interface {
	// TransformDocumentToContent converts document coordinates to content coordinates.
	// This transformation includes:
	// 1. Viewport offset (subtract viewport top/left)
	// 2. Line number offset (add 6 characters if line numbers enabled)
	TransformDocumentToContent(docPos DocumentPos) ContentPos
	
	// GetViewportInfo returns current viewport state for debugging.
	// This is essential for LLM troubleshooting of coordinate issues.
	GetViewportInfo() ViewportInfo
}

// CoordinateValidator validates coordinate positions against document bounds.
// This prevents coordinate system violations and provides clear error messages.
type CoordinateValidator interface {
	// ValidateDocumentPos checks if a document position is within bounds.
	// Returns error with clear message if position is invalid.
	ValidateDocumentPos(pos DocumentPos) error
	
	// ValidateContentPos checks if a content position is within bounds.
	// Returns error with clear message if position is invalid.
	ValidateContentPos(pos ContentPos) error
}

// CoordinateError represents an error in coordinate validation or transformation.
type CoordinateError struct {
	Type     string // "document", "content", or "screen"
	Position string // String representation of the position
	Reason   string // Human-readable reason for the error
}

// Error returns the error message.
func (e CoordinateError) Error() string {
	return fmt.Sprintf("coordinate error in %s position %s: %s", e.Type, e.Position, e.Reason)
}

// NewDocumentCoordinateError creates a coordinate error for document positions.
func NewDocumentCoordinateError(pos DocumentPos, reason string) CoordinateError {
	return CoordinateError{
		Type:     "document",
		Position: pos.String(),
		Reason:   reason,
	}
}

// NewContentCoordinateError creates a coordinate error for content positions.
func NewContentCoordinateError(pos ContentPos, reason string) CoordinateError {
	return CoordinateError{
		Type:     "content",
		Position: pos.String(),
		Reason:   reason,
	}
}