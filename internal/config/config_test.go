package config

import (
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
