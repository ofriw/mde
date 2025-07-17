// Package ast defines the cursor state management system.
//
// CURSOR MANAGER QUICK REFERENCE:
//
// WHAT: Manages cursor position and selection state
// WHERE: Position logic in CursorManager, movement logic in Document
// HOW: Use BufferPos for all position operations
//
// COMMON OPERATIONS:
//   ✅ cursor.SetBufferPos(BufferPos{Line: 10, Col: 0})
//   ✅ pos := cursor.GetBufferPos()
//   ✅ screenPos, err := cursor.GetScreenPos()
//   ✅ cursor.StartSelection(); cursor.ExtendSelection()
//   ❌ cursor.SetScreenPos(...) // Wrong - use BufferPos only
//
// MOVEMENT PATTERN:
//   1. Editor calls Document.MoveCursorRight() → returns new BufferPos
//   2. Editor calls cursor.SetBufferPos(newPos) → updates cursor state
//   3. CursorManager validates position and updates desired column
//
// SELECTION PATTERN:
//   cursor.StartSelection()    // Begin selection at current position
//   cursor.ExtendSelection()   // Extend to current position after movement
//   text := editor.GetSelectionText() // Get selected text
//
// COORDINATE TRANSFORMATION:
//   screenPos, err := cursor.GetScreenPos()
//   if err == ErrPositionNotVisible { /* handle off-screen cursor */ }
package ast


// Selection represents a text selection range using BufferPos.
type Selection struct {
	Start BufferPos
	End   BufferPos
}

// CursorManager manages cursor position state and coordinate transformations.
// DOES: Position state, coordinate transforms, selection management
// DOES NOT: Cursor movement logic (Document handles this)
type CursorManager struct {
	bufferPos   BufferPos          // Authoritative cursor position
	viewport    *Viewport          // Immutable viewport configuration
	validator   PositionValidator  // Bounds checking
	selection   *Selection         // Current selection (nil if none)
	desired     int                // Desired column for vertical movement
}

// NewCursorManager creates a new cursor manager with the given components.
// The validator is typically a Document that implements PositionValidator.
func NewCursorManager(viewport *Viewport, validator PositionValidator) *CursorManager {
	return &CursorManager{
		bufferPos: BufferPos{Line: 0, Col: 0},
		viewport:  viewport,
		validator: validator,
		desired:   0,
	}
}

// GetBufferPos returns the current cursor position in buffer coordinates.
func (c *CursorManager) GetBufferPos() BufferPos {
	return c.bufferPos
}

// GetScreenPos returns the current cursor position in screen coordinates.
// Returns error if position is not visible in current viewport.
func (c *CursorManager) GetScreenPos() (ScreenPos, error) {
	return c.viewport.BufferToScreen(c.bufferPos)
}

// SetBufferPos sets the cursor position in buffer coordinates.
// VALIDATES: Position against document bounds
// UPDATES: Desired column for vertical movement
// USAGE: cursor.SetBufferPos(BufferPos{Line: 10, Col: 0})
func (c *CursorManager) SetBufferPos(pos BufferPos) error {
	if err := c.validator.ValidateBufferPos(pos); err != nil {
		return err
	}
	
	c.bufferPos = pos
	c.desired = pos.Col
	return nil
}

// SetBufferPosWithDesiredColumn sets the cursor position while preserving desired column.
// USAGE: For vertical movement that should remember intended column
// EXAMPLE: Moving up/down tries to return to same column when possible
func (c *CursorManager) SetBufferPosWithDesiredColumn(pos BufferPos, preserveDesired bool) error {
	if err := c.validator.ValidateBufferPos(pos); err != nil {
		return err
	}
	
	c.bufferPos = pos
	if !preserveDesired {
		c.desired = pos.Col
	}
	return nil
}

// GetDesiredColumn returns the desired column for vertical movement.
// This is used by Document methods to implement proper vertical movement behavior.
func (c *CursorManager) GetDesiredColumn() int {
	return c.desired
}

// SetDesiredColumn sets the desired column for vertical movement.
// This is called by Document methods after horizontal movement.
func (c *CursorManager) SetDesiredColumn(col int) {
	c.desired = col
}

// GetSelection returns the current selection.
func (c *CursorManager) GetSelection() *Selection {
	return c.selection
}

// SetSelection sets the selection.
func (c *CursorManager) SetSelection(selection *Selection) {
	c.selection = selection
}

// ClearSelection clears the current selection.
func (c *CursorManager) ClearSelection() {
	c.selection = nil
}

// HasSelection returns true if there is an active selection.
func (c *CursorManager) HasSelection() bool {
	return c.selection != nil
}

// StartSelection starts a new selection from current position.
func (c *CursorManager) StartSelection() {
	c.selection = &Selection{
		Start: c.bufferPos,
		End:   c.bufferPos,
	}
}

// ExtendSelection extends the selection to current position.
func (c *CursorManager) ExtendSelection() {
	if c.selection == nil {
		c.StartSelection()
	} else {
		c.selection.End = c.bufferPos
	}
}

// NOTE: GetSelectionText is not implemented in CursorManager.
// Selection text extraction belongs in the Editor where both Document and
// CursorManager are available. Use Editor.GetSelectionText() instead.

// UpdateViewport updates the viewport configuration.
// This creates a new viewport instance to maintain immutability.
func (c *CursorManager) UpdateViewport(viewport *Viewport) {
	c.viewport = viewport
}

// UpdateValidator updates the position validator.
// This is used when the document changes and the cursor manager needs to validate against the new document.
func (c *CursorManager) UpdateValidator(validator PositionValidator) {
	c.validator = validator
}

// GetViewport returns the current viewport.
func (c *CursorManager) GetViewport() *Viewport {
	return c.viewport
}

// NOTE: All cursor movement logic has been moved to Document methods.
// CursorManager now only handles position state and coordinate transformations.
// This follows the document-centric architecture pattern recommended by modern
// text editor research (CodeMirror 6, Xi-editor retrospective).