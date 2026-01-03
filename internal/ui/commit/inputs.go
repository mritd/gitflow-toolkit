// Package commit provides the TUI for creating commits.
package commit

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mritd/gitflow-toolkit/v2/internal/config"
	"github.com/mritd/gitflow-toolkit/v2/internal/ui/common"
)

// Spinner frames with colors
var (
	spinnerFrame1 = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("❯")
	spinnerFrame2 = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("❯")
	spinnerFrame3 = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("❯")
)

// Input field styles
var (
	inputsTitleLayout = lipgloss.NewStyle().
				Padding(1, 0, 1, 2)

	inputsTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#333333", Dark: "#DDDDDD"}).
				Background(lipgloss.AdaptiveColor{Light: "#7653FF", Dark: "#7653FF"}).
				Bold(true).
				Padding(0, 1)

	inputsBlockLayout = lipgloss.NewStyle().
				Padding(0, 0, 0, 0)

	inputsCursorStyle = lipgloss.NewStyle().
				Foreground(common.ColorSuccess)

	inputsPromptFocusStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(common.ColorBorder).
				Foreground(common.ColorPrimary).
				Bold(true).
				Padding(0, 0, 0, 1)

	inputsPromptNormalStyle = lipgloss.NewStyle().
				Foreground(common.ColorText).
				Padding(0, 0, 0, 2)

	inputsTextFocusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#FFFDF5"}).
				Bold(true)

	inputsTextNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#FFFDF5"})

	inputsButtonLayout = lipgloss.NewStyle().
				Padding(1, 0, 1, 2)

	inputsButtonFocusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#FFFDF5", Dark: "#FFFDF5"}).
				Background(common.ColorSuccess).
				Padding(0, 1).
				Bold(true)

	inputsButtonNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#626262", Dark: "#DDDDDD"}).
				Background(lipgloss.AdaptiveColor{Light: "#DDDDDD", Dark: "#626262"}).
				Padding(0, 1)

	inputsErrStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#FF6037")).
			Padding(0, 1).
			Bold(true).
			MarginLeft(1)

	inputsHelpStyle = lipgloss.NewStyle().
			Foreground(common.ColorMuted).
			PaddingLeft(2).
			PaddingTop(1)
)

// inputField represents a single input field with validation.
type inputField struct {
	input   textinput.Model
	checker func(s string) error
}

// inputsModel is the model for the inputs screen.
type inputsModel struct {
	focusIndex int
	title      string
	inputs     []inputField
	bodyText   string // store body separately since textinput is single-line
	err        error
	errSpinner spinner.Model
	quitting   bool
	submitted  bool
	wantEditor bool // flag to open editor after exiting TUI
	width      int  // terminal width
	height     int  // terminal height
}

// inputsResult holds the result of the inputs screen.
type inputsResult struct {
	scope   string
	subject string
	body    string
	footer  string
}

func newInputsModel(commitType string) inputsModel {
	m := inputsModel{
		title:  "Commit Type: " + strings.ToUpper(commitType),
		inputs: make([]inputField, 4),
	}

	// Create error spinner with animated frames
	m.errSpinner = spinner.New()
	m.errSpinner.Spinner = spinner.Spinner{
		Frames: []string{
			spinnerFrame1 + "   ",
			spinnerFrame1 + spinnerFrame2 + "  ",
			spinnerFrame1 + spinnerFrame2 + spinnerFrame3 + " ",
			" " + spinnerFrame1 + spinnerFrame2 + spinnerFrame3,
			"  " + spinnerFrame1 + spinnerFrame2,
			"   " + spinnerFrame1,
		},
		FPS: time.Second / 10,
	}

	prompts := []struct {
		prompt      string
		placeholder string
		checker     func(string) error
	}{
		{
			prompt:      "1. SCOPE ",
			placeholder: "Specifying place of the commit change (e.g., api, ui, core)",
			checker: func(s string) error {
				if strings.TrimSpace(s) == "" {
					return errors.New("Scope cannot be empty")
				}
				if strings.ContainsAny(s, "():/\\") {
					return errors.New("Scope cannot contain ():/\\")
				}
				return nil
			},
		},
		{
			prompt:      "2. SUBJECT ",
			placeholder: "A short description, imperative mood, max 72 chars",
			checker: func(s string) error {
				if strings.TrimSpace(s) == "" {
					return errors.New("Subject cannot be empty")
				}
				if len(s) > 72 {
					return errors.New("Subject should be <= 72 chars")
				}
				return nil
			},
		},
		{
			prompt:      "3. BODY ",
			placeholder: "Detailed description (optional, Ctrl+E open editor)",
			checker:     nil,
		},
		{
			prompt:      "4. FOOTER ",
			placeholder: "BREAKING CHANGE: ... or Closes #123 (optional)",
			checker:     nil,
		},
	}

	for i, p := range prompts {
		ti := textinput.New()
		ti.Prompt = p.prompt
		ti.Placeholder = p.placeholder
		ti.CharLimit = 256
		ti.Cursor.Style = inputsCursorStyle
		ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(common.ColorMuted)

		if i == 0 {
			ti.PromptStyle = inputsPromptFocusStyle
			ti.TextStyle = inputsTextFocusStyle
			ti.Focus()
		} else {
			ti.PromptStyle = inputsPromptNormalStyle
			ti.TextStyle = inputsTextNormalStyle
		}

		m.inputs[i] = inputField{
			input:   ti,
			checker: p.checker,
		}
	}

	return m
}

