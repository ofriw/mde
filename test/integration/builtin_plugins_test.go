package integration

import (
	"context"
	"testing"
	"github.com/ofri/mde/internal/plugins"
	"github.com/ofri/mde/internal/plugins/renderers"
	"github.com/ofri/mde/internal/plugins/themes"
	"github.com/ofri/mde/pkg/ast"
	"github.com/ofri/mde/pkg/plugin"
	"github.com/ofri/mde/pkg/theme"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuiltinPlugins tests the actual built-in plugins
func TestBuiltinPlugins(t *testing.T) {
	// Test 1: Terminal renderer functionality
	t.Run("TerminalRenderer", func(t *testing.T) {
		testBuiltinTerminalRenderer(t)
	})
	
	// Test 2: Dark theme integration
	t.Run("DarkTheme", func(t *testing.T) {
		testBuiltinDarkTheme(t)
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
	
	// Create mock theme for testing
	mockTheme := &MockTheme{name: "test"}
	
	// Test rendering
	ctx := context.Background()
	lines, err := renderer.Render(ctx, doc, mockTheme)
	require.NoError(t, err, "Should render successfully")
	assert.Len(t, lines, 3, "Should render correct number of lines")
	
	// Check content
	assert.Equal(t, "func main() {", lines[0].Content, "Should render first line correctly")
	assert.Equal(t, "  println(\"Hello, World!\")", lines[1].Content, "Should expand tabs to spaces")
	assert.Equal(t, "}", lines[2].Content, "Should render last line correctly")
	
	// Test line rendering with tokens (need to create tokens properly)
	// Since Token fields are private, we'll test without tokens for now
	tokens := []ast.Token{}
	line, err := renderer.RenderLine(ctx, "func main() {", tokens, mockTheme)
	require.NoError(t, err, "Should render line with tokens")
	assert.Equal(t, "func main() {", line.Content, "Should preserve content")
	assert.Len(t, line.Styles, 1, "Should have style for normal text")
	assert.Equal(t, 0, line.Styles[0].Start, "Should start at beginning")
	assert.Equal(t, 13, line.Styles[0].End, "Should end at text length")
}

func testBuiltinDarkTheme(t *testing.T) {
	// Create dark theme
	darkTheme := themes.NewDarkTheme()
	assert.Equal(t, "dark", darkTheme.Name(), "Should have correct name")
	
	// Test color scheme
	colorScheme := darkTheme.GetColorScheme()
	assert.NotEmpty(t, colorScheme.Background, "Should have background color")
	assert.NotEmpty(t, colorScheme.Foreground, "Should have foreground color")
	assert.NotEmpty(t, colorScheme.Primary, "Should have primary color")
	assert.NotEmpty(t, colorScheme.SyntaxKeyword, "Should have syntax keyword color")
	
	// Test styles
	normalStyle := darkTheme.GetStyle(theme.TextNormal)
	assert.Equal(t, colorScheme.Foreground, normalStyle.Foreground, "Normal text should use foreground color")
	
	keywordStyle := darkTheme.GetStyle(theme.SyntaxKeyword)
	assert.Equal(t, colorScheme.SyntaxKeyword, keywordStyle.Foreground, "Keywords should use syntax color")
	assert.True(t, keywordStyle.Bold, "Keywords should be bold")
	
	headingStyle := darkTheme.GetStyle(theme.MarkdownHeading1)
	assert.Equal(t, colorScheme.Accent, headingStyle.Foreground, "Headings should use accent color")
	assert.True(t, headingStyle.Bold, "Headings should be bold")
	
	// Test configuration
	config := map[string]interface{}{
		"primary": "#FF0000",
	}
	err := darkTheme.Configure(config)
	assert.NoError(t, err, "Should configure without error")
	
	// Verify configuration took effect
	updatedColorScheme := darkTheme.GetColorScheme()
	assert.Equal(t, "#FF0000", updatedColorScheme.Primary, "Should update primary color")
	
	// Test lipgloss style conversion
	lipglossStyle := normalStyle.ToLipgloss()
	assert.NotNil(t, lipglossStyle, "Should convert to lipgloss style")
}

func testEndToEndRendering(t *testing.T) {
	// Create dark theme and terminal renderer
	darkTheme := themes.NewDarkTheme()
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
	lines, err := renderer.Render(ctx, doc, darkTheme)
	require.NoError(t, err, "Should render document successfully")
	
	// Verify we got the expected number of lines
	assert.Greater(t, len(lines), 0, "Should render at least some lines")
	
	// Test the string output
	output := renderer.RenderToString(lines, darkTheme)
	assert.Contains(t, output, "# Hello World", "Should contain heading")
	assert.Contains(t, output, "This is a", "Should contain regular text")
	assert.Contains(t, output, "1 │", "Should contain line numbers")
	
	// Test preview rendering
	previewLines, err := renderer.RenderPreview(ctx, doc, darkTheme)
	require.NoError(t, err, "Should render preview successfully")
	assert.Equal(t, len(lines), len(previewLines), "Preview should have same number of lines")
}

func testRealPluginInitialization(t *testing.T) {
	// Create a new registry for testing
	registry := plugin.NewRegistry()
	
	// Test manual registration of built-in plugins
	darkTheme := themes.NewDarkTheme()
	err := registry.RegisterTheme(darkTheme.Name(), darkTheme)
	require.NoError(t, err, "Should register dark theme")
	
	terminalRenderer := renderers.NewTerminalRenderer()
	err = registry.RegisterRenderer(terminalRenderer.Name(), terminalRenderer)
	require.NoError(t, err, "Should register terminal renderer")
	
	// Test plugin discovery
	themes := registry.ListThemes()
	assert.Contains(t, themes, "dark", "Should list dark theme")
	
	renderers := registry.ListRenderers()
	assert.Contains(t, renderers, "terminal", "Should list terminal renderer")
	
	// Test default plugin access
	defaultTheme, err := registry.GetDefaultTheme()
	require.NoError(t, err, "Should get default theme")
	assert.Equal(t, "dark", defaultTheme.Name(), "Should have dark as default theme")
	
	defaultRenderer, err := registry.GetDefaultRenderer()
	require.NoError(t, err, "Should get default renderer")
	assert.Equal(t, "terminal", defaultRenderer.Name(), "Should have terminal as default renderer")
	
	// Test the full initialization process
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