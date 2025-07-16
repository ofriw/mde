package ast

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

// Editor manages the document, cursor, and history.
// It implements CoordinateTransformer and CoordinateValidator interfaces.
type Editor struct {
	document   *Document
	cursor     *Cursor
	history    *History
	clipboard  string
	lineNumbers bool
	viewport   ViewPort
}

// ViewPort represents the visible area of the document
type ViewPort struct {
	Top    int
	Left   int
	Width  int
	Height int
}

// NewEditor creates a new editor with an empty document
func NewEditor() *Editor {
	doc := NewEmptyDocument()
	return &Editor{
		document:   doc,
		cursor:     NewCursor(doc),
		history:    NewHistory(1000),
		clipboard:  "",
		lineNumbers: false,
		viewport:   ViewPort{Top: 0, Left: 0, Width: 80, Height: 24},
	}
}

// NewEditorWithContent creates a new editor with the given content
func NewEditorWithContent(content string) *Editor {
	doc := NewDocument(content)
	return &Editor{
		document:   doc,
		cursor:     NewCursor(doc),
		history:    NewHistory(1000),
		clipboard:  "",
		lineNumbers: false,
		viewport:   ViewPort{Top: 0, Left: 0, Width: 80, Height: 24},
	}
}

// GetDocument returns the document
func (e *Editor) GetDocument() *Document {
	return e.document
}

// GetCursor returns the cursor
func (e *Editor) GetCursor() *Cursor {
	return e.cursor
}

// SetViewPort sets the viewport dimensions
func (e *Editor) SetViewPort(width, height int) {
	e.viewport.Width = width
	e.viewport.Height = height
}

// ToggleLineNumbers toggles line number display
func (e *Editor) ToggleLineNumbers() {
	e.lineNumbers = !e.lineNumbers
}

// ShowLineNumbers returns whether line numbers are enabled
func (e *Editor) ShowLineNumbers() bool {
	return e.lineNumbers
}

// LoadFile loads a file into the editor
func (e *Editor) LoadFile(filename string) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	
	e.document = NewDocument(string(content))
	e.document.SetFilename(filename)
	e.cursor = NewCursor(e.document)
	e.history.Clear()
	
	
	return nil
}


// SaveFile saves the document to a file
func (e *Editor) SaveFile(filename string) error {
	if filename == "" {
		filename = e.document.GetFilename()
	}
	
	if filename == "" {
		return fmt.Errorf("no filename specified")
	}
	
	content := e.document.GetText()
	err := ioutil.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}
	
	e.document.SetFilename(filename)
	e.document.ClearModified()
	
	return nil
}

// InsertText inserts text at the current cursor position
func (e *Editor) InsertText(text string) {
	if text == "" {
		return
	}
	
	pos := e.cursor.GetPosition()
	
	// Create change record
	change := Change{
		Type:      ChangeInsert,
		Position:  pos,
		OldText:   "",
		NewText:   text,
		Timestamp: time.Now(),
	}
	
	// Apply change to document
	newPos := pos
	for _, ch := range text {
		if ch == '\n' {
			newPos = e.document.InsertNewline(newPos)
		} else {
			newPos = e.document.InsertChar(newPos, ch)
		}
	}
	
	// Update cursor position
	e.cursor.SetPosition(newPos)
	
	// Add to history
	e.history.AddChange(change, newPos)
	
}

// DeleteText deletes text at the current cursor position
func (e *Editor) DeleteText(count int) {
	if count <= 0 {
		return
	}
	
	pos := e.cursor.GetPosition()
	
	// Collect text being deleted
	var deletedText strings.Builder
	deletePos := pos
	
	for i := 0; i < count && (deletePos.Col > 0 || deletePos.Line > 0); i++ {
		if deletePos.Col > 0 {
			ch := e.document.GetCharAt(Position{Line: deletePos.Line, Col: deletePos.Col - 1})
			deletedText.WriteRune(ch)
			deletePos = e.document.DeleteChar(deletePos)
		} else if deletePos.Line > 0 {
			deletedText.WriteRune('\n')
			deletePos = e.document.DeleteLine(deletePos)
		}
	}
	
	if deletedText.Len() == 0 {
		return
	}
	
	// Create change record
	change := Change{
		Type:      ChangeDelete,
		Position:  deletePos,
		OldText:   deletedText.String(),
		NewText:   "",
		Timestamp: time.Now(),
	}
	
	// Update cursor position
	e.cursor.SetPosition(deletePos)
	
	// Add to history
	e.history.AddChange(change, deletePos)
	
}

