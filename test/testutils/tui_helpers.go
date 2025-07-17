package testutils

import (
	"io"
	"github.com/charmbracelet/bubbletea"
	"github.com/ofri/mde/internal/tui"
	"github.com/ofri/mde/pkg/ast"
)

// LoadContentIntoModel loads content into a TUI model for testing
func LoadContentIntoModel(model *tui.Model, content string) {
	// Create a new editor with content
	_ = ast.NewEditorWithContent(content)
	
	// Replace the model's editor (this assumes the model has a way to do this)
	// For now, we'll insert the content character by character
	
	// This is a workaround - in a real implementation, you'd want a SetDocument method
	modelEditor := model.GetEditor()
	
	// Get the document and insert content
	doc := modelEditor.GetDocument()
	
	// Clear existing content and insert new content
	// This is a simplified approach - in practice, you'd want a proper SetContent method
	cursor := modelEditor.GetCursor()
	cursor.SetBufferPos(ast.BufferPos{Line: 0, Col: 0})
	
	// Insert the content character by character
	for _, char := range content {
		if char == '\n' {
			doc.InsertNewline(cursor.GetBufferPos())
			modelEditor.MoveCursorDown()
			modelEditor.MoveCursorToLineStart()
		} else {
			newPos := doc.InsertChar(cursor.GetBufferPos(), char)
			cursor.SetBufferPos(newPos)
		}
	}
	
	// Move cursor back to start
	modelEditor.MoveCursorToDocumentStart()
}

// SetModelSize sets the size of the TUI model
func SetModelSize(model *tui.Model, width, height int) {
	// Simulate WindowSizeMsg to set model dimensions
	msg := tea.WindowSizeMsg{
		Width:  width,
		Height: height,
	}
	
	// Update the model with the window size message
	model.Update(msg)
}

// Helper function to strip ANSI escape sequences for testing
func StripAnsiEscapes(s string) string {
	// Simple ANSI stripping - removes escape sequences like \033[...m
	result := ""
	inEscape := false
	
	for i, r := range s {
		if r == '\033' && i+1 < len(s) && s[i+1] == '[' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result += string(r)
	}
	
	return result
}

// Helper function to read output from a teatest output reader
func ReadOutput(output io.Reader) string {
	if output == nil {
		return ""
	}
	
	data, err := io.ReadAll(output)
	if err != nil {
		return ""
	}
	
	return string(data)
}