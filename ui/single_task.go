package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/progress"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	SingleTaskLayoutStyle = lipgloss.NewStyle().
				Padding(1, 0, 1, 2)

	SingleTaskBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#37B9FF")).
				Width(55).
				Padding(0, 1, 1, 2)

	SingleTaskSuccessStyle = lipgloss.NewStyle().
				Bold(true).
				Background(lipgloss.Color("#25A065")).
				Padding(1, 0, 0, 0)

	SingleTaskFailedStyle = SingleTaskSuccessStyle.Copy().
				Background(lipgloss.Color("#EE6FF8"))

	SingleTaskWaitingStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#37B9FF")).
				Padding(0, 0, 0, 0)

	SingleTaskSpinner = spinner.Model{
		Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#EE6FF8")).Padding(1, 0, 0, 0),
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
	Task  Task
	Width int

	Spinner  spinner.Model
	Progress progress.Model

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
		BorderStyle:  SingleTaskBorderStyle,
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

	return m.LayoutStyle.Render(m.BorderStyle.Render(view))
}
