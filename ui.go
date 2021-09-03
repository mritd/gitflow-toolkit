package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func init() {
	runewidth.EastAsianWidth = false
	runewidth.DefaultCondition.EastAsianWidth = false
}

var (
	layOutStyle = lipgloss.NewStyle().
			Padding(1, 1, 2, 1)

	doneTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 0, 1, 22)

	doneMsgStyle = lipgloss.NewStyle().
			Bold(true).
			Width(64)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#37B9FF")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#37B9FF")).
			Padding(1, 3, 1, 3)

	failedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF62DA")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF62DA")).
			Padding(1, 3, 1, 3)
)

type selectDoneMsg struct{}

func selectDone() tea.Msg {
	return selectDoneMsg{}
}

type inputsDoneMsg struct{}

func inputsDoneDone() tea.Msg {
	return inputsDoneMsg{}
}

type model struct {
	cType    string
	cScope   string
	cSubject string
	cBody    string
	cFooter  string

	cSelect bool
	cInputs bool
	cCommit bool

	err           error
	selectorModel selectorModel
	inputsModel   inputsModel
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (mod tea.Model, cmd tea.Cmd) {
	switch msg.(type) {
	case selectDoneMsg:
		m.cType = m.selectorModel.choice
		m.cSelect = true
	case inputsDoneMsg:
		m.cScope = m.inputsModel.scope
		m.cSubject = m.inputsModel.subject
		m.cBody = m.inputsModel.body
		m.cFooter = m.inputsModel.footer
		m.cInputs = true
	}

	if !m.cSelect {
		mod, cmd = m.selectorModel.Update(msg)
		m.selectorModel = mod.(selectorModel)
		return m, cmd
	}

	if !m.cInputs {
		mod, cmd = m.inputsModel.Update(msg)
		m.inputsModel = mod.(inputsModel)
		return m, cmd
	}


	m.err = execCommit(m)
	return m, tea.Quit
}

func (m model) View() string {

	if !m.cSelect {
		return m.selectorModel.View()
	}

	if !m.cInputs {
		return m.inputsModel.View()
	}

	if m.err == nil {
		title := doneTitleStyle.Render(UI_SUCCESS_TITLE)
		message := doneMsgStyle.Render(UI_SUCCESS_MSG)
		return layOutStyle.Render(successStyle.Render(lipgloss.JoinVertical(lipgloss.Left, title, message)))
	} else {
		title := doneTitleStyle.Render(UI_FAILED_TITLE)
		message := doneMsgStyle.Render(fmt.Sprintf("An error occurred during commit: %v", m.err))
		return layOutStyle.Render(failedStyle.Render(lipgloss.JoinVertical(lipgloss.Left, title, message)))
	}
}

//func main() {
//	m := model{
//		selectorModel: newSelectorModel(),
//		inputsModel:   newInputsModel(),
//	}
//	if err := tea.NewProgram(&m).Start(); err != nil {
//		fmt.Printf("could not start program: %s\n", err)
//		os.Exit(1)
//	}
//}
