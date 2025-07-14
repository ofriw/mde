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

// TokenKind represents different types of tokens
type TokenKind int

const (
	TokenText TokenKind = iota
	TokenKeyword
	TokenString
	TokenComment
	TokenNumber
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

// Position represents a cursor position in the document
type Position struct {
	Line int
	Col  int
}

// Selection represents a text selection range
type Selection struct {
	Start Position
	End   Position
}

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
func (d *Document) InsertChar(pos Position, ch rune) Position {
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
	
	return Position{Line: pos.Line, Col: pos.Col + 1}
}

// DeleteChar deletes a character at the specified position
func (d *Document) DeleteChar(pos Position) Position {
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
	
	return Position{Line: pos.Line, Col: pos.Col - 1}
}

// InsertNewline inserts a newline at the specified position
func (d *Document) InsertNewline(pos Position) Position {
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
	
	return Position{Line: pos.Line + 1, Col: 0}
}

// DeleteLine deletes a line and merges with previous if needed
func (d *Document) DeleteLine(pos Position) Position {
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
	
	return Position{Line: pos.Line - 1, Col: newCol}
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
func (d *Document) ValidatePosition(pos Position) Position {
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

// GetCharAt returns the character at the specified position
func (d *Document) GetCharAt(pos Position) rune {
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

// FindWordStart finds the start of the word at the given position
func (d *Document) FindWordStart(pos Position) Position {
	if pos.Line < 0 || pos.Line >= len(d.lines) {
		return pos
	}
	
	line := d.lines[pos.Line]
	runes := []rune(line.text)
	
	if pos.Col <= 0 {
		return Position{Line: pos.Line, Col: 0}
	}
	
	col := pos.Col
	if col > len(runes) {
		col = len(runes)
	}
	
	// Move back while we're still in the same word
	for col > 0 && !unicode.IsSpace(runes[col-1]) {
		col--
	}
	
	return Position{Line: pos.Line, Col: col}
}

// FindWordEnd finds the end of the word at the given position
func (d *Document) FindWordEnd(pos Position) Position {
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
	
	return Position{Line: pos.Line, Col: col}
}