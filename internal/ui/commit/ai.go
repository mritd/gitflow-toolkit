package commit

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mritd/gitflow-toolkit/v3/config"
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
	files           []git.FileDiff
	coreIndices     map[int]bool   // indices of core files (top N by lines changed)
	summaries       []string
	sortedFiles     []git.FileDiff // sorted files for Phase 2
	sortedSummaries []string       // summaries in sorted order
	fileStatus      []int          // 0=pending, 1=running, 2=done, -1=error
	completedCount  int
	runningCount    int
	concurrency     int
	finalMsg        string
	spinner         spinner.Model
	progressPos     int
	phase           string // "analyzing" or "generating"
	done            bool
	cancelled       bool
	err             error
	client          *llm.Client
	ctx             context.Context
	cancel          context.CancelFunc
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

	// Clear debug log for fresh start
	client.ClearDebugLog()

	// Debug: log original file list
	if client.IsDebug() {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Total files: %d\n\n", len(files)))
		for i, f := range files {
			sb.WriteString(fmt.Sprintf("%2d. %s (+%d/-%d)\n", i+1, f.Path, f.LinesAdd, f.LinesDel))
		}
		client.DebugLogSection("Phase 0: Original File List", sb.String())
	}

	// Detect core files (top 3 code files by lines changed) on original order
	originalCoreIndices := git.DetectCoreFiles(files, 3)

	// Debug: log core files detection
	if client.IsDebug() {
		var sb strings.Builder
		sb.WriteString("Core files (top 3 code files by lines changed):\n")
		for idx := range originalCoreIndices {
			f := files[idx]
			sb.WriteString(fmt.Sprintf("  - %s (+%d/-%d)\n", f.Path, f.LinesAdd, f.LinesDel))
		}
		if len(originalCoreIndices) == 0 {
			sb.WriteString("  (no core files detected)\n")
		}
		client.DebugLogSection("Phase 0: Core Files Detection", sb.String())
	}

	// Sort files by priority for better LLM attention
	sortedFiles := git.SortFilesForCommit(files, originalCoreIndices)

	// Rebuild coreIndices for sorted order
	coreIndices := make(map[int]bool)
	for i, f := range sortedFiles {
		for origIdx, origFile := range files {
			if origFile.Path == f.Path && originalCoreIndices[origIdx] {
				coreIndices[i] = true
				break
			}
		}
	}

	// Debug: log sorted file list
	if client.IsDebug() {
		var sb strings.Builder
		sb.WriteString("Sorted by priority: [CORE] code > code > test > config > doc > other\n\n")
		for i, f := range sortedFiles {
			marker := "      "
			if coreIndices[i] {
				marker = "[CORE]"
			}
			sb.WriteString(fmt.Sprintf("%2d. %s %s (+%d/-%d)\n", i+1, marker, f.Path, f.LinesAdd, f.LinesDel))
		}
		client.DebugLogSection("Phase 0: Sorted File List", sb.String())
	}

	// Initialize file status: mark first N files as running (1)
	fileStatus := make([]int, len(sortedFiles))
	runningCount := 0
	for i := range sortedFiles {
		if runningCount >= concurrency {
			break
		}
		fileStatus[i] = 1 // running
		runningCount++
	}

	return aiModel{
		files:          sortedFiles,
		coreIndices:    coreIndices,
		summaries:      make([]string, len(sortedFiles)),
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

	// Skip LLM call for truncated files
	if file.Truncated {
		summary := fmt.Sprintf("(file changes: +%d/-%d lines, content omitted)",
			file.LinesAdd, file.LinesDel)
		return func() tea.Msg {
			return aiFileAnalyzedMsg{idx: idx, summary: summary, err: nil}
		}
	}

	return func() tea.Msg {
		prompt := m.buildFilePrompt(file, idx)
		opt := llm.GenerateOptions{
			System: m.getFileAnalysisSystemPrompt(),
		}
		summary, err := m.client.Generate(m.ctx, m.client.GetModel(), prompt, opt)
		return aiFileAnalyzedMsg{idx: idx, summary: summary, err: err}
	}
}

// getFileAnalysisSystemPrompt returns the system prompt for file analysis.
func (m aiModel) getFileAnalysisSystemPrompt() string {
	if customPrompt := m.client.GetFilePrompt(); customPrompt != "" {
		return customPrompt
	}
	return consts.LLMDefaultFilePrompt
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

// buildFileListContext creates a formatted file list with [CORE] markers.
func (m aiModel) buildFileListContext() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("This commit changes %d files:\n", len(m.files)))

	for i, f := range m.files {
		marker := "      "
		if m.coreIndices[i] {
			marker = "[CORE]"
		}
		sb.WriteString(fmt.Sprintf("%s %s (+%d/-%d)\n", marker, f.Path, f.LinesAdd, f.LinesDel))
	}

	return sb.String()
}

// isCoreFile checks if a file path is in core files.
func (m aiModel) isCoreFile(path string) bool {
	for i, f := range m.files {
		if f.Path == path && m.coreIndices[i] {
			return true
		}
	}
	return false
}

// buildFilePrompt creates a prompt for analyzing a single file's changes.
func (m aiModel) buildFilePrompt(file git.FileDiff, fileIndex int) string {
	var sb strings.Builder

	// Global context: file list
	sb.WriteString(m.buildFileListContext())
	sb.WriteString("\n")

	// Current file to analyze
	marker := ""
	if m.coreIndices[fileIndex] {
		marker = " [CORE]"
	}
	sb.WriteString(fmt.Sprintf("Now analyze%s: %s\n", marker, file.Path))
	sb.WriteString(file.Diff)
	sb.WriteString("\n\n")

	// Instructions
	sb.WriteString("Describe what changed and why (max 50 words), considering this file's role in the overall commit.")

	return sb.String()
}

