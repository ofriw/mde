package terminal

// KeyInput represents keyboard input for terminal operations.
// This interface is specifically designed for Bubble Tea v2 KeyPressMsg.
type KeyInput interface {
	String() string
}

// IsWordMovement detects if a key message represents word-level cursor movement.
// Returns (left=true, right=false) for word left, (left=false, right=true) for word right,
// or (left=false, right=false) for non-word-movement keys.
// 
// Accepts KeyPressMsg or any type implementing KeyInput interface.
func IsWordMovement(msg KeyInput) (left, right bool) {
	// v2 uses string-based detection which handles Alt+Arrow correctly
	switch msg.String() {
	case "alt+left":
		return true, false
	case "alt+right":
		return false, true
	}
	
	return false, false
}