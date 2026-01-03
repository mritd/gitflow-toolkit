package git

import (
	"testing"
)

func TestCommitMessageString(t *testing.T) {
	tests := []struct {
		name     string
		msg      CommitMessage
		expected string
	}{
		{
			name: "basic message",
			msg: CommitMessage{
				Type:    "feat",
				Scope:   "api",
				Subject: "add new endpoint",
			},
			expected: "feat(api): add new endpoint\n",
		},
		{
			name: "with body",
			msg: CommitMessage{
				Type:    "fix",
				Scope:   "auth",
				Subject: "fix login bug",
				Body:    "This fixes the login issue\nwhen password contains special chars.",
			},
			expected: "fix(auth): fix login bug\n\nThis fixes the login issue\nwhen password contains special chars.\n",
		},
		{
			name: "with footer",
			msg: CommitMessage{
				Type:    "feat",
				Scope:   "core",
				Subject: "add feature",
				Footer:  "BREAKING CHANGE: API changed",
			},
			expected: "feat(core): add feature\n\nBREAKING CHANGE: API changed\n",
		},
		{
			name: "with SOB",
			msg: CommitMessage{
				Type:    "docs",
				Scope:   "readme",
				Subject: "update docs",
				SOB:     "Signed-off-by: Test User <test@example.com>",
			},
			expected: "docs(readme): update docs\n\nSigned-off-by: Test User <test@example.com>\n",
		},
		{
			name: "full message",
			msg: CommitMessage{
				Type:    "feat",
				Scope:   "api",
				Subject: "add user endpoint",
				Body:    "Add new user management endpoint.",
				Footer:  "Closes #123",
				SOB:     "Signed-off-by: Test User <test@example.com>",
			},
			expected: "feat(api): add user endpoint\n\nAdd new user management endpoint.\n\nCloses #123\n\nSigned-off-by: Test User <test@example.com>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.msg.String()
			if got != tt.expected {
				t.Errorf("CommitMessage.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}
