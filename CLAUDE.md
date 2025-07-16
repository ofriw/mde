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

### Critical Test Guardrails
**CURSOR TESTS**: Regression protection for cursor positioning bugs
- CAUTION: Cursor changes require explicit user approval
- VERIFY: Run `go test ./test/unit -run Cursor` before cursor modifications
- VALIDATE: Run `go test ./test/integration -run Cursor` for TUI validation

**Key Behaviors to Verify**:
- Cursor position (0,0) after file load
- Content length preserved in rendering
- Line number offset = 6 characters
- Cursor visible in TUI output

**User Confirmation Required For**:
- Changes to cursor positioning behavior
- Modifications to content length during rendering
- Updates to cursor position calculations

## Performance Targets
- Startup < 100ms
- Render < 50ms for 1000 lines
- Memory < 50MB typical usage

## Current Focus
Check `tickets/` for active work. One task in progress at a time.