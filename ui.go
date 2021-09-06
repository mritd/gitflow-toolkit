package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
)

const (
	SELECTOR = iota
	INPUTS
	SPINNER
	RESULT
)

func init() {
	runewidth.DefaultCondition.EastAsianWidth = false
}

type done struct {
	err error
}

type commitMsg struct {
	Type    string
	Scope   string
	Subject string
	Body    string
	Footer  string
	SOB     string
}

type model struct {
	err       error
	views     []tea.Model
	viewIndex int
}

func (m model) Init() tea.Cmd {
	return func() tea.Msg {
		err := repoCheck()
		if err != nil {
			return done{err: err}
		}

		err = hasStagedFiles()
		if err != nil {
			return done{err: err}
		}

		return nil
	}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case done: // If the view returns a done message, it means that the stage has been processed
		// Copy error
		m.err = msg.(done).err
		if m.err == nil {
			// Call the next view
			m.viewIndex++
		} else {
			m.viewIndex = RESULT
		}

		// some special views need to determine the state of the data to update
		switch m.viewIndex {
		case SPINNER:
			return m, m.commit
		case RESULT:
			return m, m.result
		default:
			return m, nil
		}
	default: // By default, the cmd returned by the view needs to be processed by itself
		var cmd tea.Cmd
		m.views[m.viewIndex], cmd = m.views[m.viewIndex].Update(msg)
		return m, cmd
	}
}

func (m model) View() string {
	return m.views[m.viewIndex].View()
}

func (m model) commit() tea.Msg {
	return commitMsg{
		Type:    m.views[SELECTOR].(selectorModel).choice,
		Scope:   m.views[INPUTS].(inputsModel).inputs[0].Value(),
		Subject: m.views[INPUTS].(inputsModel).inputs[1].Value(),
		Body:    m.views[INPUTS].(inputsModel).inputs[2].Value(),
		Footer:  m.views[INPUTS].(inputsModel).inputs[3].Value(),
		SOB:     createSOB(),
	}
}

func (m model) result() tea.Msg {
	if m.err != nil {
		return m.err
	} else {
		return successMsg
	}
}

//func main() {
//	m := model{
//		views: []tea.Model{
//			newSelectorModel(),
//			newInputsModel(),
//			newSpinnerModel(),
//			newResultModel(),
//		},
//	}
//	if err := tea.NewProgram(&m).Start(); err != nil {
//		fmt.Printf("could not start program: %s\n", err)
//		os.Exit(1)
//	}
//}
