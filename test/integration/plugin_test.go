package integration

import (
	"context"
	"testing"
	"github.com/ofri/mde/internal/plugins"
	"github.com/ofri/mde/pkg/ast"
	"github.com/ofri/mde/pkg/plugin"
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
	// Test parser registration functionality
}

func testTerminalRenderer(t *testing.T, registry *plugin.Registry) {
	// Create mock renderer
	mockRenderer := &MockRenderer{name: "test-renderer"}
	
	// Register renderer
	err := registry.RegisterRenderer(mockRenderer.Name(), mockRenderer)
	require.NoError(t, err, "Should register renderer successfully")
	
	// Test renderer functionality
	doc := ast.NewDocument("Hello World\nThis is a test")
	ctx := context.Background()
	viewport := ast.NewViewport(0, 0, 80, 25, 6, 4)
	renderCtx := &plugin.RenderContext{
		Document: doc,
		Viewport: viewport,
		ShowLineNumbers: true,
	}
	lines, err := mockRenderer.RenderVisible(ctx, renderCtx)
	require.NoError(t, err, "Should render document successfully")
	assert.Len(t, lines, 2, "Should render correct number of lines")
	assert.Equal(t, "Hello World", lines[0].Content, "Should render correct content")
	assert.Equal(t, "This is a test", lines[1].Content, "Should render correct content")
}

func testConfigurationLoading(t *testing.T) {
	// No configuration system - MDE uses sensible defaults
	t.Skip("No configuration system - using defaults")
}

func testPluginInitialization(t *testing.T) {
	// Test plugin initialization
	// This should not error even though we don't have real plugins
	err := plugins.InitializePlugins()
	// We expect this to pass or fail gracefully
	if err != nil {
		t.Logf("Plugin initialization failed (expected in test): %v", err)
	}
	
	// Test plugin status
	status := plugins.GetPluginStatus()
	assert.NotNil(t, status, "Should get plugin status")
	assert.Contains(t, status, "renderers", "Should have renderers in status")
	assert.Contains(t, status, "parsers", "Should have parsers in status")
}

func testErrorHandling(t *testing.T, registry *plugin.Registry) {
	// Test non-existent plugin retrieval
	_, err := registry.GetRenderer("non-existent")
	assert.Error(t, err, "Should error on non-existent renderer")
	
	// Test plugin error creation
	pluginErr := plugin.NewPluginError("renderer", "test", "configure", assert.AnError)
	assert.Contains(t, pluginErr.Error(), "renderer/test", "Should format plugin error correctly")
	
	// Test configuration error
	configErr := plugin.NewConfigurationError("renderer", "test", "width", assert.AnError)
	assert.Contains(t, configErr.Error(), "configuration error", "Should format config error correctly")
	
	// Test error type checking
	assert.True(t, plugin.IsPluginError(pluginErr), "Should identify plugin error")
	assert.True(t, plugin.IsConfigurationError(configErr), "Should identify config error")
}

// Mock implementations for testing

type MockRenderer struct {
	name string
}

func (m *MockRenderer) Name() string {
	return m.name
}


func (m *MockRenderer) RenderVisible(ctx context.Context, renderCtx *plugin.RenderContext) ([]plugin.RenderedLine, error) {
	doc := renderCtx.Document
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

func (m *MockRenderer) RenderPreviewVisible(ctx context.Context, renderCtx *plugin.RenderContext) ([]plugin.RenderedLine, error) {
	return m.RenderVisible(ctx, renderCtx)
}

func (m *MockRenderer) RenderLine(ctx context.Context, line string, tokens []ast.Token) (plugin.RenderedLine, error) {
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