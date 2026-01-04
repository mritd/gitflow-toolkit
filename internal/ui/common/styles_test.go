package common

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestStyleCommitType(t *testing.T) {
	rendered := StyleCommitType.Render("feat")
	if rendered == "" {
		t.Error("StyleCommitType.Render() should not return empty string")
	}
}

func TestSymbols(t *testing.T) {
	symbols := []struct {
		name  string
		value string
	}{
		{"SymbolSuccess", SymbolSuccess},
		{"SymbolError", SymbolError},
		{"SymbolWarning", SymbolWarning},
		{"SymbolPending", SymbolPending},
		{"SymbolRunning", SymbolRunning},
	}

	for _, s := range symbols {
		t.Run(s.name, func(t *testing.T) {
			if s.value == "" {
				t.Errorf("%s should not be empty", s.name)
			}
		})
	}
}

func TestStylesNotNil(t *testing.T) {
	styles := []struct {
		name  string
		style lipgloss.Style
	}{
		{"StyleSuccess", StyleSuccess},
		{"StyleWarning", StyleWarning},
		{"StyleError", StyleError},
		{"StyleMuted", StyleMuted},
		{"StylePrimary", StylePrimary},
		{"StyleTitle", StyleTitle},
		{"StyleCommitType", StyleCommitType},
	}

	for _, s := range styles {
		t.Run(s.name, func(t *testing.T) {
			rendered := s.style.Render("test")
			if rendered == "" {
				t.Errorf("%s.Render() should not return empty string", s.name)
			}
		})
	}
}
