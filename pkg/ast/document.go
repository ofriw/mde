package ast

import (
	"strings"
	"unicode"
)

// Document represents the entire document with text content and metadata
type Document struct {
	lines    []Line
	filename string
	modified bool
}

// Line represents a single line of text with metadata
type Line struct {
	text    string
	length  int
	tokens  []Token // For future syntax highlighting
}

// Token represents a syntax token for highlighting
type Token struct {
	start int
	end   int
	kind  TokenKind
}

// NewToken creates a new token
func NewToken(start, end int, kind TokenKind) Token {
	return Token{
		start: start,
		end:   end,
		kind:  kind,
	}
}

// TokenKind represents different types of tokens
type TokenKind int

const (
	TokenText TokenKind = iota
	TokenKeyword
	TokenString
	TokenComment
	TokenNumber
	// Markdown-specific tokens
	TokenHeading
	TokenBold
	TokenItalic
	TokenCode
	TokenCodeBlock
	TokenLink
	TokenLinkText
	TokenLinkURL
	TokenImage
	TokenQuote
	TokenList
	TokenTable
	TokenDelimiter
)

// Start returns the start position of the token
func (t Token) Start() int {
	return t.start
}

// End returns the end position of the token
func (t Token) End() int {
	return t.end
}

// Kind returns the kind of the token
func (t Token) Kind() TokenKind {
	return t.kind
}


// Selection is defined in cursor.go as part of the CursorManager architecture

// NewDocument creates a new document with initial content
func NewDocument(content string) *Document {
	lines := strings.Split(content, "\n")
	doc := &Document{
		lines: make([]Line, len(lines)),
	}
	
	for i, line := range lines {
		doc.lines[i] = Line{
			text:   line,
			length: len([]rune(line)), // Handle unicode properly
		}
	}
	
	return doc
}

// NewEmptyDocument creates a new empty document
func NewEmptyDocument() *Document {
	return &Document{
		lines: []Line{{text: "", length: 0}},
	}
}

// LineCount returns the number of lines in the document
func (d *Document) LineCount() int {
	return len(d.lines)
}

// GetLine returns the text content of a specific line
func (d *Document) GetLine(lineNum int) string {
	if lineNum < 0 || lineNum >= len(d.lines) {
		return ""
	}
	return d.lines[lineNum].text
}

// GetLineLength returns the length of a specific line
func (d *Document) GetLineLength(lineNum int) int {
	if lineNum < 0 || lineNum >= len(d.lines) {
		return 0
	}
	return d.lines[lineNum].length
}

// InsertChar inserts a character at the specified position
func (d *Document) InsertChar(pos BufferPos, ch rune) BufferPos {
	if pos.Line < 0 || pos.Line >= len(d.lines) {
		return pos
	}
	
	line := &d.lines[pos.Line]
	runes := []rune(line.text)
	
	// Clamp position to valid range
	if pos.Col < 0 {
		pos.Col = 0
	} else if pos.Col > len(runes) {
		pos.Col = len(runes)
	}
	
	// Insert character
	newRunes := make([]rune, len(runes)+1)
	copy(newRunes[:pos.Col], runes[:pos.Col])
	newRunes[pos.Col] = ch
	copy(newRunes[pos.Col+1:], runes[pos.Col:])
	
	line.text = string(newRunes)
	line.length = len(newRunes)
	d.modified = true
	
	return BufferPos{Line: pos.Line, Col: pos.Col + 1}
}

// DeleteChar deletes a character at the specified position
func (d *Document) DeleteChar(pos BufferPos) BufferPos {
	if pos.Line < 0 || pos.Line >= len(d.lines) {
		return pos
	}
	
	line := &d.lines[pos.Line]
	runes := []rune(line.text)
	
	if pos.Col <= 0 || pos.Col > len(runes) {
		return pos
	}
	
	// Delete character
	newRunes := make([]rune, len(runes)-1)
	copy(newRunes[:pos.Col-1], runes[:pos.Col-1])
	copy(newRunes[pos.Col-1:], runes[pos.Col:])
	
	line.text = string(newRunes)
	line.length = len(newRunes)
	d.modified = true
	
	return BufferPos{Line: pos.Line, Col: pos.Col - 1}
}

