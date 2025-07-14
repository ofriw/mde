# TUI Components

Bubble Tea UI components following Elm Architecture.

## Structure
- `app.go` - Main application model and message loop
- `viewport.go` - Scrollable content viewer
- `editor.go` - Text editing component

## Patterns
- Single flat Model approach
- Components return `tea.Cmd` for async operations
- Use Lip Gloss for styling
- Handle window resize messages

## Key Points
- Don't mutate model outside Update()
- Batch commands with `tea.Batch()`
- Use `tea.Every()` for periodic updates