package unit

import (
	"testing"

	"github.com/ofri/mde/pkg/terminal"
	"github.com/stretchr/testify/assert"
)

func TestTerminalWordMovement(t *testing.T) {
	t.Run("Alt+Arrow word movement", func(t *testing.T) {
		// Test Alt+Left - create a mock KeyPressMsg
		msg := &mockKeyPressMsg{str: "alt+left"}
		left, right := terminal.IsWordMovement(msg)
		assert.True(t, left)
		assert.False(t, right)
		
		// Test Alt+Right
		msg = &mockKeyPressMsg{str: "alt+right"}
		left, right = terminal.IsWordMovement(msg)
		assert.False(t, left)
		assert.True(t, right)
	})
	
	t.Run("non-word movement keys", func(t *testing.T) {
		// Test regular arrow keys without Alt
		msg := &mockKeyPressMsg{str: "left"}
		left, right := terminal.IsWordMovement(msg)
		assert.False(t, left)
		assert.False(t, right)
		
		// Test regular characters
		msg = &mockKeyPressMsg{str: "a"}
		left, right = terminal.IsWordMovement(msg)
		assert.False(t, left)
		assert.False(t, right)
		
		// Test other Alt combinations
		msg = &mockKeyPressMsg{str: "alt+x"}
		left, right = terminal.IsWordMovement(msg)
		assert.False(t, left)
		assert.False(t, right)
	})
}

// Mock KeyPressMsg for testing
type mockKeyPressMsg struct {
	str string
}

func (m *mockKeyPressMsg) String() string {
	return m.str
}