# 2025-07-16T12:12:49+03:00 - [FEATURE] Coordinate System Refactor for LLM-Friendly Code

**Priority**: High

## Problem

The current coordinate system is ambiguous and error-prone, leading to bugs like the double line number adjustment that broke horizontal cursor movement. The code uses raw `(int, int)` coordinates without clear documentation of which coordinate system they represent.

### Current Issues:
- **Ambiguous coordinate systems**: Document, content, and screen coordinates are mixed without clear boundaries
- **Double transformations**: `GetCursorScreenPosition()` adds line number offset, then renderer subtracts it again
- **No validation**: Coordinate assumptions are implicit and can fail silently
- **Poor LLM readability**: Future AI maintenance is hindered by unclear coordinate contracts

### Root Cause Analysis:
The coordinate transformation bug happened because:
1. `Editor.GetCursorScreenPosition()` adds 6 to column when line numbers enabled
2. `TerminalRenderer.renderLineWithStylesAndCursor()` subtracts 6 when line numbers enabled
3. **Double adjustment** causes cursor to appear at wrong position
4. **No explicit documentation** of which coordinate system each function expects

## Solution

Implement explicit coordinate type system with clear transformation boundaries:

1. **Explicit Types**: `DocumentPos`, `ContentPos`, `ScreenPos`
2. **Single Responsibility**: Each component owns one transformation
3. **Validation**: Explicit bounds checking with clear error messages
4. **Documentation**: Step-by-step transformation documentation
5. **LLM-Friendly**: Constitutional rules and examples for AI maintenance

## Implementation Plan

### Phase 1: Core Types
- [x] Define coordinate types (`DocumentPos`, `ContentPos`, `ScreenPos`)
- [x] Create transformation interfaces
- [x] Add validation functions

### Phase 2: Editor Refactor
- [x] Refactor `GetCursorScreenPosition()` → `GetCursorContentPosition()`
- [x] Implement `DocumentPos` → `ContentPos` transformation
- [x] Add coordinate validation

### Phase 3: Renderer Refactor
- [x] Update `RenderToStringWithCursor()` to use `ContentPos`
- [x] Remove double line number adjustment
- [x] Add coordinate validation

### Phase 4: Documentation
- [x] Create `COORDINATE_SYSTEM.md` with LLM context
- [x] Add constitutional rules in `COORDINATE_RULES.md`
- [x] Provide examples in `EXAMPLES.md`

### Phase 5: Testing
- [x] Update tests to use explicit coordinate types
- [x] Add coordinate transformation tests
- [x] Verify cursor movement works correctly

## Acceptance Criteria

- [x] All coordinate transformations use explicit types
- [x] Single source of truth for each coordinate transformation
- [x] Cursor movement works correctly with and without line numbers
- [x] All coordinate functions have validation
- [x] Documentation includes LLM-friendly examples and rules

## Risk Assessment

**Low Risk**: This is a refactoring that makes existing behavior explicit without changing functionality.

**Mitigation**: Comprehensive testing of cursor movement and coordinate edge cases.

# History

## 2025-07-16T12:12:49+03:00
Created ticket after discovering coordinate system ambiguity caused horizontal cursor movement bug during test fixes.

## 2025-07-16T12:23:27+03:00
**COMPLETED**: Successfully implemented explicit coordinate system with LLM-friendly documentation.

### Implementation Summary:
- **Core Types**: Created `DocumentPos`, `ContentPos`, `ScreenPos` types with validation
- **Transformation Chain**: `DocumentPos` → `ContentPos` (Editor) → `ScreenPos` (TUI)
- **Single Responsibility**: Each coordinate transformation happens exactly once
- **Validation**: All coordinate functions validate inputs with clear error messages
- **Documentation**: Comprehensive LLM-friendly guides with constitutional rules and examples
- **Testing**: Full test coverage for coordinate transformations and validation
- **Backward Compatibility**: `GetCursorScreenPosition()` deprecated but still works

### Files Created:
- `pkg/ast/coordinates.go` - Core coordinate types and interfaces
- `COORDINATE_SYSTEM.md` - LLM context and usage guide
- `COORDINATE_RULES.md` - Constitutional rules for LLM agents
- `COORDINATE_EXAMPLES.md` - Concrete examples for few-shot learning
- `test/unit/coordinates_test.go` - Comprehensive coordinate system tests

### Files Modified:
- `pkg/ast/editor.go` - Added coordinate transformation and validation methods
- `internal/plugins/renderers/terminal.go` - Updated to use ContentPos
- `internal/tui/model.go` - Updated to use new coordinate methods

### Results:
- ✅ All tests passing (42s test suite)
- ✅ Horizontal cursor movement fixed
- ✅ No double line number adjustment
- ✅ Clear error messages for coordinate violations
- ✅ LLM-friendly codebase with explicit contracts

The coordinate system is now self-documenting and prevents the class of bugs that caused the original horizontal cursor movement issue.