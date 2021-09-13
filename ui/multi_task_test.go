package ui

import (
	"errors"
	"testing"

	"github.com/mattn/go-runewidth"

	tea "github.com/charmbracelet/bubbletea"
)

func init() {
	// Some users may need to turn off the EastAsianWidth option to make the UI display correctly
	// See also: https://github.com/charmbracelet/lipgloss/issues/40#issuecomment-891167509
	runewidth.DefaultCondition.EastAsianWidth = false
}

func TestMultiTaskModel(t *testing.T) {
	m := NewMultiTaskModelWithTasks([]Task{
		{
			Title: "Clean install dir...",
			Func:  NothingFunc,
		},
		{
			Title: "Clean symlinks...",
			Func:  NothingFunc,
		},
		{
			Title: "Unset commit hooks...",
			Func:  NothingFunc,
		},
		{
			Title: "Create toolkit home...",
			Func:  NothingFunc,
		},
		{
			Title: "Install executable file...",
			Func:  NothingFunc,
		},
		{
			Title: "Create symlink...",
			Func:  func() error { return errors.New("This is a test message.") },
		},
		{
			Title: "Install success...",
			Func:  NothingFunc,
		},
	})
	if err := tea.NewProgram(&m).Start(); err != nil {
		t.Fatal(err)
	}
}
