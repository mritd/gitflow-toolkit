// Package commit provides the TUI for creating commits.
package commit

import (
	"strings"

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
		if err == errUserAborted {
			result.Cancelled = true
			return result
		}
		result.Err = err
		return result
	}

	// Step 2: Input all fields (scope, subject, body, footer)
	inputs, err := runInputs(commitType)
	if err != nil {
		if err == errUserAborted {
			result.Cancelled = true
			return result
		}
		result.Err = err
		return result
	}

	// Create SOB
	sob := git.CreateSOB()

	// Build commit message
	result.Message = git.CommitMessage{
		Type:    commitType,
		Scope:   inputs.scope,
		Subject: inputs.subject,
		Body:    inputs.body,
		Footer:  inputs.footer,
		SOB:     sob,
	}

	// Step 3: Confirm and commit
	confirmed, err := confirmCommit(result.Message)
	if err != nil {
		if err == errUserAborted {
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

	contentLayout := lipgloss.NewStyle().PaddingLeft(2)
	helpStyle := lipgloss.NewStyle().
		Foreground(common.ColorMuted).
		PaddingLeft(2).
		PaddingTop(1)

	var content strings.Builder

	// Commit message preview
	content.WriteString(renderPreview(m.msg))
	content.WriteString("\n\n")

	// Buttons
	content.WriteString(renderButtons(m.selected))

	title := titleLayout.Render(titleStyle.Render("Commit Preview"))
	body := contentLayout.Render(content.String())
	help := helpStyle.Render("y/enter confirm • n/esc cancel • ←/→ select")

	return lipgloss.JoinVertical(lipgloss.Left, title, body, help) + "\n"
}

// renderPreview renders the commit message preview.
func renderPreview(msg git.CommitMessage) string {
	var sb strings.Builder

	// Header line: type(scope): subject
	typeStyle := lipgloss.NewStyle().Foreground(common.ColorPrimary).Bold(true)
	scopeStyle := lipgloss.NewStyle().Foreground(common.ColorPrimary)
	subjectStyle := lipgloss.NewStyle().Bold(true)

	sb.WriteString(typeStyle.Render(msg.Type))
	sb.WriteString(scopeStyle.Render("(" + msg.Scope + ")"))
	sb.WriteString(": ")
	sb.WriteString(subjectStyle.Render(msg.Subject))

	// Body
	if msg.Body != "" {
		sb.WriteString("\n\n")
		bodyStyle := lipgloss.NewStyle().Foreground(common.ColorText)
		sb.WriteString(bodyStyle.Render(msg.Body))
	}

	// Footer
	if msg.Footer != "" {
		sb.WriteString("\n\n")
		footerStyle := lipgloss.NewStyle().Foreground(common.ColorMuted)
		sb.WriteString(footerStyle.Render(msg.Footer))
	}

	// Signed-off-by
	if msg.SOB != "" {
		sb.WriteString("\n\n")
		sobStyle := lipgloss.NewStyle().Foreground(common.ColorDimmed)
		sb.WriteString(sobStyle.Render(msg.SOB))
	}

	return sb.String()
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