// InsertNewline inserts a newline at the specified position
func (d *Document) InsertNewline(pos BufferPos) BufferPos {
	if pos.Line < 0 || pos.Line >= len(d.lines) {
		return pos
	}
	
	line := &d.lines[pos.Line]
	runes := []rune(line.text)
	
	// Clamp position
	if pos.Col < 0 {
		pos.Col = 0
	} else if pos.Col > len(runes) {
		pos.Col = len(runes)
	}
	
	// Split line
	leftPart := string(runes[:pos.Col])
	rightPart := string(runes[pos.Col:])
	
	// Update current line
	line.text = leftPart
	line.length = len([]rune(leftPart))
	
	// Insert new line
	newLine := Line{
		text:   rightPart,
		length: len([]rune(rightPart)),
	}
	
	newLines := make([]Line, len(d.lines)+1)
	copy(newLines[:pos.Line+1], d.lines[:pos.Line+1])
	newLines[pos.Line+1] = newLine
	copy(newLines[pos.Line+2:], d.lines[pos.Line+1:])
	
	d.lines = newLines
	d.modified = true
	
	return BufferPos{Line: pos.Line + 1, Col: 0}
}

// DeleteLine deletes a line and merges with previous if needed
func (d *Document) DeleteLine(pos BufferPos) BufferPos {
	if pos.Line <= 0 || pos.Line >= len(d.lines) {
		return pos
	}
	
	// Get content of line being deleted
	deletedLine := d.lines[pos.Line]
	
	// Merge with previous line
	prevLine := &d.lines[pos.Line-1]
	newCol := prevLine.length
	prevLine.text += deletedLine.text
	prevLine.length = len([]rune(prevLine.text))
	
	// Remove the line
	newLines := make([]Line, len(d.lines)-1)
	copy(newLines[:pos.Line], d.lines[:pos.Line])
	copy(newLines[pos.Line:], d.lines[pos.Line+1:])
	
	d.lines = newLines
	d.modified = true
	
	return BufferPos{Line: pos.Line - 1, Col: newCol}
}

// GetText returns the full text content of the document
func (d *Document) GetText() string {
	lines := make([]string, len(d.lines))
	for i, line := range d.lines {
		lines[i] = line.text
	}
	return strings.Join(lines, "\n")
}

// SetFilename sets the filename for the document
func (d *Document) SetFilename(filename string) {
	d.filename = filename
}

// GetFilename returns the document filename
func (d *Document) GetFilename() string {
	return d.filename
}

// IsModified returns whether the document has been modified
func (d *Document) IsModified() bool {
	return d.modified
}

// ClearModified clears the modified flag
func (d *Document) ClearModified() {
	d.modified = false
}

// ValidatePosition ensures a position is within document bounds
func (d *Document) ValidatePosition(pos BufferPos) BufferPos {
	if pos.Line < 0 {
		pos.Line = 0
	} else if pos.Line >= len(d.lines) {
		pos.Line = len(d.lines) - 1
	}
	
	lineLength := d.GetLineLength(pos.Line)
	if pos.Col < 0 {
		pos.Col = 0
	} else if pos.Col > lineLength {
		pos.Col = lineLength
	}
	
	return pos
}

// ValidateBufferPos implements PositionValidator interface
func (d *Document) ValidateBufferPos(pos BufferPos) error {
	if pos.Line < 0 {
		return NewBufferCoordinateError(pos, "line number cannot be negative")
	}
	if pos.Line >= len(d.lines) {
		return NewBufferCoordinateError(pos, "line number exceeds document length")
	}
	if pos.Col < 0 {
		return NewBufferCoordinateError(pos, "column number cannot be negative")
	}
	
	lineLength := d.GetLineLength(pos.Line)
	if pos.Col > lineLength {
		return NewBufferCoordinateError(pos, "column number exceeds line length")
	}
	
	return nil
}

// GetCharAt returns the character at the specified position
func (d *Document) GetCharAt(pos BufferPos) rune {
	if pos.Line < 0 || pos.Line >= len(d.lines) {
		return 0
	}
	
	line := d.lines[pos.Line]
	runes := []rune(line.text)
	
	if pos.Col < 0 || pos.Col >= len(runes) {
		return 0
	}
	
	return runes[pos.Col]
}

// SetLineTokens sets syntax highlighting tokens for a specific line
func (d *Document) SetLineTokens(lineNum int, tokens []Token) {
	if lineNum < 0 || lineNum >= len(d.lines) {
		return
	}
	
	d.lines[lineNum].tokens = tokens
}

// GetLineTokens returns syntax highlighting tokens for a specific line
func (d *Document) GetLineTokens(lineNum int) []Token {
	if lineNum < 0 || lineNum >= len(d.lines) {
		return nil
	}
	
	return d.lines[lineNum].tokens
}

// FindWordStart finds the start of the word at the given position
func (d *Document) FindWordStart(pos BufferPos) BufferPos {
	if pos.Line < 0 || pos.Line >= len(d.lines) {
		return pos
	}
	
	line := d.lines[pos.Line]
	runes := []rune(line.text)
	
	if pos.Col <= 0 {
		return BufferPos{Line: pos.Line, Col: 0}
	}
	
	col := pos.Col
	if col > len(runes) {
		col = len(runes)
	}
	
	// Move back while we're still in the same word
	for col > 0 && !unicode.IsSpace(runes[col-1]) {
		col--
	}
	
	return BufferPos{Line: pos.Line, Col: col}
}

