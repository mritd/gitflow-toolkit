package common

import (
	"strings"
	"testing"
)

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
		name       string
		content    string
		maxWidth   int
		shouldWrap bool
		checkURL   string // URL that should be preserved intact
	}{
		{
			name:       "Plain text gets wrapped",
			content:    "This is a very long error message that should be wrapped to fit within the specified width",
			maxWidth:   40,
			shouldWrap: true,
		},
		{
			name:       "URL is preserved intact",
			content:    "See: https://example.com/very/long/path/that/should/not/be/wrapped for details",
			maxWidth:   40,
			shouldWrap: true,
			checkURL:   "https://example.com/very/long/path/that/should/not/be/wrapped",
		},
		{
			name:       "Mixed content with URL",
			content:    "Error occurred at https://example.com/issue/123 please check",
			maxWidth:   30,
			shouldWrap: true,
			checkURL:   "https://example.com/issue/123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatContent(tt.content, tt.maxWidth)

			if tt.checkURL != "" {
				// URL should be preserved intact (not broken across lines)
				if !strings.Contains(result, tt.checkURL) {
					t.Errorf("URL should be preserved intact: %q\nGot: %q", tt.checkURL, result)
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

func TestSplitByURLs(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected []textSegment
	}{
		{
			name: "No URL",
			line: "Hello world",
			expected: []textSegment{
				{text: "Hello world", isURL: false},
			},
		},
		{
			name: "URL only",
			line: "https://example.com/path",
			expected: []textSegment{
				{text: "https://example.com/path", isURL: true},
			},
		},
		{
			name: "Text before URL",
			line: "See: https://example.com",
			expected: []textSegment{
				{text: "See: ", isURL: false},
				{text: "https://example.com", isURL: true},
			},
		},
		{
			name: "URL in middle",
			line: "Check https://example.com for info",
			expected: []textSegment{
				{text: "Check ", isURL: false},
				{text: "https://example.com", isURL: true},
				{text: " for info", isURL: false},
			},
		},
		{
			name: "Multiple URLs",
			line: "See https://a.com and https://b.com",
			expected: []textSegment{
				{text: "See ", isURL: false},
				{text: "https://a.com", isURL: true},
				{text: " and ", isURL: false},
				{text: "https://b.com", isURL: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitByURLs(tt.line)
			if len(result) != len(tt.expected) {
				t.Errorf("splitByURLs(%q) returned %d segments, want %d\nGot: %+v", tt.line, len(result), len(tt.expected), result)
				return
			}
			for i, seg := range result {
				if seg.text != tt.expected[i].text || seg.isURL != tt.expected[i].isURL {
					t.Errorf("splitByURLs(%q)[%d] = %+v, want %+v", tt.line, i, seg, tt.expected[i])
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
		{"Normal terminal", 100, 96},
		{"Wide terminal", 200, 120},      // MaxContentWidth = 120
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
