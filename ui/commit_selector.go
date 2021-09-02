package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
)

var (
	selectorTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	selectorNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#DDDDDD"}).
			Padding(0, 0, 0, 2)

	selectorSelectedStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}).
			Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
			Padding(0, 0, 0, 1)

	selectorPaginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
)

type selectorItem struct {
	ct    string
	title string
}

func (cti selectorItem) FilterValue() string { return cti.title }

type selectorDelegate struct{}

func (d selectorDelegate) Height() int                             { return 1 }
func (d selectorDelegate) Spacing() int                            { return 0 }
func (d selectorDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d selectorDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(selectorItem)
	if !ok {
		return
	}
	str := fmt.Sprintf("%d. %s", index+1, i.title)
	if index == m.Index() {
		_, _ = fmt.Fprintf(w, selectorSelectedStyle.Render(str))
	} else {
		_, _ = fmt.Fprintf(w, selectorNormalStyle.Render(str))
	}

}

type selectorModel struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m selectorModel) Init() tea.Cmd {
	return nil
}

func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(selectorItem)
			if ok {
				m.choice = i.ct
			}
			return m, selectDone

		default:
			if !m.list.SettingFilter() && (keypress == "q" || keypress == "esc") {
				m.quitting = true
				return m, tea.Quit
			}

			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

	default:
		return m, nil
	}
}

func (m selectorModel) View() string {
	if m.choice != "" {
		return m.choice
	}
	if m.quitting {
		return "exit"
	}
	return "\n" + m.list.View()
}

func newSelectorModel() selectorModel {
	l := list.NewModel([]list.Item{
		selectorItem{ct: "feat", title: "FEAT (Introducing new features)"},
		selectorItem{ct: "fix", title: "FIX (Bug fix)"},
		selectorItem{ct: "docs", title: "DOCS (Writing docs)"},
		selectorItem{ct: "style", title: "STYLE (Improving structure/format of the code)"},
		selectorItem{ct: "refactor", title: "REFACTOR (Refactoring code)"},
		selectorItem{ct: "test", title: "TEST (When adding missing tests)"},
		selectorItem{ct: "chore", title: "CHORE (Changing CI/CD)"},
		selectorItem{ct: "perf", title: "PERF (Improving performance)"},
		selectorItem{ct: "hotfix", title: "HOTFIX (Bug fix urgently)"},
	}, selectorDelegate{}, 20, 12)

	l.Title = "Select Commit Type"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = selectorTitleStyle
	l.Styles.PaginationStyle = selectorPaginationStyle

	return selectorModel{list: l}
}