// FindWordEnd finds the end of the word at the given position
func (d *Document) FindWordEnd(pos BufferPos) BufferPos {
	if pos.Line < 0 || pos.Line >= len(d.lines) {
		return pos
	}
	
	line := d.lines[pos.Line]
	runes := []rune(line.text)
	
	col := pos.Col
	if col < 0 {
		col = 0
	}
	
	// Move forward while we're still in the same word
	for col < len(runes) && !unicode.IsSpace(runes[col]) {
		col++
	}
	
	return BufferPos{Line: pos.Line, Col: col}
}

// ============================================================================
// CURSOR MOVEMENT METHODS
// ============================================================================
//
// ARCHITECTURAL DESIGN:
// Following modern text editor best practices (CodeMirror 6, Xi-editor research),
// cursor movement logic resides in the Document since it requires content-aware
// operations. The Document knows its structure and can make informed decisions
// about cursor positioning.
//
// DESIGN PRINCIPLES:
// 1. Document-Centric: Content-aware operations belong with content
// 2. Pure Functions: Movement methods return new positions without side effects
// 3. Bounds Checking: All methods ensure positions remain within document bounds
// 4. Desired Column: Vertical movement preserves intended column position
// 5. Unicode Aware: Proper handling of multi-byte characters
//
// SOURCES:
// - CodeMirror 6 architecture: https://codemirror.net/docs/ref
// - Xi-editor retrospective: https://raphlinus.github.io/xi/2020/06/27/xi-retrospective.html
// - Cursor movement subtleties: https://munificent.github.io/
//
// USAGE PATTERN:
// The Editor calls these methods and updates the CursorManager with the result:
//   newPos := editor.document.MoveCursorRight(currentPos)
//   editor.cursorManager.SetBufferPos(newPos)

// MoveCursorRight moves cursor right by one character.
// Handles line wrapping: moves to start of next line if at end of current line.
func (d *Document) MoveCursorRight(pos BufferPos) BufferPos {
	pos = d.ValidatePosition(pos)
	
	lineLength := d.GetLineLength(pos.Line)
	if pos.Col < lineLength {
		return BufferPos{Line: pos.Line, Col: pos.Col + 1}
	}
	
	if pos.Line < d.LineCount()-1 {
		return BufferPos{Line: pos.Line + 1, Col: 0}
	}
	
	return pos
}

// MoveCursorLeft moves cursor left by one character.
// Handles line wrapping: moves to end of previous line if at start of current line.
func (d *Document) MoveCursorLeft(pos BufferPos) BufferPos {
	pos = d.ValidatePosition(pos)
	
	if pos.Col > 0 {
		return BufferPos{Line: pos.Line, Col: pos.Col - 1}
	}
	
	if pos.Line > 0 {
		prevLine := pos.Line - 1
		return BufferPos{Line: prevLine, Col: d.GetLineLength(prevLine)}
	}
	
	return pos
}

// MoveCursorUp moves cursor up by one line with desired column preservation.
// Returns new position and whether desired column was preserved.
func (d *Document) MoveCursorUp(pos BufferPos, desiredCol int) (BufferPos, bool) {
	pos = d.ValidatePosition(pos)
	
	if pos.Line <= 0 {
		return pos, false
	}
	
	newLine := pos.Line - 1
	lineLength := d.GetLineLength(newLine)
	
	newCol := desiredCol
	if newCol > lineLength {
		newCol = lineLength
	}
	
	preservedDesired := (newCol == desiredCol)
	return BufferPos{Line: newLine, Col: newCol}, preservedDesired
}

// MoveCursorDown moves cursor down by one line with desired column preservation.
// Returns new position and whether desired column was preserved.
func (d *Document) MoveCursorDown(pos BufferPos, desiredCol int) (BufferPos, bool) {
	pos = d.ValidatePosition(pos)
	
	if pos.Line >= d.LineCount()-1 {
		return pos, false
	}
	
	newLine := pos.Line + 1
	lineLength := d.GetLineLength(newLine)
	
	newCol := desiredCol
	if newCol > lineLength {
		newCol = lineLength
	}
	
	preservedDesired := (newCol == desiredCol)
	return BufferPos{Line: newLine, Col: newCol}, preservedDesired
}

// MoveCursorToLineStart moves cursor to beginning of current line.
func (d *Document) MoveCursorToLineStart(pos BufferPos) BufferPos {
	pos = d.ValidatePosition(pos)
	return BufferPos{Line: pos.Line, Col: 0}
}

