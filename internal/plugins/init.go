package plugins

import (
	"fmt"
	"github.com/ofri/mde/internal/plugins/parsers"
	"github.com/ofri/mde/internal/plugins/renderers"
	"github.com/ofri/mde/internal/plugins/themes"
	"github.com/ofri/mde/pkg/plugin"
)

// InitializePlugins initializes all built-in plugins
func InitializePlugins() error {
	// Initialize themes
	if err := initializeThemes(); err != nil {
		return fmt.Errorf("failed to initialize themes: %w", err)
	}
	
	// Initialize renderers
	if err := initializeRenderers(); err != nil {
		return fmt.Errorf("failed to initialize renderers: %w", err)
	}
	
	// Initialize parsers
	if err := initializeParsers(); err != nil {
		return fmt.Errorf("failed to initialize parsers: %w", err)
	}
	
	// Set default plugins
	if err := setDefaultPlugins(); err != nil {
		return fmt.Errorf("failed to set default plugins: %w", err)
	}
	
	return nil
}

// initializeThemes registers all built-in themes
func initializeThemes() error {
	registry := plugin.GetRegistry()
	
	// Register dark theme
	darkTheme := themes.NewDarkTheme()
	if err := registry.RegisterTheme(darkTheme.Name(), darkTheme); err != nil {
		return fmt.Errorf("failed to register dark theme: %w", err)
	}
	
	// Register light theme
	lightTheme := themes.NewLightTheme()
	if err := registry.RegisterTheme(lightTheme.Name(), lightTheme); err != nil {
		return fmt.Errorf("failed to register light theme: %w", err)
	}
	
	return nil
}

// initializeRenderers registers all built-in renderers
func initializeRenderers() error {
	registry := plugin.GetRegistry()
	
	// Register terminal renderer
	terminalRenderer := renderers.NewTerminalRenderer()
	if err := registry.RegisterRenderer(terminalRenderer.Name(), terminalRenderer); err != nil {
		return fmt.Errorf("failed to register terminal renderer: %w", err)
	}
	
	// Configure renderer with sensible defaults
	rendererConfig := map[string]interface{}{
		"showLineNumbers": true,
		"tabWidth":        4,
	}
	
	if err := terminalRenderer.Configure(rendererConfig); err != nil {
		return fmt.Errorf("failed to configure terminal renderer: %w", err)
	}
	
	return nil
}

// initializeParsers registers all built-in parsers
func initializeParsers() error {
	registry := plugin.GetRegistry()
	
	// Register CommonMark parser
	commonMarkParser := parsers.NewCommonMarkParser()
	if err := registry.RegisterParser(commonMarkParser.Name(), commonMarkParser); err != nil {
		return fmt.Errorf("failed to register CommonMark parser: %w", err)
	}
	
	return nil
}

// setDefaultPlugins sets the default plugins
func setDefaultPlugins() error {
	registry := plugin.GetRegistry()
	
	// Set default theme
	if err := registry.SetDefaultTheme("dark"); err != nil {
		return fmt.Errorf("failed to set default theme: %w", err)
	}
	
	// Set default renderer (always terminal for now)
	if err := registry.SetDefaultRenderer("terminal"); err != nil {
		return fmt.Errorf("failed to set default renderer: %w", err)
	}
	
	// Set default parser (always commonmark for now)
	if err := registry.SetDefaultParser("commonmark"); err != nil {
		return fmt.Errorf("failed to set default parser: %w", err)
	}
	
	return nil
}

// GetPluginStatus returns the status of all registered plugins
func GetPluginStatus() map[string]interface{} {
	registry := plugin.GetRegistry()
	
	return map[string]interface{}{
		"parsers":   registry.ListParsers(),
		"renderers": registry.ListRenderers(),
		"themes":    registry.ListThemes(),
	}
}

