package main

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"time"
)

var (
	spinnerStyle = lipgloss.NewStyle().
			Padding(1, 1, 1, 2)

	spinnerBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#25A065")).
				Padding(1, 2, 1, 2)

	spinnerTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#25A065"))
)

type spinnerModel struct {
	err     error
	spinner spinner.Model
}

func newSpinnerModel() spinnerModel {
	s := spinner.NewModel()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return spinnerModel{spinner: s}
}

func (m spinnerModel) Init() tea.Cmd {
	return nil
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		default:
			return m, nil
		}
	case commitMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(spinner.Tick())
		return m, tea.Batch(cmd, func() tea.Msg {
			time.Sleep(500 * time.Millisecond)
			return done{err: execCommit(msg)}
		})
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m spinnerModel) View() string {
	spinnerView := m.spinner.View() + spinnerTextStyle.Render("Committing... Please wait...")
	return spinnerStyle.Render(spinnerBorderStyle.Render(spinnerView))
}
