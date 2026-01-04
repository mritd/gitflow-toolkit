package commit

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mritd/gitflow-toolkit/v3/consts"
	"github.com/mritd/gitflow-toolkit/v3/internal/git"
	"github.com/mritd/gitflow-toolkit/v3/internal/llm"
	"github.com/mritd/gitflow-toolkit/v3/internal/ui/common"
)

// aiResult represents the result of AI generation.
type aiResult struct {
	Message   string
	Cancelled bool
	Err       error
}

// aiModel is the bubbletea model for AI generation progress.
type aiModel struct {
	files          []git.FileDiff
	summaries      []string
	fileStatus     []int // 0=pending, 1=running, 2=done, -1=error
	completedCount int
	runningCount   int
	concurrency    int
	finalMsg       string
	spinner        spinner.Model
	progressPos    int
	phase          string // "analyzing" or "generating"
	done           bool
	cancelled      bool
	err            error
	client         *llm.Client
	ctx            context.Context
	cancel         context.CancelFunc
}

// Messages for async operations
type aiFileAnalyzedMsg struct {
	idx     int
	summary string
	err     error
}

type aiFinalGeneratedMsg struct {
	message string
	err     error
}

type aiTickMsg struct{}

func newAIModel(files []git.FileDiff, client *llm.Client) aiModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(common.ColorPrimary)

	ctx, cancel := context.WithCancel(context.Background())
	concurrency := llm.GetConcurrency()

	// Initialize file status: mark first N files as running (1)
	fileStatus := make([]int, len(files))
	runningCount := 0
	for i := range files {
		if runningCount >= concurrency {
			break
		}
		fileStatus[i] = 1 // running
		runningCount++
	}

	return aiModel{
		files:          files,
		summaries:      make([]string, len(files)),
		fileStatus:     fileStatus,
		completedCount: 0,
		runningCount:   runningCount,
		concurrency:    concurrency,
		spinner:        s,
		phase:          "analyzing",
		client:         client,
		ctx:            ctx,
		cancel:         cancel,
	}
}

func (m aiModel) Init() tea.Cmd {
	// Start analyzing files with concurrency limit
	// File status already set in constructor
	cmds := []tea.Cmd{m.spinner.Tick, m.tickAnimation()}

	for i := range m.files {
		if m.fileStatus[i] == 1 { // running
			cmds = append(cmds, m.analyzeFile(i))
		}
	}

	return tea.Batch(cmds...)
}

func (m aiModel) tickAnimation() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return aiTickMsg{}
	})
}

// startNextFile finds the next pending file and starts analyzing it.
// Updates the model in place and returns the command to run.
// Returns -1, nil if no pending files or concurrency limit reached.
func (m aiModel) startNextFile() (int, tea.Cmd) {
	if m.runningCount >= m.concurrency {
		return -1, nil
	}

	for i := range m.files {
		if m.fileStatus[i] == 0 { // pending
			return i, m.analyzeFile(i)
		}
	}

	return -1, nil
}

func (m aiModel) analyzeFile(idx int) tea.Cmd {
	file := m.files[idx]

	return func() tea.Msg {
		prompt := m.buildFilePrompt(file)
		opt := llm.GenerateOptions{
			System: consts.LLMDefaultFilePrompt,
		}
		// Use custom file prompt as system prompt if configured
		if customPrompt := m.client.GetFilePrompt(); customPrompt != "" {
			opt.System = customPrompt
		}
		summary, err := m.client.Generate(m.ctx, m.client.GetModel(), prompt, opt)
		return aiFileAnalyzedMsg{idx: idx, summary: summary, err: err}
	}
}

func (m aiModel) generateFinalMessage() tea.Cmd {
	return func() tea.Msg {
		prompt := m.buildCommitPrompt()
		lang := m.client.GetLang()

		// Select system prompt based on language (custom prompt takes precedence)
		var systemPrompt string
		if customPrompt := m.client.GetCommitPrompt(lang); customPrompt != "" {
			systemPrompt = customPrompt
		} else {
			switch lang {
			case consts.LLMLangZH:
				systemPrompt = consts.LLMCommitPromptZH
			case consts.LLMLangBilingual:
				systemPrompt = consts.LLMCommitPromptBilingual
			default:
				systemPrompt = consts.LLMCommitPromptEN
			}
		}

		opt := llm.GenerateOptions{
			System: systemPrompt,
		}
		message, err := m.client.Generate(m.ctx, m.client.GetModel(), prompt, opt)
		return aiFinalGeneratedMsg{message: message, err: err}
	}
}

