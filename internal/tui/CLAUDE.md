# TUI Components

Bubble Tea v2 UI components following Elm Architecture.

## Structure
- `model.go` - Main application state
- `update.go` - Message handling and state updates
- `view.go` - Rendering logic
- `file.go` - File operations

## Message Types (v2)
- `tea.KeyPressMsg` - Keyboard input
- `tea.MouseClickMsg` - Mouse clicks
- `tea.MouseWheelMsg` - Scroll events
- `tea.WindowSizeMsg` - Terminal resize

## Key Points
- Single Model with all state
- Update() handles all messages
- View() renders current state
- Commands for async operations