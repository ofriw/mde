package ast

import "strings"

// Cursor manages cursor position and movement within a document
type Cursor struct {
	pos       Position
	document  *Document
	selection *Selection
	desired   int // Desired column for vertical movement
}

// NewCursor creates a new cursor for the given document
func NewCursor(doc *Document) *Cursor {
	return &Cursor{
		pos:      Position{Line: 0, Col: 0},
		document: doc,
		desired:  0,
	}
}

// GetPosition returns the current cursor position
func (c *Cursor) GetPosition() Position {
	return c.pos
}

// SetPosition sets the cursor position
func (c *Cursor) SetPosition(pos Position) {
	c.pos = c.document.ValidatePosition(pos)
	c.desired = c.pos.Col
}

// MoveLeft moves cursor left by one character
func (c *Cursor) MoveLeft() {
	if c.pos.Col > 0 {
		c.pos.Col--
	} else if c.pos.Line > 0 {
		c.pos.Line--
		c.pos.Col = c.document.GetLineLength(c.pos.Line)
	}
	c.desired = c.pos.Col
}

// MoveRight moves cursor right by one character
func (c *Cursor) MoveRight() {
	lineLength := c.document.GetLineLength(c.pos.Line)
	if c.pos.Col < lineLength {
		c.pos.Col++
	} else if c.pos.Line < c.document.LineCount()-1 {
		c.pos.Line++
		c.pos.Col = 0
	}
	c.desired = c.pos.Col
}

// MoveUp moves cursor up by one line
func (c *Cursor) MoveUp() {
	if c.pos.Line > 0 {
		c.pos.Line--
		c.pos.Col = c.desired
		lineLength := c.document.GetLineLength(c.pos.Line)
		if c.pos.Col > lineLength {
			c.pos.Col = lineLength
		}
	}
}

// MoveDown moves cursor down by one line
func (c *Cursor) MoveDown() {
	if c.pos.Line < c.document.LineCount()-1 {
		c.pos.Line++
		c.pos.Col = c.desired
		lineLength := c.document.GetLineLength(c.pos.Line)
		if c.pos.Col > lineLength {
			c.pos.Col = lineLength
		}
	}
}

// MoveToLineStart moves cursor to beginning of current line
func (c *Cursor) MoveToLineStart() {
	c.pos.Col = 0
	c.desired = 0
}

// MoveToLineEnd moves cursor to end of current line
func (c *Cursor) MoveToLineEnd() {
	c.pos.Col = c.document.GetLineLength(c.pos.Line)
	c.desired = c.pos.Col
}

// MoveToDocumentStart moves cursor to beginning of document
func (c *Cursor) MoveToDocumentStart() {
	c.pos = Position{Line: 0, Col: 0}
	c.desired = 0
}

// MoveToDocumentEnd moves cursor to end of document
func (c *Cursor) MoveToDocumentEnd() {
	c.pos.Line = c.document.LineCount() - 1
	c.pos.Col = c.document.GetLineLength(c.pos.Line)
	c.desired = c.pos.Col
}

// MoveWordLeft moves cursor to start of previous word
func (c *Cursor) MoveWordLeft() {
	// If at start of line, move to end of previous line
	if c.pos.Col == 0 {
		if c.pos.Line > 0 {
			c.pos.Line--
			c.pos.Col = c.document.GetLineLength(c.pos.Line)
		}
		c.desired = c.pos.Col
		return
	}
	
	// Find word start
	c.pos = c.document.FindWordStart(c.pos)
	c.desired = c.pos.Col
}

// MoveWordRight moves cursor to start of next word
func (c *Cursor) MoveWordRight() {
	lineLength := c.document.GetLineLength(c.pos.Line)
	
	// If at end of line, move to start of next line
	if c.pos.Col >= lineLength {
		if c.pos.Line < c.document.LineCount()-1 {
			c.pos.Line++
			c.pos.Col = 0
		}
		c.desired = c.pos.Col
		return
	}
	
	// Find word end
	c.pos = c.document.FindWordEnd(c.pos)
	c.desired = c.pos.Col
}

// GetSelection returns the current selection
func (c *Cursor) GetSelection() *Selection {
	return c.selection
}

// SetSelection sets the selection
func (c *Cursor) SetSelection(selection *Selection) {
	c.selection = selection
}

// ClearSelection clears the current selection
func (c *Cursor) ClearSelection() {
	c.selection = nil
}

// HasSelection returns true if there is an active selection
func (c *Cursor) HasSelection() bool {
	return c.selection != nil
}

// StartSelection starts a new selection from current position
func (c *Cursor) StartSelection() {
	c.selection = &Selection{
		Start: c.pos,
		End:   c.pos,
	}
}

// ExtendSelection extends the selection to current position
func (c *Cursor) ExtendSelection() {
	if c.selection == nil {
		c.StartSelection()
	} else {
		c.selection.End = c.pos
	}
}

// GetSelectionText returns the selected text
func (c *Cursor) GetSelectionText() string {
	if c.selection == nil {
		return ""
	}
	
	start := c.selection.Start
	end := c.selection.End
	
	// Ensure start is before end
	if start.Line > end.Line || (start.Line == end.Line && start.Col > end.Col) {
		start, end = end, start
	}
	
	if start.Line == end.Line {
		// Single line selection
		line := c.document.GetLine(start.Line)
		runes := []rune(line)
		if start.Col < len(runes) && end.Col <= len(runes) {
			return string(runes[start.Col:end.Col])
		}
		return ""
	}
	
	// Multi-line selection
	var result []string
	
	// First line
	firstLine := c.document.GetLine(start.Line)
	firstRunes := []rune(firstLine)
	if start.Col < len(firstRunes) {
		result = append(result, string(firstRunes[start.Col:]))
	}
	
	// Middle lines
	for i := start.Line + 1; i < end.Line; i++ {
		result = append(result, c.document.GetLine(i))
	}
	
	// Last line
	lastLine := c.document.GetLine(end.Line)
	lastRunes := []rune(lastLine)
	if end.Col <= len(lastRunes) {
		result = append(result, string(lastRunes[:end.Col]))
	}
	
	return strings.Join(result, "\n")
}