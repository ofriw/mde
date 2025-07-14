package tui

import (
	"bufio"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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
	if m.filename == "" {
		m.showMessage("No filename specified")
		return nil
	}

	return func() tea.Msg {
		content := strings.Join(m.content, "\n")
		err := os.WriteFile(m.filename, []byte(content), 0644)
		return fileSavedMsg{filename: m.filename, err: err}
	}
}

func (m *Model) showMessage(msg string) {
	m.message = msg
	m.messageTimer = 60
}

func (m *Model) handleFileMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case fileLoadedMsg:
		if msg.err != nil {
			m.showMessage("Error loading file: " + msg.err.Error())
			return m, nil
		}
		m.content = msg.content
		m.filename = msg.filename
		m.cursor = position{0, 0}
		m.offset = 0
		m.modified = false
		m.showMessage("Loaded " + msg.filename)
		return m, nil

	case fileSavedMsg:
		if msg.err != nil {
			m.showMessage("Error saving file: " + msg.err.Error())
			return m, nil
		}
		m.modified = false
		m.showMessage("Saved " + msg.filename)
		return m, nil

	case fileOpenPromptMsg:
		m.showMessage("Open file functionality coming soon")
		return m, nil
	}

	return m, nil
}