// Package branch provides the TUI for creating branches.
package branch

import (
	"fmt"
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
	StateCreating State = iota
	StateSuccess
	StateFailed
)

// Model is the branch creation UI model.
type Model struct {
	branchType string
	branchName string
	fullName   string
	state      State
	spinner    spinner.Model
	err        error
	result     string
}

// NewModel creates a new branch model.
func NewModel(branchType, branchName string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(common.ColorPrimary)

	return Model{
		branchType: branchType,
		branchName: branchName,
		fullName:   fmt.Sprintf("%s/%s", branchType, branchName),
		state:      StateCreating,
		spinner:    s,
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.createBranch(),
	)
}

// branchDoneMsg is sent when branch creation is complete.
type branchDoneMsg struct {
	result string
	err    error
}

// createBranch creates the branch.
func (m Model) createBranch() tea.Cmd {
	return func() tea.Msg {
		result, err := git.CreateTypedBranch(m.branchType, m.branchName)
		return branchDoneMsg{result: result, err: err}
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
		if m.state == StateCreating {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case branchDoneMsg:
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

	branchStyle := common.StyleCommitType

	switch m.state {
	case StateCreating:
		sb.WriteString(m.spinner.View())
		sb.WriteString(" Creating branch ")
		sb.WriteString(branchStyle.Render(m.fullName))
		sb.WriteString("...")

	case StateSuccess:
		sb.WriteString(common.StyleSuccess.Render(common.SymbolSuccess))
		sb.WriteString(" Branch ")
		sb.WriteString(branchStyle.Render(m.fullName))
		sb.WriteString(" created successfully!")
		if m.result != "" {
			sb.WriteString("\n")
			sb.WriteString(common.StyleMuted.Render(m.result))
		}

	case StateFailed:
		sb.WriteString(common.StyleError.Render(common.SymbolError))
		sb.WriteString(" Failed to create branch ")
		sb.WriteString(branchStyle.Render(m.fullName))
		if m.err != nil {
			sb.WriteString("\n")
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

// IsSuccess returns true if branch was created successfully.
func (m Model) IsSuccess() bool {
	return m.state == StateSuccess
}
