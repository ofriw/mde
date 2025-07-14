# 2025-07-14 - [FEATURE] Markdown Parsing
**Priority**: High

Integrate goldmark parser with syntax highlighting and preview mode.

## Deliverables
- CommonMark compliant parsing
- Syntax highlighted markdown in editor
- Toggle-able preview mode

## Tasks
- Implement CommonMarkParser plugin using goldmark
- Add markdown syntax highlighting to editor view
- Synchronize AST with text changes
- Implement preview mode (Ctrl+P toggle)
- Add proper markdown element styling
- Handle goldmark extensions configuration
- Optimize parsing for large documents
- Add markdown-specific cursor behaviors

## Success Criteria
- Markdown files display with syntax highlighting
- Preview mode shows rendered markdown
- Parsing is fast and accurate
- CommonMark spec compliance