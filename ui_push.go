package main

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	pushLayout = lipgloss.NewStyle().
			Padding(1, 1, 1, 2)

	pushSuccessTextStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#57FFBA"))

	pushSuccessMsgStyle = lipgloss.NewStyle().
				Bold(true).
				Width(75).
				Foreground(lipgloss.Color("#2FD0FF"))

	pushFailedTextStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#D63B3A"))

	pushFailedErrStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FF6037"))
)

type pushModel struct {
	done    bool
	msg     string
	err     error
	spinner spinner.Model
}

type pushMsg struct {
	err error
	msg string
}

func newPushModel() pushModel {
	s := spinner.NewModel()
	s.Spinner = spinner.Spinner{
		Frames: []string{
			"●∙∙∙ P",
			"∙●∙∙ Pu",
			"∙∙●∙ Pus",
			"∙∙∙● Push",
			"∙∙●∙ Pushi",
			"∙●∙∙ Pushin",
			"●∙∙∙ Pushing",
			"∙●∙∙ Pushing, p",
			"∙∙●∙ Pushing, pl",
			"∙∙∙● Pushing, ple",
			"∙∙●∙ Pushing, plea",
			"∙●∙∙ Pushing, pleas",
			"●∙∙∙ Pushing, please w",
			"∙●∙∙ Pushing, please wa",
			"∙∙●∙ Pushing, please wai",
			"∙∙∙● Pushing, please wait",
			"∙∙●∙ Pushing, please wait.",
			"∙●∙∙ Pushing, please wait..",
			"●∙∙∙ Pushing, please wait...",
		},
		FPS: time.Second / 15,
	}
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#19F896")).Bold(true)
	return pushModel{
		spinner: s,
	}
}

func (m pushModel) Init() tea.Cmd {
	return tea.Batch(spinner.Tick, func() tea.Msg { return "running" })
}

func (m pushModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		}
	case string:
		return m, func() tea.Msg {
			time.Sleep(1500 * time.Millisecond)
			var ps pushMsg
			ps.msg, ps.err = push()
			return ps
		}
	case pushMsg:
		m.msg, m.err = msg.msg, msg.err
		m.done = true
		return m, tea.Quit
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m pushModel) View() string {
	view := pushLayout.Render(m.spinner.View())

	if m.done {
		if m.err == nil {
			view = pushLayout.Render(
				lipgloss.JoinVertical(
					lipgloss.Left,
					pushSuccessTextStyle.Render("(￣▽￣)ノ The code has been pushed successfully.\n"),
					pushSuccessMsgStyle.Render(m.msg),
				),
			)
		} else {
			view = pushLayout.Render(
				lipgloss.JoinVertical(
					lipgloss.Left,
					pushFailedTextStyle.Render("(｡•́︿•̀｡) Something went wrong while pushing:\n"),
					pushFailedErrStyle.Render(m.err.Error()),
				),
			)
		}
	}

	return view
}
