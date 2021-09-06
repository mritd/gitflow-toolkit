package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	successMsg = `Always code as if the guy who ends up maintaining your code will be a violent psychopath who knows where you live.`
)

var (
	layOutStyle = lipgloss.NewStyle().
			Padding(1, 0, 1, 2)

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Width(64).
			Foreground(lipgloss.Color("#37B9FF")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#37B9FF")).
			Padding(1, 3, 0, 3)

	failedStyle = successStyle.Copy().
			Foreground(lipgloss.Color("#FF62DA")).
			BorderForeground(lipgloss.Color("#FF62DA"))
)

type resultModel struct {
	message string
	err     error
}

func newResultModel() resultModel {
	return resultModel{}
}

func (m resultModel) Init() tea.Cmd {
	return nil
}

func (m resultModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case error:
		m.err = msg.(error)
	case string:
		m.message = msg.(string)
	}
	return m, tea.Quit
}

func (m resultModel) View() string {
	if m.err == nil {
		return layOutStyle.Render(successStyle.Render(m.message))
	} else {
		return layOutStyle.Render(failedStyle.Render(m.err.Error()))
	}
}
