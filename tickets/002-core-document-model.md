# 2025-07-14 - [FEATURE] Core Document Model
**Priority**: High

Implement AST-based document model with cursor management and editing operations.

## Deliverables
- Document model with efficient text operations
- Undo/redo functionality
- Text selection and clipboard support

## Tasks
- Design AST-based document structure
- Implement cursor position tracking
- Add text insertion/deletion with proper cursor updates
- Create undo/redo stack with change tracking
- Implement text selection (shift+arrows, mouse)
- Add clipboard operations (Ctrl+C/X/V)
- Implement line numbers display toggle
- Add efficient text rendering from AST

## Success Criteria
- Smooth text editing with proper cursor behavior
- Working undo/redo (Ctrl+Z/Y)
- Can select, copy, cut, and paste text
- Line numbers can be toggled on/off