// MoveCursorToLineEnd moves cursor to end of current line.
func (d *Document) MoveCursorToLineEnd(pos BufferPos) BufferPos {
	pos = d.ValidatePosition(pos)
	return BufferPos{Line: pos.Line, Col: d.GetLineLength(pos.Line)}
}

// MoveCursorToDocumentStart moves cursor to beginning of document.
func (d *Document) MoveCursorToDocumentStart(pos BufferPos) BufferPos {
	return BufferPos{Line: 0, Col: 0}
}

// MoveCursorToDocumentEnd moves cursor to end of document.
func (d *Document) MoveCursorToDocumentEnd(pos BufferPos) BufferPos {
	lastLine := d.LineCount() - 1
	return BufferPos{Line: lastLine, Col: d.GetLineLength(lastLine)}
}

// MoveCursorWordLeft moves cursor to start of previous word.
func (d *Document) MoveCursorWordLeft(pos BufferPos) BufferPos {
	pos = d.ValidatePosition(pos)
	
	line := d.lines[pos.Line]
	runes := []rune(line.text)
	col := pos.Col
	
	if col > len(runes) {
		col = len(runes)
	}
	
	// Step 1: Skip backwards over whitespace (may cross lines)
	for {
		// Skip whitespace on current line
		for col > 0 && unicode.IsSpace(runes[col-1]) {
			col--
		}
		
		// If we found a non-space character, break to handle the word
		if col > 0 {
			break
		}
		
		// If we're at start of line, move to previous line
		if pos.Line > 0 {
			pos.Line--
			line = d.lines[pos.Line]
			runes = []rune(line.text)
			col = len(runes)
			
			// Continue loop to skip whitespace on previous line
		} else {
			// We're at the start of the document
			return BufferPos{Line: 0, Col: 0}
		}
	}
	
	// Step 2: Skip backwards over the current word
	for col > 0 && !unicode.IsSpace(runes[col-1]) {
		col--
	}
	
	return BufferPos{Line: pos.Line, Col: col}
}

// MoveCursorWordRight moves cursor to start of next word.
func (d *Document) MoveCursorWordRight(pos BufferPos) BufferPos {
	pos = d.ValidatePosition(pos)
	
	// Step 1: Skip forward over the current word on current line
	line := d.lines[pos.Line]
	runes := []rune(line.text)
	col := pos.Col
	
	// Skip over current word
	for col < len(runes) && !unicode.IsSpace(runes[col]) {
		col++
	}
	
	// Step 2: Skip forward over whitespace (may cross lines)
	for {
		// Skip whitespace on current line
		for col < len(runes) && unicode.IsSpace(runes[col]) {
			col++
		}
		
		// If we found a non-space character, we're done
		if col < len(runes) {
			return BufferPos{Line: pos.Line, Col: col}
		}
		
		// If we're at end of line, move to next line
		if pos.Line < d.LineCount()-1 {
			pos.Line++
			col = 0
			line = d.lines[pos.Line]
			runes = []rune(line.text)
			
			// If next line is empty or all whitespace, continue loop
			// If next line has content, we'll find it in the next iteration
		} else {
			// We're at the last line and reached the end
			return BufferPos{Line: pos.Line, Col: len(runes)}
		}
	}
}

// GetSelectionText returns the text content of a selection.
func (d *Document) GetSelectionText(selection *Selection) string {
	if selection == nil {
		return ""
	}
	
	start := selection.Start
	end := selection.End
	
	// Normalize selection direction
	if start.Line > end.Line || (start.Line == end.Line && start.Col > end.Col) {
		start, end = end, start
	}
	
	start = d.ValidatePosition(start)
	end = d.ValidatePosition(end)
	
	// Single line selection
	if start.Line == end.Line {
		line := d.GetLine(start.Line)
		runes := []rune(line)
		
		if start.Col >= len(runes) || end.Col > len(runes) || start.Col >= end.Col {
			return ""
		}
		
		return string(runes[start.Col:end.Col])
	}
	
	// Multi-line selection
	var result []string
	
	// First line
	firstLine := d.GetLine(start.Line)
	firstRunes := []rune(firstLine)
	if start.Col < len(firstRunes) {
		result = append(result, string(firstRunes[start.Col:]))
	}
	
	// Middle lines
	for i := start.Line + 1; i < end.Line; i++ {
		result = append(result, d.GetLine(i))
	}
	
	// Last line
	lastLine := d.GetLine(end.Line)
	lastRunes := []rune(lastLine)
	if end.Col <= len(lastRunes) {
		result = append(result, string(lastRunes[:end.Col]))
	}
	
	return strings.Join(result, "\n")
}