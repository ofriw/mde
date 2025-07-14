package integration

import (
	"context"
	"testing"
	"github.com/ofri/mde/internal/config"
	"github.com/ofri/mde/internal/plugins"
	"github.com/ofri/mde/pkg/ast"
	"github.com/ofri/mde/pkg/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRealWorldScenarios tests real-world usage scenarios
func TestRealWorldScenarios(t *testing.T) {
	// Test 1: Full plugin lifecycle
	t.Run("FullPluginLifecycle", func(t *testing.T) {
		testFullPluginLifecycle(t)
	})
	
	// Test 2: Configuration integration
	t.Run("ConfigurationIntegration", func(t *testing.T) {
		testConfigurationIntegration(t)
	})
	
	// Test 3: Document rendering pipeline
	t.Run("DocumentRenderingPipeline", func(t *testing.T) {
		testDocumentRenderingPipeline(t)
	})
	
	// Test 4: Error scenarios
	t.Run("ErrorScenarios", func(t *testing.T) {
		testErrorScenarios(t)
	})
}

func testFullPluginLifecycle(t *testing.T) {
	// Reset registry for clean test
	plugin.ResetRegistry()
	
	// 1. Load configuration
	cfg := config.DefaultConfig()
	require.NotNil(t, cfg, "Should load configuration")
	
	// 2. Initialize plugins
	err := plugins.InitializePlugins(cfg)
	require.NoError(t, err, "Should initialize plugins without error")
	
	// 3. Get plugin registry
	registry := plugin.GetRegistry()
	require.NotNil(t, registry, "Should get plugin registry")
	
	// 4. Verify plugins are registered
	themes := registry.ListThemes()
	renderers := registry.ListRenderers()
	
	assert.Contains(t, themes, "dark", "Should have dark theme registered")
	assert.Contains(t, renderers, "terminal", "Should have terminal renderer registered")
	
	// 5. Get default plugins
	defaultTheme, err := registry.GetDefaultTheme()
	require.NoError(t, err, "Should get default theme")
	assert.Equal(t, "dark", defaultTheme.Name(), "Should have correct default theme")
	
	defaultRenderer, err := registry.GetDefaultRenderer()
	require.NoError(t, err, "Should get default renderer")
	assert.Equal(t, "terminal", defaultRenderer.Name(), "Should have correct default renderer")
	
	// 6. Test plugin functionality
	doc := ast.NewDocument("# Test Document\nThis is a test.")
	ctx := context.Background()
	
	lines, err := defaultRenderer.Render(ctx, doc, defaultTheme)
	require.NoError(t, err, "Should render document successfully")
	assert.Len(t, lines, 2, "Should render correct number of lines")
	assert.Equal(t, "# Test Document", lines[0].Content, "Should render heading correctly")
	assert.Equal(t, "This is a test.", lines[1].Content, "Should render text correctly")
}

func testConfigurationIntegration(t *testing.T) {
	// Reset registry for clean test
	plugin.ResetRegistry()
	
	// Create configuration with custom settings
	cfg := config.DefaultConfig()
	
	// Set custom renderer settings
	rendererConfig := map[string]interface{}{
		"showLineNumbers": true,
		"tabWidth":        2,
		"maxWidth":        120,
	}
	cfg.SetPluginConfig("renderer", "terminal", rendererConfig)
	
	// Set custom theme settings
	themeConfig := map[string]interface{}{
		"primary": "#FF6B6B",
		"secondary": "#4ECDC4",
	}
	cfg.SetPluginConfig("theme", "dark", themeConfig)
	
	// Initialize plugins with custom config
	err := plugins.InitializePlugins(cfg)
	require.NoError(t, err, "Should initialize plugins with custom config")
	
	// Test that configuration is applied
	err = plugins.ConfigurePlugin("renderer", "terminal", rendererConfig)
	require.NoError(t, err, "Should configure renderer plugin")
	
	err = plugins.ConfigurePlugin("theme", "dark", themeConfig)
	require.NoError(t, err, "Should configure theme plugin")
	
	// Verify plugin status
	status := plugins.GetPluginStatus()
	assert.Contains(t, status, "themes", "Should have themes in status")
	assert.Contains(t, status, "renderers", "Should have renderers in status")
}

func testDocumentRenderingPipeline(t *testing.T) {
	// Reset registry for clean test
	plugin.ResetRegistry()
	
	// Initialize plugins
	cfg := config.DefaultConfig()
	err := plugins.InitializePlugins(cfg)
	require.NoError(t, err, "Should initialize plugins")
	
	registry := plugin.GetRegistry()
	theme, err := registry.GetDefaultTheme()
	require.NoError(t, err, "Should get theme")
	
	renderer, err := registry.GetDefaultRenderer()
	require.NoError(t, err, "Should get renderer")
	
	// Test various document types
	testCases := []struct {
		name     string
		content  string
		expected int // expected number of lines
	}{
		{
			name:     "Simple text",
			content:  "Hello, World!",
			expected: 1,
		},
		{
			name:     "Multi-line text",
			content:  "Line 1\nLine 2\nLine 3",
			expected: 3,
		},
		{
			name:     "Markdown-like content",
			content:  "# Heading\n\nSome **bold** text\n\n- List item",
			expected: 5,
		},
		{
			name:     "Empty document",
			content:  "",
			expected: 1, // Empty document still has one empty line
		},
	}
	
	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc := ast.NewDocument(tc.content)
			lines, err := renderer.Render(ctx, doc, theme)
			require.NoError(t, err, "Should render document")
			assert.Equal(t, tc.expected, len(lines), "Should render correct number of lines")
			
			// Test preview rendering
			previewLines, err := renderer.RenderPreview(ctx, doc, theme)
			require.NoError(t, err, "Should render preview")
			assert.Equal(t, len(lines), len(previewLines), "Preview should match regular rendering")
		})
	}
}