// buildFilePrompt creates a prompt for analyzing a single file's changes.
func (m aiModel) buildFilePrompt(file git.FileDiff) string {
	return fmt.Sprintf(`Summarize the changes in this git diff in 1-2 sentences.
Focus on WHAT changed and WHY (if apparent). Be concise.

File: %s
Diff:
%s

Summary:`, file.Path, file.Diff)
}

// buildCommitPrompt creates a prompt for generating the final commit message.
func (m aiModel) buildCommitPrompt() string {
	var sb strings.Builder

	// Few-shot example based on language
	lang := m.client.GetLang()
	switch lang {
	case consts.LLMLangZH:
		sb.WriteString(`示例:
输入:
- auth.go: 添加了 JWT 验证
- user.go: 添加了用户资料接口
- docs.md: 更新了 API 文档

输出:
feat(api): 添加用户认证和资料功能

- 实现 JWT token 验证
- 添加用户资料接口
- 更新 API 文档

输入:
`)
	case consts.LLMLangBilingual:
		sb.WriteString(`Example:
Input:
- auth.go: Added JWT validation
- user.go: Added profile endpoint
- docs.md: Updated API docs

Output:
feat(api): add authentication and user profile (添加用户认证和资料功能)

- 实现 JWT token 验证
- 添加用户资料接口
- 更新 API 文档

Input:
`)
	default:
		sb.WriteString(`Example:
Input:
- auth.go: Added JWT validation
- user.go: Added profile endpoint
- docs.md: Updated API docs

Output:
feat(api): add authentication and user profile

- implement JWT token validation
- add user profile endpoint
- update API documentation

Input:
`)
	}

	for i, summary := range m.summaries {
		if summary != "" {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", m.files[i].Path, strings.TrimSpace(summary)))
		}
	}

	sb.WriteString("\nOutput:")

	return sb.String()
}

func (m aiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.cancelled = true
			m.cancel()
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case aiTickMsg:
		if !m.done && !m.cancelled {
			m.progressPos = (m.progressPos + 1) % 20
			return m, m.tickAnimation()
		}

	case aiFileAnalyzedMsg:
		if m.cancelled {
			return m, tea.Quit
		}

		m.runningCount--

		if msg.err != nil {
			m.fileStatus[msg.idx] = -1 // error
			m.err = fmt.Errorf("failed to analyze %s: %w", m.files[msg.idx].Path, msg.err)
			m.done = true
			return m, tea.Quit
		}

		m.summaries[msg.idx] = msg.summary
		m.fileStatus[msg.idx] = 2 // done
		m.completedCount++

		if m.completedCount >= len(m.files) {
			// All files analyzed, generate final message
			m.phase = "generating"
			return m, m.generateFinalMessage()
		}

		// Start next pending file if under concurrency limit
		nextIdx, cmd := m.startNextFile()
		if nextIdx >= 0 {
			m.fileStatus[nextIdx] = 1 // running
			m.runningCount++
		}
		return m, cmd

	case aiFinalGeneratedMsg:
		if m.cancelled {
			return m, tea.Quit
		}
		m.done = true
		if msg.err != nil {
			m.err = fmt.Errorf("failed to generate commit message: %w", msg.err)
		} else {
			m.finalMsg = msg.message
		}
		return m, tea.Quit
	}

	return m, nil
}

