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
	Note    string // Optional note displayed at the bottom (e.g., quote)
}

func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80
	}
	return width
}

// containsANSI checks if a string contains ANSI escape codes.
func containsANSI(s string) bool {
	return strings.Contains(s, "\x1b[")
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

	// Format content with line wrapping (skip if content already contains ANSI codes)
	formattedContent := r.Content
	if !containsANSI(r.Content) {
		termWidth := getTerminalWidth()
		contentWidth := GetContentWidth(termWidth)
		// Account for padding (2 left) and border (2 = "│ ")
		maxContentWidth := contentWidth - 4
		formattedContent = formatContent(r.Content, maxContentWidth)
	}

	title := titleLayout.Render(titleStyle.Render(symbol + " " + r.Title))
	content := contentLayout.Render(contentStyle.Render(formattedContent))

	// Build result with optional note
	var sb strings.Builder
	sb.WriteString(lipgloss.JoinVertical(lipgloss.Left, title, content))
	sb.WriteString("\n")

	// Add note if present (with bullet prefix and bold green style)
	if r.Note != "" {
		termWidth := getTerminalWidth()
		contentWidth := GetContentWidth(termWidth)
		// Account for padding (2) and prefix (5 = "◉◉◉◉ ")
		noteWidth := contentWidth - 2 - 5

		noteStyle := lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true).
			PaddingLeft(2)

		// Wrap the note text
		wrappedLines := wrapLine(r.Note, noteWidth)
		sb.WriteString("\n")
		for _, line := range wrappedLines {
			sb.WriteString(noteStyle.Render("◉◉◉◉ " + line))
			sb.WriteString("\n")
		}
	}

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
		if len(line) > maxWidth {
			// Smart wrap: preserve URLs, wrap other parts
			result = append(result, smartWrapLine(line, maxWidth)...)
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// smartWrapLine wraps a line while preserving URLs intact.
// It splits the line into URL and non-URL segments, wrapping only the non-URL parts.
func smartWrapLine(line string, maxWidth int) []string {
	// Find all URLs in the line
	segments := splitByURLs(line)

	var result []string
	var currentLine strings.Builder

	for _, seg := range segments {
		if seg.isURL {
			// URL segment: don't break it
			if currentLine.Len() > 0 && currentLine.Len()+len(seg.text) > maxWidth {
				// Current line is too long, flush it first
				result = append(result, wrapLine(currentLine.String(), maxWidth)...)
				currentLine.Reset()
			}
			// If URL itself is longer than maxWidth, put it on its own line
			if currentLine.Len() == 0 {
				result = append(result, seg.text)
			} else {
				currentLine.WriteString(seg.text)
			}
		} else {
			// Non-URL segment: can be wrapped
			currentLine.WriteString(seg.text)
		}
	}

	// Flush remaining content
	if currentLine.Len() > 0 {
		result = append(result, wrapLine(currentLine.String(), maxWidth)...)
	}

	return result
}

type textSegment struct {
	text  string
	isURL bool
}

// splitByURLs splits a line into URL and non-URL segments.
func splitByURLs(line string) []textSegment {
	var segments []textSegment
	remaining := line

	for len(remaining) > 0 {
		// Find the next URL (http:// or https://)
		httpIdx := strings.Index(remaining, "http://")
		httpsIdx := strings.Index(remaining, "https://")

		urlStart := -1
		if httpIdx >= 0 && (httpsIdx < 0 || httpIdx < httpsIdx) {
			urlStart = httpIdx
		} else if httpsIdx >= 0 {
			urlStart = httpsIdx
		}

		if urlStart < 0 {
			// No more URLs, add the rest as non-URL
			if len(remaining) > 0 {
				segments = append(segments, textSegment{text: remaining, isURL: false})
			}
			break
		}

		// Add text before URL
		if urlStart > 0 {
			segments = append(segments, textSegment{text: remaining[:urlStart], isURL: false})
		}

		// Find URL end (space, or end of string)
		urlEnd := urlStart
		for urlEnd < len(remaining) && !isURLTerminator(remaining[urlEnd]) {
			urlEnd++
		}

		// Add URL segment
		segments = append(segments, textSegment{text: remaining[urlStart:urlEnd], isURL: true})
		remaining = remaining[urlEnd:]
	}

	return segments
}

// isURLTerminator checks if a character terminates a URL.
func isURLTerminator(c byte) bool {
	// Common URL terminators: space, quotes, brackets, etc.
	return c == ' ' || c == '\t' || c == '"' || c == '\'' ||
		c == '<' || c == '>' || c == '[' || c == ']' ||
		c == '(' || c == ')' || c == '{' || c == '}'
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
// maxWidth: maximum line width for wrapping (0 = no wrapping)
func FormatCommitMessage(msg CommitMessageContent, maxWidth int) string {
	var sb strings.Builder

	// Header: type(scope): subject - each part with distinct color
	typeStyle := lipgloss.NewStyle().Foreground(ColorCommitType)
	scopeStyle := lipgloss.NewStyle().Foreground(ColorCommitScope)
	subjectStyle := lipgloss.NewStyle().Foreground(ColorCommitSubject)

	// Calculate header prefix length for subject wrapping
	headerPrefix := msg.Type + "(" + msg.Scope + "): "
	prefixLen := len(headerPrefix)

	sb.WriteString(typeStyle.Render(msg.Type))
	sb.WriteString(scopeStyle.Render("(" + msg.Scope + ")"))
	sb.WriteString(": ")

	// Wrap subject if needed, applying color to each line
	if maxWidth > 0 && len(headerPrefix)+len(msg.Subject) > maxWidth {
		subjectWidth := maxWidth - prefixLen
		if subjectWidth < 20 {
			subjectWidth = 20
		}
		subjectLines := wrapLine(msg.Subject, subjectWidth)
		for i, line := range subjectLines {
			if i > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(subjectStyle.Render(line))
		}
	} else {
		sb.WriteString(subjectStyle.Render(msg.Subject))
	}

	// Body - wrap each line and apply color
	if msg.Body != "" {
		sb.WriteString("\n\n")
		bodyStyle := lipgloss.NewStyle().Foreground(ColorCommitBody)
		bodyLines := strings.Split(msg.Body, "\n")
		for i, line := range bodyLines {
			if i > 0 {
				sb.WriteString("\n")
			}
			if maxWidth > 0 && len(line) > maxWidth {
				wrapped := wrapLine(line, maxWidth)
				for j, wl := range wrapped {
					if j > 0 {
						sb.WriteString("\n")
					}
					sb.WriteString(bodyStyle.Render(wl))
				}
			} else {
				sb.WriteString(bodyStyle.Render(line))
			}
		}
	}

	// Footer - wrap and apply color
	if msg.Footer != "" {
		sb.WriteString("\n\n")
		footerStyle := lipgloss.NewStyle().Foreground(ColorCommitFooter)
		if maxWidth > 0 && len(msg.Footer) > maxWidth {
			footerLines := wrapLine(msg.Footer, maxWidth)
			for i, line := range footerLines {
				if i > 0 {
					sb.WriteString("\n")
				}
				sb.WriteString(footerStyle.Render(line))
			}
		} else {
			sb.WriteString(footerStyle.Render(msg.Footer))
		}
	}

	// Signed-off-by
	if msg.SOB != "" {
		sb.WriteString("\n\n")
		sobStyle := lipgloss.NewStyle().Foreground(ColorCommitSOB)
		sb.WriteString(sobStyle.Render(msg.SOB))
	}

	return sb.String()
}
