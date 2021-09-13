package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
)

func init() {
	// Some users may need to turn off the EastAsianWidth option to make the UI display correctly
	// See also: https://github.com/charmbracelet/lipgloss/issues/40#issuecomment-891167509
	runewidth.DefaultCondition.EastAsianWidth = false
}

func TestSingleTask(t *testing.T) {
	m := NewSingleTaskModel(Task{
		Title: "This is a test message.",
		Func:  NothingFunc,
	})

	if err := tea.NewProgram(&m).Start(); err != nil {
		t.Fatal(err)
	}
}
