package main

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	committingStyle = lipgloss.NewStyle().
			Padding(1, 1, 1, 2)

	committingTypeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#7653FF"))

	committingScopeStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#2AD67F"))

	committingSubjectStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color("#EE6FF8"))

	committingBodyStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#2AD67F"))

	committingFooterStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#2AD67F"))

	committingSuccessStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#19F896"))

	committingFailedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#D63B3A"))
)

type committingModel struct {
	err     error
	done    bool
	msg     commitMsg
	spinner spinner.Model
}

func newCommittingModel() committingModel {
	s := spinner.NewModel()
	s.Spinner = spinner.Spinner{
		Frames: []string{
			"(●    ) C",
			"( ●   ) Co",
			"(  ●  ) Com",
			"(   ● ) Comm",
			"(    ●) Commi",
			"(    ●) Commit",
			"(   ● ) Committ",
			"(  ●  ) Committi",
			"( ●   ) Committin",
			"(●    ) Committing",
			"( ●   ) Committing.",
			"(  ●  ) Committing..",
			"(   ● ) Committing...",
			"(    ●) Committing...",
			"(   ● ) Committing...",
			"(  ●  ) Committing...",
			"( ●   ) Committing...",
			"(●    ) Committing...",
		},
		FPS: time.Second / 15,
	}
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#19F896")).Bold(true)
	return committingModel{spinner: s}
}

func (m committingModel) Init() tea.Cmd {
	return nil
}

func (m committingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		default:
			return m, nil
		}
	case commitMsg:
		m.msg = msg
		return m, func() tea.Msg {
			time.Sleep(time.Second)
			return commit(msg)
		}
	case error:
		m.done = true
		m.err = msg
		return m, tea.Quit
	case nil:
		m.done = true
		return m, tea.Quit
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m committingModel) View() string {
	header := committingTypeStyle.Render(m.msg.Type) + committingScopeStyle.Render("("+m.msg.Scope+")") + committingSubjectStyle.Render(": "+m.msg.Subject) + "\n"
	body := committingBodyStyle.Render(m.msg.Body)
	footer := committingFooterStyle.Render(m.msg.Footer+"\n"+m.msg.SOB) + "\n"

	msg := m.spinner.View()
	if m.done {
		if m.err != nil {
			msg = committingFailedStyle.Render("( ●●● ) Commit Failed: " + m.err.Error())
		} else {
			msg = committingSuccessStyle.Render("◉◉◉◉ Always code as if the guy who ends up maintaining your \n◉◉◉◉ code will be a violent psychopath who knows where you live...")
		}
	}

	return committingStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, body, footer, msg))
}
