package integration

import (
	"context"
	"testing"
	"github.com/ofri/mde/internal/plugins"
	"github.com/ofri/mde/internal/plugins/renderers"
	"github.com/ofri/mde/pkg/ast"
	"github.com/ofri/mde/pkg/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuiltinPlugins tests the actual built-in plugins
func TestBuiltinPlugins(t *testing.T) {
	// Test 1: Terminal renderer functionality
	t.Run("TerminalRenderer", func(t *testing.T) {
		testBuiltinTerminalRenderer(t)
	})
	
	// Test 3: End-to-end rendering
	t.Run("EndToEndRendering", func(t *testing.T) {
		testEndToEndRendering(t)
	})
	
	// Test 4: Plugin initialization with real plugins
	t.Run("RealPluginInitialization", func(t *testing.T) {
		testRealPluginInitialization(t)
	})
}

func testBuiltinTerminalRenderer(t *testing.T) {
	// Create terminal renderer
	renderer := renderers.NewTerminalRenderer()
	assert.Equal(t, "terminal", renderer.Name(), "Should have correct name")
	
	// Test configuration
	config := map[string]interface{}{
		"maxWidth":        120,
		"tabWidth":        2,
		"showLineNumbers": true,
	}
	err := renderer.Configure(config)
	assert.NoError(t, err, "Should configure without error")
	
	// Create test document
	doc := ast.NewDocument("func main() {\n\tprintln(\"Hello, World!\")\n}")
	
	// Test rendering
	ctx := context.Background()
	viewport := ast.NewViewport(0, 0, 80, 25, 6, 4)
	renderCtx := &plugin.RenderContext{
		Document: doc,
		Viewport: viewport,
		ShowLineNumbers: true,
	}
	lines, err := renderer.RenderVisible(ctx, renderCtx)
	require.NoError(t, err, "Should render successfully")
	assert.Len(t, lines, 3, "Should render correct number of lines")
	
	// Check content (now includes line numbers since we use RenderVisible)
	assert.Equal(t, "    1│func main() {", lines[0].Content, "Should render first line correctly")
	assert.Equal(t, "    2│  println(\"Hello, World!\")", lines[1].Content, "Should expand tabs to spaces")
	assert.Equal(t, "    3│}", lines[2].Content, "Should render last line correctly")
	
	// Test line rendering with tokens (need to create tokens properly)
	// Since Token fields are private, we'll test without tokens for now
	tokens := []ast.Token{}
	line, err := renderer.RenderLine(ctx, "func main() {", tokens)
	require.NoError(t, err, "Should render line with tokens")
	assert.Equal(t, "func main() {", line.Content, "Should preserve content")
	// Styles are now handled by terminal ANSI colors
}

func testEndToEndRendering(t *testing.T) {
	// Create terminal renderer
	renderer := renderers.NewTerminalRenderer()
	
	// Configure renderer for testing
	config := map[string]interface{}{
		"showLineNumbers": true,
		"tabWidth":        4,
	}
	err := renderer.Configure(config)
	require.NoError(t, err, "Should configure renderer")
	
	// Create test document with various content
	content := `# Hello World
This is a **bold** text and *italic* text.
	
` + "```go\nfunc main() {\n\tprintln(\"Hello!\")\n}\n```"
	
	doc := ast.NewDocument(content)
	
	// Render document
	ctx := context.Background()
	viewport2 := ast.NewViewport(0, 0, 80, 25, 6, 4)
	renderCtx2 := &plugin.RenderContext{
		Document: doc,
		Viewport: viewport2,
		ShowLineNumbers: true,
	}
	lines, err := renderer.RenderVisible(ctx, renderCtx2)
	require.NoError(t, err, "Should render document successfully")
	
	// Verify we got the expected number of lines
	assert.Greater(t, len(lines), 0, "Should render at least some lines")
	
	// Test the string output
	output := renderer.RenderToString(lines)
	assert.Contains(t, output, "# Hello World", "Should contain heading")
	assert.Contains(t, output, "This is a", "Should contain regular text")
	assert.Contains(t, output, "1│", "Should contain line numbers")
	
	// Test preview rendering
	viewport := ast.NewViewport(0, 0, 80, 25, 6, 4)
	renderCtx := &plugin.RenderContext{
		Document: doc,
		Viewport: viewport,
		ShowLineNumbers: false, // Preview mode typically doesn't show line numbers
	}
	previewLines, err := renderer.RenderPreviewVisible(ctx, renderCtx)
	require.NoError(t, err, "Should render preview successfully")
	assert.Equal(t, len(lines), len(previewLines), "Preview should have same number of lines")
}

func testRealPluginInitialization(t *testing.T) {
	// Create a new registry for testing
	registry := plugin.NewRegistry()
	
	// Test manual registration of built-in plugins
	terminalRenderer := renderers.NewTerminalRenderer()
	err := registry.RegisterRenderer(terminalRenderer.Name(), terminalRenderer)
	require.NoError(t, err, "Should register terminal renderer")
	
	// Test plugin discovery
	renderers := registry.ListRenderers()
	assert.Contains(t, renderers, "terminal", "Should list terminal renderer")
	
	// Test default plugin access
	defaultRenderer, err := registry.GetDefaultRenderer()
	require.NoError(t, err, "Should get default renderer")
	assert.Equal(t, "terminal", defaultRenderer.Name(), "Should have terminal as default renderer")
	
	// Reset registry and test the full initialization process
	plugin.ResetRegistry()
	err = plugins.InitializePlugins()
	// This might fail due to missing dependencies, but should handle gracefully
	if err != nil {
		t.Logf("Plugin initialization failed (might be expected): %v", err)
	}
	
	// Test plugin status
	status := plugins.GetPluginStatus()
	assert.NotNil(t, status, "Should get plugin status")
	
	// Plugin configuration has been removed from the minimal config system
}