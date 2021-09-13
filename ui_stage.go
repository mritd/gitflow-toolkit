package main

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	stageLayoutStyle = lipgloss.NewStyle().
				Padding(0, 0, 1, 2)

	stageBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#37B9FF")).
				Width(55).
				Padding(0, 1, 1, 2)

	stageMsgSuccessStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#25A065"))

	stageMsgFailedStyle = stageMsgSuccessStyle.Copy().
				Foreground(lipgloss.Color("#EE6FF8"))

	stageMsgWaitingStyle = stageMsgSuccessStyle.Copy().
				Foreground(lipgloss.Color("#37B9FF"))

	stageSpinner = spinner.Model{
		Style: lipgloss.NewStyle().Foreground(lipgloss.Color("#f8ca61")),
		Spinner: spinner.Spinner{
			Frames: []string{
				"[     ]",
				"[=    ]",
				"[==   ]",
				"[===  ]",
				"[ === ]",
				"[  ===]",
				"[   ==]",
				"[    =]",
				"[     ]",
				"[    =]",
				"[   ==]",
				"[  ===]",
				"[ === ]",
				"[===  ]",
				"[==   ]",
				"[=    ]",
				"[     ]",
			},
			FPS: time.Second / 10,
		}}
)

type stageFunc func() error

type stage struct {
	title string
	err   error
	f     stageFunc
}

type stageDoneMsg struct{}

type stageModel struct {
	index   int
	stages  []stage
	spinner spinner.Model
}

func (m stageModel) Init() tea.Cmd {
	return tea.Batch(spinner.Tick, func() tea.Msg {
		return m.stages[0]
	})
}

func (m stageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case stage:
		return m, func() tea.Msg {
			time.Sleep(500 * time.Millisecond)
			m.stages[m.index].err = msg.(stage).f()
			return stageDoneMsg{}
		}
	case stageDoneMsg:
		if m.stages[m.index].err != nil {
			return m, tea.Quit
		}

		m.index++
		return m, func() tea.Msg {
			if m.index == len(m.stages) {
				return tea.Quit()
			}
			return m.stages[m.index]
		}
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

}

func (m stageModel) View() string {
	var view string
	for i, stage := range m.stages {
		if stage.err == nil {
			if i < m.index {
				view = lipgloss.JoinVertical(lipgloss.Left, view, stageMsgSuccessStyle.Render("[  ✔  ] "+stage.title))
			} else {
				view = lipgloss.JoinVertical(lipgloss.Left, view, m.spinner.View()+" "+stageMsgWaitingStyle.Render(stage.title))
			}
		} else {
			view = lipgloss.JoinVertical(lipgloss.Left, view, stageMsgFailedStyle.Render("[  ✗  ] "+stage.err.Error()))
		}

	}
	return stageLayoutStyle.Render(stageBorderStyle.Render(view))
}

//func main() {
//	m := stageModel{
//		stages: []stage{
//			{
//				title: "Clean install dir...",
//				f:     func() error { return nil },
//			},
//			{
//				title: "Clean symlinks...",
//				f:     func() error { return nil },
//			},
//			{
//				title: "Unset commit hooks...",
//				f:     func() error { return nil },
//			},
//			{
//				title: "Create toolkit home...",
//				f:     func() error { return nil },
//			},
//			{
//				title: "Install executable file...",
//				f:     func() error { return nil },
//			},
//			{
//				title: "Create symlink...",
//				f:     func() error { return errors.New("This is a test message.") },
//			},
//			{
//				title: "Install success...",
//				f:     func() error { return nil },
//			},
//		},
//		spinner: stageSpinner,
//	}
//	if err := tea.NewProgram(&m).Start(); err != nil {
//		fmt.Printf("could not start program: %s\n", err)
//		os.Exit(1)
//	}
//}
