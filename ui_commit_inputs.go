package main

import (
	"fmt"
	"strings"

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
)

type inputsModel struct {
	focusIndex int
	title      string
	inputs     []textinput.Model
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

			if s == "enter" && m.focusIndex == len(m.inputs) {
				return m, func() tea.Msg { return done{nextView: COMMIT} }
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
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = inputsPromptStyle
					m.inputs[i].TextStyle = inputsTextStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = inputsPromptNormalStyle
				m.inputs[i].TextStyle = inputsTextNormalStyle
			}

			return m, tea.Batch(cmds...)
		}
	case string:
		m.title = "✔ Commit Type: " + msg
		return m, nil
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *inputsModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m inputsModel) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := inputsButtonNormalStyle.Render("➜ Submit")
	if m.focusIndex == len(m.inputs) {
		button = inputsButtonStyle.Render("➜ Submit")
	}
	_, _ = fmt.Fprint(&b, inputsButtonBlockStyle.Render(button))

	title := inputsTitleBarStyle.Render(inputsTitleStyle.Render(m.title))
	inputs := inputsBlockStyle.Render(b.String())

	return lipgloss.JoinVertical(lipgloss.Left, title, inputs)
}

func newInputsModel() inputsModel {
	m := inputsModel{
		inputs: make([]textinput.Model, 4),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.NewModel()
		t.CursorStyle = inputsCursorStyle
		t.CharLimit = 128

		switch i {
		case 0:
			t.Prompt = "1. SCOPE "
			t.Placeholder = "Specifying place of the commit change."
			t.PromptStyle = inputsPromptStyle
			t.TextStyle = inputsTextStyle
			t.Focus()
		case 1:
			t.Prompt = "2. SUBJECT "
			t.PromptStyle = inputsPromptNormalStyle
			t.Placeholder = "A very short description of the change."
		case 2:
			t.Prompt = "3. BODY "
			t.PromptStyle = inputsPromptNormalStyle
			t.Placeholder = "Motivation and contrasts for the change."
		case 3:
			t.Prompt = "4. FOOTER "
			t.PromptStyle = inputsPromptNormalStyle
			t.Placeholder = "Description of the change, justification and migration notes."
		}

		m.inputs[i] = t
	}

	return m
}
