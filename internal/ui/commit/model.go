// Package commit provides the TUI for creating commits.
package commit

import (
	"errors"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mritd/gitflow-toolkit/v2/internal/git"
	"github.com/mritd/gitflow-toolkit/v2/internal/ui/common"
)

// Result represents the result of the commit flow.
type Result struct {
	Cancelled bool
	Err       error
	Message   git.CommitMessage
}

// Run runs the interactive commit flow.
// Returns the result of the commit operation.
func Run() Result {
	var result Result

	// Step 1: Select commit type
	commitType, err := runSelector()
	if err != nil {
		if errors.Is(err, errUserAborted) {
			result.Cancelled = true
			return result
		}
		result.Err = err
		return result
	}

	// Step 2: Input all fields (scope, subject, body, footer)
	inputs, err := runInputs(commitType)
	if err != nil {
		if errors.Is(err, errUserAborted) {
			result.Cancelled = true
			return result
		}
		result.Err = err
		return result
	}

	// Create SOB
	sob := git.CreateSOB()

	// If body is empty, use subject as body
	body := inputs.body
	if body == "" {
		body = inputs.subject
	}

	// Build commit message
	result.Message = git.CommitMessage{
		Type:    commitType,
		Scope:   inputs.scope,
		Subject: inputs.subject,
		Body:    body,
		Footer:  inputs.footer,
		SOB:     sob,
	}

	// Step 3: Confirm and commit
	confirmed, err := confirmCommit(result.Message)
	if err != nil {
		if errors.Is(err, errUserAborted) {
			result.Cancelled = true
			return result
		}
		result.Err = err
		return result
	}

	if !confirmed {
		result.Cancelled = true
		return result
	}

	// Perform commit
	if err := git.Commit(result.Message); err != nil {
		result.Err = err
		return result
	}

	return result
}

// confirmCommit shows a preview and asks for confirmation using a TUI.
func confirmCommit(msg git.CommitMessage) (bool, error) {
	m := newConfirmModel(msg)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return false, err
	}

	result := finalModel.(confirmModel)
	if result.cancelled {
		return false, errUserAborted
	}

	return result.confirmed, nil
}

// confirmModel is the bubbletea model for commit confirmation.
type confirmModel struct {
	msg       git.CommitMessage
	confirmed bool
	cancelled bool
	selected  int // 0 = Commit, 1 = Cancel
}

func newConfirmModel(msg git.CommitMessage) confirmModel {
	return confirmModel{msg: msg, selected: 0}
}

func (m confirmModel) Init() tea.Cmd {
	return nil
}

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "n":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			if m.selected == 0 {
				m.confirmed = true
			} else {
				m.cancelled = true
			}
			return m, tea.Quit
		case "y":
			m.confirmed = true
			return m, tea.Quit
		case "left", "h":
			m.selected = 0
		case "right", "l":
			m.selected = 1
		case "tab":
			m.selected = (m.selected + 1) % 2
		}
	}
	return m, nil
}

func (m confirmModel) View() string {
	if m.confirmed || m.cancelled {
		return ""
	}

	// Use same layout structure as inputs.go
	titleLayout := lipgloss.NewStyle().Padding(1, 0, 1, 2)
	titleStyle := lipgloss.NewStyle().
		Foreground(common.ColorTitleFg).
		Background(common.ColorTitleBg).
		Bold(true).
		Padding(0, 1)

	// Content with left border (same as Result)
	contentLayout := lipgloss.NewStyle().PaddingLeft(2)
	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(common.ColorSuccess).
		PaddingLeft(1)

	// Buttons without border
	buttonLayout := lipgloss.NewStyle().PaddingLeft(2).PaddingTop(1)

	helpStyle := lipgloss.NewStyle().
		Foreground(common.ColorMuted).
		PaddingLeft(2).
		PaddingTop(1)

	// Commit message preview with border
	preview := contentLayout.Render(contentStyle.Render(renderPreview(m.msg)))

	// Buttons
	buttons := buttonLayout.Render(renderButtons(m.selected))

	title := titleLayout.Render(titleStyle.Render("Commit Preview"))
	help := helpStyle.Render("y/enter confirm • n/esc cancel • ←/→ select")

	return lipgloss.JoinVertical(lipgloss.Left, title, preview, buttons, help) + "\n"
}

// renderPreview renders the commit message preview using the same colors as result.
func renderPreview(msg git.CommitMessage) string {
	return common.FormatCommitMessage(common.CommitMessageContent{
		Type:    msg.Type,
		Scope:   msg.Scope,
		Subject: msg.Subject,
		Body:    msg.Body,
		Footer:  msg.Footer,
		SOB:     msg.SOB,
	})
}

// renderButtons renders the confirm/cancel buttons.
func renderButtons(selected int) string {
	activeStyle := lipgloss.NewStyle().
		Foreground(common.ColorTitleFg).
		Background(common.ColorSuccess).
		Bold(true).
		Padding(0, 2)

	inactiveStyle := lipgloss.NewStyle().
		Foreground(common.ColorMuted).
		Background(lipgloss.Color("#3a3a3a")).
		Padding(0, 2)

	cancelActiveStyle := lipgloss.NewStyle().
		Foreground(common.ColorTitleFg).
		Background(common.ColorError).
		Bold(true).
		Padding(0, 2)

	var commitBtn, cancelBtn string
	if selected == 0 {
		commitBtn = activeStyle.Render("  Commit  ")
		cancelBtn = inactiveStyle.Render("  Cancel  ")
	} else {
		commitBtn = inactiveStyle.Render("  Commit  ")
		cancelBtn = cancelActiveStyle.Render("  Cancel  ")
	}

	return commitBtn + "  " + cancelBtn
}
