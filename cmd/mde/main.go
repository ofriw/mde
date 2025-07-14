package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ofri/mde/internal/tui"
)

func main() {
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