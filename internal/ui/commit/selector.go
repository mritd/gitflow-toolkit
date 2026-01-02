// Package commit provides the TUI for creating commits.
package commit

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mritd/gitflow-toolkit/v2/internal/config"
	"github.com/mritd/gitflow-toolkit/v2/internal/ui/common"
)

// Styles for the selector
var (
	selectorTitleStyle = lipgloss.NewStyle().
				Foreground(common.ColorTitleFg).
				Background(common.ColorTitleBg).
				Bold(true).
				Padding(0, 1)

	selectorNormalStyle = lipgloss.NewStyle().
				Foreground(common.ColorText).
				Padding(0, 0, 0, 2)

	selectorSelectedStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(common.ColorBorder).
				Foreground(common.ColorPrimary).
				Bold(true).
				Padding(0, 0, 0, 1)

	selectorHelpStyle = lipgloss.NewStyle().
				Foreground(common.ColorMuted)
)

// selectorItem represents a commit type item in the list.
type selectorItem struct {
	commitType  string
	description string
}

func (i selectorItem) FilterValue() string { return i.description }

// selectorDelegate handles rendering of list items.
type selectorDelegate struct{}

func (d selectorDelegate) Height() int                             { return 1 }
func (d selectorDelegate) Spacing() int                            { return 0 }
func (d selectorDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d selectorDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(selectorItem)
	if !ok {
		return
	}

	// Format: "1. FEAT (Introducing new features)"
	str := fmt.Sprintf("%d. %s (%s)", index+1, strings.ToUpper(i.commitType), i.description)

	if index == m.Index() {
		_, _ = fmt.Fprint(w, selectorSelectedStyle.Render(str))
	} else {
		_, _ = fmt.Fprint(w, selectorNormalStyle.Render(str))
	}
}

// selectorModel is the bubbletea model for commit type selection.
type selectorModel struct {
	list     list.Model
	choice   string
	quitting bool
}

func newSelectorModel() selectorModel {
	items := make([]list.Item, len(config.CommitTypes))
	for i, ct := range config.CommitTypes {
		items[i] = selectorItem{
			commitType:  ct.Name,
			description: ct.Description,
		}
	}

	// Create list with reasonable defaults
	l := list.New(items, selectorDelegate{}, 40, 14)
	l.Title = "Select Commit Type"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = selectorTitleStyle
	l.Styles.PaginationStyle = lipgloss.NewStyle().PaddingLeft(2)
	l.Styles.HelpStyle = selectorHelpStyle

	return selectorModel{list: l}
}

func (m selectorModel) Init() tea.Cmd {
	return nil
}

func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		// Adjust height based on terminal size, leave room for title and help
		h := msg.Height - 4
		if h < 5 {
			h = 5
		}
		if h > len(config.CommitTypes)+2 {
			h = len(config.CommitTypes) + 2
		}
		m.list.SetHeight(h)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if item, ok := m.list.SelectedItem().(selectorItem); ok {
				m.choice = item.commitType
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectorModel) View() string {
	if m.choice != "" {
		return ""
	}
	return "\n" + m.list.View()
}

// runSelector shows a selector for commit type using bubbles/list.
func runSelector() (string, error) {
	m := newSelectorModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run selector: %w", err)
	}

	result := finalModel.(selectorModel)
	if result.quitting {
		return "", errUserAborted
	}

	return result.choice, nil
}

// errUserAborted is returned when the user cancels the selection.
var errUserAborted = fmt.Errorf("user aborted")
