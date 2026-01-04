package commit

import (
	"strings"
	"testing"

	"github.com/mritd/gitflow-toolkit/v3/internal/git"
	"github.com/mritd/gitflow-toolkit/v3/internal/ui/common"
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
	// Test without initial type
	m := newSelectorModel("")

	if m.list.Title != "Select Commit Type" {
		t.Errorf("list.Title = %q, want 'Select Commit Type'", m.list.Title)
	}

	// Should have 9 commit types
	if len(m.list.Items()) != 9 {
		t.Errorf("len(items) = %d, want 9", len(m.list.Items()))
	}

	// First item should be selected by default
	if m.list.Index() != 0 {
		t.Errorf("list.Index() = %d, want 0", m.list.Index())
	}
}

func TestNewSelectorModelWithInitialType(t *testing.T) {
	// Test with initial type "fix" (should be index 1)
	m := newSelectorModel("fix")

	if m.list.Index() != 1 {
		t.Errorf("list.Index() = %d, want 1 (fix)", m.list.Index())
	}

	// Test with initial type "docs" (should be index 2)
	m = newSelectorModel("docs")

	if m.list.Index() != 2 {
		t.Errorf("list.Index() = %d, want 2 (docs)", m.list.Index())
	}

	// Test with invalid type (should default to 0)
	m = newSelectorModel("invalid")

	if m.list.Index() != 0 {
		t.Errorf("list.Index() = %d, want 0 (default)", m.list.Index())
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
	}

	for _, s := range styles {
		if s.style == "" {
			t.Errorf("%s.Render() returned empty string", s.name)
		}
	}
}

func TestParseAIMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    git.CommitMessage
	}{
		{
			name:    "standard format",
			message: "feat(api): add user endpoint\n\nAdd new user management endpoint.",
			want: git.CommitMessage{
				Type:    "feat",
				Scope:   "api",
				Subject: "add user endpoint",
				Body:    "Add new user management endpoint.",
			},
		},
		{
			name:    "type without scope",
			message: "fix: correct typo",
			want: git.CommitMessage{
				Type:    "fix",
				Scope:   "general", // default scope when not provided
				Subject: "correct typo",
				Body:    "",
			},
		},
		{
			name:    "header only",
			message: "docs(readme): update installation guide",
			want: git.CommitMessage{
				Type:    "docs",
				Scope:   "readme",
				Subject: "update installation guide",
				Body:    "",
			},
		},
		{
			name:    "multiline body",
			message: "refactor(core): simplify logic\n\nRemoved redundant code.\nImproved performance.",
			want: git.CommitMessage{
				Type:    "refactor",
				Scope:   "core",
				Subject: "simplify logic",
				Body:    "Removed redundant code.\nImproved performance.",
			},
		},
		{
			name:    "plain text fallback",
			message: "update readme",
			want: git.CommitMessage{
				Type:    "feat",
				Scope:   "general",
				Subject: "update readme",
				Body:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAIMessage(tt.message)
			// Note: SOB is auto-generated, so we don't compare it
			if got.Type != tt.want.Type {
				t.Errorf("Type = %q, want %q", got.Type, tt.want.Type)
			}
			if got.Scope != tt.want.Scope {
				t.Errorf("Scope = %q, want %q", got.Scope, tt.want.Scope)
			}
			if got.Subject != tt.want.Subject {
				t.Errorf("Subject = %q, want %q", got.Subject, tt.want.Subject)
			}
			if got.Body != tt.want.Body {
				t.Errorf("Body = %q, want %q", got.Body, tt.want.Body)
			}
		})
	}
}

func TestParseHeader(t *testing.T) {
	tests := []struct {
		name        string
		header      string
		wantType    string
		wantScope   string
		wantSubject string
	}{
		{
			name:        "full format",
			header:      "feat(api): add endpoint",
			wantType:    "feat",
			wantScope:   "api",
			wantSubject: "add endpoint",
		},
		{
			name:        "no scope",
			header:      "fix: bug fix",
			wantType:    "fix",
			wantScope:   "general", // default scope when not provided
			wantSubject: "bug fix",
		},
		{
			name:        "nested scope",
			header:      "refactor(ui/commit): simplify",
			wantType:    "refactor",
			wantScope:   "ui/commit",
			wantSubject: "simplify",
		},
		{
			name:        "plain text",
			header:      "update something",
			wantType:    "feat",
			wantScope:   "general",
			wantSubject: "update something",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotScope, gotSubject := parseHeader(tt.header)
			if gotType != tt.wantType {
				t.Errorf("type = %q, want %q", gotType, tt.wantType)
			}
			if gotScope != tt.wantScope {
				t.Errorf("scope = %q, want %q", gotScope, tt.wantScope)
			}
			if gotSubject != tt.wantSubject {
				t.Errorf("subject = %q, want %q", gotSubject, tt.wantSubject)
			}
		})
	}
}

func TestCalcVisibleRange(t *testing.T) {
	tests := []struct {
		name       string
		fileCount  int
		status     []int // 0=pending, 1=running, 2=done
		maxVisible int
		wantStart  int
		wantEnd    int
	}{
		{
			name:       "fewer files than max",
			fileCount:  5,
			status:     []int{2, 2, 1, 0, 0},
			maxVisible: 10,
			wantStart:  0,
			wantEnd:    5,
		},
		{
			name:       "running at start",
			fileCount:  15,
			status:     []int{1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			maxVisible: 10,
			wantStart:  0,
			wantEnd:    10,
		},
		{
			name:       "running in middle",
			fileCount:  15,
			status:     []int{2, 2, 2, 2, 2, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0},
			maxVisible: 10,
			wantStart:  3, // 5 - 2 (contextAbove)
			wantEnd:    13,
		},
		{
			name:       "running near end",
			fileCount:  15,
			status:     []int{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1, 0, 0},
			maxVisible: 10,
			wantStart:  5, // 15 - 10
			wantEnd:    15,
		},
		{
			name:       "all done shows from start",
			fileCount:  15,
			status:     []int{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
			maxVisible: 10,
			wantStart:  0,
			wantEnd:    10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := aiModel{
				files:      make([]git.FileDiff, tt.fileCount),
				fileStatus: tt.status,
			}
			gotStart, gotEnd := m.calcVisibleRange(tt.maxVisible)
			if gotStart != tt.wantStart || gotEnd != tt.wantEnd {
				t.Errorf("calcVisibleRange() = (%d, %d), want (%d, %d)", gotStart, gotEnd, tt.wantStart, tt.wantEnd)
			}
		})
	}
}
