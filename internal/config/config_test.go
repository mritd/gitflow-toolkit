package config

import (
	"regexp"
	"testing"
)

func TestCommitTypeNames(t *testing.T) {
	names := CommitTypeNames()

	if len(names) != len(CommitTypes) {
		t.Errorf("CommitTypeNames() returned %d names, want %d", len(names), len(CommitTypes))
	}

	expected := []string{Feat, Fix, Docs, Style, Refactor, Test, Chore, Perf, Hotfix}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("CommitTypeNames()[%d] = %q, want %q", i, name, expected[i])
		}
	}
}

func TestCommitMessagePattern(t *testing.T) {
	reg := regexp.MustCompile(CommitMessagePattern)

	tests := []struct {
		name    string
		message string
		valid   bool
	}{
		{"valid feat", "feat(scope): add new feature", true},
		{"valid fix", "fix(api): fix null pointer", true},
		{"valid docs", "docs(readme): update installation", true},
		{"valid style", "style(lint): format code", true},
		{"valid refactor", "refactor(core): simplify logic", true},
		{"valid test", "test(unit): add more tests", true},
		{"valid chore", "chore(ci): update workflow", true},
		{"valid perf", "perf(db): optimize query", true},
		{"valid hotfix", "hotfix(auth): fix login bug", true},
		{"valid merge", "Merge branch 'main' into feature", true},
		{"invalid no type", "(scope): message", false},
		{"invalid no scope", "feat: message", false},
		{"invalid no colon", "feat(scope) message", false},
		{"invalid empty scope", "feat(): message", false},
		{"invalid empty subject", "feat(scope): ", false},
		{"invalid random", "random commit message", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := reg.FindStringSubmatch(tt.message)
			got := len(matches) == 4 || (tt.message[:5] == "Merge" && reg.MatchString(tt.message))

			// For merge commits, just check if it matches
			if tt.message[:5] == "Merge" {
				got = reg.MatchString(tt.message)
			} else {
				got = len(matches) == 4
			}

			if got != tt.valid {
				t.Errorf("pattern match for %q = %v, want %v", tt.message, got, tt.valid)
			}
		})
	}
}

func TestSymlinkCommands(t *testing.T) {
	cmds := SymlinkCommands()

	// Should contain CmdCommit and CmdPush
	hasCommit := false
	hasPush := false
	for _, cmd := range cmds {
		if cmd == CmdCommit {
			hasCommit = true
		}
		if cmd == CmdPush {
			hasPush = true
		}
	}

	if !hasCommit {
		t.Errorf("SymlinkCommands() should contain %q", CmdCommit)
	}
	if !hasPush {
		t.Errorf("SymlinkCommands() should contain %q", CmdPush)
	}

	// Should contain all commit types
	for _, ct := range CommitTypes {
		found := false
		for _, cmd := range cmds {
			if cmd == ct.Name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SymlinkCommands() should contain %q", ct.Name)
		}
	}
}