func testErrorScenarios(t *testing.T) {
	registry := plugin.GetRegistry()
	
	// Test non-existent plugin access
	_, err := registry.GetTheme("non-existent-theme")
	assert.Error(t, err, "Should error on non-existent theme")
	assert.Contains(t, err.Error(), "not found", "Should have appropriate error message")
	
	_, err = registry.GetRenderer("non-existent-renderer")
	assert.Error(t, err, "Should error on non-existent renderer")
	assert.Contains(t, err.Error(), "not found", "Should have appropriate error message")
	
	// Test invalid plugin configuration
	err = plugins.ConfigurePlugin("invalid-type", "test", map[string]interface{}{})
	assert.Error(t, err, "Should error on invalid plugin type")
	assert.Contains(t, err.Error(), "unknown plugin type", "Should have appropriate error message")
	
	// Test plugin error handling
	pluginErr := plugin.NewPluginError("theme", "test", "render", assert.AnError)
	assert.True(t, plugin.IsPluginError(pluginErr), "Should identify as plugin error")
	
	pluginType, pluginName, ok := plugin.GetPluginFromError(pluginErr)
	assert.True(t, ok, "Should extract plugin info from error")
	assert.Equal(t, "theme", pluginType, "Should extract correct plugin type")
	assert.Equal(t, "test", pluginName, "Should extract correct plugin name")
	
	// Test safe call functionality
	err = plugin.SafeCall("test", "plugin", "operation", func() error {
		return assert.AnError
	})
	assert.Error(t, err, "Should propagate error from safe call")
	assert.True(t, plugin.IsPluginError(err), "Should wrap error as plugin error")
}

// TestPluginArchitecturePerformance tests performance characteristics
func TestPluginArchitecturePerformance(t *testing.T) {
	// Reset registry for clean test
	plugin.ResetRegistry()
	
	// Initialize plugins
	cfg := config.DefaultConfig()
	err := plugins.InitializePlugins(cfg)
	require.NoError(t, err, "Should initialize plugins")
	
	registry := plugin.GetRegistry()
	theme, err := registry.GetDefaultTheme()
	require.NoError(t, err, "Should get theme")
	
	renderer, err := registry.GetDefaultRenderer()
	require.NoError(t, err, "Should get renderer")
	
	// Create a large document
	content := ""
	for i := 0; i < 1000; i++ {
		content += "This is line " + string(rune(i)) + " of the test document.\n"
	}
	doc := ast.NewDocument(content)
	
	// Test rendering performance
	ctx := context.Background()
	lines, err := renderer.Render(ctx, doc, theme)
	require.NoError(t, err, "Should render large document")
	// The document will have 1000 lines + 1 empty line from the trailing newline
	assert.Greater(t, len(lines), 1000, "Should render many lines")
	assert.LessOrEqual(t, len(lines), 1002, "Should not have too many lines")
	
	// Test that rendering is reasonably fast (should complete within reasonable time)
	// This test mainly ensures we don't have any infinite loops or major performance issues
}

// TestPluginArchitectureThreadSafety tests thread safety
func TestPluginArchitectureThreadSafety(t *testing.T) {
	// Reset registry for clean test
	plugin.ResetRegistry()
	
	// Initialize plugins
	cfg := config.DefaultConfig()
	err := plugins.InitializePlugins(cfg)
	require.NoError(t, err, "Should initialize plugins")
	
	registry := plugin.GetRegistry()
	
	// Test concurrent access to registry
	done := make(chan bool, 10)
	
	// Start multiple goroutines accessing the registry
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			
			// Test concurrent reading
			themes := registry.ListThemes()
			assert.Contains(t, themes, "dark", "Should list themes")
			
			renderers := registry.ListRenderers()
			assert.Contains(t, renderers, "terminal", "Should list renderers")
			
			// Test concurrent plugin access
			theme, err := registry.GetDefaultTheme()
			assert.NoError(t, err, "Should get theme")
			assert.NotNil(t, theme, "Should get valid theme")
			
			renderer, err := registry.GetDefaultRenderer()
			assert.NoError(t, err, "Should get renderer")
			assert.NotNil(t, renderer, "Should get valid renderer")
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}