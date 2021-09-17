package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	layOutStyle = lipgloss.NewStyle().
			Padding(1, 0, 1, 2)

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			MaxWidth(64).
			Foreground(lipgloss.Color("#FF62DA")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF62DA")).
			Padding(1, 3, 1, 3)
)

type errorModel struct {
	err error
}

func newErrorModel() errorModel {
	return errorModel{}
}

func (m errorModel) Init() tea.Cmd {
	return nil
}

func (m errorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case error:
		m.err = msg.(error)
	}
	return m, tea.Quit
}

func (m errorModel) View() string {
	if m.err == nil {
		return ""
	}
	return layOutStyle.Render(errorStyle.Render(m.err.Error()))
}
