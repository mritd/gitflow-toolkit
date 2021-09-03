package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"time"
)

type spinnerModel struct {
	spinner spinner.Model
	err     error
}

func newSpinnerModel() spinnerModel {
	s := spinner.NewModel()
	s.Spinner = spinner.Spinner{
		Frames: []string{"∙∙∙∙", "●∙∙∙", "∙●∙∙", "∙∙●∙", "∙∙∙●"},
		FPS:    time.Second / 7,
	}
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return spinnerModel{spinner: s}
}

func (m spinnerModel) Init() tea.Cmd {
	return spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		default:
			return m, nil
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m spinnerModel) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	return fmt.Sprintf("\n\n   %s Commit processing...\n\n", m.spinner.View())
}
