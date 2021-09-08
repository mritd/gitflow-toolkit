package main

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	commitStyle = lipgloss.NewStyle().
			Padding(1, 1, 1, 2)

	commitTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#25A065"))

	commitBorderStyle = commitTextStyle.Copy().
				Border(lipgloss.RoundedBorder()).
				BorderBottomBackground(lipgloss.Color("#25A065")).
				Padding(1, 2, 1, 2)
)

type commitModel struct {
	err     error
	spinner spinner.Model
}

func newCommitModel() commitModel {
	s := spinner.NewModel()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return commitModel{spinner: s}
}

func (m commitModel) Init() tea.Cmd {
	return nil
}

func (m commitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			return done{err: commit(msg)}
		})
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m commitModel) View() string {
	spinnerView := m.spinner.View() + commitTextStyle.Render("Committing... Please wait...")
	return commitStyle.Render(commitBorderStyle.Render(spinnerView))
}
