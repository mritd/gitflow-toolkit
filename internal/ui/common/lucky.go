package common

import (
	"crypto/rand"
	"encoding/hex"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LuckyResult represents the result of a lucky commit operation.
type LuckyResult struct {
	Cancelled bool
	Err       error
	Hash      string
}

// luckyModel is the bubbletea model for lucky commit spinner.
type luckyModel struct {
	prefix      string
	cmd         *exec.Cmd
	spinner     spinner.Model
	progressPos int
	fakeHash    string
	done        bool
	success     bool
	cancelled   bool
	err         error
	hash        string
	getHash     func() (string, error)
}

// luckyDoneMsg is sent when lucky_commit completes.
type luckyDoneMsg struct {
	err error
}

// luckyTickMsg is sent for animation updates.
type luckyTickMsg struct{}

// newLuckyModel creates a new lucky commit model.
func newLuckyModel(prefix string, cmd *exec.Cmd, getHash func() (string, error)) luckyModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)

	return luckyModel{
		prefix:      prefix,
		cmd:         cmd,
		spinner:     s,
		progressPos: 0,
		fakeHash:    generateFakeHash(),
		getHash:     getHash,
	}
}

// generateFakeHash generates a random 40-char hex hash.
func generateFakeHash() string {
	randomBytes := make([]byte, 20)
	_, _ = rand.Read(randomBytes)
	return hex.EncodeToString(randomBytes)
}

// Init initializes the model.
func (m luckyModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.runLuckyCommit(),
		m.tickAnimation(),
	)
}

func (m luckyModel) runLuckyCommit() tea.Cmd {
	return func() tea.Msg {
		err := m.cmd.Run()
		return luckyDoneMsg{err: err}
	}
}

func (m luckyModel) tickAnimation() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return luckyTickMsg{}
	})
}

// Update handles messages.
func (m luckyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.cancelled = true
			// Send SIGTERM to the process
			if m.cmd.Process != nil {
				_ = m.cmd.Process.Signal(syscall.SIGTERM)
			}
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case luckyTickMsg:
		if !m.done {
			m.progressPos = (m.progressPos + 1) % 20
			m.fakeHash = generateFakeHash()
			return m, m.tickAnimation()
		}

	case luckyDoneMsg:
		m.done = true
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.success = true
			// Get the actual hash
			if hash, err := m.getHash(); err == nil {
				m.hash = hash
			}
		}
		return m, tea.Quit
	}

	return m, nil
}

// View renders the model.
func (m luckyModel) View() string {
	if m.done || m.cancelled {
		return ""
	}

	var sb strings.Builder

	// Title with background color
	titleLayout := lipgloss.NewStyle().Padding(1, 0, 1, 2)
	titleStyle := lipgloss.NewStyle().
		Foreground(ColorTitleFg).
		Background(ColorTitleBg).
		Bold(true).
		Padding(0, 1)
	sb.WriteString(titleLayout.Render(titleStyle.Render("Lucky Commit")))
	sb.WriteString("\n")

	// Progress bar with status
	contentLayout := lipgloss.NewStyle().PaddingLeft(2)
	progressBar := m.renderProgressBar()
	sb.WriteString(contentLayout.Render(progressBar + "  Searching: " + m.prefix + "..."))
	sb.WriteString("\n\n")

	// Current hash animation
	hashStyle := lipgloss.NewStyle().Foreground(ColorMuted)
	sb.WriteString(contentLayout.Render("Current: " + hashStyle.Render(m.fakeHash)))
	sb.WriteString("\n")

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(ColorMuted).
		PaddingLeft(2).
		PaddingTop(1)
	sb.WriteString(helpStyle.Render("Press Ctrl+C to skip"))
	sb.WriteString("\n")

	return sb.String()
}

func (m luckyModel) renderProgressBar() string {
	width := 20
	pulseWidth := 6

	var bar strings.Builder
	for i := 0; i < width; i++ {
		// Create pulse effect - check if position is within pulse window
		inPulse := false
		for j := 0; j < pulseWidth; j++ {
			if (m.progressPos+j)%width == i {
				inPulse = true
				break
			}
		}
		if inPulse {
			bar.WriteString(lipgloss.NewStyle().Foreground(ColorSuccess).Render("█"))
		} else {
			bar.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("░"))
		}
	}

	return bar.String()
}

// RunLuckyCommit runs lucky_commit with an animated spinner.
// Returns the result of the operation.
func RunLuckyCommit(prefix string, cmd *exec.Cmd, getHash func() (string, error)) LuckyResult {
	m := newLuckyModel(prefix, cmd, getHash)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return LuckyResult{Err: err}
	}

	result := finalModel.(luckyModel)
	return LuckyResult{
		Cancelled: result.cancelled,
		Err:       result.err,
		Hash:      result.hash,
	}
}