// Copy copies the selected text to clipboard
func (e *Editor) Copy() {
	if e.cursor.HasSelection() {
		e.clipboard = e.cursor.GetSelectionText()
	}
}

// Cut cuts the selected text to clipboard
func (e *Editor) Cut() {
	if e.cursor.HasSelection() {
		e.clipboard = e.cursor.GetSelectionText()
		e.DeleteSelection()
	}
}

// Paste pastes text from clipboard
func (e *Editor) Paste() {
	if e.clipboard != "" {
		e.InsertText(e.clipboard)
	}
}

// DeleteSelection deletes the selected text
func (e *Editor) DeleteSelection() {
	if !e.cursor.HasSelection() {
		return
	}
	
	selection := e.cursor.GetSelection()
	start := selection.Start
	end := selection.End
	
	// Ensure start is before end
	if start.Line > end.Line || (start.Line == end.Line && start.Col > end.Col) {
		start, end = end, start
	}
	
	// Get selected text
	selectedText := e.cursor.GetSelectionText()
	
	// Create change record
	change := Change{
		Type:      ChangeDelete,
		Position:  start,
		OldText:   selectedText,
		NewText:   "",
		Timestamp: time.Now(),
	}
	
	// Delete the selected text
	// This is a simplified implementation - in practice you'd want to
	// delete the entire selection range more efficiently
	e.cursor.SetPosition(start)
	e.cursor.ClearSelection()
	
	for range selectedText {
		if e.cursor.GetPosition().Col > 0 || e.cursor.GetPosition().Line > 0 {
			pos := e.cursor.GetPosition()
			if pos.Col > 0 {
				pos = e.document.DeleteChar(pos)
			} else if pos.Line > 0 {
				pos = e.document.DeleteLine(pos)
			}
			e.cursor.SetPosition(pos)
		}
	}
	
	// Add to history
	e.history.AddChange(change, e.cursor.GetPosition())
	
}

// Undo undoes the last change
func (e *Editor) Undo() {
	// Force end current group before undo
	e.history.ForceEndGroup()
	
	entry, ok := e.history.Undo()
	if !ok {
		return
	}
	
	// Apply reverse changes
	for i := len(entry.Changes) - 1; i >= 0; i-- {
		change := entry.Changes[i]
		reversed := ReverseChange(change)
		ApplyChange(e.document, reversed)
	}
	
	// Restore cursor position
	e.cursor.SetPosition(entry.Cursor)
	
}

// Redo redoes the next change
func (e *Editor) Redo() {
	// Force end current group before redo
	e.history.ForceEndGroup()
	
	entry, ok := e.history.Redo()
	if !ok {
		return
	}
	
	// Apply changes
	for _, change := range entry.Changes {
		ApplyChange(e.document, change)
	}
	
	// Restore cursor position
	e.cursor.SetPosition(entry.Cursor)
	
}

// CanUndo returns true if there are changes to undo
func (e *Editor) CanUndo() bool {
	// Force end current group to get accurate undo state
	e.history.ForceEndGroup()
	return e.history.CanUndo()
}

// CanRedo returns true if there are changes to redo
func (e *Editor) CanRedo() bool {
	// Force end current group to get accurate redo state
	e.history.ForceEndGroup()
	return e.history.CanRedo()
}

