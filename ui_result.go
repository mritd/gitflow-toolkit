package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

const (
	successTitle = `🟢 COMMIT SUCCESS`
	failedTitle  = `🔴 COMMIT FAILED`
	successMsg   = `Always code as if the guy who ends up maintaining your code will be a violent psychopath who knows where you live.`
)

var (
	layOutStyle = lipgloss.NewStyle().
			Padding(1, 0, 1, 2)

	doneTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 0, 1, 22)

	doneMsgStyle = lipgloss.NewStyle().
			Bold(true).
			Width(64)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#37B9FF")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#37B9FF")).
			Padding(1, 3, 1, 3)

	failedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF62DA")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF62DA")).
			Padding(1, 3, 1, 3)
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
		title := doneTitleStyle.Render(successTitle)
		message := doneMsgStyle.Render(strings.TrimSpace(m.message))
		return layOutStyle.Render(successStyle.Render(lipgloss.JoinVertical(lipgloss.Left, title, message)))
	} else {
		title := doneTitleStyle.Render(failedTitle)
		message := doneMsgStyle.Render(strings.TrimSpace(m.err.Error()))
		return layOutStyle.Render(failedStyle.Render(lipgloss.JoinVertical(lipgloss.Left, title, message)))
	}
}
