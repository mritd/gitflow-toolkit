package common

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TaskState represents the state of a task.
type TaskState int

const (
	TaskPending TaskState = iota
	TaskRunning
	TaskSuccess
	TaskWarning
	TaskFailed
)

// Task represents a single task with name and execution function.
type Task struct {
	Name string
	Run  func() error
}

// WarnErr is an error that indicates a warning (non-fatal).
type WarnErr struct {
	Msg string
}

func (e WarnErr) Error() string {
	return e.Msg
}

// IsWarnErr checks if an error is a warning.
func IsWarnErr(err error) bool {
	var warnErr WarnErr
	return errors.As(err, &warnErr)
}

// taskDoneMsg is sent when a task completes.
type taskDoneMsg struct {
	index int
	err   error
}

// MultiTaskModel is a model for running multiple tasks sequentially with progress display.
type MultiTaskModel struct {
	tasks       []Task
	states      []TaskState
	errors      []error
	currentTask int
	spinner     spinner.Model
	done        bool
	title       string
}

// NewMultiTaskModel creates a new multi-task model.
func NewMultiTaskModel(title string, tasks []Task) MultiTaskModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	states := make([]TaskState, len(tasks))
	if len(tasks) > 0 {
		states[0] = TaskRunning
	}

	return MultiTaskModel{
		tasks:       tasks,
		states:      states,
		errors:      make([]error, len(tasks)),
		currentTask: 0,
		spinner:     s,
		title:       title,
	}
}

// Init initializes the model.
func (m MultiTaskModel) Init() tea.Cmd {
	if len(m.tasks) == 0 {
		return m.spinner.Tick
	}

	task := m.tasks[0]
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			err := task.Run()
			return taskDoneMsg{index: 0, err: err}
		},
	)
}

// Update handles messages.
func (m MultiTaskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case taskDoneMsg:
		if msg.err != nil {
			if IsWarnErr(msg.err) {
				m.states[msg.index] = TaskWarning
			} else {
				m.states[msg.index] = TaskFailed
			}
			m.errors[msg.index] = msg.err
		} else {
			m.states[msg.index] = TaskSuccess
		}

		// Stop on fatal error, continue on warning
		if msg.err != nil && !IsWarnErr(msg.err) {
			m.done = true
			return m, tea.Quit
		}

		// Run next task
		m.currentTask++
		if m.currentTask >= len(m.tasks) {
			m.done = true
			return m, tea.Quit
		}

		m.states[m.currentTask] = TaskRunning
		task := m.tasks[m.currentTask]
		index := m.currentTask
		return m, func() tea.Msg {
			err := task.Run()
			return taskDoneMsg{index: index, err: err}
		}
	}

	return m, nil
}

// View renders the model.
func (m MultiTaskModel) View() string {
	var sb strings.Builder

	if m.title != "" {
		sb.WriteString(StyleTitle.Render(m.title))
		sb.WriteString("\n")
	}

	for i, task := range m.tasks {
		var icon string
		var style lipgloss.Style

		switch m.states[i] {
		case TaskPending:
			icon = StyleMuted.Render(SymbolPending)
			style = StyleMuted
		case TaskRunning:
			icon = m.spinner.View()
			style = StylePrimary
		case TaskSuccess:
			icon = StyleSuccess.Render(SymbolSuccess)
			style = StyleSuccess
		case TaskWarning:
			icon = StyleWarning.Render(SymbolWarning)
			style = StyleWarning
		case TaskFailed:
			icon = StyleError.Render(SymbolError)
			style = StyleError
		}

		sb.WriteString(fmt.Sprintf("  %s %s", icon, style.Render(task.Name)))

		if m.errors[i] != nil && m.states[i] != TaskRunning {
			sb.WriteString(StyleMuted.Render(fmt.Sprintf(" (%s)", m.errors[i].Error())))
		}

		sb.WriteString("\n")
	}

	if m.done {
		sb.WriteString("\n")
		if m.HasError() {
			sb.WriteString(StyleError.Render("Some tasks failed."))
		} else {
			sb.WriteString(StyleSuccess.Render("All tasks completed successfully."))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// HasError returns true if any task failed (non-warning).
func (m MultiTaskModel) HasError() bool {
	for i, err := range m.errors {
		if err != nil && m.states[i] == TaskFailed {
			return true
		}
	}
	return false
}

// SingleTaskModel is a model for running a single task with a spinner.
type SingleTaskModel struct {
	task    Task
	spinner spinner.Model
	state   TaskState
	err     error
	message string
}

// NewSingleTaskModel creates a new single-task model.
func NewSingleTaskModel(message string, task Task) SingleTaskModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	return SingleTaskModel{
		task:    task,
		spinner: s,
		state:   TaskPending,
		message: message,
	}
}

// Init initializes the model.
func (m SingleTaskModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.runTask(),
	)
}

// runTask returns a command to run the task.
func (m SingleTaskModel) runTask() tea.Cmd {
	return func() tea.Msg {
		err := m.task.Run()
		return taskDoneMsg{index: 0, err: err}
	}
}

// Update handles messages.
func (m SingleTaskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case spinner.TickMsg:
		if m.state == TaskRunning || m.state == TaskPending {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case taskDoneMsg:
		if msg.err != nil {
			if IsWarnErr(msg.err) {
				m.state = TaskWarning
			} else {
				m.state = TaskFailed
			}
			m.err = msg.err
		} else {
			m.state = TaskSuccess
		}
		return m, tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
			return tea.Quit()
		})
	}

	if m.state == TaskPending {
		m.state = TaskRunning
	}

	return m, nil
}

// View renders the model.
func (m SingleTaskModel) View() string {
	var sb strings.Builder

	var icon string
	var style lipgloss.Style

	switch m.state {
	case TaskPending, TaskRunning:
		icon = m.spinner.View()
		style = StylePrimary
	case TaskSuccess:
		icon = StyleSuccess.Render(SymbolSuccess)
		style = StyleSuccess
	case TaskWarning:
		icon = StyleWarning.Render(SymbolWarning)
		style = StyleWarning
	case TaskFailed:
		icon = StyleError.Render(SymbolError)
		style = StyleError
	}

	sb.WriteString(fmt.Sprintf("%s %s", icon, style.Render(m.message)))

	if m.err != nil && m.state != TaskRunning && m.state != TaskPending {
		sb.WriteString("\n")
		sb.WriteString(StyleMuted.Render(m.err.Error()))
	}

	sb.WriteString("\n")
	return sb.String()
}
