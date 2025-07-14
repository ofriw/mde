# MDE - Markdown Editor

Terminal markdown editor with plugin architecture for parsing, rendering, and theming.

## Architecture
- **Bubble Tea** TUI with Elm Architecture - single flat Model
- **Internal plugins** - compiled into binary (no dynamic loading)
- **AST-based** document model with goldmark parser
- **Non-modal** editing with micro-style keybindings

## Key Commands
```bash
make build      # Build binary
make test       # Run tests  
make lint       # Run linters
make install    # Install locally
```

## Testing
- Use `testify` for assertions and mocks
- Target 80% coverage
- VHS for terminal recordings

## Performance Targets
- Startup < 100ms
- Render < 50ms for 1000 lines
- Memory < 50MB typical usage

## Current Focus
Check `tickets/` for active work. One task in progress at a time.