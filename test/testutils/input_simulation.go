package testutils

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/ofri/mde/pkg/ast"
)

// InputSimulator provides utilities for simulating user input in TUI tests
type InputSimulator struct {
	tm *teatest.TestModel
}

// NewInputSimulator creates a new input simulator
func NewInputSimulator(tm *teatest.TestModel) *InputSimulator {
	return &InputSimulator{tm: tm}
}

// KeySequence represents a sequence of key presses
type KeySequence struct {
	Keys  []string
	Delay time.Duration
}

// NewKeySequence creates a new key sequence
func NewKeySequence(keys ...string) *KeySequence {
	return &KeySequence{
		Keys:  keys,
		Delay: 0,
	}
}

// WithDelay sets a delay between key presses
func (ks *KeySequence) WithDelay(delay time.Duration) *KeySequence {
	ks.Delay = delay
	return ks
}

// ExecuteKeySequence executes a sequence of key presses
func (is *InputSimulator) ExecuteKeySequence(sequence *KeySequence) {
	for _, key := range sequence.Keys {
		is.SendKey(key)
		if sequence.Delay > 0 {
			time.Sleep(sequence.Delay)
		}
	}
}

// SendKey sends a single key press
func (is *InputSimulator) SendKey(key string) {
	msg := is.keyToMessage(key)
	is.tm.Send(msg)
}

// SendKeys sends multiple key presses
func (is *InputSimulator) SendKeys(keys ...string) {
	for _, key := range keys {
		is.SendKey(key)
	}
}

// SendMouseClick sends a mouse click at the specified coordinates
func (is *InputSimulator) SendMouseClick(x, y int) {
	msg := tea.MouseMsg{
		Type: tea.MouseLeft,
		X:    x,
		Y:    y,
	}
	is.tm.Send(msg)
}

// SendMouseDrag sends a mouse drag event
func (is *InputSimulator) SendMouseDrag(startX, startY, endX, endY int) {
	// Start drag
	msg := tea.MouseMsg{
		Type: tea.MouseLeft,
		X:    startX,
		Y:    startY,
	}
	is.tm.Send(msg)
	
	// Drag to end position
	msg = tea.MouseMsg{
		Type: tea.MouseLeft,
		X:    endX,
		Y:    endY,
	}
	is.tm.Send(msg)
}

// SendMouseWheel sends a mouse wheel event
func (is *InputSimulator) SendMouseWheel(x, y int, up bool) {
	var wheelType tea.MouseEventType
	if up {
		wheelType = tea.MouseWheelUp
	} else {
		wheelType = tea.MouseWheelDown
	}
	
	msg := tea.MouseMsg{
		Type: wheelType,
		X:    x,
		Y:    y,
	}
	is.tm.Send(msg)
}

// TypeText simulates typing text character by character
func (is *InputSimulator) TypeText(text string) {
	for _, char := range text {
		is.SendKey(string(char))
	}
}

// SimulateResize simulates a terminal resize event
func (is *InputSimulator) SimulateResize(width, height int) {
	msg := tea.WindowSizeMsg{
		Width:  width,
		Height: height,
	}
	is.tm.Send(msg)
}

// keyToMessage converts a key string to a tea.KeyMsg
func (is *InputSimulator) keyToMessage(key string) tea.KeyMsg {
	switch key {
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "home":
		return tea.KeyMsg{Type: tea.KeyHome}
	case "end":
		return tea.KeyMsg{Type: tea.KeyEnd}
	case "pageup":
		return tea.KeyMsg{Type: tea.KeyPgUp}
	case "pagedown":
		return tea.KeyMsg{Type: tea.KeyPgDown}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case "delete":
		return tea.KeyMsg{Type: tea.KeyDelete}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "escape":
		return tea.KeyMsg{Type: tea.KeyEscape}
	case "space":
		return tea.KeyMsg{Type: tea.KeySpace}
	case "ctrl+a":
		return tea.KeyMsg{Type: tea.KeyCtrlA}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "ctrl+v":
		return tea.KeyMsg{Type: tea.KeyCtrlV}
	case "ctrl+x":
		return tea.KeyMsg{Type: tea.KeyCtrlX}
	case "ctrl+z":
		return tea.KeyMsg{Type: tea.KeyCtrlZ}
	case "ctrl+y":
		return tea.KeyMsg{Type: tea.KeyCtrlY}
	case "ctrl+f":
		return tea.KeyMsg{Type: tea.KeyCtrlF}
	case "ctrl+g":
		return tea.KeyMsg{Type: tea.KeyCtrlG}
	case "ctrl+s":
		return tea.KeyMsg{Type: tea.KeyCtrlS}
	case "ctrl+o":
		return tea.KeyMsg{Type: tea.KeyCtrlO}
	case "ctrl+n":
		return tea.KeyMsg{Type: tea.KeyCtrlN}
	case "ctrl+home":
		return tea.KeyMsg{Type: tea.KeyCtrlHome}
	case "ctrl+end":
		return tea.KeyMsg{Type: tea.KeyCtrlEnd}
	case "ctrl+left":
		return tea.KeyMsg{Type: tea.KeyCtrlLeft}
	case "ctrl+right":
		return tea.KeyMsg{Type: tea.KeyCtrlRight}
	case "shift+left":
		return tea.KeyMsg{Type: tea.KeyShiftLeft}
	case "shift+right":
		return tea.KeyMsg{Type: tea.KeyShiftRight}
	case "shift+up":
		return tea.KeyMsg{Type: tea.KeyShiftUp}
	case "shift+down":
		return tea.KeyMsg{Type: tea.KeyShiftDown}
	case "shift+home":
		return tea.KeyMsg{Type: tea.KeyShiftHome}
	case "shift+end":
		return tea.KeyMsg{Type: tea.KeyShiftEnd}
	default:
		if len(key) == 1 {
			return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		}
		panic(fmt.Sprintf("Unknown key: %s", key))
	}
}

