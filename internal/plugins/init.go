package plugins

import (
	"fmt"
	"github.com/ofri/mde/internal/config"
	"github.com/ofri/mde/internal/plugins/parsers"
	"github.com/ofri/mde/internal/plugins/renderers"
	"github.com/ofri/mde/internal/plugins/themes"
	"github.com/ofri/mde/pkg/plugin"
)

// InitializePlugins initializes all built-in plugins
func InitializePlugins(cfg *config.Config) error {
	// Initialize themes
	if err := initializeThemes(cfg); err != nil {
		return fmt.Errorf("failed to initialize themes: %w", err)
	}
	
	// Initialize renderers
	if err := initializeRenderers(cfg); err != nil {
		return fmt.Errorf("failed to initialize renderers: %w", err)
	}
	
	// Initialize parsers (when we have them)
	if err := initializeParsers(cfg); err != nil {
		return fmt.Errorf("failed to initialize parsers: %w", err)
	}
	
	// Set default plugins
	if err := setDefaultPlugins(cfg); err != nil {
		return fmt.Errorf("failed to set default plugins: %w", err)
	}
	
	return nil
}

// initializeThemes registers all built-in themes
func initializeThemes(cfg *config.Config) error {
	registry := plugin.GetRegistry()
	
	// Register dark theme
	darkTheme := themes.NewDarkTheme()
	if err := registry.RegisterTheme(darkTheme.Name(), darkTheme); err != nil {
		return fmt.Errorf("failed to register dark theme: %w", err)
	}
	
	// Configure theme with user settings
	themeConfig := cfg.GetPluginConfig("theme", darkTheme.Name())
	if err := darkTheme.Configure(themeConfig); err != nil {
		return fmt.Errorf("failed to configure dark theme: %w", err)
	}
	
	return nil
}

// initializeRenderers registers all built-in renderers
func initializeRenderers(cfg *config.Config) error {
	registry := plugin.GetRegistry()
	
	// Register terminal renderer
	terminalRenderer := renderers.NewTerminalRenderer()
	if err := registry.RegisterRenderer(terminalRenderer.Name(), terminalRenderer); err != nil {
		return fmt.Errorf("failed to register terminal renderer: %w", err)
	}
	
	// Configure renderer with user settings
	rendererConfig := cfg.GetPluginConfig("renderer", terminalRenderer.Name())
	if err := terminalRenderer.Configure(rendererConfig); err != nil {
		return fmt.Errorf("failed to configure terminal renderer: %w", err)
	}
	
	return nil
}

// initializeParsers registers all built-in parsers
func initializeParsers(cfg *config.Config) error {
	registry := plugin.GetRegistry()
	
	// Register CommonMark parser
	commonMarkParser := parsers.NewCommonMarkParser()
	if err := registry.RegisterParser(commonMarkParser.Name(), commonMarkParser); err != nil {
		return fmt.Errorf("failed to register CommonMark parser: %w", err)
	}
	
	// Configure parser with user settings
	parserConfig := cfg.GetPluginConfig("parser", commonMarkParser.Name())
	if err := commonMarkParser.Configure(parserConfig); err != nil {
		return fmt.Errorf("failed to configure CommonMark parser: %w", err)
	}
	
	return nil
}

// setDefaultPlugins sets the default plugins based on configuration
func setDefaultPlugins(cfg *config.Config) error {
	registry := plugin.GetRegistry()
	
	// Set default theme
	if cfg.Plugins.DefaultTheme != "" {
		if err := registry.SetDefaultTheme(cfg.Plugins.DefaultTheme); err != nil {
			return fmt.Errorf("failed to set default theme '%s': %w", cfg.Plugins.DefaultTheme, err)
		}
	}
	
	// Set default renderer
	if cfg.Plugins.DefaultRenderer != "" {
		if err := registry.SetDefaultRenderer(cfg.Plugins.DefaultRenderer); err != nil {
			return fmt.Errorf("failed to set default renderer '%s': %w", cfg.Plugins.DefaultRenderer, err)
		}
	}
	
	// Set default parser
	if cfg.Plugins.DefaultParser != "" {
		if err := registry.SetDefaultParser(cfg.Plugins.DefaultParser); err != nil {
			return fmt.Errorf("failed to set default parser '%s': %w", cfg.Plugins.DefaultParser, err)
		}
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

// ConfigurePlugin configures a specific plugin with new settings
func ConfigurePlugin(pluginType, pluginName string, options map[string]interface{}) error {
	registry := plugin.GetRegistry()
	
	switch pluginType {
	case "renderer":
		renderer, err := registry.GetRenderer(pluginName)
		if err != nil {
			return fmt.Errorf("renderer '%s' not found: %w", pluginName, err)
		}
		return renderer.Configure(options)
		
	case "theme":
		theme, err := registry.GetTheme(pluginName)
		if err != nil {
			return fmt.Errorf("theme '%s' not found: %w", pluginName, err)
		}
		return theme.Configure(options)
		
	case "parser":
		parser, err := registry.GetParser(pluginName)
		if err != nil {
			return fmt.Errorf("parser '%s' not found: %w", pluginName, err)
		}
		return parser.Configure(options)
		
	default:
		return fmt.Errorf("unknown plugin type: %s", pluginType)
	}
}