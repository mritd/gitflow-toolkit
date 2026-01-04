// Package commit provides the TUI for creating commits.
package commit

import (
	"errors"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mritd/gitflow-toolkit/v3/internal/git"
	"github.com/mritd/gitflow-toolkit/v3/internal/ui/common"
)

// Result represents the result of the commit flow.
type Result struct {
	Cancelled    bool
	Err          error
	Message      git.CommitMessage
	LuckySkipped bool   // true if lucky commit was skipped (Ctrl+C)
	LuckyFailed  error  // error if lucky commit failed
	Hash         string // final commit hash (may be lucky hash)
}

// Run runs the interactive commit flow.
// The luckyPrefix parameter is the validated lucky commit prefix (empty if not enabled).
func Run(luckyPrefix string) Result {
	var result Result

	// Step 1: Select commit type or AI generate
	choice, err := runSelector()
	if err != nil {
		if errors.Is(err, errUserAborted) {
			result.Cancelled = true
			return result
		}
		result.Err = err
		return result
	}

	// Check if user selected AI generate
	if choice == aiGenerateChoice {
		return runAIFlow(luckyPrefix)
	}

	// Original flow for manual commit type selection
	return runManualFlow(choice, luckyPrefix)
}

// runManualFlow runs the manual commit flow with the given commit type.
func runManualFlow(commitType, luckyPrefix string) Result {
	var result Result

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

	// Perform commit and handle lucky commit
	return performCommit(result.Message, luckyPrefix)
}

// runAIFlow runs the AI-powered commit flow.
func runAIFlow(luckyPrefix string) Result {
	var currentMessage string

	for {
		// Run AI generation (only if we don't have a message yet, or user requested retry)
		if currentMessage == "" {
			aiResult := runAIGenerate()
			if aiResult.Cancelled {
				return Result{Cancelled: true}
			}
			if aiResult.Err != nil {
				return Result{Err: aiResult.Err}
			}
			currentMessage = aiResult.Message
		}

		// Show preview
		previewResult := runAIPreview(currentMessage)

		switch previewResult.Action {
		case "commit":
			// Parse the AI message and commit
			msg := parseAIMessage(previewResult.Message)
			return performCommit(msg, luckyPrefix)

		case "edit":
			// Open editor
			edited, err := runExternalEditor(previewResult.Message)
			if err != nil {
				return Result{Err: err}
			}
			currentMessage = edited
			// Loop continues to show preview with edited message

		case "retry":
			// Clear message and regenerate
			currentMessage = ""
			// Loop continues

		case "cancel":
			return Result{Cancelled: true}
		}
	}
}

// parseAIMessage parses the AI-generated message into a CommitMessage struct.
func parseAIMessage(message string) git.CommitMessage {
	lines := splitLines(message)
	if len(lines) == 0 {
		return git.CommitMessage{}
	}

	// Parse header: type(scope): subject
	header := lines[0]
	msgType, scope, subject := parseHeader(header)

	// Rest is body
	var body string
	if len(lines) > 1 {
		// Skip empty line after header if present
		startIdx := 1
		if startIdx < len(lines) && lines[startIdx] == "" {
			startIdx++
		}
		if startIdx < len(lines) {
			bodyLines := lines[startIdx:]
			body = joinLines(bodyLines)
		}
	}

	// Create SOB
	sob := git.CreateSOB()

	return git.CommitMessage{
		Type:    msgType,
		Scope:   scope,
		Subject: subject,
		Body:    body,
		SOB:     sob,
	}
}

// parseHeader parses "type(scope): subject" format.
func parseHeader(header string) (msgType, scope, subject string) {
	// Default values
	msgType = "feat"
	scope = "general"
	subject = header

	// Try to parse "type(scope): subject"
	colonIdx := -1
	for i, c := range header {
		if c == ':' {
			colonIdx = i
			break
		}
	}

	if colonIdx > 0 {
		prefix := header[:colonIdx]
		subject = trimLeft(header[colonIdx+1:])

		// Parse type and scope from prefix
		parenStart := -1
		parenEnd := -1
		for i, c := range prefix {
			if c == '(' {
				parenStart = i
			} else if c == ')' {
				parenEnd = i
			}
		}

		if parenStart > 0 && parenEnd > parenStart {
			msgType = prefix[:parenStart]
			scope = prefix[parenStart+1 : parenEnd]
		} else {
			msgType = prefix
		}
	}

	return msgType, scope, subject
}

// splitLines splits a string into lines.
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// joinLines joins lines with newline.
func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	result := lines[0]
	for i := 1; i < len(lines); i++ {
		result += "\n" + lines[i]
	}
	return result
}

// trimLeft trims leading whitespace.
func trimLeft(s string) string {
	for i, c := range s {
		if c != ' ' && c != '\t' {
			return s[i:]
		}
	}
	return ""
}

// performCommit commits the message and handles lucky commit.
func performCommit(msg git.CommitMessage, luckyPrefix string) Result {
	var result Result
	result.Message = msg

	// Perform commit
	if err := git.Commit(msg); err != nil {
		result.Err = err
		return result
	}

	// Run lucky commit if prefix is set
	if luckyPrefix != "" {
		cmd := git.LuckyCommitCmd(luckyPrefix)
		luckyResult := common.RunLuckyCommit(luckyPrefix, cmd, git.GetHeadHash)

		if luckyResult.Cancelled {
			result.LuckySkipped = true
		} else if luckyResult.Err != nil {
			result.LuckyFailed = luckyResult.Err
		}
		result.Hash = luckyResult.Hash
	}

	// Get hash if not already set
	if result.Hash == "" {
		result.Hash, _ = git.GetHeadHash()
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
