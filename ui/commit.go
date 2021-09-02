package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
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
	cSOB     string

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
		return m, cmd
	case inputsDoneMsg:
		m.cScope = m.inputsModel.scope
		m.cSubject = m.inputsModel.subject
		m.cBody = m.inputsModel.body
		m.cFooter = m.inputsModel.footer
		return m, tea.Quit
	}

	if m.cType == "" {
		mod, cmd = m.selectorModel.Update(msg)
		m.selectorModel = mod.(selectorModel)
	} else {
		mod, cmd = m.inputsModel.Update(msg)
		m.inputsModel = mod.(inputsModel)
	}
	return m, cmd
}

func (m model) View() string {
	if m.cType == "" {
		return m.selectorModel.View()
	} else {
		return m.inputsModel.View()
	}
}

func main() {
	m := model{
		selectorModel: newSelectorModel(),
		inputsModel:   newInputsModel(),
	}
	if err := tea.NewProgram(&m).Start(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("%s(%s): %s\n%s\n%s\n", m.cType, m.cScope, m.cSubject, m.cBody, m.cFooter)
}
