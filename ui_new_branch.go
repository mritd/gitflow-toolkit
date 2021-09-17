package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	newBranchLayout = lipgloss.NewStyle().
			Padding(1, 1, 1, 2)

	newBranchSuccessTextStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#2AD67F"))

	newBranchSuccessNameStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#FFFDF5")).
					Background(lipgloss.Color("#7653FF"))

	newBranchFailedTextStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#D63B3A"))

	newBranchFailedErrStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#D63B3A"))
)

type branchModel struct {
	branch string
	done   bool
	err    error
}

func newBranchModel(branch string) branchModel {
	return branchModel{branch: branch}
}

func (m branchModel) Init() tea.Cmd {
	return func() tea.Msg { return m.branch }
}

func (m branchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		}
	case string:
		time.Sleep(500 * time.Millisecond)
		_, m.err = createBranch(msg)
		m.done = true
		return m, tea.Quit
	}

	return m, nil
}

func (m branchModel) View() string {
	msg := newBranchLayout.Render(newBranchSuccessTextStyle.Render(" (⊙ˍ⊙) Creating new branch: " + m.branch + " "))
	if m.done {
		if m.err == nil {
			msg = newBranchLayout.Render(
				newBranchSuccessTextStyle.Render(" (^_~) Switched to a new branch: ") +
					newBranchSuccessNameStyle.Render(m.branch) + " ")
		} else {
			msg = newBranchLayout.Render(
				newBranchFailedTextStyle.Render(" (｡•́︿•̀｡) Create failed: ") +
					newBranchFailedErrStyle.Render(m.err.Error()) + " ")
		}
	}
	return msg
}
