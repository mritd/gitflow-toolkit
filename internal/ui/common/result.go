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
	Content string
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
	termWidth := getTerminalWidth()
	contentWidth := GetContentWidth(termWidth)

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

	// Title bar style (like TUI title)
	titleStyle := lipgloss.NewStyle().
		Foreground(fgColor).
		Background(bgColor).
		Bold(true).
		Padding(0, 1)

	// Content style with left border indicator
	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(bgColor).
		PaddingLeft(1)

	var sb strings.Builder

	// Title bar
	sb.WriteString("\n")
	sb.WriteString(titleStyle.Render(symbol + " " + r.Title))
	sb.WriteString("\n\n")

	// Format and render content
	formattedContent := formatContent(r.Content, contentWidth-4)
	sb.WriteString(contentStyle.Render(formattedContent))
	sb.WriteString("\n")

	return sb.String()
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
