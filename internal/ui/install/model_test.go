package install

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mritd/gitflow-toolkit/v2/internal/config"
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

	if paths.Binary != filepath.Join(tmpDir, config.BinaryName) {
		t.Errorf("Binary = %q, want %q", paths.Binary, filepath.Join(tmpDir, config.BinaryName))
	}

	// Home should be in user's home directory
	home, _ := os.UserHomeDir()
	expectedHome := filepath.Join(home, ".gitflow-toolkit")
	if paths.Home != expectedHome {
		t.Errorf("Home = %q, want %q", paths.Home, expectedHome)
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
	expectedCmds := config.SymlinkCommands()
	if len(symlinks) != len(expectedCmds) {
		t.Errorf("SymlinkPaths() returned %d links, want %d", len(symlinks), len(expectedCmds))
	}

	// Check each symlink path
	for i, cmd := range expectedCmds {
		expected := filepath.Join(tmpDir, config.GitCommandPrefix+cmd)
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

	// Test without hook
	tasks := InstallTasks(paths, false)
	if len(tasks) == 0 {
		t.Error("InstallTasks() returned no tasks")
	}

	// Verify task names
	expectedNames := []string{
		"Clean existing installation",
		"Remove existing symlinks",
		"Unset global git hooks",
		"Create toolkit directory",
		"Install binary",
		"Create command symlinks",
		"Configure git hooks",
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
		"Remove toolkit directory",
		"Remove symlinks",
		"Remove binary",
		"Unset global git hooks",
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

	// Test with /usr/local/bin (may need sudo depending on system)
	// We can't reliably test this as it depends on system configuration
}

func TestGetRecommendedInstallDir(t *testing.T) {
	dir := GetRecommendedInstallDir()
	if dir == "" {
		t.Error("GetRecommendedInstallDir() returned empty string")
	}
}

func TestInstallTasks_CreateToolkitDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	installDir := filepath.Join(tmpDir, "bin")
	homeDir := filepath.Join(tmpDir, "home")

	// Create custom paths for testing
	paths := &Paths{
		Home:       filepath.Join(homeDir, ".gitflow-toolkit"),
		Binary:     filepath.Join(installDir, config.BinaryName),
		HooksDir:   filepath.Join(homeDir, ".gitflow-toolkit", "hooks"),
		InstallDir: installDir,
	}

	tasks := InstallTasks(paths, false)

	// Find and run the "Create toolkit directory" task
	for _, task := range tasks {
		if task.Name == "Create toolkit directory" {
			if err := task.Run(); err != nil {
				t.Errorf("Create toolkit directory failed: %v", err)
			}

			// Verify directory was created
			if _, err := os.Stat(paths.Home); os.IsNotExist(err) {
				t.Error("Toolkit directory was not created")
			}
			break
		}
	}
}

func TestUninstallTasks_RemoveToolkitDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, ".gitflow-toolkit")

	// Create directory first
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	paths := &Paths{
		Home:       homeDir,
		Binary:     filepath.Join(tmpDir, "bin", config.BinaryName),
		HooksDir:   filepath.Join(homeDir, "hooks"),
		InstallDir: filepath.Join(tmpDir, "bin"),
	}

	tasks := UninstallTasks(paths)

	// Find and run the "Remove toolkit home" task
	for _, task := range tasks {
		if task.Name == "Remove toolkit home" {
			if err := task.Run(); err != nil {
				t.Errorf("Remove toolkit home failed: %v", err)
			}

			// Verify directory was removed
			if _, err := os.Stat(homeDir); !os.IsNotExist(err) {
				t.Error("Toolkit directory was not removed")
			}
			break
		}
	}
}
