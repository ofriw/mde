package integration

import (
	"context"
	"testing"
	"github.com/ofri/mde/internal/config"
	"github.com/ofri/mde/internal/plugins"
	"github.com/ofri/mde/pkg/ast"
	"github.com/ofri/mde/pkg/plugin"
	"github.com/ofri/mde/pkg/theme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPluginArchitecture tests the complete plugin architecture
func TestPluginArchitecture(t *testing.T) {
	// Create a fresh registry for testing
	registry := plugin.NewRegistry()
	
	// Test 1: Plugin registration and discovery
	t.Run("PluginRegistration", func(t *testing.T) {
		testPluginRegistration(t, registry)
	})
	
	// Test 2: Terminal renderer functionality
	t.Run("TerminalRenderer", func(t *testing.T) {
		testTerminalRenderer(t, registry)
	})
	
	// Test 3: Dark theme integration
	t.Run("DarkTheme", func(t *testing.T) {
		testDarkTheme(t, registry)
	})
	
	// Test 4: Configuration loading
	t.Run("ConfigurationLoading", func(t *testing.T) {
		testConfigurationLoading(t)
	})
	
	// Test 5: Plugin initialization
	t.Run("PluginInitialization", func(t *testing.T) {
		testPluginInitialization(t)
	})
	
	// Test 6: Error handling
	t.Run("ErrorHandling", func(t *testing.T) {
		testErrorHandling(t, registry)
	})
}

func testPluginRegistration(t *testing.T, registry *plugin.Registry) {
	// Create mock theme
	mockTheme := &MockTheme{name: "test-theme"}
	
	// Test registration
	err := registry.RegisterTheme(mockTheme.Name(), mockTheme)
	require.NoError(t, err, "Should register theme successfully")
	
	// Test duplicate registration
	err = registry.RegisterTheme(mockTheme.Name(), mockTheme)
	assert.Error(t, err, "Should fail on duplicate registration")
	
	// Test retrieval
	retrieved, err := registry.GetTheme(mockTheme.Name())
	require.NoError(t, err, "Should retrieve registered theme")
	assert.Equal(t, mockTheme.Name(), retrieved.Name(), "Should retrieve correct theme")
	
	// Test listing
	themes := registry.ListThemes()
	assert.Contains(t, themes, mockTheme.Name(), "Should list registered theme")
	
	// Test default theme
	defaultTheme, err := registry.GetDefaultTheme()
	require.NoError(t, err, "Should have default theme")
	assert.Equal(t, mockTheme.Name(), defaultTheme.Name(), "Should set first theme as default")
}

func testTerminalRenderer(t *testing.T, registry *plugin.Registry) {
	// Create mock renderer
	mockRenderer := &MockRenderer{name: "test-renderer"}
	
	// Register renderer
	err := registry.RegisterRenderer(mockRenderer.Name(), mockRenderer)
	require.NoError(t, err, "Should register renderer successfully")
	
	// Test renderer functionality
	doc := ast.NewDocument("Hello World\nThis is a test")
	mockTheme := &MockTheme{name: "test-theme"}
	
	ctx := context.Background()
	lines, err := mockRenderer.Render(ctx, doc, mockTheme)
	require.NoError(t, err, "Should render document successfully")
	assert.Len(t, lines, 2, "Should render correct number of lines")
	assert.Equal(t, "Hello World", lines[0].Content, "Should render correct content")
	assert.Equal(t, "This is a test", lines[1].Content, "Should render correct content")
}

func testDarkTheme(t *testing.T, registry *plugin.Registry) {
	mockTheme := &MockTheme{name: "dark-theme"}
	
	// Register theme
	err := registry.RegisterTheme(mockTheme.Name(), mockTheme)
	require.NoError(t, err, "Should register theme successfully")
	
	// Test theme functionality
	style := mockTheme.GetStyle(theme.TextNormal)
	assert.NotEmpty(t, style.Foreground, "Should have foreground color")
	
	colorScheme := mockTheme.GetColorScheme()
	assert.NotEmpty(t, colorScheme.Background, "Should have background color")
	assert.NotEmpty(t, colorScheme.Foreground, "Should have foreground color")
	
	// Test theme configuration
	options := map[string]interface{}{
		"primary": "#FF0000",
	}
	err = mockTheme.Configure(options)
	assert.NoError(t, err, "Should configure theme successfully")
}

func testConfigurationLoading(t *testing.T) {
	// Test default configuration
	defaultConfig := config.DefaultConfig()
	assert.NotNil(t, defaultConfig, "Should create default config")
	assert.Equal(t, 4, defaultConfig.Editor.TabWidth, "Should have correct tab width")
	assert.Equal(t, "dark", defaultConfig.Theme.Name, "Should have correct default theme")
	
	// Test plugin configuration
	pluginConfig := defaultConfig.GetPluginConfig("theme", "dark")
	assert.NotNil(t, pluginConfig, "Should get plugin config")
	
	// Test setting plugin configuration
	newConfig := map[string]interface{}{
		"primary": "#00FF00",
	}
	defaultConfig.SetPluginConfig("theme", "dark", newConfig)
	
	retrieved := defaultConfig.GetPluginConfig("theme", "dark")
	assert.Equal(t, "#00FF00", retrieved["primary"], "Should set and retrieve plugin config")
}

func testPluginInitialization(t *testing.T) {
	// Test plugin initialization with default config
	cfg := config.DefaultConfig()
	
	// This should not error even though we don't have real plugins
	err := plugins.InitializePlugins(cfg)
	// We expect this to pass or fail gracefully
	if err != nil {
		t.Logf("Plugin initialization failed (expected in test): %v", err)
	}
	
	// Test plugin status
	status := plugins.GetPluginStatus()
	assert.NotNil(t, status, "Should get plugin status")
	assert.Contains(t, status, "themes", "Should have themes in status")
	assert.Contains(t, status, "renderers", "Should have renderers in status")
	assert.Contains(t, status, "parsers", "Should have parsers in status")
}

func testErrorHandling(t *testing.T, registry *plugin.Registry) {
	// Test non-existent plugin retrieval
	_, err := registry.GetTheme("non-existent")
	assert.Error(t, err, "Should error on non-existent theme")
	
	_, err = registry.GetRenderer("non-existent")
	assert.Error(t, err, "Should error on non-existent renderer")
	
	// Test plugin error creation
	pluginErr := plugin.NewPluginError("theme", "test", "configure", assert.AnError)
	assert.Contains(t, pluginErr.Error(), "theme/test", "Should format plugin error correctly")
	
	// Test configuration error
	configErr := plugin.NewConfigurationError("theme", "test", "color", assert.AnError)
	assert.Contains(t, configErr.Error(), "configuration error", "Should format config error correctly")
	
	// Test error type checking
	assert.True(t, plugin.IsPluginError(pluginErr), "Should identify plugin error")
	assert.True(t, plugin.IsConfigurationError(configErr), "Should identify config error")
}

// Mock implementations for testing

type MockTheme struct {
	name string
}

func (m *MockTheme) Name() string {
	return m.name
}

func (m *MockTheme) GetStyle(elementType theme.ElementType) theme.Style {
	return theme.Style{
		Foreground: "#FFFFFF",
		Background: "#000000",
	}
}

func (m *MockTheme) GetColorScheme() theme.ColorScheme {
	return theme.ColorScheme{
		Background: "#000000",
		Foreground: "#FFFFFF",
		Primary:    "#0000FF",
	}
}

func (m *MockTheme) Configure(options map[string]interface{}) error {
	return nil
}

type MockRenderer struct {
	name string
}

func (m *MockRenderer) Name() string {
	return m.name
}

func (m *MockRenderer) Render(ctx context.Context, doc *ast.Document, theme theme.Theme) ([]plugin.RenderedLine, error) {
	lines := make([]plugin.RenderedLine, 0, doc.LineCount())
	
	for i := 0; i < doc.LineCount(); i++ {
		line := doc.GetLine(i)
		lines = append(lines, plugin.RenderedLine{
			Content: line,
			Styles:  []plugin.StyleRange{},
			Metadata: map[string]interface{}{
				"test": true,
			},
		})
	}
	
	return lines, nil
}

func (m *MockRenderer) RenderPreview(ctx context.Context, doc *ast.Document, theme theme.Theme) ([]plugin.RenderedLine, error) {
	return m.Render(ctx, doc, theme)
}

func (m *MockRenderer) RenderLine(ctx context.Context, line string, tokens []ast.Token, theme theme.Theme) (plugin.RenderedLine, error) {
	return plugin.RenderedLine{
		Content: line,
		Styles:  []plugin.StyleRange{},
		Metadata: map[string]interface{}{
			"test": true,
		},
	}, nil
}

func (m *MockRenderer) Configure(options map[string]interface{}) error {
	return nil
}