func (m aiModel) View() string {
	if m.done || m.cancelled {
		return ""
	}

	var sb strings.Builder

	// Title
	titleLayout := lipgloss.NewStyle().Padding(1, 0, 1, 2)
	titleStyle := lipgloss.NewStyle().
		Foreground(common.ColorTitleFg).
		Background(common.ColorTitleBg).
		Bold(true).
		Padding(0, 1)
	sb.WriteString(titleLayout.Render(titleStyle.Render("Auto Generate")))
	sb.WriteString("\n")

	// Progress bar with status
	contentLayout := lipgloss.NewStyle().PaddingLeft(2)
	progressBar := m.renderProgressBar()

	var status string
	if m.phase == "analyzing" {
		status = fmt.Sprintf("Analyzing... (%d/%d)", m.completedCount, len(m.files))
	} else {
		status = "Generating commit message..."
	}
	sb.WriteString(contentLayout.Render(progressBar + "  " + status))
	sb.WriteString("\n\n")

	// File list with status (auto-scroll to show running files)
	const maxVisibleFiles = 10
	start, end := m.calcVisibleRange(maxVisibleFiles)

	// Show "N more above" indicator
	if start > 0 {
		moreStyle := lipgloss.NewStyle().Foreground(common.ColorMuted)
		sb.WriteString(contentLayout.Render(moreStyle.Render(fmt.Sprintf("  ↑ %d more above", start))))
		sb.WriteString("\n")
	}

	for i := start; i < end; i++ {
		file := m.files[i]
		var icon string
		var style lipgloss.Style
		switch m.fileStatus[i] {
		case 2: // done
			icon = common.SymbolSuccess
			style = lipgloss.NewStyle().Foreground(common.ColorSuccess)
		case 1: // running
			icon = common.SymbolRunning
			style = lipgloss.NewStyle().Foreground(common.ColorWarning)
		case 0: // pending
			icon = common.SymbolPending
			style = lipgloss.NewStyle().Foreground(common.ColorMuted)
		default: // error
			icon = common.SymbolError
			style = lipgloss.NewStyle().Foreground(common.ColorError)
		}
		sb.WriteString(contentLayout.Render(style.Render(icon + " " + file.Path)))
		sb.WriteString("\n")
	}

	// Show "N more below" indicator
	if end < len(m.files) {
		moreStyle := lipgloss.NewStyle().Foreground(common.ColorMuted)
		sb.WriteString(contentLayout.Render(moreStyle.Render(fmt.Sprintf("  ↓ %d more below", len(m.files)-end))))
		sb.WriteString("\n")
	}

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(common.ColorMuted).
		PaddingLeft(2).
		PaddingTop(1)
	sb.WriteString(helpStyle.Render("Press Ctrl+C to cancel"))
	sb.WriteString("\n")

	return sb.String()
}

// calcVisibleRange calculates the visible file range for auto-scrolling.
// It prioritizes showing running files in the visible window.
func (m aiModel) calcVisibleRange(maxVisible int) (start, end int) {
	total := len(m.files)
	if total <= maxVisible {
		return 0, total
	}

	// Find the first running file
	firstRunning := -1
	for i, status := range m.fileStatus {
		if status == 1 { // running
			firstRunning = i
			break
		}
	}

	// If no running file, show from the first pending or from start
	if firstRunning < 0 {
		// Find first pending
		for i, status := range m.fileStatus {
			if status == 0 { // pending
				firstRunning = i
				break
			}
		}
	}

	// If still not found, show from start
	if firstRunning < 0 {
		return 0, maxVisible
	}

	// Center the running file in the visible window
	// But keep some context (show a few completed files above)
	contextAbove := 2
	start = firstRunning - contextAbove
	if start < 0 {
		start = 0
	}

	end = start + maxVisible
	if end > total {
		end = total
		start = end - maxVisible
		if start < 0 {
			start = 0
		}
	}

	return start, end
}

func (m aiModel) renderProgressBar() string {
	width := 20
	pulseWidth := 6

	var bar strings.Builder
	for i := 0; i < width; i++ {
		inPulse := false
		for j := 0; j < pulseWidth; j++ {
			if (m.progressPos+j)%width == i {
				inPulse = true
				break
			}
		}
		if inPulse {
			bar.WriteString(lipgloss.NewStyle().Foreground(common.ColorSuccess).Render("█"))
		} else {
			bar.WriteString(lipgloss.NewStyle().Foreground(common.ColorMuted).Render("░"))
		}
	}

	return bar.String()
}

// aiPreviewResult represents the result of preview interaction.
type aiPreviewResult struct {
	Message string
	Action  string // "commit", "edit", "retry", "cancel"
}

// aiPreviewModel is the bubbletea model for AI preview.
type aiPreviewModel struct {
	message   string // original AI message (without SOB)
	sob       string // Signed-off-by line
	selected  int    // 0=Commit, 1=Edit, 2=Retry
	committed bool
	edit      bool
	retry     bool
	cancelled bool
}

func newAIPreviewModel(message string) aiPreviewModel {
	return aiPreviewModel{
		message:  message,
		sob:      git.CreateSOB(),
		selected: 0,
	}
}

func (m aiPreviewModel) Init() tea.Cmd {
	return nil
}

func (m aiPreviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			switch m.selected {
			case 0:
				m.committed = true
			case 1:
				m.edit = true
			case 2:
				m.retry = true
			}
			return m, tea.Quit
		case "c":
			m.committed = true
			return m, tea.Quit
		case "e":
			m.edit = true
			return m, tea.Quit
		case "r":
			m.retry = true
			return m, tea.Quit
		case "left", "h":
			if m.selected > 0 {
				m.selected--
			}
		case "right", "l":
			if m.selected < 2 {
				m.selected++
			}
		case "tab":
			m.selected = (m.selected + 1) % 3
		}
	}
	return m, nil
}

