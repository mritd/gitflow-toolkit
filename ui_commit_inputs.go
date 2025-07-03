package main

import (
	"errors"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	inputsTitleLayout = lipgloss.NewStyle().
				Padding(1, 0, 1, 2)

	inputsTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#333333", Dark: "#DDDDDD"}).
				Background(lipgloss.AdaptiveColor{Light: "#7653FF", Dark: "#7653FF"}).
				Bold(true).
				Padding(0, 1, 0, 1)

	inputsBlockLayout = lipgloss.NewStyle().
				Padding(0, 0, 1, 0)

	inputsCursorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#25A065", Dark: "#25A065"})

	inputsPromptFocusStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(lipgloss.AdaptiveColor{Light: "#9F72FF", Dark: "#AD58B4"}).
				Foreground(lipgloss.AdaptiveColor{Light: "#9A4AFF", Dark: "#EE6FF8"}).
				Bold(true).
				Padding(0, 0, 0, 1)

	inputsPromptNormalStyle = lipgloss.NewStyle().
				Padding(0, 0, 0, 2)

	inputsTextFocusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#FFFDF5"}).
				Bold(true)

	inputsTextNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#FFFDF5"})

	inputsButtonLayout = lipgloss.NewStyle().
				Padding(2, 0, 1, 2)

	inputsButtonFocusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#FFFDF5", Dark: "#FFFDF5"}).
				Background(lipgloss.AdaptiveColor{Light: "#25A065", Dark: "#25A065"}).
				Padding(0, 1, 0, 1).
				Bold(true)

	inputsButtonNormalStyle = inputsButtonFocusStyle.Copy().
				Foreground(lipgloss.AdaptiveColor{Light: "#626262", Dark: "#DDDDDD"}).
				Background(lipgloss.AdaptiveColor{Light: "#DDDDDD", Dark: "#626262"})

	inputsErrLayout = lipgloss.NewStyle().
			Padding(0, 0, 0, 1)

	inputsErrStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#FF6037")).
			Padding(0, 1, 0, 1).
			Bold(true)

	spinnerMetaFrame1 = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("❯")
	spinnerMetaFrame2 = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("❯")
	spinnerMetaFrame3 = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("❯")
)

type inputWithCheck struct {
	input   textinput.Model
	checker func(s string) error
}

type inputsModel struct {
	focusIndex int
	title      string
	inputs     []inputWithCheck
	err        error
	errSpinner spinner.Model
	editMode   bool
}

func (m inputsModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, spinner.Tick)
}

func (m inputsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.editMode {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		var renderCursor bool
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if m.focusIndex == len(m.inputs) {
				for _, iwc := range m.inputs {
					if iwc.checker != nil {
						m.err = iwc.checker(iwc.input.Value())
						if m.err != nil {
							return m, spinner.Tick
						}
					}
				}
				return m, func() tea.Msg { return done{nextView: COMMIT} }
			}
			fallthrough
		case "tab", "down":
			m.focusIndex++
			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			}
			renderCursor = true
		case "shift+tab", "up":
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}
			renderCursor = true
		}

		if renderCursor {
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].input.Focus()
					m.inputs[i].input.PromptStyle = inputsPromptFocusStyle
					m.inputs[i].input.TextStyle = inputsTextFocusStyle
					continue
				}
				// Remove focused state
				m.inputs[i].input.Blur()
				m.inputs[i].input.PromptStyle = inputsPromptNormalStyle
				m.inputs[i].input.TextStyle = inputsTextNormalStyle
			}

			return m, tea.Batch(cmds...)
		}

	case string:
		m.title = "✔ Commit Type: " + msg
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.errSpinner, cmd = m.errSpinner.Update(msg)
		return m, cmd
	}

	// Handle character input and blinking
	return m, m.updateInputs(msg)
}

func (m *inputsModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, len(m.inputs)+1)

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i].input, cmds[i] = m.inputs[i].input.Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m inputsModel) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].input.View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := inputsButtonNormalStyle.Render("➜ Submit")
	if m.focusIndex == len(m.inputs) {
		button = inputsButtonFocusStyle.Render("➜ Submit")
	}

	// check input value
	for _, iwc := range m.inputs {
		if iwc.checker != nil {
			m.err = iwc.checker(iwc.input.Value())
			if m.err != nil {
				button += inputsErrLayout.Render(m.errSpinner.View() + " " + inputsErrStyle.Render(m.err.Error()))
				break
			}
		}
	}

	b.WriteString(inputsButtonLayout.Render(button))

	title := inputsTitleLayout.Render(inputsTitleStyle.Render(m.title))
	inputs := inputsBlockLayout.Render(b.String())

	return lipgloss.JoinVertical(lipgloss.Left, title, inputs)
}

func newInputsModel() inputsModel {
	m := inputsModel{
		inputs: make([]inputWithCheck, 4),
	}

	for i := range m.inputs {
		var iwc inputWithCheck

		iwc.input = textinput.NewModel()
		iwc.input.CursorStyle = inputsCursorStyle
		iwc.input.CharLimit = 128

		switch i {
		case 0:
			iwc.input.Prompt = "1. SCOPE "
			iwc.input.Placeholder = "Specifying place of the commit change."
			iwc.input.PromptStyle = inputsPromptFocusStyle
			iwc.input.TextStyle = inputsTextFocusStyle
			iwc.input.Focus()
		case 1:
			iwc.input.Prompt = "2. SUBJECT "
			iwc.input.PromptStyle = inputsPromptNormalStyle
			iwc.input.Placeholder = "A very short description of the change."
			iwc.checker = func(s string) error {
				if strings.TrimSpace(s) == "" {
					return errors.New("Subject cannot be empty")
				}
				return nil
			}
		case 2:
			iwc.input.Prompt = "3. BODY "
			iwc.input.PromptStyle = inputsPromptNormalStyle
			iwc.input.Placeholder = "Motivation and contrasts for the change."
		case 3:
			iwc.input.Prompt = "4. FOOTER "
			iwc.input.PromptStyle = inputsPromptNormalStyle
			iwc.input.Placeholder = "Description of the change, justification and migration notes."
		}

		m.inputs[i] = iwc
	}

	m.errSpinner = spinner.NewModel()
	m.errSpinner.Spinner = spinner.Spinner{
		Frames: []string{
			// "❯   "
			spinnerMetaFrame1 + "   ",
			// "❯❯  "
			spinnerMetaFrame1 + spinnerMetaFrame2 + "  ",
			// "❯❯❯ "
			spinnerMetaFrame1 + spinnerMetaFrame2 + spinnerMetaFrame3 + " ",
			// " ❯❯❯"
			" " + spinnerMetaFrame1 + spinnerMetaFrame2 + spinnerMetaFrame3,
			// "  ❯❯"
			"  " + spinnerMetaFrame1 + spinnerMetaFrame2,
			// "   ❯"
			"   " + spinnerMetaFrame1,
		},
		FPS: time.Second / 10,
	}

	return m
}
