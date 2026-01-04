package git

import (
	"testing"

	"github.com/mritd/gitflow-toolkit/v3/consts"
)

func TestParseBranchType(t *testing.T) {
	tests := []struct {
		name     string
		branch   string
		expected string
	}{
		// Standard formats with /
		{"feat with slash", "feat/login", consts.Feat},
		{"fix with slash", "fix/bug-123", consts.Fix},
		{"docs with slash", "docs/readme", consts.Docs},
		{"style with slash", "style/format", consts.Style},
		{"refactor with slash", "refactor/cleanup", consts.Refactor},
		{"test with slash", "test/unit", consts.Test},
		{"chore with slash", "chore/deps", consts.Chore},
		{"perf with slash", "perf/optimize", consts.Perf},
		{"hotfix with slash", "hotfix/urgent", consts.Hotfix},

		// Aliases with /
		{"feature alias", "feature/login", consts.Feat},
		{"bugfix alias", "bugfix/issue-42", consts.Fix},
		{"bug alias", "bug/crash", consts.Fix},
		{"doc alias", "doc/api", consts.Docs},
		{"document alias", "document/guide", consts.Docs},
		{"refact alias", "refact/module", consts.Refactor},
		{"testing alias", "testing/e2e", consts.Test},
		{"performance alias", "performance/cache", consts.Perf},

		// Formats with -
		{"feat with dash", "feat-login", consts.Feat},
		{"fix with dash", "fix-bug-123", consts.Fix},
		{"feature with dash", "feature-new-ui", consts.Feat},
		{"bugfix with dash", "bugfix-issue", consts.Fix},

		// Formats with _
		{"feat with underscore", "feat_login", consts.Feat},
		{"fix with underscore", "fix_bug_123", consts.Fix},

		// Case insensitivity
		{"uppercase FEAT", "FEAT/login", consts.Feat},
		{"mixed case Feature", "Feature/login", consts.Feat},

		// Non-matching branches
		{"main branch", "main", ""},
		{"master branch", "master", ""},
		{"develop branch", "develop", ""},
		{"release branch", "release/v1.0", ""},
		{"random branch", "random-branch", ""},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseBranchType(tt.branch)
			if result != tt.expected {
				t.Errorf("ParseBranchType(%q) = %q, want %q", tt.branch, result, tt.expected)
			}
		})
	}
}
