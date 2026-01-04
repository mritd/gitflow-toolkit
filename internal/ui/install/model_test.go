package install

import (
	"path/filepath"
	"testing"

	"github.com/mritd/gitflow-toolkit/v3/consts"
)

func TestNewPaths(t *testing.T) {
	tmpDir := t.TempDir()

	paths, err := NewPaths(tmpDir)
	if err != nil {
		t.Fatalf("NewPaths() error = %v", err)
	}

	if paths.InstallDir != tmpDir {
		t.Errorf("InstallDir = %q, want %q", paths.InstallDir, tmpDir)
	}

	if paths.Binary != filepath.Join(tmpDir, consts.BinaryName) {
		t.Errorf("Binary = %q, want %q", paths.Binary, filepath.Join(tmpDir, consts.BinaryName))
	}
}

func TestPaths_SymlinkPaths(t *testing.T) {
	tmpDir := t.TempDir()

	paths, err := NewPaths(tmpDir)
	if err != nil {
		t.Fatalf("NewPaths() error = %v", err)
	}

	symlinks := paths.SymlinkPaths()

	// Should have symlinks for all commands
	expectedCmds := consts.SymlinkCommands()
	if len(symlinks) != len(expectedCmds) {
		t.Errorf("SymlinkPaths() returned %d links, want %d", len(symlinks), len(expectedCmds))
	}

	// Check each symlink path
	for i, cmd := range expectedCmds {
		expected := filepath.Join(tmpDir, consts.GitCommandPrefix+cmd)
		if symlinks[i] != expected {
			t.Errorf("SymlinkPaths()[%d] = %q, want %q", i, symlinks[i], expected)
		}
	}
}

func TestInstallTasks(t *testing.T) {
	tmpDir := t.TempDir()

	paths, err := NewPaths(tmpDir)
	if err != nil {
		t.Fatalf("NewPaths() error = %v", err)
	}

	tasks := InstallTasks(paths)
	if len(tasks) == 0 {
		t.Error("InstallTasks() returned no tasks")
	}

	// Verify task names
	expectedNames := []string{
		"Remove existing symlinks",
		"Install binary",
		"Create command symlinks",
	}

	if len(tasks) != len(expectedNames) {
		t.Errorf("InstallTasks() returned %d tasks, want %d", len(tasks), len(expectedNames))
	}

	for i, task := range tasks {
		if i < len(expectedNames) && task.Name != expectedNames[i] {
			t.Errorf("tasks[%d].Name = %q, want %q", i, task.Name, expectedNames[i])
		}
	}
}

func TestUninstallTasks(t *testing.T) {
	tmpDir := t.TempDir()

	paths, err := NewPaths(tmpDir)
	if err != nil {
		t.Fatalf("NewPaths() error = %v", err)
	}

	tasks := UninstallTasks(paths)
	if len(tasks) == 0 {
		t.Error("UninstallTasks() returned no tasks")
	}

	expectedNames := []string{
		"Remove symlinks",
		"Remove binary",
	}

	if len(tasks) != len(expectedNames) {
		t.Errorf("UninstallTasks() returned %d tasks, want %d", len(tasks), len(expectedNames))
	}

	for i, task := range tasks {
		if i < len(expectedNames) && task.Name != expectedNames[i] {
			t.Errorf("tasks[%d].Name = %q, want %q", i, task.Name, expectedNames[i])
		}
	}
}

func TestNeedsSudo(t *testing.T) {
	// Test with temp directory (should not need sudo)
	tmpDir := t.TempDir()
	if NeedsSudo(tmpDir) {
		t.Error("NeedsSudo() should return false for temp directory")
	}
}

func TestGetRecommendedInstallDir(t *testing.T) {
	dir := GetRecommendedInstallDir()
	if dir == "" {
		t.Error("GetRecommendedInstallDir() returned empty string")
	}
}
