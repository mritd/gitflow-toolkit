package main

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type installModel struct {
	index   int
	err     error
	stages  map[int]func() error
	spinner spinner.Model
}

func (m installModel) Init() tea.Cmd {
	return nil
}

func (m installModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case done:
	default:

	}

	return m, nil
}

func (m installModel) View() string {
	return ""
}
