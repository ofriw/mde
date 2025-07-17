# MDE Test Suite

Critical regression tests for cursor positioning and TUI behavior validation.

## ⚠️ CRITICAL BUG DETECTION

### Ghost Line Cursor Bug
**Problem**: Cursor appears at end of line instead of ON first character  
**Symptoms**: `"Hello World█"` pattern in output  
**Tests**: All cursor rendering tests catch this bug

### Key Invariants Tested
1. **Position (0,0)**: Cursor initializes at first character after file load
2. **Content Length**: Cursor replaces character, never appends
3. **Screen Position**: Correct coordinate transformation with/without line numbers
4. **Visual Presence**: Cursor visible in actual TUI output

## AI Agent Guardrails

### ⚠️ CAUTION: Requires User Approval
- Cursor initialization logic (`pkg/ast/cursor.go`)
- Terminal renderer (`internal/plugins/renderers/terminal.go`)
- TUI rendering pipeline (`internal/tui/model.go`)

### ✅ VERIFY: Test Validation Required
- Cursor position (0,0) after file load
- Content length preservation in rendering
- Line number offset calculation (viewport configuration)
- Cursor visibility in TUI output

### 🔍 VALIDATE: User Confirmation Needed For
- Changes to cursor positioning behavior
- Modifications to content length during rendering
- Updates to cursor position calculations
- Removal of cursor from TUI output

## Test Architecture

Comprehensive testing framework for the MDE markdown editor, with focus on cursor management and TUI testing.

## Structure

```
test/
├── unit/                           # Unit tests for core functionality
│   ├── cursor_test.go              # Core cursor logic tests (16 tests)
│   ├── coordinate_system_test.go   # BufferPos and viewport transformation tests
│   └── cursor_rendering_test.go    # Cursor rendering and visual positioning tests
├── integration/                    # Integration tests
│   ├── performance_test.go         # Performance benchmarks and tests (7 tests + 3 benchmarks)
│   ├── tui_cursor_test.go          # TUI cursor integration tests (8 tests)
│   ├── cursor_fixes_test.go        # Cursor positioning regression tests
│   ├── plugin_test.go              # Plugin integration tests (existing)
│   ├── builtin_plugins_test.go     # Built-in plugin tests (existing)
│   └── real_world_test.go          # Real-world scenario tests (existing)
├── testutils/                      # Test utilities and helpers
│   ├── cursor_helpers.go           # Cursor testing utilities
│   ├── input_simulation.go         # Input simulation framework
│   ├── visual_assertions.go        # Visual artifact detection
│   └── tui_helpers.go              # TUI model helpers
└── demo/                           # Demo and example code
    └── plugin_demo.go              # Plugin demonstration
```

## Key Features

### 1. Comprehensive Unit Testing
- **cursor_test.go**: Tests all cursor movement operations, selection handling, and edge cases
- **coordinate_system_test.go**: Tests BufferPos validation and viewport coordinate transformation
- **cursor_rendering_test.go**: Tests cursor rendering and visual positioning

### 2. TUI Integration Framework
- **teatest Integration**: Uses `github.com/charmbracelet/x/exp/teatest` for deterministic TUI testing
- **Input Simulation**: Comprehensive keyboard and mouse input simulation
- **Visual Assertions**: Detection of visual artifacts and cursor rendering issues

### 3. Performance Testing
- **Benchmarks**: Cursor movement benchmarks (1.757 ns/op, exceeds 60fps target)
- **Large Document Testing**: Tests with 10k+ lines
- **Memory Usage**: Concurrent access and memory leak detection

### 4. Coordinate System Testing
- **BufferPos Validation**: Tests buffer position validation and bounds checking
- **Viewport Transformation**: Tests BufferPos → ScreenPos coordinate transformation
- **Edge Cases**: Boundary conditions, Unicode content, line number handling
- **Visibility Checks**: Ensures positions are correctly marked as visible/invisible

## Running Tests

```bash
# Run all unit tests
go test ./test/unit/... -v

# Run specific test categories
go test ./test/unit/cursor_test.go -v
go test ./test/unit/coordinate_system_test.go -v
go test ./test/integration/performance_test.go -v

# Run benchmarks
go test ./test/integration/performance_test.go -bench=. -v

# Run with coverage
go test ./test/unit/... -cover

# Run specific test
go test ./test/unit/cursor_test.go -run TestCursor_MoveLeft -v
```

## Test Categories

### Unit Tests (33 tests)
- **Basic Movement**: Left, right, up, down, word movement
- **Document Navigation**: Start/end of line/document
- **Selection Handling**: Text selection, multi-line selection
- **Edge Cases**: Empty documents, single character, Unicode content
- **Position Validation**: Boundary checking, coordinate transformation
- **Property-based**: Invariant testing with random inputs

### Integration Tests (22 tests)
- **Cursor Positioning**: CursorManager integration and coordinate transformation
- **Mouse Handling**: Click coordinate transformation using new coordinate system
- **Performance**: Movement speed, memory usage, concurrent access
- **TUI Integration**: Full TUI testing with teatest framework
- **Visual Artifacts**: Detection of cursor rendering issues

### Performance Benchmarks (3 benchmarks)
- **Cursor Movement**: 1.757 ns/op (exceeds 60fps requirement)
- **Screen Position Calculation**: Sub-millisecond performance
- **Position Updates**: Efficient cursor positioning

## Test Utilities

### CursorTestHelper
Provides utilities for cursor testing in TUI context:
- Position verification and assertions
- Screen coordinate validation
- Input simulation and state management

### InputSimulator
Comprehensive input simulation framework:
- Keyboard input simulation (arrow keys, shortcuts)
- Mouse input simulation (clicks, drags, wheel)
- Test scenario execution

### VisualAssertion
Visual testing and artifact detection:
- Cursor visibility verification
- Visual artifact detection (duplicate cursors, line number fragments)
- Golden file testing support

## Issues Addressed

This test suite specifically addresses the cursor management issues identified in:
- `internal/tui/model.go:197-212` - Cursor rendering artifacts in fallback mode
- `internal/tui/update.go:344-400` - Mouse click coordinate transformation bugs
- `pkg/ast/cursor.go` - BufferPos validation and coordinate transformation
- `pkg/ast/viewport.go` - Viewport coordinate transformation edge cases

## Future Enhancements

1. **Golden File Testing**: Complete implementation of visual regression testing
2. **Real TUI Integration**: Full teatest integration with actual TUI rendering
3. **Accessibility Testing**: Cursor visibility and navigation for accessibility
4. **Multi-platform Testing**: Platform-specific cursor behavior testing
5. **Stress Testing**: Extended performance testing under load

## Contributing

When adding new cursor-related features:
1. Add unit tests to `test/unit/cursor_test.go`
2. Add coordinate transformation tests to `test/unit/coordinate_system_test.go`
3. Add integration tests to `test/integration/cursor_fixes_test.go` or `test/integration/tui_cursor_test.go`
4. Add performance benchmarks if needed
5. Update CursorManager tests for new functionality

All tests should pass individually and the test suite should maintain high coverage of cursor-related functionality.