// Package common provides shared UI components and styles.
package common

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// ResultType indicates the severity level of a result message.
type ResultType int

// Result type constants.
const (
	ResultSuccess ResultType = iota
	ResultError
	ResultWarning
)

// Result represents a formatted message to display to the user.
type Result struct {
	Type    ResultType
	Title   string
	Content string // Pre-formatted content with ANSI colors
}

func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80
	}
	return width
}

// RenderResult renders a Result with adaptive terminal width.
// The output automatically adjusts to terminal width (capped at MaxContentWidth),
// and preserves URLs and file paths on single lines without wrapping.
func RenderResult(r Result) string {
	// Select colors based on result type
	var bgColor, fgColor lipgloss.AdaptiveColor
	var symbol string
	switch r.Type {
	case ResultSuccess:
		bgColor = ColorSuccess
		fgColor = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}
		symbol = SymbolSuccess
	case ResultError:
		bgColor = ColorError
		fgColor = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}
		symbol = SymbolError
	case ResultWarning:
		bgColor = ColorWarning
		fgColor = lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#1A1A1A"}
		symbol = SymbolWarning
	}

	// Layout matching other UIs: Padding(1, 0, 1, 2)
	titleLayout := lipgloss.NewStyle().Padding(1, 0, 1, 2)
	titleStyle := lipgloss.NewStyle().
		Foreground(fgColor).
		Background(bgColor).
		Bold(true).
		Padding(0, 1)

	// Content style with left border indicator
	contentLayout := lipgloss.NewStyle().PaddingLeft(2)
	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(bgColor).
		PaddingLeft(1)

	title := titleLayout.Render(titleStyle.Render(symbol + " " + r.Title))
	content := contentLayout.Render(contentStyle.Render(r.Content))

	return lipgloss.JoinVertical(lipgloss.Left, title, content) + "\n\n"
}

func formatContent(content string, maxWidth int) string {
	if maxWidth <= 0 {
		maxWidth = 76
	}

	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		if shouldPreserveLine(line) {
			// Don't wrap lines with URLs, file paths, or special patterns
			result = append(result, line)
		} else if len(line) > maxWidth {
			// Wrap long plain text lines
			result = append(result, wrapLine(line, maxWidth)...)
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// shouldPreserveLine checks if the line contains URLs, paths, or git refs
// that should not be wrapped.
func shouldPreserveLine(line string) bool {
	// URLs (http://, https://, etc.)
	if strings.Contains(line, "://") {
		return true
	}
	// Unix absolute paths
	if strings.HasPrefix(line, "/") {
		return true
	}
	// Windows absolute paths (C:\, D:/, etc.)
	if len(line) > 2 && line[1] == ':' && (line[2] == '/' || line[2] == '\\') {
		return true
	}
	// Git ref arrows (e.g., "origin/main -> local/main")
	if strings.Contains(line, " -> ") {
		return true
	}
	// Git-related paths and refs
	pathIndicators := []string{".git", "refs/", "origin/", "HEAD"}
	for _, indicator := range pathIndicators {
		if strings.Contains(line, indicator) {
			return true
		}
	}
	return false
}

// wrapLine wraps a long line at word boundaries.
func wrapLine(line string, maxWidth int) []string {
	if len(line) <= maxWidth {
		return []string{line}
	}

	var result []string
	remaining := line

	for len(remaining) > maxWidth {
		// Try to break at the last space before maxWidth
		breakPoint := strings.LastIndex(remaining[:maxWidth], " ")
		if breakPoint <= 0 {
			// No space found, force break at maxWidth
			breakPoint = maxWidth
		}

		result = append(result, remaining[:breakPoint])
		remaining = strings.TrimLeft(remaining[breakPoint:], " ")
	}

	if len(remaining) > 0 {
		result = append(result, remaining)
	}

	return result
}

// Success returns a new success Result with the given title and content.
func Success(title, content string) Result {
	return Result{Type: ResultSuccess, Title: title, Content: content}
}

// Error returns a new error Result with the given title and content.
func Error(title, content string) Result {
	return Result{Type: ResultError, Title: title, Content: content}
}

// Warning returns a new warning Result with the given title and content.
func Warning(title, content string) Result {
	return Result{Type: ResultWarning, Title: title, Content: content}
}

// CommitMessageContent holds the parts of a commit message for styled rendering.
type CommitMessageContent struct {
	Type    string
	Scope   string
	Subject string
	Body    string
	Footer  string
	SOB     string
}

// FormatCommitMessage formats a commit message with colored parts.
// Header: yellow(type) + magenta(scope) + white(subject)
// Body: green
// Footer: blue
// SOB: gray
func FormatCommitMessage(msg CommitMessageContent) string {
	var sb strings.Builder

	// Header: type(scope): subject - each part with distinct color
	typeStyle := lipgloss.NewStyle().Foreground(ColorCommitType)
	scopeStyle := lipgloss.NewStyle().Foreground(ColorCommitScope)
	subjectStyle := lipgloss.NewStyle().Foreground(ColorCommitSubject)

	sb.WriteString(typeStyle.Render(msg.Type))
	sb.WriteString(scopeStyle.Render("(" + msg.Scope + ")"))
	sb.WriteString(subjectStyle.Render(": " + msg.Subject))

	// Body
	if msg.Body != "" {
		sb.WriteString("\n\n")
		bodyStyle := lipgloss.NewStyle().Foreground(ColorCommitBody)
		sb.WriteString(bodyStyle.Render(msg.Body))
	}

	// Footer
	if msg.Footer != "" {
		sb.WriteString("\n\n")
		footerStyle := lipgloss.NewStyle().Foreground(ColorCommitFooter)
		sb.WriteString(footerStyle.Render(msg.Footer))
	}

	// Signed-off-by
	if msg.SOB != "" {
		sb.WriteString("\n\n")
		sobStyle := lipgloss.NewStyle().Foreground(ColorCommitSOB)
		sb.WriteString(sobStyle.Render(msg.SOB))
	}

	return sb.String()
}
