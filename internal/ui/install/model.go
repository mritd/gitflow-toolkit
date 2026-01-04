// Package install provides the TUI for installation tasks.
package install

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mritd/gitflow-toolkit/v3/consts"
	"github.com/mritd/gitflow-toolkit/v3/internal/ui/common"
)

// Paths holds all installation paths.
type Paths struct {
	Binary     string // /usr/local/bin/gitflow-toolkit
	InstallDir string // /usr/local/bin
}

// NewPaths creates installation paths.
func NewPaths(installDir string) (*Paths, error) {
	return &Paths{
		Binary:     filepath.Join(installDir, consts.BinaryName),
		InstallDir: installDir,
	}, nil
}

// SymlinkPaths returns all symlink paths.
func (p *Paths) SymlinkPaths() []string {
	var links []string
	for _, cmd := range consts.SymlinkCommands() {
		links = append(links, filepath.Join(p.InstallDir, consts.GitCommandPrefix+cmd))
	}
	return links
}

// InstallTasks returns all installation tasks.
func InstallTasks(paths *Paths) []common.Task {
	return []common.Task{
		{
			Name: "Remove existing symlinks",
			Run: func() error {
				for _, link := range paths.SymlinkPaths() {
					if _, err := os.Lstat(link); err == nil {
						if err := os.Remove(link); err != nil {
							return fmt.Errorf("failed to remove %s: %w", filepath.Base(link), err)
						}
					} else if !os.IsNotExist(err) {
						return fmt.Errorf("failed to check %s: %w", filepath.Base(link), err)
					}
				}
				return nil
			},
		},
		{
			Name: "Install binary",
			Run: func() error {
				binPath, err := exec.LookPath(os.Args[0])
				if err != nil {
					return fmt.Errorf("failed to locate current binary: %w", err)
				}

				src, err := os.Open(binPath)
				if err != nil {
					return fmt.Errorf("failed to open source binary: %w", err)
				}
				defer func() { _ = src.Close() }()

				dst, err := os.OpenFile(paths.Binary, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
				if err != nil {
					return fmt.Errorf("failed to create target binary: %w", err)
				}
				defer func() { _ = dst.Close() }()

				if _, err := io.Copy(dst, src); err != nil {
					return fmt.Errorf("failed to copy binary: %w", err)
				}
				return nil
			},
		},
		{
			Name: "Create command symlinks",
			Run: func() error {
				for _, link := range paths.SymlinkPaths() {
					if err := os.Symlink(paths.Binary, link); err != nil {
						return fmt.Errorf("failed to create %s: %w", filepath.Base(link), err)
					}
				}
				return nil
			},
		},
	}
}

// UninstallTasks returns all uninstallation tasks.
func UninstallTasks(paths *Paths) []common.Task {
	return []common.Task{
		{
			Name: "Remove symlinks",
			Run: func() error {
				for _, link := range paths.SymlinkPaths() {
					if _, err := os.Lstat(link); err == nil {
						if err := os.Remove(link); err != nil {
							return fmt.Errorf("failed to remove %s: %w", filepath.Base(link), err)
						}
					}
				}
				return nil
			},
		},
		{
			Name: "Remove binary",
			Run: func() error {
				if err := os.Remove(paths.Binary); err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("failed to remove binary: %w", err)
				}
				return nil
			},
		},
	}
}

// NeedsSudo checks if the installation directory requires elevated privileges.
func NeedsSudo(installDir string) bool {
	testFile := filepath.Join(installDir, "."+consts.TempFilePrefix+"-test")
	f, err := os.Create(testFile)
	if err != nil {
		return true
	}
	_ = f.Close()
	_ = os.Remove(testFile)
	return false
}

// GetRecommendedInstallDir returns a recommended installation directory.
func GetRecommendedInstallDir() string {
	homeDir := os.Getenv("HOME")
	candidates := []string{
		consts.DefaultInstallDir,
		filepath.Join(homeDir, ".local", "bin"),
		filepath.Join(homeDir, "bin"),
	}

	for _, dir := range candidates {
		if !NeedsSudo(dir) {
			return dir
		}
	}

	return consts.DefaultInstallDir
}
