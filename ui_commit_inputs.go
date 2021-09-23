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
	inputsTitleBarStyle = lipgloss.NewStyle().
				Padding(1, 0, 1, 2)

	inputsTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#7653FF")).
				Bold(true).
				Padding(0, 1)

	inputsBlockStyle = lipgloss.NewStyle().
				Padding(0, 0, 1, 0)

	inputsTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#FFFDF5"}).
			Bold(true)

	inputsTextNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#313131", Dark: "#DDDDDD"})

	inputsCursorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#25A065"))

	inputsPromptNormalStyle = lipgloss.NewStyle().
				Padding(0, 0, 0, 2)

	inputsPromptStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}).
				Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
				Bold(true).
				Padding(0, 0, 0, 1)

	inputsButtonBlockStyle = lipgloss.NewStyle().
				Padding(2, 0, 1, 2)

	inputsButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#25A065")).
				Padding(0, 1, 0, 1).
				Bold(true)

	inputsButtonNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#626262", Dark: "#DDDDDD"}).
				Background(lipgloss.AdaptiveColor{Light: "#DDDDDD", Dark: "#626262"}).
				Padding(0, 1, 0, 1).
				Bold(true)

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
}

func (m inputsModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" {
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

			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].input.Focus()
					m.inputs[i].input.PromptStyle = inputsPromptStyle
					m.inputs[i].input.TextStyle = inputsTextStyle
					continue
				}
				// Remove focused state
				m.inputs[i].input.Blur()
				m.inputs[i].input.PromptStyle = inputsPromptNormalStyle
				m.inputs[i].input.TextStyle = inputsTextNormalStyle
			}

			if m.focusIndex < len(m.inputs) && m.inputs[m.focusIndex].checker != nil {
				m.err = m.inputs[m.focusIndex].checker(m.inputs[m.focusIndex].input.Value())
				cmds = append(cmds, spinner.Tick)
			} else {
				m.err = nil
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
	cmd := m.updateInputs(msg)

	return m, cmd
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
		button = inputsButtonStyle.Render("➜ Submit")
	}

	if m.err != nil {
		b.WriteString(inputsButtonBlockStyle.Render(button + inputsErrLayout.Render(m.errSpinner.View()+" "+inputsErrStyle.Render(m.err.Error()))))
	} else {
		b.WriteString(inputsButtonBlockStyle.Render(button))
	}

	title := inputsTitleBarStyle.Render(inputsTitleStyle.Render(m.title))
	inputs := inputsBlockStyle.Render(b.String())

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
			iwc.input.PromptStyle = inputsPromptStyle
			iwc.input.TextStyle = inputsTextStyle
			iwc.input.Focus()
			iwc.checker = func(s string) error {
				if strings.TrimSpace(s) == "" {
					return errors.New("Scope cannot be empty")
				}
				return nil
			}
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
