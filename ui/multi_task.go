package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var NothingFunc = func() error { return nil }

var (
	MultiTaskLayoutStyle = lipgloss.NewStyle().
				Padding(0, 0, 1, 2)

	MultiTaskBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#37B9FF")).
				Width(55).
				Padding(0, 1, 1, 2)

	MultiTaskMsgSuccessStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#25A065"))

	MultiTaskMsgFailedStyle = MultiTaskMsgSuccessStyle.Copy().
				Foreground(lipgloss.Color("#EE6FF8"))

	MultiTaskMsgWaitingStyle = MultiTaskMsgSuccessStyle.Copy().
					Foreground(lipgloss.Color("#37B9FF"))

	MultiTaskSpinner = spinner.Model{
		Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#F8CA61")),
		Spinner: spinner.Spinner{
			Frames: []string{
				"[     ]",
				"[=    ]",
				"[==   ]",
				"[===  ]",
				"[ === ]",
				"[  ===]",
				"[   ==]",
				"[    =]",
				"[     ]",
				"[    =]",
				"[   ==]",
				"[  ===]",
				"[ === ]",
				"[===  ]",
				"[==   ]",
				"[=    ]",
				"[     ]",
			},
			FPS: time.Second / 10,
		}}
)

type TaskFunc func() error

type TaskDoneMsg struct{}

type Task struct {
	Title string
	Func  TaskFunc

	err error
}

type MultiTaskModel struct {
	Tasks   []Task
	Spinner spinner.Model

	TaskDelay       time.Duration
	LayoutStyle     lipgloss.Style
	BorderStyle     lipgloss.Style
	MsgSuccessStyle lipgloss.Style
	MsgFailedStyle  lipgloss.Style
	MsgWaitingStyle lipgloss.Style

	index int
}

func NewMultiTaskModel() MultiTaskModel {
	return NewMultiTaskModelWithTasks(nil)
}

func NewMultiTaskModelWithTasks(tasks []Task) MultiTaskModel {
	return MultiTaskModel{
		Tasks:           tasks,
		Spinner:         MultiTaskSpinner,
		TaskDelay:       300 * time.Millisecond,
		LayoutStyle:     MultiTaskLayoutStyle,
		BorderStyle:     MultiTaskBorderStyle,
		MsgSuccessStyle: MultiTaskMsgSuccessStyle,
		MsgFailedStyle:  MultiTaskMsgFailedStyle,
		MsgWaitingStyle: MultiTaskMsgWaitingStyle,
	}
}

func (m MultiTaskModel) Init() tea.Cmd {
	return tea.Batch(spinner.Tick, func() tea.Msg {
		if len(m.Tasks) == 0 {
			panic("The number of tasks cannot be 0")
		}
		return m.Tasks[0]
	})
}

func (m MultiTaskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case Task:
		return m, func() tea.Msg {
			if m.TaskDelay > 0 {
				time.Sleep(m.TaskDelay)
			}
			m.Tasks[m.index].err = msg.(Task).Func()
			return TaskDoneMsg{}
		}
	case TaskDoneMsg:
		if m.Tasks[m.index].err != nil {
			return m, tea.Quit
		}

		m.index++
		return m, func() tea.Msg {
			if m.index == len(m.Tasks) {
				return tea.Quit()
			}
			return m.Tasks[m.index]
		}
	default:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	}

}

func (m MultiTaskModel) View() string {
	var view string
	for i, task := range m.Tasks {
		if task.err == nil {
			if i < m.index {
				view = lipgloss.JoinVertical(lipgloss.Left, view, m.MsgSuccessStyle.Render("[  ✔  ] "+task.Title))
			} else {
				view = lipgloss.JoinVertical(lipgloss.Left, view, m.Spinner.View()+" "+m.MsgWaitingStyle.Render(task.Title))
			}
		} else {
			view = lipgloss.JoinVertical(lipgloss.Left, view, m.MsgFailedStyle.Render("[  ✗  ] "+task.err.Error()))
		}

	}
	return m.LayoutStyle.Render(m.BorderStyle.Render(view))
}
