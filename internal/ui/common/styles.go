// Package common provides shared UI components and styles.
package common

import (
	"github.com/charmbracelet/lipgloss"
)

// Adaptive colors for light and dark terminals.
var (
	ColorPrimary = lipgloss.AdaptiveColor{Light: "#9A4AFF", Dark: "#EE6FF8"}
	ColorBorder  = lipgloss.AdaptiveColor{Light: "#9F72FF", Dark: "#AD58B4"}
	ColorTitleBg = lipgloss.AdaptiveColor{Light: "#19A04B", Dark: "#19A04B"}
	ColorTitleFg = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}
	ColorSuccess = lipgloss.AdaptiveColor{Light: "#19A04B", Dark: "#25A065"}
	ColorWarning = lipgloss.AdaptiveColor{Light: "#D97706", Dark: "#FBBF24"}
	ColorError   = lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#F87171"}
	ColorText    = lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#DDDDDD"}
	ColorMuted   = lipgloss.AdaptiveColor{Light: "#6F6C6C", Dark: "#7A7A7A"}
	ColorDimmed  = lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#6B7280"}
)

// Reusable lipgloss styles.
var (
	StyleBold       = lipgloss.NewStyle().Bold(true)
	StyleSuccess    = lipgloss.NewStyle().Foreground(ColorSuccess)
	StyleWarning    = lipgloss.NewStyle().Foreground(ColorWarning)
	StyleError      = lipgloss.NewStyle().Foreground(ColorError)
	StyleMuted      = lipgloss.NewStyle().Foreground(ColorMuted)
	StylePrimary    = lipgloss.NewStyle().Foreground(ColorPrimary)
	StyleDimmed     = lipgloss.NewStyle().Foreground(ColorDimmed)
	StyleRoundedBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1)
	StyleErrorBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorError).
			Padding(0, 1)
	StyleSuccessBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSuccess).
			Padding(0, 1)
	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)
	StyleHelp = lipgloss.NewStyle().
			Foreground(ColorDimmed).
			MarginTop(1)
	StyleCommitType = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)
)

// Status indicator symbols.
const (
	SymbolSuccess = "✓"
	SymbolError   = "✗"
	SymbolWarning = "⚠"
	SymbolPending = "○"
	SymbolRunning = "●"
	SymbolArrow   = "→"
	SymbolBullet  = "•"
)

// MaxContentWidth defines the maximum display width for content.
const MaxContentWidth = 80

// GetContentWidth calculates the appropriate content width for the given terminal width.
// It returns the smaller of (terminalWidth - 4) or MaxContentWidth, with a minimum of 40.
func GetContentWidth(terminalWidth int) int {
	width := terminalWidth - 4
	if width > MaxContentWidth {
		width = MaxContentWidth
	}
	if width < 40 {
		width = 40
	}
	return width
}
