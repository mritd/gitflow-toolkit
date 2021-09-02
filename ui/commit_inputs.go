package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

var (
	inputsTitleBarStyle = lipgloss.NewStyle().
				Padding(1, 0, 1, 2)

	inputsTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#7653FF")).
				Padding(0, 1)

	inputsBlockStyle = lipgloss.NewStyle().
				Padding(0, 0, 1, 2)

	inputsNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#313131", Dark: "#DDDDDD"})

	inputsFocusedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#FFFDF5"})

	inputsCursorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#25A065"))

	inputsPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F793FF"))

	inputsFocusedButtonStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#FFFDF5")).
					Background(lipgloss.Color("#25A065")).
					Padding(0, 1, 0, 1)

	inputsBlurredButtonStyle = lipgloss.NewStyle().
					Foreground(lipgloss.AdaptiveColor{Light: "#626262", Dark: "#DDDDDD"}).
					Background(lipgloss.AdaptiveColor{Light: "#DDDDDD", Dark: "#626262"}).
					Padding(0, 1, 0, 1)
)

type inputsModel struct {
	scope   string
	subject string
	body    string
	footer  string

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
				m.scope = m.inputs[0].Value()
				m.subject = m.inputs[1].Value()
				m.body = m.inputs[2].Value()
				m.footer = m.inputs[3].Value()
				return m, inputsDoneDone
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
					m.inputs[i].TextStyle = inputsFocusedStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = inputsNormalStyle
				m.inputs[i].TextStyle = inputsNormalStyle
			}

			return m, tea.Batch(cmds...)
		}
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

	button := inputsBlurredButtonStyle.Render("Submit")
	if m.focusIndex == len(m.inputs) {
		button = inputsFocusedButtonStyle.Render("Submit")
	}
	_, _ = fmt.Fprintf(&b, "\n\n%s\n\n", button)

	//var pl string
	//if m.focusIndex < len(m.inputs) {
	//	pl = m.inputs[m.focusIndex].Placeholder
	//} else {
	//	pl = "Message"
	//}
	title := inputsTitleBarStyle.Render(inputsTitleStyle.Render("Input Other Message"))
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
			t.Prompt = "1. SCOPE: "
			t.Placeholder = "Specifying place of the commit change."
			t.PromptStyle = inputsPromptStyle
			t.TextStyle = inputsFocusedStyle
			t.Focus()
		case 1:
			t.Prompt = "2. SUBJECT: "
			t.Placeholder = "A very short description of the change."
		case 2:
			t.Prompt = "3. BODY: "
			t.Placeholder = "Motivation and contrasts for the change."
		case 3:
			t.Prompt = "4. FOOTER: "
			t.Placeholder = "Description of the change, justification and migration notes."
		}

		m.inputs[i] = t
	}

	return m
}
