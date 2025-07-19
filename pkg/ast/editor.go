package ast

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"unicode/utf8"
)

// Editor manages the document and cursor.
// Uses CursorManager for unified coordinate handling.
type Editor struct {
	document      *Document
	cursorManager *CursorManager
	clipboard     string
	lineNumbers   bool
	viewport      *Viewport
}

// GetViewport returns the current viewport
func (e *Editor) GetViewport() *Viewport {
	return e.viewport
}

// calculateLineNumberWidth computes the display width needed for line numbers
func calculateLineNumberWidth(doc *Document) int {
	maxLines := doc.LineCount()
	if maxLines == 0 {
		maxLines = 1 // Minimum for empty documents
	}
	digits := len(fmt.Sprintf("%d", maxLines))
	formatStr := fmt.Sprintf("%%%dd │ ", digits)
	sample := fmt.Sprintf(formatStr, maxLines)
	return utf8.RuneCountInString(sample)
}

// NewEditor creates a new editor with an empty document
func NewEditor() *Editor {
	doc := NewEmptyDocument()
	lineNumberWidth := calculateLineNumberWidth(doc)
	viewport := NewViewport(0, 0, 80, 24, lineNumberWidth, 4) // Default: with line numbers, 4-space tabs
	cursorManager := NewCursorManager(viewport, doc)
	
	return &Editor{
		document:      doc,
		cursorManager: cursorManager,
		clipboard:     "",
		lineNumbers:   true,
		viewport:      viewport,
	}
}

// NewEditorWithContent creates a new editor with the given content
func NewEditorWithContent(content string) *Editor {
	doc := NewDocument(content)
	lineNumberWidth := calculateLineNumberWidth(doc)
	viewport := NewViewport(0, 0, 80, 24, lineNumberWidth, 4) // Default: with line numbers, 4-space tabs
	cursorManager := NewCursorManager(viewport, doc)
	
	return &Editor{
		document:      doc,
		cursorManager: cursorManager,
		clipboard:     "",
		lineNumbers:   true,
		viewport:      viewport,
	}
}

// GetDocument returns the document
func (e *Editor) GetDocument() *Document {
	return e.document
}

// GetCursor returns the cursor manager
func (e *Editor) GetCursor() *CursorManager {
	return e.cursorManager
}

// SetViewPort sets the viewport dimensions
func (e *Editor) SetViewPort(width, height int) {
	// Create new viewport with updated dimensions
	newViewport := e.viewport.WithDimensions(width, height)
	e.viewport = newViewport
	e.cursorManager.UpdateViewport(newViewport)
}

// ToggleLineNumbers toggles line number display
func (e *Editor) ToggleLineNumbers() {
	e.lineNumbers = !e.lineNumbers
	
	// Update viewport with calculated line number width
	lineNumberWidth := 0
	if e.lineNumbers {
		lineNumberWidth = e.calculateLineNumberWidth()
	}
	
	newViewport := NewViewport(
		e.viewport.GetTopLine(),
		e.viewport.GetLeftColumn(),
		e.viewport.GetWidth(),
		e.viewport.GetHeight(),
		lineNumberWidth,
		e.viewport.GetTabWidth(),
	)
	
	e.viewport = newViewport
	e.cursorManager.UpdateViewport(newViewport)
}

// calculateLineNumberWidth calculates the width needed for line number display
func (e *Editor) calculateLineNumberWidth() int {
	maxLines := e.document.LineCount()
	if maxLines == 0 {
		maxLines = 1 // Minimum for empty documents
	}
	
	// Calculate digits needed: log10(maxLines) + 1
	digits := len(fmt.Sprintf("%d", maxLines))
	
	// Format string: "%Nd │ " where N is the digit count
	formatStr := fmt.Sprintf("%%%dd │ ", digits)
	
	// Calculate actual width by measuring formatted output in runes (not bytes)
	sample := fmt.Sprintf(formatStr, maxLines)
	return utf8.RuneCountInString(sample)
}

// ShowLineNumbers returns whether line numbers are enabled
func (e *Editor) ShowLineNumbers() bool {
	return e.lineNumbers
}

