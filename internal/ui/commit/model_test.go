package commit

import (
	"strings"
	"testing"

	"github.com/mritd/gitflow-toolkit/v2/internal/git"
	"github.com/mritd/gitflow-toolkit/v2/internal/ui/common"
)

func TestRenderPreview(t *testing.T) {
	tests := []struct {
		name     string
		msg      git.CommitMessage
		contains []string
	}{
		{
			name: "basic message",
			msg: git.CommitMessage{
				Type:    "feat",
				Scope:   "api",
				Subject: "add new endpoint",
			},
			contains: []string{"feat", "api", "add new endpoint"},
		},
		{
			name: "message with body",
			msg: git.CommitMessage{
				Type:    "fix",
				Scope:   "auth",
				Subject: "fix login bug",
				Body:    "This fixes the login issue.",
			},
			contains: []string{"fix", "auth", "fix login bug", "This fixes the login issue"},
		},
		{
			name: "message with footer",
			msg: git.CommitMessage{
				Type:    "feat",
				Scope:   "core",
				Subject: "add feature",
				Footer:  "BREAKING CHANGE: API changed",
			},
			contains: []string{"feat", "core", "add feature", "BREAKING CHANGE"},
		},
		{
			name: "full message",
			msg: git.CommitMessage{
				Type:    "feat",
				Scope:   "api",
				Subject: "add user endpoint",
				Body:    "Add new user management endpoint.",
				Footer:  "Closes #123",
				SOB:     "Signed-off-by: Test <test@example.com>",
			},
			contains: []string{"feat", "api", "add user endpoint", "Add new user management", "Closes #123", "Signed-off-by"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderPreview(tt.msg)

			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("renderPreview() should contain %q, got:\n%s", s, result)
				}
			}
		})
	}
}

func TestNewInputsModel(t *testing.T) {
	m := newInputsModel("feat")

	if m.title != "Commit Type: FEAT" {
		t.Errorf("title = %q, want 'Commit Type: FEAT'", m.title)
	}

	if len(m.inputs) != 4 {
		t.Errorf("len(inputs) = %d, want 4", len(m.inputs))
	}

	// First input should be focused
	if !m.inputs[0].input.Focused() {
		t.Error("first input should be focused")
	}
}

func TestNewSelectorModel(t *testing.T) {
	m := newSelectorModel()

	if m.list.Title != "Select Commit Type" {
		t.Errorf("list.Title = %q, want 'Select Commit Type'", m.list.Title)
	}

	// Should have 9 commit types
	if len(m.list.Items()) != 9 {
		t.Errorf("len(items) = %d, want 9", len(m.list.Items()))
	}
}

func TestResultStruct(t *testing.T) {
	// Test default values
	var result Result

	if result.Cancelled {
		t.Error("Result.Cancelled should be false by default")
	}

	if result.Err != nil {
		t.Error("Result.Err should be nil by default")
	}

	// Test with values
	result = Result{
		Cancelled: true,
		Err:       nil,
		Message: git.CommitMessage{
			Type:    "feat",
			Scope:   "test",
			Subject: "test subject",
		},
	}

	if !result.Cancelled {
		t.Error("Result.Cancelled should be true")
	}

	if result.Message.Type != "feat" {
		t.Errorf("Result.Message.Type = %q, want 'feat'", result.Message.Type)
	}
}

func TestCommonStylesUsedInCommit(t *testing.T) {
	// Verify the styles we use are available
	styles := []struct {
		name  string
		style string
	}{
		{"StyleTitle", common.StyleTitle.Render("test")},
		{"StyleMuted", common.StyleMuted.Render("test")},
		{"StylePrimary", common.StylePrimary.Render("test")},
		{"StyleDimmed", common.StyleDimmed.Render("test")},
	}

	for _, s := range styles {
		if s.style == "" {
			t.Errorf("%s.Render() returned empty string", s.name)
		}
	}
}
