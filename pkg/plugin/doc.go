// Package plugin provides the internal plugin architecture for MDE.
//
// IMPORTANT: This is an INTERNAL modularization technique, NOT a public plugin API.
// All plugins are compiled directly into the binary and registered at startup.
// There is no dynamic loading or external plugin support.
//
// # Architecture
//
// The plugin system serves as an internal code organization pattern that:
//   - Separates concerns between parsing and rendering
//   - Provides clear interfaces for different components
//   - Enables compile-time modularity without runtime complexity
//
// # Plugin Types
//
// Two types of plugins are supported:
//   - Parser plugins: Convert raw text to AST with syntax tokens
//   - Renderer plugins: Convert AST to terminal output with ANSI colors
//
// # Error Handling
//
// Since all plugins are internal and compiled into the binary:
//   - Plugin registration failures indicate programming errors
//   - Plugin execution failures indicate bugs in the implementation
//   - All plugin errors should cause explicit failures with detailed error messages
//   - There are NO fallback mechanisms - the system must work correctly
//
// # Registration
//
// Plugins must be registered at startup in main.go:
//
//	func init() {
//	    registry := plugin.GetRegistry()
//	    registry.RegisterParser("markdown", &markdownParser{})
//	    registry.RegisterRenderer("terminal", &terminalRenderer{})
//	}
//
// # Usage
//
// The TUI layer gets plugins from the registry and uses them directly:
//
//	renderer, err := registry.GetDefaultRenderer()
//	if err != nil {
//	    panic("FATAL: No renderer registered - programming error")
//	}
//
// # Design Rationale
//
// This internal plugin architecture provides:
//   - Clear separation of concerns
//   - Testable components with defined interfaces
//   - Future extensibility without current complexity
//   - Compile-time safety with no runtime plugin loading
package plugin