func (m inputsModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.errSpinner.Tick)
}

func (m inputsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Update input widths (subtract prompt width ~12 and padding ~4)
		inputWidth := msg.Width - 16
		if inputWidth < 20 {
			inputWidth = 20
		}
		for i := range m.inputs {
			m.inputs[i].input.Width = inputWidth
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.errSpinner, cmd = m.errSpinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "ctrl+e":
			// Open editor for body field (index 2)
			if m.focusIndex == 2 {
				m.wantEditor = true
				return m, tea.Quit
			}

		case "enter":
			if m.focusIndex == len(m.inputs) {
				// On submit button - validate all and submit
				for _, f := range m.inputs {
					if f.checker != nil {
						if err := f.checker(f.input.Value()); err != nil {
							m.err = err
							return m, nil
						}
					}
				}
				m.submitted = true
				return m, tea.Quit
			}
			// Move to next field
			fallthrough

		case "tab", "down":
			m.focusIndex++
			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			}
			return m, m.updateFocus()

		case "shift+tab", "up":
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}
			return m, m.updateFocus()
		}
	}

	// Update the focused input
	return m, m.updateInputs(msg)
}

func (m inputsModel) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		if i == m.focusIndex {
			cmds[i] = m.inputs[i].input.Focus()
			m.inputs[i].input.PromptStyle = inputsPromptFocusStyle
			m.inputs[i].input.TextStyle = inputsTextFocusStyle
		} else {
			m.inputs[i].input.Blur()
			m.inputs[i].input.PromptStyle = inputsPromptNormalStyle
			m.inputs[i].input.TextStyle = inputsTextNormalStyle
		}
	}
	return tea.Batch(cmds...)
}

func (m inputsModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i].input, cmds[i] = m.inputs[i].input.Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m inputsModel) View() string {
	if m.quitting || m.submitted || m.wantEditor {
		return ""
	}

	var b strings.Builder

	// Calculate available height for inputs
	// Title takes ~3 lines, button ~2 lines, help ~2 lines
	headerLines := 3
	footerLines := 4
	availableHeight := m.height - headerLines - footerLines
	if availableHeight < 1 {
		availableHeight = 1
	}

	// Determine which inputs to show based on focus and available height
	startIdx := 0
	endIdx := len(m.inputs)

	if availableHeight < len(m.inputs) {
		// Need to scroll - keep focused item visible
		if m.focusIndex < len(m.inputs) {
			// Center the focused input if possible
			startIdx = m.focusIndex - availableHeight/2
			if startIdx < 0 {
				startIdx = 0
			}
			endIdx = startIdx + availableHeight
			if endIdx > len(m.inputs) {
				endIdx = len(m.inputs)
				startIdx = endIdx - availableHeight
				if startIdx < 0 {
					startIdx = 0
				}
			}
		}
	}

	// Show scroll indicator if needed
	if startIdx > 0 {
		b.WriteString(inputsHelpStyle.Render("  ↑ more above"))
		b.WriteRune('\n')
	}

	// Input fields
	for i := startIdx; i < endIdx; i++ {
		b.WriteString(m.inputs[i].input.View())
		// Show indicator if body has multi-line content from editor
		if i == 2 && m.bodyText != "" && strings.Contains(m.bodyText, "\n") {
			lineCount := strings.Count(m.bodyText, "\n") + 1
			indicator := lipgloss.NewStyle().Foreground(common.ColorSuccess).Render(
				fmt.Sprintf(" (%d lines)", lineCount))
			b.WriteString(indicator)
		}
		b.WriteRune('\n')
	}

	// Show scroll indicator if needed
	if endIdx < len(m.inputs) {
		b.WriteString(inputsHelpStyle.Render("  ↓ more below"))
		b.WriteRune('\n')
	}

	// Submit button
	button := inputsButtonNormalStyle.Render("Submit")
	if m.focusIndex == len(m.inputs) {
		button = inputsButtonFocusStyle.Render("Submit")
	}

	// Validate and show error with animated spinner
	m.err = nil
	for _, f := range m.inputs {
		if f.checker != nil {
			if err := f.checker(f.input.Value()); err != nil {
				m.err = err
				break
			}
		}
	}

	if m.err != nil {
		button += " " + m.errSpinner.View() + inputsErrStyle.Render(m.err.Error())
	}

	b.WriteString(inputsButtonLayout.Render(button))

	// Help text
	b.WriteString(inputsHelpStyle.Render("↑/↓ navigate • enter next/submit • ctrl+c quit • ctrl+e editor (body)"))

	title := inputsTitleLayout.Render(inputsTitleStyle.Render(m.title))
	inputs := inputsBlockLayout.Render(b.String())

	return lipgloss.JoinVertical(lipgloss.Left, title, inputs)
}

