package ast

import "time"

// ChangeType represents the type of change made to the document
type ChangeType int

const (
	ChangeInsert ChangeType = iota
	ChangeDelete
	ChangeReplace
)

// Change represents a single change to the document
type Change struct {
	Type      ChangeType
	BufferPos  BufferPos
	OldText   string
	NewText   string
	Timestamp time.Time
}

// HistoryEntry represents a group of changes that can be undone/redone together
type HistoryEntry struct {
	Changes []Change
	Cursor  BufferPos
}

// History manages the undo/redo stack for a document
type History struct {
	entries     []HistoryEntry
	current     int
	maxEntries  int
	groupTimer  *time.Timer
	grouping    bool
	tempEntry   *HistoryEntry
}

// NewHistory creates a new history manager
func NewHistory(maxEntries int) *History {
	if maxEntries <= 0 {
		maxEntries = 1000
	}
	
	return &History{
		entries:    make([]HistoryEntry, 0, maxEntries),
		current:    -1,
		maxEntries: maxEntries,
	}
}

// AddChange adds a change to the history
func (h *History) AddChange(change Change, cursor BufferPos) {
	// Start a new group if not already grouping
	if !h.grouping {
		h.startGroup(cursor)
	}
	
	// Add change to current group
	h.tempEntry.Changes = append(h.tempEntry.Changes, change)
	h.tempEntry.Cursor = cursor
	
	// Reset group timer
	h.resetGroupTimer()
}

// startGroup starts a new change group
func (h *History) startGroup(cursor BufferPos) {
	h.grouping = true
	h.tempEntry = &HistoryEntry{
		Changes: make([]Change, 0),
		Cursor:  cursor,
	}
}

// endGroup ends the current change group and adds it to history
func (h *History) endGroup() {
	if !h.grouping || h.tempEntry == nil || len(h.tempEntry.Changes) == 0 {
		return
	}
	
	// Remove any entries after current position (for when we're in middle of history)
	h.entries = h.entries[:h.current+1]
	
	// Add new entry
	h.entries = append(h.entries, *h.tempEntry)
	h.current++
	
	// Limit history size
	if len(h.entries) > h.maxEntries {
		h.entries = h.entries[1:]
		h.current--
	}
	
	// Reset grouping state
	h.grouping = false
	h.tempEntry = nil
	
	if h.groupTimer != nil {
		h.groupTimer.Stop()
		h.groupTimer = nil
	}
}

// resetGroupTimer resets the timer for automatic group ending
func (h *History) resetGroupTimer() {
	if h.groupTimer != nil {
		h.groupTimer.Stop()
	}
	
	h.groupTimer = time.AfterFunc(500*time.Millisecond, func() {
		h.endGroup()
	})
}

// ForceEndGroup forces the current group to end
func (h *History) ForceEndGroup() {
	h.endGroup()
}

// CanUndo returns true if there are changes to undo
func (h *History) CanUndo() bool {
	return h.current >= 0
}

// CanRedo returns true if there are changes to redo
func (h *History) CanRedo() bool {
	return h.current < len(h.entries)-1
}

// Undo returns the changes needed to undo the last operation
func (h *History) Undo() (*HistoryEntry, bool) {
	if !h.CanUndo() {
		return nil, false
	}
	
	// End current group if active
	h.ForceEndGroup()
	
	entry := h.entries[h.current]
	h.current--
	
	return &entry, true
}

// Redo returns the changes needed to redo the next operation
func (h *History) Redo() (*HistoryEntry, bool) {
	if !h.CanRedo() {
		return nil, false
	}
	
	h.current++
	entry := h.entries[h.current]
	
	return &entry, true
}

// Clear clears all history
func (h *History) Clear() {
	h.entries = h.entries[:0]
	h.current = -1
	h.grouping = false
	h.tempEntry = nil
	
	if h.groupTimer != nil {
		h.groupTimer.Stop()
		h.groupTimer = nil
	}
}

// ApplyChange applies a change to the document
func ApplyChange(doc *Document, change Change) {
	switch change.Type {
	case ChangeInsert:
		// Insert text at position
		pos := change.BufferPos
		for _, ch := range change.NewText {
			if ch == '\n' {
				pos = doc.InsertNewline(pos)
			} else {
				pos = doc.InsertChar(pos, ch)
			}
		}
		
	case ChangeDelete:
		// Delete text from position
		pos := change.BufferPos
		for range change.OldText {
			if pos.Col > 0 {
				pos = doc.DeleteChar(pos)
			} else if pos.Line > 0 {
				pos = doc.DeleteLine(pos)
			}
		}
		
	case ChangeReplace:
		// Delete old text and insert new text
		pos := change.BufferPos
		
		// Delete old text
		for range change.OldText {
			if pos.Col > 0 {
				pos = doc.DeleteChar(pos)
			} else if pos.Line > 0 {
				pos = doc.DeleteLine(pos)
			}
		}
		
		// Insert new text
		for _, ch := range change.NewText {
			if ch == '\n' {
				pos = doc.InsertNewline(pos)
			} else {
				pos = doc.InsertChar(pos, ch)
			}
		}
	}
}

// ReverseChange creates the reverse of a change for undo
func ReverseChange(change Change) Change {
	reversed := Change{
		BufferPos:  change.BufferPos,
		Timestamp: time.Now(),
	}
	
	switch change.Type {
	case ChangeInsert:
		reversed.Type = ChangeDelete
		reversed.OldText = change.NewText
		reversed.NewText = ""
		
	case ChangeDelete:
		reversed.Type = ChangeInsert
		reversed.OldText = ""
		reversed.NewText = change.OldText
		
	case ChangeReplace:
		reversed.Type = ChangeReplace
		reversed.OldText = change.NewText
		reversed.NewText = change.OldText
	}
	
	return reversed
}