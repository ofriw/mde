package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ofri/mde/internal/plugins"
	"github.com/ofri/mde/internal/tui"
)

func main() {
	// Initialize plugins with defaults
	err := plugins.InitializePlugins()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing plugins: %v\n", err)
		os.Exit(1)
	}
	
	app := tui.New()
	
	if len(os.Args) > 1 {
		app.SetFilename(os.Args[1])
	}
	
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
	
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}