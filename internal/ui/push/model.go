// Package push provides the TUI for pushing branches.
package push

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mritd/gitflow-toolkit/v2/internal/git"
	"github.com/mritd/gitflow-toolkit/v2/internal/ui/common"
)

// State represents the current state.
type State int

const (
	StatePushing State = iota
	StateSuccess
	StateFailed
)

// Model is the push UI model.
type Model struct {
	state   State
	spinner spinner.Model
	err     error
	result  string
}

// NewModel creates a new push model.
func NewModel() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(common.ColorPrimary)

	return Model{
		state:   StatePushing,
		spinner: s,
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.doPush(),
	)
}

// pushDoneMsg is sent when push is complete.
type pushDoneMsg struct {
	result string
	err    error
}

// doPush performs the push.
func (m Model) doPush() tea.Cmd {
	return func() tea.Msg {
		result, err := git.Push()
		return pushDoneMsg{result: result, err: err}
	}
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc", "enter":
			return m, tea.Quit
		}

	case spinner.TickMsg:
		if m.state == StatePushing {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case pushDoneMsg:
		if msg.err != nil {
			m.state = StateFailed
			m.err = msg.err
		} else {
			m.state = StateSuccess
			m.result = msg.result
		}
		// Auto-quit after a short delay
		return m, tea.Tick(time.Millisecond*800, func(t time.Time) tea.Msg {
			return tea.Quit()
		})
	}

	return m, nil
}

// View renders the model.
func (m Model) View() string {
	var sb strings.Builder

	switch m.state {
	case StatePushing:
		sb.WriteString(m.spinner.View())
		sb.WriteString(" Pushing to origin...")

	case StateSuccess:
		sb.WriteString(common.StyleSuccess.Render(common.SymbolSuccess))
		sb.WriteString(" Push completed!")
		if m.result != "" {
			sb.WriteString("\n\n")
			sb.WriteString(common.StyleMuted.Render(m.result))
		}

	case StateFailed:
		sb.WriteString(common.StyleError.Render(common.SymbolError))
		sb.WriteString(" Push failed")
		if m.err != nil {
			sb.WriteString("\n\n")
			sb.WriteString(common.StyleMuted.Render(m.err.Error()))
		}
	}

	sb.WriteString("\n")
	return sb.String()
}

// Error returns any error that occurred.
func (m Model) Error() error {
	return m.err
}

// IsSuccess returns true if push was successful.
func (m Model) IsSuccess() bool {
	return m.state == StateSuccess
}
