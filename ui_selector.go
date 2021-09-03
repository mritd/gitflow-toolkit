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
				Bold(true).
				Padding(0, 1)

	selectorNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#DDDDDD"}).
				Padding(0, 0, 0, 2)

	selectorSelectedStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}).
				Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
				Bold(true).
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
	list   list.Model
	choice string

	done bool
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
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "enter":
			m.choice = m.list.SelectedItem().(selectorItem).ct
			m.done = true
			return m, nil

		default:
			if !m.list.SettingFilter() && (keypress == "q" || keypress == "esc") {
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
	return "\n" + m.list.View()
}

func newSelectorModel() selectorModel {
	l := list.NewModel([]list.Item{
		selectorItem{ct: FEAT, title: FEAT_DESC},
		selectorItem{ct: FIX, title: FIX_DESC},
		selectorItem{ct: DOCS, title: DOCS_DESC},
		selectorItem{ct: STYLE, title: STYLE_DESC},
		selectorItem{ct: REFACTOR, title: REFACTOR_DESC},
		selectorItem{ct: TEST, title: TEST_DESC},
		selectorItem{ct: CHORE, title: CHORE_DESC},
		selectorItem{ct: PERF, title: PERF_DESC},
		selectorItem{ct: HOTFIX, title: HOTFIX_DESC},
	}, selectorDelegate{}, 20, 12)

	l.Title = "Select Commit Type"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = selectorTitleStyle
	l.Styles.PaginationStyle = selectorPaginationStyle

	return selectorModel{list: l}
}
