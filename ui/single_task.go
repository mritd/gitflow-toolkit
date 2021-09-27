package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	SingleTaskLayoutStyle = lipgloss.NewStyle().
				Padding(0, 0, 1, 2)

	SingleTaskMsgLayout = lipgloss.NewStyle().
				Padding(1, 0, 1, 0)

	SingleTaskSuccessStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.AdaptiveColor{Light: "#1C9518", Dark: "#2AFFA3"})

	SingleTaskFailedStyle = SingleTaskSuccessStyle.Copy().
				Background(lipgloss.AdaptiveColor{Light: "#E11C9C", Dark: "#EE6FF8"})

	SingleTaskWaitingStyle = SingleTaskSuccessStyle.Copy().
				Foreground(lipgloss.AdaptiveColor{Light: "#2B53AF", Dark: "#37B9FF"})

	SingleTaskSpinner = spinner.Model{
		Style: lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#E11C9C", Dark: "#EE6FF8"}),
		Spinner: spinner.Spinner{
			Frames: []string{
				"[∙∙∙]",
				"[●∙∙]",
				"[∙●∙]",
				"[∙∙●]",
				"[∙∙∙]",
			},
			FPS: time.Second / 10,
		}}
)

type SingleTaskModel struct {
	Task    Task
	Spinner spinner.Model

	TaskDelay    time.Duration
	LayoutStyle  lipgloss.Style
	BorderStyle  lipgloss.Style
	SuccessStyle lipgloss.Style
	FailedStyle  lipgloss.Style
	RunningStyle lipgloss.Style

	err  error
	done bool
}

func NewSingleTaskModel(task Task) SingleTaskModel {
	return SingleTaskModel{
		Task:         task,
		Spinner:      SingleTaskSpinner,
		TaskDelay:    time.Second,
		LayoutStyle:  SingleTaskLayoutStyle,
		SuccessStyle: SingleTaskSuccessStyle,
		FailedStyle:  SingleTaskFailedStyle,
		RunningStyle: SingleTaskWaitingStyle,
	}
}

func (m SingleTaskModel) Init() tea.Cmd {
	return tea.Batch(spinner.Tick, func() tea.Msg {
		if m.TaskDelay > 0 {
			time.Sleep(m.TaskDelay)
		}
		return m.Task.Func
	})
}

func (m SingleTaskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case TaskFunc:
		m.Task.err = msg.(TaskFunc)()
		m.done = true
		return m, tea.Quit
	default:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	}
}

func (m SingleTaskModel) View() string {
	var view string
	if m.done {
		if m.Task.err == nil {
			view = m.SuccessStyle.Render("[ ✔ ] " + m.Task.Title)
		} else {
			view = m.FailedStyle.Render("[ ✗ ] " + m.Task.Title)
		}
	} else {
		view = m.Spinner.View() + " " + m.RunningStyle.Render(m.Task.Title)
	}

	return m.LayoutStyle.Render(m.BorderStyle.Render(SingleTaskMsgLayout.Render(view)))
}