// GetVisibleLines returns the lines that should be visible in the viewport
func (e *Editor) GetVisibleLines() []string {
	lines := make([]string, 0, e.viewport.Height)
	
	for i := 0; i < e.viewport.Height; i++ {
		lineNum := e.viewport.Top + i
		if lineNum >= e.document.LineCount() {
			break
		}
		
		line := e.document.GetLine(lineNum)
		
		// Add line numbers if enabled
		if e.lineNumbers {
			lineNumStr := fmt.Sprintf("%4d │ ", lineNum+1)
			line = lineNumStr + line
		}
		
		lines = append(lines, line)
	}
	
	return lines
}

// AdjustViewPort adjusts the viewport to ensure cursor is visible
func (e *Editor) AdjustViewPort() {
	pos := e.cursor.GetPosition()
	
	// Adjust vertical position
	if pos.Line < e.viewport.Top {
		e.viewport.Top = pos.Line
	} else if pos.Line >= e.viewport.Top+e.viewport.Height {
		e.viewport.Top = pos.Line - e.viewport.Height + 1
		if e.viewport.Top < 0 {
			e.viewport.Top = 0
		}
	}
	
	// Adjust horizontal position
	if pos.Col < e.viewport.Left {
		e.viewport.Left = pos.Col
	} else if pos.Col >= e.viewport.Left+e.viewport.Width {
		e.viewport.Left = pos.Col - e.viewport.Width + 1
		if e.viewport.Left < 0 {
			e.viewport.Left = 0
		}
	}
}

// GetCursorScreenPosition returns the cursor position relative to viewport
// DEPRECATED: Use GetCursorContentPosition() instead for explicit coordinate types
func (e *Editor) GetCursorScreenPosition() (int, int) {
	pos := e.cursor.GetPosition()
	screenRow := pos.Line - e.viewport.Top
	screenCol := pos.Col - e.viewport.Left
	
	// Account for line numbers
	if e.lineNumbers {
		screenCol += 6 // "1234 │ "
	}
	
	// Note: We return the calculated screen position even if it's outside viewport bounds
	// The caller is responsible for checking if the cursor is visible
	return screenRow, screenCol
}

// GetCursorContentPosition returns the cursor position in content coordinates.
// This is the explicit coordinate transformation from DocumentPos to ContentPos.
func (e *Editor) GetCursorContentPosition() ContentPos {
	docPos := e.GetCursorDocumentPosition()
	return e.TransformDocumentToContent(docPos)
}

// GetCursorDocumentPosition returns the cursor position in document coordinates.
func (e *Editor) GetCursorDocumentPosition() DocumentPos {
	pos := e.cursor.GetPosition()
	return DocumentPos{Line: pos.Line, Col: pos.Col}
}

// TransformDocumentToContent converts document coordinates to content coordinates.
// This transformation includes viewport offset and line number offset.
func (e *Editor) TransformDocumentToContent(docPos DocumentPos) ContentPos {
	// STEP 1: Apply viewport offset (document → viewport)
	contentPos := ContentPos{
		Line: docPos.Line - e.viewport.Top,
		Col:  docPos.Col - e.viewport.Left,
	}
	
	// STEP 2: Apply line number offset (viewport → content)
	if e.lineNumbers {
		contentPos.Col += 6 // "  1 │ " prefix
	}
	
	return contentPos
}

// GetViewportInfo returns current viewport state for debugging.
func (e *Editor) GetViewportInfo() ViewportInfo {
	return ViewportInfo{
		Top:         e.viewport.Top,
		Left:        e.viewport.Left,
		Width:       e.viewport.Width,
		Height:      e.viewport.Height,
		LineNumbers: e.lineNumbers,
	}
}

// ValidateDocumentPos checks if a document position is within bounds.
func (e *Editor) ValidateDocumentPos(pos DocumentPos) error {
	if !pos.IsValid() {
		return NewDocumentCoordinateError(pos, "negative coordinates")
	}
	
	if pos.Line >= e.document.LineCount() {
		return NewDocumentCoordinateError(pos, 
			fmt.Sprintf("line %d >= document line count %d", pos.Line, e.document.LineCount()))
	}
	
	lineLength := e.document.GetLineLength(pos.Line)
	if pos.Col > lineLength {
		return NewDocumentCoordinateError(pos,
			fmt.Sprintf("column %d > line length %d", pos.Col, lineLength))
	}
	
	return nil
}

