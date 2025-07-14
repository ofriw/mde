package config

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	// Editor settings
	Editor EditorConfig `mapstructure:"editor"`
	
	// Plugin settings
	Plugins PluginConfig `mapstructure:"plugins"`
	
	// Theme settings
	Theme ThemeConfig `mapstructure:"theme"`
}

// EditorConfig holds editor-specific configuration
type EditorConfig struct {
	// Tab width
	TabWidth int `mapstructure:"tab_width"`
	
	// Line numbers
	ShowLineNumbers bool `mapstructure:"show_line_numbers"`
	
	// Viewport settings
	ViewportWidth  int `mapstructure:"viewport_width"`
	ViewportHeight int `mapstructure:"viewport_height"`
	
	// Auto-save settings
	AutoSave bool `mapstructure:"auto_save"`
	
	// History settings
	HistorySize int `mapstructure:"history_size"`
}

// PluginConfig holds plugin-specific configuration
type PluginConfig struct {
	// Default plugins
	DefaultParser   string `mapstructure:"default_parser"`
	DefaultRenderer string `mapstructure:"default_renderer"`
	DefaultTheme    string `mapstructure:"default_theme"`
	
	// Plugin-specific settings
	ParserSettings   map[string]interface{} `mapstructure:"parser_settings"`
	RendererSettings map[string]interface{} `mapstructure:"renderer_settings"`
	ThemeSettings    map[string]interface{} `mapstructure:"theme_settings"`
}

// ThemeConfig holds theme-specific configuration
type ThemeConfig struct {
	// Current theme
	Name string `mapstructure:"name"`
	
	// Theme customizations
	Colors map[string]string `mapstructure:"colors"`
	
	// Style overrides
	Styles map[string]interface{} `mapstructure:"styles"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Editor: EditorConfig{
			TabWidth:        4,
			ShowLineNumbers: false,
			ViewportWidth:   80,
			ViewportHeight:  24,
			AutoSave:        false,
			HistorySize:     1000,
		},
		Plugins: PluginConfig{
			DefaultParser:    "commonmark",
			DefaultRenderer:  "terminal",
			DefaultTheme:     "dark",
			ParserSettings:   make(map[string]interface{}),
			RendererSettings: make(map[string]interface{}),
			ThemeSettings:    make(map[string]interface{}),
		},
		Theme: ThemeConfig{
			Name:   "dark",
			Colors: make(map[string]string),
			Styles: make(map[string]interface{}),
		},
	}
}

// Load loads configuration from file and environment
func Load() (*Config, error) {
	// Set up viper
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	
	// Add config paths
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.config/mde")
	v.AddConfigPath("/etc/mde")
	
	// Set environment variable prefix
	v.SetEnvPrefix("MDE")
	v.AutomaticEnv()
	
	// Set defaults
	setDefaults(v)
	
	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, use defaults
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}
	
	// Unmarshal into config struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}
	
	return &config, nil
}

// Save saves configuration to file
func (c *Config) Save() error {
	// Create config directory if it doesn't exist
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "mde")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}
	
	// Set up viper for writing
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configDir)
	
	// Set values from config
	v.Set("editor", c.Editor)
	v.Set("plugins", c.Plugins)
	v.Set("theme", c.Theme)
	
	// Write config file
	configFile := filepath.Join(configDir, "config.yaml")
	if err := v.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}
	
	return nil
}

// setDefaults sets default values in viper
func setDefaults(v *viper.Viper) {
	defaults := DefaultConfig()
	
	// Editor defaults
	v.SetDefault("editor.tab_width", defaults.Editor.TabWidth)
	v.SetDefault("editor.show_line_numbers", defaults.Editor.ShowLineNumbers)
	v.SetDefault("editor.viewport_width", defaults.Editor.ViewportWidth)
	v.SetDefault("editor.viewport_height", defaults.Editor.ViewportHeight)
	v.SetDefault("editor.auto_save", defaults.Editor.AutoSave)
	v.SetDefault("editor.history_size", defaults.Editor.HistorySize)
	
	// Plugin defaults
	v.SetDefault("plugins.default_parser", defaults.Plugins.DefaultParser)
	v.SetDefault("plugins.default_renderer", defaults.Plugins.DefaultRenderer)
	v.SetDefault("plugins.default_theme", defaults.Plugins.DefaultTheme)
	v.SetDefault("plugins.parser_settings", defaults.Plugins.ParserSettings)
	v.SetDefault("plugins.renderer_settings", defaults.Plugins.RendererSettings)
	v.SetDefault("plugins.theme_settings", defaults.Plugins.ThemeSettings)
	
	// Theme defaults
	v.SetDefault("theme.name", defaults.Theme.Name)
	v.SetDefault("theme.colors", defaults.Theme.Colors)
	v.SetDefault("theme.styles", defaults.Theme.Styles)
}

// GetPluginConfig returns plugin-specific configuration
func (c *Config) GetPluginConfig(pluginType, pluginName string) map[string]interface{} {
	var settings map[string]interface{}
	
	switch pluginType {
	case "parser":
		settings = c.Plugins.ParserSettings
	case "renderer":
		settings = c.Plugins.RendererSettings
	case "theme":
		settings = c.Plugins.ThemeSettings
	default:
		return make(map[string]interface{})
	}
	
	if pluginSettings, ok := settings[pluginName].(map[string]interface{}); ok {
		return pluginSettings
	}
	
	return make(map[string]interface{})
}

// SetPluginConfig sets plugin-specific configuration
func (c *Config) SetPluginConfig(pluginType, pluginName string, config map[string]interface{}) {
	var settings map[string]interface{}
	
	switch pluginType {
	case "parser":
		if c.Plugins.ParserSettings == nil {
			c.Plugins.ParserSettings = make(map[string]interface{})
		}
		settings = c.Plugins.ParserSettings
	case "renderer":
		if c.Plugins.RendererSettings == nil {
			c.Plugins.RendererSettings = make(map[string]interface{})
		}
		settings = c.Plugins.RendererSettings
	case "theme":
		if c.Plugins.ThemeSettings == nil {
			c.Plugins.ThemeSettings = make(map[string]interface{})
		}
		settings = c.Plugins.ThemeSettings
	default:
		return
	}
	
	settings[pluginName] = config
}