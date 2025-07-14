# Plugin System

Internal plugins compiled into binary. No dynamic loading in v1.

## Plugin Types
- **Parsers**: Convert markdown to AST (goldmark-based)
- **Renderers**: Convert AST to styled output  
- **Themes**: Provide styling for elements

## Adding Plugins
1. Implement interface in `pkg/plugin/`
2. Create plugin in appropriate subdirectory
3. Register in `init()` function
4. Add tests

## Interfaces
```go
ParserPlugin.Parse(content string) (ast.Node, error)
RendererPlugin.Render(node ast.Node, theme Theme) string
Theme.GetStyle(element string) Style
```