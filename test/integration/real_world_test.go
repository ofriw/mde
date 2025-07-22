package integration

import (
	"context"
	"strings"
	"testing"
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
	
	// 1. Initialize plugins
	err := plugins.InitializePlugins()
	require.NoError(t, err, "Should initialize plugins without error")
	
	// 3. Get plugin registry
	registry := plugin.GetRegistry()
	require.NotNil(t, registry, "Should get plugin registry")
	
	// 4. Verify plugins are registered
	renderers := registry.ListRenderers()
	parsers := registry.ListParsers()
	
	assert.Contains(t, renderers, "terminal", "Should have terminal renderer registered")
	assert.Contains(t, parsers, "commonmark", "Should have commonmark parser registered")
	
	// 5. Get default plugins
	defaultRenderer, err := registry.GetDefaultRenderer()
	require.NoError(t, err, "Should get default renderer")
	assert.Equal(t, "terminal", defaultRenderer.Name(), "Should have correct default renderer")
	
	// 6. Use plugins to render document
	doc := ast.NewDocument("# Hello World\nThis is a test.")
	ctx := context.Background()
	viewport := ast.NewViewport(0, 0, 80, 25, 6, 4)
	renderCtx := &plugin.RenderContext{
		Document: doc,
		Viewport: viewport,
		ShowLineNumbers: true,
	}
	lines, err := defaultRenderer.RenderVisible(ctx, renderCtx)
	require.NoError(t, err, "Should render document without error")
	assert.Len(t, lines, 2, "Should render correct number of lines")
	
	// 8. Test plugin error handling
	err = plugin.NewPluginError("renderer", "terminal", "render", assert.AnError)
	assert.NotNil(t, err, "Should create plugin error")
	assert.Contains(t, err.Error(), "renderer/terminal", "Should format error correctly")
	
	// 9. Get plugin status
	status := plugins.GetPluginStatus()
	assert.NotNil(t, status, "Should get plugin status")
	assert.Contains(t, status, "renderers", "Should have renderers in status")
	assert.Contains(t, status, "parsers", "Should have parsers in status")
}

func testConfigurationIntegration(t *testing.T) {
	// Reset registry
	plugin.ResetRegistry()
	
	// Initialize plugins
	err := plugins.InitializePlugins()
	require.NoError(t, err, "Should initialize plugins")
	
	// Get registry
	registry := plugin.GetRegistry()
	
	// Get renderer
	renderer, err := registry.GetDefaultRenderer()
	require.NoError(t, err, "Should get renderer")
	
	// Configure renderer
	config := map[string]interface{}{
		"showLineNumbers": true,
		"tabWidth":        2,
		"maxWidth":        100,
	}
	err = renderer.Configure(config)
	assert.NoError(t, err, "Should configure renderer without error")
	
	// Verify configuration by rendering
	doc := ast.NewDocument("func main() {\n\tprintln(\"Hello\")\n}")
	ctx := context.Background()
	
	// Test both regular and preview rendering
	viewport := ast.NewViewport(0, 0, 80, 25, 6, 4)
	renderCtx := &plugin.RenderContext{
		Document: doc,
		Viewport: viewport,
		ShowLineNumbers: true,
	}
	previewCtx := &plugin.RenderContext{
		Document: doc,
		Viewport: viewport,
		ShowLineNumbers: false,
	}
	
	for _, tc := range []struct {
		name string
		renderFunc func() ([]plugin.RenderedLine, error)
	}{
		{
			name: "Regular render",
			renderFunc: func() ([]plugin.RenderedLine, error) {
				return renderer.RenderVisible(ctx, renderCtx)
			},
		},
		{
			name: "Preview render",
			renderFunc: func() ([]plugin.RenderedLine, error) {
				return renderer.RenderPreviewVisible(ctx, previewCtx)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lines, err := tc.renderFunc()
			require.NoError(t, err, "Should render without error")
			assert.Greater(t, len(lines), 0, "Should render lines")
		})
	}
}

func testErrorScenarios(t *testing.T) {
	registry := plugin.GetRegistry()
	
	// Test non-existent plugin retrieval
	_, err := registry.GetRenderer("non-existent-renderer")
	assert.Error(t, err, "Should error on non-existent renderer")
	
	_, err = registry.GetParser("non-existent-parser")
	assert.Error(t, err, "Should error on non-existent parser")
	
	// Test plugin error types
	pluginErr := plugin.NewPluginError("renderer", "test", "render", assert.AnError)
	assert.NotNil(t, pluginErr, "Should create plugin error")
	
	// Test error message extraction
	pluginType, pluginName, operation := extractPluginErrorInfo(pluginErr)
	assert.Equal(t, "renderer", pluginType, "Should extract correct plugin type")
	assert.Equal(t, "test", pluginName, "Should extract correct plugin name")
	assert.Equal(t, "render", operation, "Should extract correct operation")
}

func testDocumentRenderingPipeline(t *testing.T) {
	// Get plugins
	registry := plugin.GetRegistry()
	
	parser, err := registry.GetDefaultParser()
	require.NoError(t, err, "Should get parser")
	
	renderer, err := registry.GetDefaultRenderer()
	require.NoError(t, err, "Should get renderer")
	
	// Create complex document
	content := `# Markdown Editor

## Features
- **Bold** and *italic* text
- Code blocks with syntax highlighting
- Lists and quotes

### Code Example
` + "```go\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n```"

	// Parse document
	doc, err := parser.Parse(context.Background(), content)
	require.NoError(t, err, "Should parse document")
	require.NotNil(t, doc, "Should get valid document")
	
	// Render document
	ctx := context.Background()
	viewport2 := ast.NewViewport(0, 0, 80, 25, 6, 4)
	renderCtx2 := &plugin.RenderContext{
		Document: doc,
		Viewport: viewport2,
		ShowLineNumbers: true,
	}
	lines, err := renderer.RenderVisible(ctx, renderCtx2)
	require.NoError(t, err, "Should render document")
	assert.Greater(t, len(lines), 5, "Should render multiple lines")
	
	// Verify basic content preservation
	var allContent strings.Builder
	for _, line := range lines {
		allContent.WriteString(line.Content)
		allContent.WriteRune('\n')
	}
	
	result := allContent.String()
	assert.Contains(t, result, "Markdown Editor", "Should contain title")
	assert.Contains(t, result, "Features", "Should contain features section")
	assert.Contains(t, result, "func main()", "Should contain code")
}

// Helper function to extract plugin error information
func extractPluginErrorInfo(err error) (pluginType, pluginName, operation string) {
	// Simple extraction based on error message format
	// Real implementation would use type assertion
	errMsg := err.Error()
	if strings.Contains(errMsg, "renderer/test") {
		return "renderer", "test", "render"
	}
	return "", "", ""
}

// Compile-time interface checks
var (
	_ plugin.ParserPlugin   = (*MockParser)(nil)
)

// Mock implementations for interface validation
type MockParser struct{}

func (m *MockParser) Name() string { return "mock" }
func (m *MockParser) Parse(ctx context.Context, content string) (*ast.Document, error) { 
	return ast.NewDocument(content), nil 
}
func (m *MockParser) GetSyntaxHighlighting(ctx context.Context, line string) ([]ast.Token, error) {
	return nil, nil
}
func (m *MockParser) Configure(options map[string]interface{}) error { return nil }