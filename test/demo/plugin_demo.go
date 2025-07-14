package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"github.com/ofri/mde/internal/config"
	"github.com/ofri/mde/internal/plugins"
	"github.com/ofri/mde/pkg/ast"
	"github.com/ofri/mde/pkg/plugin"
)

func main() {
	fmt.Println("=== MDE Plugin Architecture Demo ===")
	
	// 1. Load configuration
	fmt.Println("\n1. Loading configuration...")
	cfg, err := config.Load()
	if err != nil {
		// Use defaults if config file not found
		cfg = config.DefaultConfig()
		fmt.Printf("   Using default configuration (no config file found)\n")
	} else {
		fmt.Printf("   Configuration loaded successfully\n")
	}
	
	// 2. Initialize plugins
	fmt.Println("\n2. Initializing plugins...")
	if err := plugins.InitializePlugins(cfg); err != nil {
		log.Fatalf("Failed to initialize plugins: %v", err)
	}
	fmt.Printf("   Plugins initialized successfully\n")
	
	// 3. Show plugin status
	fmt.Println("\n3. Plugin Status:")
	status := plugins.GetPluginStatus()
	fmt.Printf("   Themes: %v\n", status["themes"])
	fmt.Printf("   Renderers: %v\n", status["renderers"])
	fmt.Printf("   Parsers: %v\n", status["parsers"])
	
	// 4. Get plugin registry
	fmt.Println("\n4. Getting plugin registry...")
	registry := plugin.GetRegistry()
	
	// 5. Test theme functionality
	fmt.Println("\n5. Testing theme functionality...")
	theme, err := registry.GetDefaultTheme()
	if err != nil {
		log.Fatalf("Failed to get default theme: %v", err)
	}
	fmt.Printf("   Default theme: %s\n", theme.Name())
	
	// Test theme styles
	colorScheme := theme.GetColorScheme()
	fmt.Printf("   Theme colors - Background: %s, Foreground: %s\n", 
		colorScheme.Background, colorScheme.Foreground)
	
	// 6. Test renderer functionality
	fmt.Println("\n6. Testing renderer functionality...")
	renderer, err := registry.GetDefaultRenderer()
	if err != nil {
		log.Fatalf("Failed to get default renderer: %v", err)
	}
	fmt.Printf("   Default renderer: %s\n", renderer.Name())
	
	// 7. Create a test document
	fmt.Println("\n7. Creating test document...")
	testContent := `# Welcome to MDE Plugin Architecture
This is a **demonstration** of the plugin system.

## Features
- Plugin registration and discovery
- Theme support with color schemes
- Terminal renderer with styling
- Configuration management

## Code Example
` + "```go\nfunc main() {\n\tfmt.Println(\"Hello, MDE!\")\n}\n```"
	
	doc := ast.NewDocument(testContent)
	fmt.Printf("   Document created with %d lines\n", doc.LineCount())
	
	// 8. Render the document
	fmt.Println("\n8. Rendering document...")
	ctx := context.Background()
	lines, err := renderer.Render(ctx, doc, theme)
	if err != nil {
		log.Fatalf("Failed to render document: %v", err)
	}
	fmt.Printf("   Document rendered successfully into %d lines\n", len(lines))
	
	// 9. Show rendered output
	fmt.Println("\n9. Rendered Output:")
	fmt.Println("   " + strings.Repeat("=", 60))
	for i, line := range lines {
		// Show first 10 lines to avoid too much output
		if i >= 10 {
			fmt.Printf("   ... (%d more lines)\n", len(lines)-i)
			break
		}
		fmt.Printf("   %s\n", line.Content)
	}
	fmt.Println("   " + strings.Repeat("=", 60))
	
	// 10. Test configuration
	fmt.Println("\n10. Testing configuration...")
	rendererConfig := map[string]interface{}{
		"showLineNumbers": true,
		"tabWidth":        2,
	}
	
	err = plugins.ConfigurePlugin("renderer", "terminal", rendererConfig)
	if err != nil {
		log.Fatalf("Failed to configure renderer: %v", err)
	}
	fmt.Printf("   Renderer configured successfully\n")
	
	// 11. Test error handling
	fmt.Println("\n11. Testing error handling...")
	_, err = registry.GetTheme("non-existent-theme")
	if err != nil {
		fmt.Printf("   Error handling works: %v\n", err)
	}
	
	// 12. Performance test
	fmt.Println("\n12. Performance test...")
	largeContent := ""
	for i := 0; i < 1000; i++ {
		largeContent += fmt.Sprintf("Line %d: This is a test line with some content.\n", i+1)
	}
	
	largeDoc := ast.NewDocument(largeContent)
	largeLines, err := renderer.Render(ctx, largeDoc, theme)
	if err != nil {
		log.Fatalf("Failed to render large document: %v", err)
	}
	fmt.Printf("   Large document (%d lines) rendered successfully\n", len(largeLines))
	
	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("The plugin architecture is working correctly!")
}