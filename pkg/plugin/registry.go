package plugin

import (
	"fmt"
	"sync"
	"github.com/ofri/mde/pkg/theme"
)

// Registry manages plugin registration and discovery
type Registry struct {
	mu sync.RWMutex
	
	// Registered plugins
	parsers   map[string]ParserPlugin
	renderers map[string]RendererPlugin
	themes    map[string]theme.Theme
	
	// Default plugins
	defaultParser   string
	defaultRenderer string
	defaultTheme    string
}

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{
		parsers:   make(map[string]ParserPlugin),
		renderers: make(map[string]RendererPlugin),
		themes:    make(map[string]theme.Theme),
	}
}

// RegisterParser registers a parser plugin
func (r *Registry) RegisterParser(name string, plugin ParserPlugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.parsers[name]; exists {
		return fmt.Errorf("parser plugin '%s' already registered", name)
	}
	
	r.parsers[name] = plugin
	
	// Set as default if it's the first parser
	if len(r.parsers) == 1 {
		r.defaultParser = name
	}
	
	return nil
}

// RegisterRenderer registers a renderer plugin
func (r *Registry) RegisterRenderer(name string, plugin RendererPlugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.renderers[name]; exists {
		return fmt.Errorf("renderer plugin '%s' already registered", name)
	}
	
	r.renderers[name] = plugin
	
	// Set as default if it's the first renderer
	if len(r.renderers) == 1 {
		r.defaultRenderer = name
	}
	
	return nil
}

// RegisterTheme registers a theme plugin
func (r *Registry) RegisterTheme(name string, theme theme.Theme) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.themes[name]; exists {
		return fmt.Errorf("theme '%s' already registered", name)
	}
	
	r.themes[name] = theme
	
	// Set as default if it's the first theme
	if len(r.themes) == 1 {
		r.defaultTheme = name
	}
	
	return nil
}

// GetParser retrieves a parser plugin by name
func (r *Registry) GetParser(name string) (ParserPlugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	plugin, exists := r.parsers[name]
	if !exists {
		return nil, fmt.Errorf("parser plugin '%s' not found", name)
	}
	
	return plugin, nil
}

// GetRenderer retrieves a renderer plugin by name
func (r *Registry) GetRenderer(name string) (RendererPlugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	plugin, exists := r.renderers[name]
	if !exists {
		return nil, fmt.Errorf("renderer plugin '%s' not found", name)
	}
	
	return plugin, nil
}

// GetTheme retrieves a theme by name
func (r *Registry) GetTheme(name string) (theme.Theme, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	themePlugin, exists := r.themes[name]
	if !exists {
		return nil, fmt.Errorf("theme '%s' not found", name)
	}
	
	return themePlugin, nil
}

// GetDefaultParser returns the default parser plugin
func (r *Registry) GetDefaultParser() (ParserPlugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if r.defaultParser == "" {
		return nil, fmt.Errorf("no default parser registered")
	}
	
	return r.parsers[r.defaultParser], nil
}

// GetDefaultRenderer returns the default renderer plugin
func (r *Registry) GetDefaultRenderer() (RendererPlugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if r.defaultRenderer == "" {
		return nil, fmt.Errorf("no default renderer registered")
	}
	
	return r.renderers[r.defaultRenderer], nil
}

// GetDefaultTheme returns the default theme
func (r *Registry) GetDefaultTheme() (theme.Theme, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if r.defaultTheme == "" {
		return nil, fmt.Errorf("no default theme registered")
	}
	
	return r.themes[r.defaultTheme], nil
}

// SetDefaultParser sets the default parser plugin
func (r *Registry) SetDefaultParser(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.parsers[name]; !exists {
		return fmt.Errorf("parser plugin '%s' not registered", name)
	}
	
	r.defaultParser = name
	return nil
}

// SetDefaultRenderer sets the default renderer plugin
func (r *Registry) SetDefaultRenderer(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.renderers[name]; !exists {
		return fmt.Errorf("renderer plugin '%s' not registered", name)
	}
	
	r.defaultRenderer = name
	return nil
}

// SetDefaultTheme sets the default theme
func (r *Registry) SetDefaultTheme(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.themes[name]; !exists {
		return fmt.Errorf("theme '%s' not registered", name)
	}
	
	r.defaultTheme = name
	return nil
}

// ListParsers returns a list of registered parser names
func (r *Registry) ListParsers() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	names := make([]string, 0, len(r.parsers))
	for name := range r.parsers {
		names = append(names, name)
	}
	
	return names
}

// ListRenderers returns a list of registered renderer names
func (r *Registry) ListRenderers() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	names := make([]string, 0, len(r.renderers))
	for name := range r.renderers {
		names = append(names, name)
	}
	
	return names
}

// ListThemes returns a list of registered theme names
func (r *Registry) ListThemes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	names := make([]string, 0, len(r.themes))
	for name := range r.themes {
		names = append(names, name)
	}
	
	return names
}

// Global registry instance
var globalRegistry = NewRegistry()

// RegisterParser registers a parser plugin globally
func RegisterParser(name string, plugin ParserPlugin) error {
	return globalRegistry.RegisterParser(name, plugin)
}

// RegisterRenderer registers a renderer plugin globally
func RegisterRenderer(name string, plugin RendererPlugin) error {
	return globalRegistry.RegisterRenderer(name, plugin)
}

// RegisterTheme registers a theme globally
func RegisterTheme(name string, theme theme.Theme) error {
	return globalRegistry.RegisterTheme(name, theme)
}

// GetRegistry returns the global registry instance
func GetRegistry() *Registry {
	return globalRegistry
}

// ResetRegistry resets the global registry (for testing)
func ResetRegistry() {
	globalRegistry = NewRegistry()
}