package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ofri/mde/internal/config"
	"github.com/ofri/mde/internal/plugins"
	"github.com/ofri/mde/internal/tui"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	
	// Initialize plugins
	err = plugins.InitializePlugins(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing plugins: %v\n", err)
		os.Exit(1)
	}
	
	app := tui.New()
	
	if len(os.Args) > 1 {
		app.SetFilename(os.Args[1])
	}
	
	p := tea.NewProgram(app, tea.WithAltScreen())
	
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}