// Common key sequences for testing
var (
	// Navigation sequences
	MoveToStartOfLine = NewKeySequence("home")
	MoveToEndOfLine   = NewKeySequence("end")
	MoveToStartOfDoc  = NewKeySequence("ctrl+home")
	MoveToEndOfDoc    = NewKeySequence("ctrl+end")
	
	// Selection sequences
	SelectAll           = NewKeySequence("ctrl+a")
	SelectToStartOfLine = NewKeySequence("shift+home")
	SelectToEndOfLine   = NewKeySequence("shift+end")
	SelectWordLeft      = NewKeySequence("ctrl+shift+left")
	SelectWordRight     = NewKeySequence("ctrl+shift+right")
	
	// Editing sequences
	DeleteLine     = NewKeySequence("ctrl+l")
	DuplicateLine  = NewKeySequence("ctrl+d")
	CopyLine       = NewKeySequence("ctrl+c")
	PasteLine      = NewKeySequence("ctrl+v")
	Undo           = NewKeySequence("ctrl+z")
	Redo           = NewKeySequence("ctrl+y")
	
	// Common text patterns
	LoremIpsum = "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
	Markdown   = "# Header\n\nThis is **bold** and *italic* text.\n\n- List item 1\n- List item 2"
	CodeBlock  = "```go\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}\n```"
)

// TestScenario represents a complete test scenario
type TestScenario struct {
	Name        string
	InitialText string
	Steps       []TestStep
	Expected    ExpectedState
}

// TestStep represents a single step in a test scenario
type TestStep struct {
	Type        string // "key", "mouse", "resize", "wait"
	Description string
	Data        interface{}
}

// ExpectedState represents the expected state after a test scenario
type ExpectedState struct {
	CursorPosition ast.BufferPos
	ScreenPosition struct{ Row, Col int }
	Selection      *ast.Selection
	Content        string
	VisibleLines   []string
}

// NewTestScenario creates a new test scenario
func NewTestScenario(name string) *TestScenario {
	return &TestScenario{
		Name:  name,
		Steps: make([]TestStep, 0),
	}
}

// WithInitialText sets the initial text content
func (ts *TestScenario) WithInitialText(text string) *TestScenario {
	ts.InitialText = text
	return ts
}

// AddKeyStep adds a key press step
func (ts *TestScenario) AddKeyStep(description string, keys ...string) *TestScenario {
	ts.Steps = append(ts.Steps, TestStep{
		Type:        "key",
		Description: description,
		Data:        keys,
	})
	return ts
}

// AddMouseStep adds a mouse interaction step
func (ts *TestScenario) AddMouseStep(description string, x, y int, action string) *TestScenario {
	ts.Steps = append(ts.Steps, TestStep{
		Type:        "mouse",
		Description: description,
		Data: map[string]interface{}{
			"x":      x,
			"y":      y,
			"action": action,
		},
	})
	return ts
}

// AddResizeStep adds a resize step
func (ts *TestScenario) AddResizeStep(description string, width, height int) *TestScenario {
	ts.Steps = append(ts.Steps, TestStep{
		Type:        "resize",
		Description: description,
		Data: map[string]interface{}{
			"width":  width,
			"height": height,
		},
	})
	return ts
}

// WithExpectedState sets the expected final state
func (ts *TestScenario) WithExpectedState(expected ExpectedState) *TestScenario {
	ts.Expected = expected
	return ts
}

// Execute executes the test scenario
func (ts *TestScenario) Execute(is *InputSimulator) {
	for _, step := range ts.Steps {
		switch step.Type {
		case "key":
			keys := step.Data.([]string)
			is.SendKeys(keys...)
		case "mouse":
			data := step.Data.(map[string]interface{})
			x := data["x"].(int)
			y := data["y"].(int)
			action := data["action"].(string)
			
			switch action {
			case "click":
				is.SendMouseClick(x, y)
			case "drag":
				// For drag, we need start and end positions
				// This is a simplified implementation
				is.SendMouseDrag(x, y, x+1, y+1)
			}
		case "resize":
			data := step.Data.(map[string]interface{})
			width := data["width"].(int)
			height := data["height"].(int)
			is.SimulateResize(width, height)
		}
	}
}