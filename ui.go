package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"time"
)

func init() {
	runewidth.EastAsianWidth = false
	runewidth.DefaultCondition.EastAsianWidth = false
}

var (
	layOutStyle = lipgloss.NewStyle().
			Padding(0, 1, 2, 1)

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

type model struct {
	cType    string
	cScope   string
	cSubject string
	cBody    string
	cFooter  string

	stage      int
	committing bool

	err      error
	spinner  spinnerModel
	selector selectorModel
	inputs   inputsModel
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (mod tea.Model, cmd tea.Cmd) {

	switch m.stage {
	case 0:
		mod, cmd = m.selector.Update(msg)
		m.selector = mod.(selectorModel)
		if m.selector.done {
			m.cType = m.selector.choice
			m.stage++
		}

		return m, cmd
	case 1:
		mod, cmd = m.inputs.Update(msg)
		m.inputs = mod.(inputsModel)

		if m.inputs.done {
			m.cScope = m.inputs.scope
			m.cSubject = m.inputs.subject
			m.cBody = m.inputs.body
			m.cFooter = m.inputs.footer
			m.stage++
		}

		return m, spinner.Tick
	case 2:
		switch msg.(type) {
		case error:
			m.err = msg.(error)
			m.stage++
			return m, nil
		case string:
			m.stage++
			return m, nil
		}

		mod, cmd = m.spinner.Update(msg)
		m.spinner = mod.(spinnerModel)
		return m, tea.Batch(cmd, func() tea.Msg {
			if !m.committing {
				m.committing = true
				time.Sleep(3*time.Second)
				if err := execCommit(m);err != nil {
					return err
				}else {
					return "ok"
				}
			}
			return nil
		})
	}

	return m, tea.Quit
}

func (m model) View() string {
	switch m.stage {
	case 0:
		return m.selector.View()
	case 1:
		return m.inputs.View()
	case 2:
		return m.spinner.View()
	default:
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
}

//func main() {
//	m := model{
//		selector: newSelectorModel(),
//		inputs:   newInputsModel(),
//		spinner:  newSpinnerModel(),
//	}
//	if err := tea.NewProgram(&m).Start(); err != nil {
//		fmt.Printf("could not start program: %s\n", err)
//		os.Exit(1)
//	}
//}