// buildCommitPrompt creates a prompt for generating the final commit message.
func (m aiModel) buildCommitPrompt() string {
	var sb strings.Builder

	// Few-shot examples based on language (trivial/medium/large)
	lang := m.client.GetLang()
	switch lang {
	case consts.LLMLangZH:
		sb.WriteString(`示例 1 (单一目的, 1 文件):
输入:
- config.go: 修复超时变量名拼写错误

输出:
fix(config): 修复超时变量名拼写错误

- 修复超时变量名拼写错误

示例 2 (单一目的, 多文件 - body 仍然只需 1 行):
输入:
- cmd/root.go: 更新 import 路径
- internal/service/service.go: 更新 import 路径
- pkg/crypto/processor.go: 更新 import 路径
- go.mod: 更新模块路径

输出:
chore(module): 迁移模块路径

- 迁移模块路径以适配新仓库地址

示例 3 (多个目的 - 每个目的 1 行):
输入:
- auth.go: 添加 JWT 验证
- config.go: 添加认证配置
- logger.go: 修复日志格式问题

输出:
feat(auth): 添加 JWT 认证并修复日志

- 添加 JWT 认证功能
- 修复日志格式问题

输入:
`)
	case consts.LLMLangBilingual:
		sb.WriteString(`Example 1 (single purpose, 1 file):
Input:
- config.go: Fix typo in timeout variable name

Output:
fix(config): correct timeout variable typo (修复超时变量名拼写错误)

- 修复超时变量名拼写错误

Example 2 (single purpose, many files - still 1 line body):
Input:
- cmd/root.go: Update import path
- internal/service/service.go: Update import path
- pkg/crypto/processor.go: Update import path
- go.mod: Update module path

Output:
chore(module): migrate module path (迁移模块路径)

- 迁移模块路径以适配新仓库地址

Example 3 (multiple purposes - 1 line per purpose):
Input:
- auth.go: Add JWT validation
- config.go: Add auth config
- logger.go: Fix log format issue

Output:
feat(auth): add JWT auth and fix logging (添加 JWT 认证并修复日志)

- 添加 JWT 认证功能
- 修复日志格式问题

Input:
`)
	default:
		sb.WriteString(`Example 1 (single purpose, 1 file):
Input:
- config.go: Fix typo in timeout variable name

Output:
fix(config): correct timeout variable typo

- Correct timeout variable typo

Example 2 (IMPORTANT - many files with SAME action, body is ONE summary line, NOT file list):
Input:
- cmd/root.go: Update import path
- internal/service/service.go: Update import path
- pkg/crypto/processor.go: Update import path
- pkg/utils/helper.go: Update import path
- main.go: Update import path
- go.mod: Update module path
(... 20+ more files with same change ...)

Output:
chore(module): migrate module paths

- Migrate module paths for new repository location

Example 3 (multiple purposes - 1 line per purpose):
Input:
- cache.go: Add Redis cache support
- config.go: Add cache config
- metrics.go: Fix counter reset bug

Output:
feat(cache): add Redis cache and fix metrics

- Add Redis cache support
- Fix counter reset bug

Input:
`)
	}

	// Use sorted files and summaries with [CORE] markers
	for i, f := range m.sortedFiles {
		summary := m.sortedSummaries[i]
		if summary != "" {
			marker := "      "
			if m.isCoreFile(f.Path) {
				marker = "[CORE]"
			}
			sb.WriteString(fmt.Sprintf("%s %s: %s\n", marker, f.Path, strings.TrimSpace(summary)))
		}
	}

	sb.WriteString("\nOutput (MUST include body with \"- \" lines):")

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

		// Debug: log each file analysis result
		if m.client.IsDebug() {
			m.client.DebugLog("Phase 1: [%d/%d] %s => %s",
				m.completedCount, len(m.files), m.files[msg.idx].Path, strings.TrimSpace(msg.summary))
		}

		if m.completedCount >= len(m.files) {
			// Files already sorted in newAIModel, just copy for Phase 2
			m.sortedFiles = m.files
			m.sortedSummaries = m.summaries

			// Debug: log Phase 1 complete summary
			if m.client.IsDebug() {
				var sb strings.Builder
				sb.WriteString("All file summaries:\n\n")
				for i, f := range m.files {
					marker := "      "
					if m.coreIndices[i] {
						marker = "[CORE]"
					}
					sb.WriteString(fmt.Sprintf("%s %s:\n  %s\n\n", marker, f.Path, strings.TrimSpace(m.summaries[i])))
				}
				m.client.DebugLogSection("Phase 1 Complete: File Summaries", sb.String())
			}

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
			// Debug: log error
			if m.client.IsDebug() {
				m.client.DebugLogSection("Phase 2 Error", msg.err.Error())
			}
		} else {
			m.finalMsg = msg.message
			// Debug: log final commit message
			if m.client.IsDebug() {
				m.client.DebugLogSection("Phase 2 Complete: Generated Commit Message", msg.message)
			}
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

	// Debug log path (if debug mode is enabled)
	if config.GetBool(config.GitConfigLLMAPIDebug, false) {
		debugStyle := lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#666666", Dark: "#888888"}).
			PaddingLeft(2).
			PaddingTop(1)
		sb.WriteString(debugStyle.Render("Debug log: " + llm.GetDebugLogPath()))
		sb.WriteString("\n")
	}

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

	// Filter and truncate diffs
	maxDiffLines := config.GetInt(config.GitConfigLLMMaxDiffLines, consts.LLMDefaultMaxDiffLines)
	files = git.FilterAndTruncateDiffs(files, maxDiffLines)
	if len(files) == 0 {
		return aiResult{Err: fmt.Errorf("no files after filtering")}
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