// GetLineNumberWidth returns the viewport's line number width, or 0 if disabled
func (e *Editor) GetLineNumberWidth() int {
	if !e.lineNumbers {
		return 0
	}
	return e.viewport.GetLineNumberWidth()
}

// FormatLineNumber formats a line number using the appropriate width
func (e *Editor) FormatLineNumber(lineNum int) string {
	if !e.lineNumbers {
		return ""
	}
	
	maxLines := e.document.LineCount()
	if maxLines == 0 {
		maxLines = 1
	}
	
	// Calculate digits needed for the total number of lines
	digits := len(fmt.Sprintf("%d", maxLines))
	
	// Create format string: "%Nd │ " where N is the digit count
	formatStr := fmt.Sprintf("%%%dd │ ", digits)
	
	return fmt.Sprintf(formatStr, lineNum)
}

// LoadFile loads a file into the editor
func (e *Editor) LoadFile(filename string) error {
	var content []byte
	
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		content = []byte{}
	} else {
		var err error
		content, err = ioutil.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", filename, err)
		}
	}
	
	e.document = NewDocument(string(content))
	e.document.SetFilename(filename)
	// Update cursor manager to use the new document for validation
	e.cursorManager.UpdateValidator(e.document)
	// Reset cursor position to start of document
	e.cursorManager.SetBufferPos(BufferPos{Line: 0, Col: 0})
	
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
	
	pos := e.cursorManager.GetBufferPos()
	
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
	e.cursorManager.SetBufferPos(newPos)
	
}

// DeleteText deletes text at the current cursor position
func (e *Editor) DeleteText(count int) {
	if count <= 0 {
		return
	}
	
	pos := e.cursorManager.GetBufferPos()
	
	// Delete text
	deletePos := pos
	
	for i := 0; i < count && (deletePos.Col > 0 || deletePos.Line > 0); i++ {
		if deletePos.Col > 0 {
			deletePos = e.document.DeleteChar(deletePos)
		} else if deletePos.Line > 0 {
			deletePos = e.document.DeleteLine(deletePos)
		}
	}
	
	// Update cursor position
	e.cursorManager.SetBufferPos(deletePos)
	
}

// Copy copies the selected text to clipboard
func (e *Editor) Copy() {
	if e.cursorManager.HasSelection() {
		e.clipboard = e.GetSelectionText()
	}
}

