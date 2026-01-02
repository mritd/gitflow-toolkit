package common

import (
	"strings"
	"testing"
)

func TestShouldPreserveLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{"HTTP URL", "See: https://example.com/path", true},
		{"HTTPS URL", "Error at https://github.com/user/repo/issues/123", true},
		{"Absolute path", "/usr/local/bin/gitflow-toolkit", true},
		{"Windows path", "C:/Users/test/file.txt", true},
		{"Git ref", "origin/main -> local/main", true},
		{"Git HEAD", "HEAD is now at abc123", true},
		{"Git refs path", "refs/heads/master", true},
		{".git path", "fatal: .git/hooks/commit-msg failed", true},
		{"Plain text", "This is a simple error message", false},
		{"Plain text with colon", "Error: something went wrong", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldPreserveLine(tt.line)
			if result != tt.expected {
				t.Errorf("shouldPreserveLine(%q) = %v, want %v", tt.line, result, tt.expected)
			}
		})
	}
}

func TestWrapLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		width    int
		expected []string
	}{
		{
			name:     "Short line",
			line:     "Hello world",
			width:    20,
			expected: []string{"Hello world"},
		},
		{
			name:     "Exact width",
			line:     "Hello world",
			width:    11,
			expected: []string{"Hello world"},
		},
		{
			name:     "Wrap at word boundary",
			line:     "Hello world foo bar",
			width:    12,
			expected: []string{"Hello world", "foo bar"},
		},
		{
			name:     "Multiple wraps",
			line:     "The quick brown fox jumps over the lazy dog",
			width:    15,
			expected: []string{"The quick", "brown fox", "jumps over the", "lazy dog"},
		},
		{
			name:     "No space to break",
			line:     "Superlongwordwithoutspaces",
			width:    10,
			expected: []string{"Superlongw", "ordwithout", "spaces"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapLine(tt.line, tt.width)
			if len(result) != len(tt.expected) {
				t.Errorf("wrapLine(%q, %d) returned %d lines, want %d", tt.line, tt.width, len(result), len(tt.expected))
				return
			}
			for i, line := range result {
				if line != tt.expected[i] {
					t.Errorf("wrapLine(%q, %d)[%d] = %q, want %q", tt.line, tt.width, i, line, tt.expected[i])
				}
			}
		})
	}
}

func TestFormatContent(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		maxWidth    int
		shouldWrap  bool
		preserveURL bool
	}{
		{
			name:       "Plain text gets wrapped",
			content:    "This is a very long error message that should be wrapped to fit within the specified width",
			maxWidth:   40,
			shouldWrap: true,
		},
		{
			name:        "URL is preserved",
			content:     "See: https://example.com/very/long/path/that/should/not/be/wrapped",
			maxWidth:    40,
			preserveURL: true,
		},
		{
			name:        "Mixed content",
			content:     "Error occurred\nSee: https://example.com/issue/123\nPlease try again",
			maxWidth:    40,
			preserveURL: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatContent(tt.content, tt.maxWidth)

			if tt.preserveURL {
				// URLs should be preserved intact
				if strings.Contains(tt.content, "https://") {
					for _, line := range strings.Split(tt.content, "\n") {
						if strings.Contains(line, "https://") {
							if !strings.Contains(result, line) {
								t.Errorf("URL line should be preserved: %q", line)
							}
						}
					}
				}
			}

			if tt.shouldWrap {
				lines := strings.Split(result, "\n")
				if len(lines) <= 1 && len(tt.content) > tt.maxWidth {
					t.Errorf("Content should have been wrapped")
				}
			}
		})
	}
}

func TestRenderResult(t *testing.T) {
	tests := []struct {
		name           string
		result         Result
		containsSymbol string
		containsTitle  bool
	}{
		{
			name:           "Success result",
			result:         Success("Operation completed", "All done"),
			containsSymbol: SymbolSuccess,
			containsTitle:  true,
		},
		{
			name:           "Error result",
			result:         Error("Operation failed", "Something went wrong"),
			containsSymbol: SymbolError,
			containsTitle:  true,
		},
		{
			name:           "Warning result",
			result:         Warning("Warning", "Be careful"),
			containsSymbol: SymbolWarning,
			containsTitle:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := RenderResult(tt.result)

			if !strings.Contains(output, tt.containsSymbol) {
				t.Errorf("Output should contain symbol %q", tt.containsSymbol)
			}

			if tt.containsTitle && !strings.Contains(output, tt.result.Title) {
				t.Errorf("Output should contain title %q", tt.result.Title)
			}

			if !strings.Contains(output, tt.result.Content) {
				t.Errorf("Output should contain content %q", tt.result.Content)
			}
		})
	}
}

func TestGetContentWidth(t *testing.T) {
	tests := []struct {
		name          string
		terminalWidth int
		expected      int
	}{
		{"Narrow terminal", 50, 46},
		{"Normal terminal", 100, 80},
		{"Wide terminal", 200, 80},
		{"Very narrow terminal", 30, 40}, // Minimum width
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetContentWidth(tt.terminalWidth)
			if result != tt.expected {
				t.Errorf("GetContentWidth(%d) = %d, want %d", tt.terminalWidth, result, tt.expected)
			}
		})
	}
}
