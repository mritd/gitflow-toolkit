// Package commit provides the TUI for creating commits.
package commit

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mritd/gitflow-toolkit/v3/consts"
	"github.com/mritd/gitflow-toolkit/v3/internal/ui/common"
)

// aiGenerateChoice is the special choice value for AI generation.
const aiGenerateChoice = "__ai_generate__"

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

	selectorInactiveSelectedStyle = lipgloss.NewStyle().
					Foreground(common.ColorMuted).
					Padding(0, 0, 0, 2)

	selectorHelpStyle = lipgloss.NewStyle().
				Foreground(common.ColorMuted)

	selectorButtonLayout = lipgloss.NewStyle().
				PaddingLeft(2).
				PaddingTop(1)

	selectorAIButtonStyle = lipgloss.NewStyle().
				Foreground(common.ColorMuted).
				Background(lipgloss.AdaptiveColor{Light: "#E0E0E0", Dark: "#3a3a3a"}).
				Padding(0, 2)

	selectorAIButtonActiveStyle = lipgloss.NewStyle().
					Foreground(common.ColorTitleFg).
					Background(common.ColorPrimary).
					Bold(true).
					Padding(0, 2)
)

// selectorItem represents a commit type item in the list.
type selectorItem struct {
	commitType  string
	description string
}

func (i selectorItem) FilterValue() string { return i.description }

// selectorDelegate handles rendering of list items.
type selectorDelegate struct {
	inactive *bool // pointer to track if list is inactive (AI button selected)
}

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

	isInactive := d.inactive != nil && *d.inactive

	if index == m.Index() {
		if isInactive {
			// Selected but inactive (AI button is focused)
			_, _ = fmt.Fprint(w, selectorInactiveSelectedStyle.Render(str))
		} else {
			_, _ = fmt.Fprint(w, selectorSelectedStyle.Render(str))
		}
	} else {
		_, _ = fmt.Fprint(w, selectorNormalStyle.Render(str))
	}
}

// selectorModel is the bubbletea model for commit type selection.
type selectorModel struct {
	list       list.Model
	delegate   *selectorDelegate
	choice     string
	aiSelected bool // true when AI option is focused
	quitting   bool
	width      int
}

// findTypeIndex returns the index of the commit type in CommitTypes.
// Returns 0 if not found.
func findTypeIndex(commitType string) int {
	for i, ct := range consts.CommitTypes {
		if ct.Name == commitType {
			return i
		}
	}
	return 0
}

func newSelectorModel(initialType string) selectorModel {
	items := make([]list.Item, len(consts.CommitTypes))
	for i, ct := range consts.CommitTypes {
		items[i] = selectorItem{
			commitType:  ct.Name,
			description: ct.Description,
		}
	}

	// Create delegate with pointer to track inactive state
	aiSelected := false
	delegate := &selectorDelegate{inactive: &aiSelected}

	// Create list with reasonable defaults
	l := list.New(items, delegate, 40, 12)
	l.Title = "Select Commit Type"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false) // Disable built-in help, we render our own
	l.Styles.Title = selectorTitleStyle
	l.Styles.PaginationStyle = lipgloss.NewStyle().PaddingLeft(2)

	// Set initial selection based on branch type
	if initialType != "" {
		l.Select(findTypeIndex(initialType))
	}

	return selectorModel{list: l, delegate: delegate, aiSelected: false}
}

func (m selectorModel) Init() tea.Cmd {
	return nil
}

func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.list.SetWidth(msg.Width)
		// Adjust height based on terminal size, leave room for title, help, and AI option
		h := msg.Height - 6
		if h < 5 {
			h = 5
		}
		if h > len(consts.CommitTypes)+2 {
			h = len(consts.CommitTypes) + 2
		}
		m.list.SetHeight(h)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if m.aiSelected {
				m.choice = aiGenerateChoice
			} else if item, ok := m.list.SelectedItem().(selectorItem); ok {
				m.choice = item.commitType
			}
			return m, tea.Quit

		case "tab":
			// Toggle between list and AI button
			m.aiSelected = !m.aiSelected
			*m.delegate.inactive = m.aiSelected
			return m, nil

		case "down", "j":
			if m.aiSelected {
				return m, nil // Already at AI button, don't move
			}
			// If at the last item in the list, move to AI button
			if m.list.Index() == len(m.list.Items())-1 {
				m.aiSelected = true
				*m.delegate.inactive = true
				return m, nil
			}

		case "up", "k":
			// If at AI button, move back to list
			if m.aiSelected {
				m.aiSelected = false
				*m.delegate.inactive = false
				return m, nil
			}

		case "a":
			// Quick select AI option
			m.choice = aiGenerateChoice
			return m, tea.Quit
		}
	}

	// Only update list if not on AI option
	if !m.aiSelected {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m selectorModel) View() string {
	if m.choice != "" || m.quitting {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(m.list.View())
	sb.WriteString("\n")

	// AI Generate button
	aiText := "Auto Generate"
	var button string
	if m.aiSelected {
		button = selectorAIButtonActiveStyle.Render(aiText)
	} else {
		button = selectorAIButtonStyle.Render(aiText)
	}
	sb.WriteString(selectorButtonLayout.Render(button))
	sb.WriteString("\n")

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(common.ColorMuted).
		PaddingLeft(2).
		PaddingTop(1)
	sb.WriteString(helpStyle.Render("↑/↓ navigate • tab switch • enter select • a auto generate"))
	sb.WriteString("\n")

	return sb.String()
}

// runSelector shows a selector for commit type using bubbles/list.
// initialType is the commit type to pre-select (can be empty).
func runSelector(initialType string) (string, error) {
	m := newSelectorModel(initialType)
	p := tea.NewProgram(m)

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