func (m aiPreviewModel) View() string {
	if m.committed || m.edit || m.retry || m.cancelled {
		return ""
	}

	var sb strings.Builder

	// Title
	titleLayout := lipgloss.NewStyle().Padding(1, 0, 1, 2)
	titleStyle := lipgloss.NewStyle().
		Foreground(common.ColorTitleFg).
		Background(common.ColorTitleBg).
		Bold(true).
		Padding(0, 1)
	sb.WriteString(titleLayout.Render(titleStyle.Render("Auto Generated Commit")))
	sb.WriteString("\n")

	// Content with left border
	contentLayout := lipgloss.NewStyle().PaddingLeft(2)
	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(common.ColorSuccess).
		PaddingLeft(1)

	// Show message with SOB in preview
	displayMsg := m.message
	if m.sob != "" {
		displayMsg += "\n\n" + m.sob
	}
	sb.WriteString(contentLayout.Render(contentStyle.Render(displayMsg)))
	sb.WriteString("\n")

	// Buttons
	buttonLayout := lipgloss.NewStyle().PaddingLeft(2).PaddingTop(1)
	sb.WriteString(buttonLayout.Render(m.renderButtons()))
	sb.WriteString("\n")

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(common.ColorMuted).
		PaddingLeft(2).
		PaddingTop(1)
	sb.WriteString(helpStyle.Render("←/→ select • enter confirm • c commit • e edit • r retry • q quit"))
	sb.WriteString("\n")

	return sb.String()
}

func (m aiPreviewModel) renderButtons() string {
	activeStyle := lipgloss.NewStyle().
		Foreground(common.ColorTitleFg).
		Background(common.ColorSuccess).
		Bold(true).
		Padding(0, 2)

	inactiveStyle := lipgloss.NewStyle().
		Foreground(common.ColorMuted).
		Background(lipgloss.AdaptiveColor{Light: "#DDDDDD", Dark: "#3a3a3a"}).
		Padding(0, 2)

	editActiveStyle := lipgloss.NewStyle().
		Foreground(common.ColorTitleFg).
		Background(common.ColorPrimary).
		Bold(true).
		Padding(0, 2)

	retryActiveStyle := lipgloss.NewStyle().
		Foreground(common.ColorTitleFg).
		Background(common.ColorWarning).
		Bold(true).
		Padding(0, 2)

	var commitBtn, editBtn, retryBtn string
	switch m.selected {
	case 0:
		commitBtn = activeStyle.Render("  Commit  ")
		editBtn = inactiveStyle.Render("  Edit  ")
		retryBtn = inactiveStyle.Render("  Retry  ")
	case 1:
		commitBtn = inactiveStyle.Render("  Commit  ")
		editBtn = editActiveStyle.Render("  Edit  ")
		retryBtn = inactiveStyle.Render("  Retry  ")
	case 2:
		commitBtn = inactiveStyle.Render("  Commit  ")
		editBtn = inactiveStyle.Render("  Edit  ")
		retryBtn = retryActiveStyle.Render("  Retry  ")
	}

	return commitBtn + "  " + editBtn + "  " + retryBtn
}

// runAIPreview shows the AI-generated message and returns user action.
func runAIPreview(message string) aiPreviewResult {
	m := newAIPreviewModel(message)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return aiPreviewResult{Message: message, Action: "cancel"}
	}

	result := finalModel.(aiPreviewModel)
	if result.cancelled {
		return aiPreviewResult{Message: message, Action: "cancel"}
	}
	if result.edit {
		return aiPreviewResult{Message: message, Action: "edit"}
	}
	if result.retry {
		return aiPreviewResult{Message: message, Action: "retry"}
	}
	return aiPreviewResult{Message: message, Action: "commit"}
}

// runAIGenerate runs the AI generation flow.
func runAIGenerate() aiResult {
	// Get staged diff
	contextLines := llm.GetDiffContext()
	diff, err := git.GetStagedDiff(contextLines)
	if err != nil {
		return aiResult{Err: fmt.Errorf("failed to get staged diff: %w", err)}
	}
	if diff == "" {
		return aiResult{Err: fmt.Errorf("no staged changes")}
	}

	// Split diff by file
	files := git.SplitDiffByFile(diff)
	if len(files) == 0 {
		return aiResult{Err: fmt.Errorf("no files in diff")}
	}

	// Create LLM client
	client := llm.NewClient()

	// Run the TUI
	m := newAIModel(files, client)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return aiResult{Err: err}
	}

	result := finalModel.(aiModel)
	if result.cancelled {
		return aiResult{Cancelled: true}
	}
	if result.err != nil {
		return aiResult{Err: result.err}
	}

	return aiResult{Message: result.finalMsg}
}
