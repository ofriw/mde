# Bubble Tea v2 Migration

MDE uses Bubble Tea v2.0.0-beta.4 for the TUI.

## Key Changes

### Message Types
- `tea.KeyMsg` → `tea.KeyPressMsg`
- `tea.MouseMsg` → Split into specific types:
  - `tea.MouseClickMsg`
  - `tea.MouseReleaseMsg`
  - `tea.MouseMotionMsg`
  - `tea.MouseWheelMsg`

### Key Detection
- String-based: `msg.String() == "ctrl+s"`
- Alt+Arrow: Handled via terminal package helper

### Dependencies
- Direct: `github.com/charmbracelet/bubbletea/v2 v2.0.0-beta.4`
- Imports: Use `/v2` suffix

### Word Navigation
- Unicode-based word boundaries
- Crosses line boundaries when needed
- Skips whitespace intelligently