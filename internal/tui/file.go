package tui

import (
	"bufio"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/ofri/mde/pkg/ast"
)

type fileLoadedMsg struct {
	filename string
	content  []string
	err      error
}

type fileSavedMsg struct {
	filename string
	err      error
}

type fileOpenPromptMsg struct{}

func (m *Model) openFile() tea.Cmd {
	return func() tea.Msg {
		return fileOpenPromptMsg{}
	}
}

func (m *Model) loadFile(filename string) tea.Cmd {
	return func() tea.Msg {
		file, err := os.Open(filename)
		if err != nil {
			return fileLoadedMsg{filename: filename, err: err}
		}
		defer file.Close()

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return fileLoadedMsg{filename: filename, err: err}
		}

		if len(lines) == 0 {
			lines = []string{""}
		}

		return fileLoadedMsg{filename: filename, content: lines}
	}
}

func (m *Model) saveFile() tea.Cmd {
	filename := m.editor.GetDocument().GetFilename()
	if filename == "" {
		m.showMessage("No filename specified")
		return nil
	}

	return func() tea.Msg {
		err := m.editor.SaveFile(filename)
		return fileSavedMsg{filename: filename, err: err}
	}
}


func (m *Model) handleFileMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case fileLoadedMsg:
		if msg.err != nil {
			m.showMessage("Error loading file: " + msg.err.Error())
			return m, nil
		}
		// Load content into editor
		content := strings.Join(msg.content, "\n")
		m.editor = ast.NewEditorWithContent(content)
		m.editor.GetDocument().SetFilename(msg.filename)
		m.editor.GetDocument().ClearModified()
		m.showMessage("Loaded " + msg.filename)
		return m, nil

	case fileSavedMsg:
		if msg.err != nil {
			m.showMessage("Error saving file: " + msg.err.Error())
			return m, nil
		}
		m.showMessage("Saved " + msg.filename)
		return m, nil

	case fileOpenPromptMsg:
		m.showMessage("Open file functionality coming soon")
		return m, nil
	}

	return m, nil
}