// ValidateContentPos checks if a content position is within bounds.
func (e *Editor) ValidateContentPos(pos ContentPos) error {
	if !pos.IsValid() {
		return NewContentCoordinateError(pos, "negative coordinates")
	}
	
	if pos.Line >= e.viewport.Height {
		return NewContentCoordinateError(pos,
			fmt.Sprintf("line %d >= viewport height %d", pos.Line, e.viewport.Height))
	}
	
	// CRITICAL CHECK: Content position should account for line numbers
	if e.lineNumbers && pos.Col < 6 {
		return NewContentCoordinateError(pos,
			fmt.Sprintf("column %d < 6 but line numbers enabled (missing line number offset)", pos.Col))
	}
	
	// Calculate maximum content width
	maxContentWidth := e.viewport.Width
	if pos.Col > maxContentWidth {
		return NewContentCoordinateError(pos,
			fmt.Sprintf("column %d > viewport width %d", pos.Col, maxContentWidth))
	}
	
	return nil
}

// FindText searches for text in the document starting from current cursor position
func (e *Editor) FindText(searchText string, caseSensitive bool) *Position {
	if searchText == "" {
		return nil
	}
	
	pos := e.cursor.GetPosition()
	text := e.document.GetText()
	
	if !caseSensitive {
		searchText = strings.ToLower(searchText)
		text = strings.ToLower(text)
	}
	
	// Convert position to text offset
	lines := strings.Split(text, "\n")
	offset := 0
	for i := 0; i < pos.Line && i < len(lines); i++ {
		offset += len(lines[i]) + 1 // +1 for newline
	}
	offset += pos.Col
	
	// Search from current position
	index := strings.Index(text[offset:], searchText)
	if index == -1 {
		// Wrap around search
		index = strings.Index(text, searchText)
		if index == -1 {
			return nil
		}
	} else {
		index += offset
	}
	
	// Convert back to position
	return e.offsetToPosition(index)
}

// ReplaceText replaces text at the current cursor position
func (e *Editor) ReplaceText(oldText, newText string, caseSensitive bool) bool {
	if oldText == "" {
		return false
	}
	
	pos := e.cursor.GetPosition()
	text := e.document.GetText()
	
	searchText := oldText
	if !caseSensitive {
		searchText = strings.ToLower(oldText)
		text = strings.ToLower(text)
	}
	
	// Convert position to text offset
	lines := strings.Split(text, "\n")
	offset := 0
	for i := 0; i < pos.Line && i < len(lines); i++ {
		offset += len(lines[i]) + 1 // +1 for newline
	}
	offset += pos.Col
	
	// Check if text at cursor matches
	if offset+len(searchText) <= len(text) && text[offset:offset+len(searchText)] == searchText {
		// Delete old text
		for i := 0; i < len(oldText); i++ {
			e.DeleteText(1)
		}
		// Insert new text
		e.InsertText(newText)
		return true
	}
	
	return false
}

// GotoLine moves cursor to specified line
func (e *Editor) GotoLine(lineNum int) {
	if lineNum < 1 {
		lineNum = 1
	}
	if lineNum > e.document.LineCount() {
		lineNum = e.document.LineCount()
	}
	
	newPos := Position{Line: lineNum - 1, Col: 0}
	e.cursor.SetPosition(newPos)
}

// offsetToPosition converts text offset to Position
func (e *Editor) offsetToPosition(offset int) *Position {
	text := e.document.GetText()
	lines := strings.Split(text, "\n")
	
	currentOffset := 0
	for lineNum, line := range lines {
		if currentOffset+len(line) >= offset {
			return &Position{
				Line: lineNum,
				Col:  offset - currentOffset,
			}
		}
		currentOffset += len(line) + 1 // +1 for newline
	}
	
	// If we get here, offset is at end of document
	return &Position{
		Line: len(lines) - 1,
		Col:  len(lines[len(lines)-1]),
	}
}
// GetViewPort returns the current viewport
func (e *Editor) GetViewPort() ViewPort {
	return e.viewport
}