func (m inputsModel) result() inputsResult {
	// Use bodyText if it has content (from editor), otherwise use input field
	body := m.bodyText
	if body == "" {
		body = m.inputs[2].input.Value()
	}
	return inputsResult{
		scope:   strings.TrimSpace(m.inputs[0].input.Value()),
		subject: strings.TrimSpace(m.inputs[1].input.Value()),
		body:    strings.TrimSpace(body),
		footer:  strings.TrimSpace(m.inputs[3].input.Value()),
	}
}

// runInputs runs the inputs screen and returns the result.
func runInputs(commitType string) (inputsResult, error) {
	m := newInputsModel(commitType)

	for {
		p := tea.NewProgram(m)
		finalModel, err := p.Run()
		if err != nil {
			return inputsResult{}, err
		}

		m = finalModel.(inputsModel)

		if m.quitting {
			return inputsResult{}, errUserAborted
		}

		if m.submitted {
			return m.result(), nil
		}

		if m.wantEditor {
			// Run editor outside of TUI
			// Use existing bodyText or input field value as initial content
			initialContent := m.bodyText
			if initialContent == "" {
				initialContent = m.inputs[2].input.Value()
			}
			content, err := runExternalEditor(initialContent)
			if err == nil {
				m.bodyText = content
				// Show truncated preview in input field
				preview := strings.ReplaceAll(content, "\n", " ")
				if len(preview) > 50 {
					preview = preview[:50] + "..."
				}
				m.inputs[2].input.SetValue(preview)
			}
			m.wantEditor = false
			// Continue the loop to restart TUI
		}
	}
}

// runExternalEditor opens the default editor and returns its content.
func runExternalEditor(currentContent string) (string, error) {
	// Get editor from environment
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		for _, e := range []string{"vim", "vi", "nano"} {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}
	if editor == "" {
		return "", fmt.Errorf("no editor found, please set EDITOR environment variable")
	}

	// Create temp file
	f, err := os.CreateTemp("", config.TempFilePrefix+"-body-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tempName := f.Name()
	defer func() { _ = os.Remove(tempName) }()

	// Write template
	template := "# Enter commit body (lines starting with # will be removed)\n# Explain WHAT and WHY, not HOW\n\n"
	if currentContent != "" {
		template += currentContent
	}
	if _, err := f.WriteString(template); err != nil {
		_ = f.Close()
		return "", fmt.Errorf("failed to write template: %w", err)
	}
	_ = f.Close()

	// Run editor
	cmd := exec.Command(editor, tempName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor exited with error: %w", err)
	}

	// Read result
	content, err := os.ReadFile(tempName)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Remove comment lines
	lines := strings.Split(string(content), "\n")
	var result []string
	for _, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "#") {
			result = append(result, line)
		}
	}

	return strings.TrimSpace(strings.Join(result, "\n")), nil
}
