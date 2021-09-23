package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	SELECTOR = iota
	INPUTS
	COMMIT
	ERROR
)

type done struct {
	nextView int
	err      error
}

type commitMsg struct {
	Type    string
	Scope   string
	Subject string
	Body    string
	Footer  string
	SOB     string
}

type commitModel struct {
	err       error
	views     []tea.Model
	viewIndex int
}

func (m commitModel) Init() tea.Cmd {
	return func() tea.Msg {
		err := repoCheck()
		if err != nil {
			return done{nextView: ERROR, err: err}
		}

		err = hasStagedFiles()
		if err != nil {
			return done{nextView: ERROR, err: err}
		}

		return nil
	}
}

func (m *commitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case done: // If the view returns a done message, it means that the stage has been processed
		m.err = msg.(done).err
		m.viewIndex = msg.(done).nextView

		// some special views need to determine the state of the data to update
		switch m.viewIndex {
		case INPUTS:
			return m, m.inputs
		case COMMIT:
			return m, m.commit
		case ERROR:
			return m, m.showErr
		default:
			return m, tea.Quit
		}
	default: // By default, the cmd returned by the view needs to be processed by itself
		var cmd tea.Cmd
		m.views[m.viewIndex], cmd = m.views[m.viewIndex].Update(msg)
		return m, cmd
	}
}

func (m commitModel) View() string {
	return m.views[m.viewIndex].View()
}

func (m commitModel) inputs() tea.Msg {
	return strings.ToUpper(m.views[SELECTOR].(selectorModel).choice)
}

func (m commitModel) commit() tea.Msg {
	sob, err := createSOB()
	if err != nil {
		return done{err: err}
	}

	msg := commitMsg{
		Type:    m.views[SELECTOR].(selectorModel).choice,
		Scope:   m.views[INPUTS].(inputsModel).inputs[0].input.Value(),
		Subject: m.views[INPUTS].(inputsModel).inputs[1].input.Value(),
		Body:    m.views[INPUTS].(inputsModel).inputs[2].input.Value(),
		Footer:  m.views[INPUTS].(inputsModel).inputs[3].input.Value(),
		SOB:     sob,
	}

	if msg.Body == "" {
		msg.Body = msg.Subject
	}

	return msg
}

func (m commitModel) showErr() tea.Msg {
	return m.err
}

//func main() {
//	m := commitModel{
//		views: []tea.Model{
//			newSelectorModel(),
//			newInputsModel(),
//			newCommitModel(),
//			newErrorModel(),
//		},
//	}
//	if err := tea.NewProgram(&m).Start(); err != nil {
//		fmt.Printf("could not start program: %s\n", err)
//		os.Exit(1)
//	}
//}