// Cut cuts the selected text to clipboard
func (e *Editor) Cut() {
	if e.cursorManager.HasSelection() {
		e.clipboard = e.GetSelectionText()
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
	if !e.cursorManager.HasSelection() {
		return
	}
	
	selection := e.cursorManager.GetSelection()
	start := selection.Start
	end := selection.End
	
	// Ensure start is before end
	if start.Line > end.Line || (start.Line == end.Line && start.Col > end.Col) {
		start, end = end, start
	}
	
	// Get selected text
	selectedText := e.GetSelectionText()
	
	// Delete the selected text
	// This is a simplified implementation - in practice you'd want to
	// delete the entire selection range more efficiently
	e.cursorManager.SetBufferPos(start)
	e.cursorManager.ClearSelection()
	
	for range selectedText {
		if e.cursorManager.GetBufferPos().Col > 0 || e.cursorManager.GetBufferPos().Line > 0 {
			pos := e.cursorManager.GetBufferPos()
			if pos.Col > 0 {
				pos = e.document.DeleteChar(pos)
			} else if pos.Line > 0 {
				pos = e.document.DeleteLine(pos)
			}
			e.cursorManager.SetBufferPos(pos)
		}
	}
	
}


// GetVisibleLines returns the lines that should be visible in the viewport
func (e *Editor) GetVisibleLines() []string {
	lines := make([]string, 0, e.viewport.GetHeight())
	
	for i := 0; i < e.viewport.GetHeight(); i++ {
		lineNum := e.viewport.GetTopLine() + i
		if lineNum >= e.document.LineCount() {
			break
		}
		
		line := e.document.GetLine(lineNum)
		
		// Add line numbers if enabled
		if e.lineNumbers {
			lineNumStr := e.FormatLineNumber(lineNum + 1)
			line = lineNumStr + line
		}
		
		lines = append(lines, line)
	}
	
	return lines
}

// AdjustViewPort adjusts the viewport to ensure cursor is visible
func (e *Editor) AdjustViewPort() {
	pos := e.cursorManager.GetBufferPos()
	
	newTopLine := e.viewport.GetTopLine()
	newLeftColumn := e.viewport.GetLeftColumn()
	
	// Adjust vertical position
	if pos.Line < newTopLine {
		newTopLine = pos.Line
	} else if pos.Line >= newTopLine+e.viewport.GetHeight() {
		newTopLine = pos.Line - e.viewport.GetHeight() + 1
		if newTopLine < 0 {
			newTopLine = 0
		}
	}
	
	// Adjust horizontal position
	if pos.Col < newLeftColumn {
		newLeftColumn = pos.Col
	} else if pos.Col >= newLeftColumn+e.viewport.GetWidth()-e.viewport.GetLineNumberWidth() {
		newLeftColumn = pos.Col - e.viewport.GetWidth() + e.viewport.GetLineNumberWidth() + 1
		if newLeftColumn < 0 {
			newLeftColumn = 0
		}
	}
	
	// Update viewport if needed
	if newTopLine != e.viewport.GetTopLine() || newLeftColumn != e.viewport.GetLeftColumn() {
		newViewport := NewViewport(
			newTopLine,
			newLeftColumn,
			e.viewport.GetWidth(),
			e.viewport.GetHeight(),
			e.viewport.GetLineNumberWidth(),
			e.viewport.GetTabWidth(),
		)
		e.viewport = newViewport
		e.cursorManager.UpdateViewport(newViewport)
	}
}


// GetCursorBufferPosition returns the cursor position in buffer coordinates
func (e *Editor) GetCursorBufferPosition() BufferPos {
	return e.cursorManager.GetBufferPos()
}

// ============================================================================
// CURSOR MOVEMENT METHODS
// ============================================================================
//
// ARCHITECTURAL PATTERN:
// These methods implement the Editor orchestration pattern where:
// 1. Editor calls Document methods for content-aware movement logic
// 2. Editor updates CursorManager with new positions
// 3. Editor adjusts viewport to keep cursor visible
//
// This follows the document-centric architecture recommended by modern text
// editor research, avoiding the Xi-editor pitfall of over-modularization.
//
// DESIGN PRINCIPLE:
// Each method follows the pattern:
//   currentPos := e.cursorManager.GetBufferPos()
//   newPos := e.document.MoveCursor[Direction](currentPos, ...)
//   e.cursorManager.SetBufferPos(newPos)
//   e.AdjustViewPort()

// MoveCursorRight moves cursor right by one character with line wrapping.
func (e *Editor) MoveCursorRight() {
	currentPos := e.cursorManager.GetBufferPos()
	newPos := e.document.MoveCursorRight(currentPos)
	e.cursorManager.SetBufferPos(newPos)
	e.cursorManager.SetDesiredColumn(newPos.Col)
	e.AdjustViewPort()
}

// MoveCursorLeft moves cursor left by one character with line wrapping.
func (e *Editor) MoveCursorLeft() {
	currentPos := e.cursorManager.GetBufferPos()
	newPos := e.document.MoveCursorLeft(currentPos)
	e.cursorManager.SetBufferPos(newPos)
	e.cursorManager.SetDesiredColumn(newPos.Col)
	e.AdjustViewPort()
}

// MoveCursorUp moves cursor up by one line with desired column preservation.
func (e *Editor) MoveCursorUp() {
	currentPos := e.cursorManager.GetBufferPos()
	desiredCol := e.cursorManager.GetDesiredColumn()
	newPos, _ := e.document.MoveCursorUp(currentPos, desiredCol)
	e.cursorManager.SetBufferPosWithDesiredColumn(newPos, true) // Preserve desired column
	e.AdjustViewPort()
}

// MoveCursorDown moves cursor down by one line with desired column preservation.
func (e *Editor) MoveCursorDown() {
	currentPos := e.cursorManager.GetBufferPos()
	desiredCol := e.cursorManager.GetDesiredColumn()
	newPos, _ := e.document.MoveCursorDown(currentPos, desiredCol)
	e.cursorManager.SetBufferPosWithDesiredColumn(newPos, true) // Preserve desired column
	e.AdjustViewPort()
}

// MoveCursorToLineStart moves cursor to beginning of current line.
func (e *Editor) MoveCursorToLineStart() {
	currentPos := e.cursorManager.GetBufferPos()
	newPos := e.document.MoveCursorToLineStart(currentPos)
	e.cursorManager.SetBufferPos(newPos)
	e.cursorManager.SetDesiredColumn(newPos.Col)
	e.AdjustViewPort()
}

// MoveCursorToLineEnd moves cursor to end of current line.
func (e *Editor) MoveCursorToLineEnd() {
	currentPos := e.cursorManager.GetBufferPos()
	newPos := e.document.MoveCursorToLineEnd(currentPos)
	e.cursorManager.SetBufferPos(newPos)
	e.cursorManager.SetDesiredColumn(newPos.Col)
	e.AdjustViewPort()
}

// MoveCursorToDocumentStart moves cursor to beginning of document.
func (e *Editor) MoveCursorToDocumentStart() {
	currentPos := e.cursorManager.GetBufferPos()
	newPos := e.document.MoveCursorToDocumentStart(currentPos)
	e.cursorManager.SetBufferPos(newPos)
	e.cursorManager.SetDesiredColumn(newPos.Col)
	e.AdjustViewPort()
}

// MoveCursorToDocumentEnd moves cursor to end of document.
func (e *Editor) MoveCursorToDocumentEnd() {
	currentPos := e.cursorManager.GetBufferPos()
	newPos := e.document.MoveCursorToDocumentEnd(currentPos)
	e.cursorManager.SetBufferPos(newPos)
	e.cursorManager.SetDesiredColumn(newPos.Col)
	e.AdjustViewPort()
}

// MoveCursorWordLeft moves cursor to start of previous word.
func (e *Editor) MoveCursorWordLeft() {
	currentPos := e.cursorManager.GetBufferPos()
	newPos := e.document.MoveCursorWordLeft(currentPos)
	e.cursorManager.SetBufferPos(newPos)
	e.cursorManager.SetDesiredColumn(newPos.Col)
	e.AdjustViewPort()
}

// MoveCursorWordRight moves cursor to start of next word.
func (e *Editor) MoveCursorWordRight() {
	currentPos := e.cursorManager.GetBufferPos()
	newPos := e.document.MoveCursorWordRight(currentPos)
	e.cursorManager.SetBufferPos(newPos)
	e.cursorManager.SetDesiredColumn(newPos.Col)
	e.AdjustViewPort()
}

// GetSelectionText returns the text content of the current selection.
// This method properly implements the document-centric architecture where
// the Editor orchestrates between Document (content) and CursorManager (selection state).
func (e *Editor) GetSelectionText() string {
	selection := e.cursorManager.GetSelection()
	return e.document.GetSelectionText(selection)
}

// FindText searches for text in the document starting from current cursor position
func (e *Editor) FindText(searchText string, caseSensitive bool) *BufferPos {
	if searchText == "" {
		return nil
	}
	
	pos := e.cursorManager.GetBufferPos()
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
	
	pos := e.cursorManager.GetBufferPos()
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
	
	newPos := BufferPos{Line: lineNum - 1, Col: 0}
	e.cursorManager.SetBufferPos(newPos)
}

// offsetToPosition converts text offset to BufferPos
func (e *Editor) offsetToPosition(offset int) *BufferPos {
	text := e.document.GetText()
	lines := strings.Split(text, "\n")
	
	currentOffset := 0
	for lineNum, line := range lines {
		if currentOffset+len(line) >= offset {
			return &BufferPos{
				Line: lineNum,
				Col:  offset - currentOffset,
			}
		}
		currentOffset += len(line) + 1 // +1 for newline
	}
	
	// If we get here, offset is at end of document
	return &BufferPos{
		Line: len(lines) - 1,
		Col:  len(lines[len(lines)-1]),
	}
}
