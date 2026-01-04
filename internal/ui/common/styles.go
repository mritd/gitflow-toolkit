// Package common provides shared UI components and styles.
package common

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Adaptive colors for light and dark terminals.
var (
	ColorPrimary = lipgloss.AdaptiveColor{Light: "#9A4AFF", Dark: "#EE6FF8"}
	ColorBorder  = lipgloss.AdaptiveColor{Light: "#9F72FF", Dark: "#AD58B4"}
	ColorTitleBg = lipgloss.AdaptiveColor{Light: "#16A34A", Dark: "#22C55E"} // Unified green
	ColorTitleFg = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}
	ColorSuccess = lipgloss.AdaptiveColor{Light: "#16A34A", Dark: "#22C55E"} // Same as TitleBg
	ColorWarning = lipgloss.AdaptiveColor{Light: "#D97706", Dark: "#FBBF24"}
	ColorError   = lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#F87171"}
	ColorText    = lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#DDDDDD"}
	ColorMuted   = lipgloss.AdaptiveColor{Light: "#6F6C6C", Dark: "#7A7A7A"}

	// Commit message colors
	ColorCommitType    = lipgloss.AdaptiveColor{Light: "#CA8A04", Dark: "#FACC15"} // Yellow for type
	ColorCommitScope   = lipgloss.AdaptiveColor{Light: "#C026D3", Dark: "#E879F9"} // Magenta for scope
	ColorCommitSubject = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"} // White for subject
	ColorCommitBody    = lipgloss.AdaptiveColor{Light: "#16A34A", Dark: "#4ADE80"} // Green for body
	ColorCommitFooter  = lipgloss.AdaptiveColor{Light: "#2563EB", Dark: "#60A5FA"} // Blue for footer
	ColorCommitSOB     = lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"} // Gray for signed-off-by
)

// Reusable lipgloss styles.
var (
	StyleSuccess    = lipgloss.NewStyle().Foreground(ColorSuccess)
	StyleWarning    = lipgloss.NewStyle().Foreground(ColorWarning)
	StyleError      = lipgloss.NewStyle().Foreground(ColorError)
	StyleMuted      = lipgloss.NewStyle().Foreground(ColorMuted)
	StylePrimary    = lipgloss.NewStyle().Foreground(ColorPrimary)
	StyleTitle      = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).MarginBottom(1)
	StyleCommitType = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
)

// Status indicator symbols.
const (
	SymbolSuccess = "✓"
	SymbolError   = "✗"
	SymbolWarning = "⚠"
	SymbolPending = "○"
	SymbolRunning = "●"
)

// MaxContentWidth defines the maximum display width for content.
const MaxContentWidth = 120

// GetContentWidth calculates the appropriate content width for the given terminal width.
// If terminalWidth is 0, it auto-detects the terminal width.
// It returns the smaller of (terminalWidth - 4) or MaxContentWidth, with a minimum of 40.
func GetContentWidth(terminalWidth int) int {
	if terminalWidth <= 0 {
		w, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil && w > 0 {
			terminalWidth = w
		} else {
			terminalWidth = 80
		}
	}
	width := terminalWidth - 4
	if width > MaxContentWidth {
		width = MaxContentWidth
	}
	if width < 40 {
		width = 40
	}
	return width
}
