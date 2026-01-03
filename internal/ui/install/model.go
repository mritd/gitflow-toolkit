// Package install provides the TUI for installation tasks.
package install

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/mritd/gitflow-toolkit/v3/internal/config"
	"github.com/mritd/gitflow-toolkit/v3/internal/ui/common"
)

// Paths holds all installation paths.
type Paths struct {
	Home       string // ~/.gitflow-toolkit (real user's home)
	Binary     string // /usr/local/bin/gitflow-toolkit
	HooksDir   string // ~/.gitflow-toolkit/hooks
	InstallDir string // /usr/local/bin
	RealUser   *user.User
}

// NewPaths creates installation paths.
// It correctly handles sudo by using SUDO_USER to find the real user's home.
func NewPaths(installDir string) (*Paths, error) {
	realUser, err := getRealUser()
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	toolkitHome := filepath.Join(realUser.HomeDir, config.ToolkitHomeDir)
	return &Paths{
		Home:       toolkitHome,
		Binary:     filepath.Join(installDir, config.BinaryName),
		HooksDir:   filepath.Join(toolkitHome, config.HooksDir),
		InstallDir: installDir,
		RealUser:   realUser,
	}, nil
}

// getRealUser returns the real user, even when running under sudo.
func getRealUser() (*user.User, error) {
	// Check if running under sudo
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		return user.Lookup(sudoUser)
	}

	// Not running under sudo, use current user
	return user.Current()
}

// SymlinkPaths returns all symlink paths.
func (p *Paths) SymlinkPaths() []string {
	var links []string
	for _, cmd := range config.SymlinkCommands() {
		links = append(links, filepath.Join(p.InstallDir, config.GitCommandPrefix+cmd))
	}
	return links
}

// chownToRealUser changes ownership of a path to the real user (not root when using sudo).
func (p *Paths) chownToRealUser(path string) error {
	if p.RealUser == nil {
		return nil
	}

	// Only need to chown if running as root
	if os.Geteuid() != 0 {
		return nil
	}

	uid, err := strconv.Atoi(p.RealUser.Uid)
	if err != nil {
		return fmt.Errorf("invalid uid: %w", err)
	}

	gid, err := strconv.Atoi(p.RealUser.Gid)
	if err != nil {
		return fmt.Errorf("invalid gid: %w", err)
	}

	return chownRecursive(path, uid, gid)
}

// chownRecursive changes ownership recursively.
func chownRecursive(path string, uid, gid int) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return os.Chown(name, uid, gid)
	})
}

// InstallTasks returns all installation tasks.
func InstallTasks(paths *Paths, withHook bool) []common.Task {
	return []common.Task{
		{
			Name: "Clean existing installation",
			Run: func() error {
				if err := os.RemoveAll(paths.Home); err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("failed to remove %s: %w", paths.Home, err)
				}
				return nil
			},
		},
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
			Name: "Unset global git hooks",
			Run: func() error {
				// Run as real user to modify their git config
				// Ignore error if config doesn't exist
				_ = runAsRealUser(paths.RealUser, "git", "config", "--global", "--unset", "core.hooksPath")
				return nil
			},
		},
		{
			Name: "Create toolkit directory",
			Run: func() error {
				if err := os.MkdirAll(paths.Home, 0755); err != nil {
					return fmt.Errorf("failed to create directory: %w", err)
				}
				return paths.chownToRealUser(paths.Home)
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
		{
			Name: "Configure git hooks",
			Run: func() error {
				if !withHook {
					return common.WarnErr{Msg: "hook not installed (use --hook to enable)"}
				}

				if err := os.MkdirAll(paths.HooksDir, 0755); err != nil {
					return fmt.Errorf("failed to create hooks directory: %w", err)
				}
				if err := paths.chownToRealUser(paths.HooksDir); err != nil {
					return fmt.Errorf("failed to set hooks directory ownership: %w", err)
				}

				hookPath := filepath.Join(paths.HooksDir, config.CommitMsgHook)
				if err := os.Symlink(paths.Binary, hookPath); err != nil {
					return fmt.Errorf("failed to create hook symlink: %w", err)
				}

				if err := runAsRealUser(paths.RealUser, "git", "config", "--global", "core.hooksPath", paths.HooksDir); err != nil {
					return fmt.Errorf("failed to configure git hooks: %w", err)
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
			Name: "Remove toolkit directory",
			Run: func() error {
				if err := os.RemoveAll(paths.Home); err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("failed to remove toolkit directory: %w", err)
				}
				return nil
			},
		},
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
		{
			Name: "Unset global git hooks",
			Run: func() error {
				_ = runAsRealUser(paths.RealUser, "git", "config", "--global", "--unset", "core.hooksPath")
				return nil
			},
		},
	}
}

// runAsRealUser runs a command as the real user (not root) when under sudo.
func runAsRealUser(realUser *user.User, name string, args ...string) error {
	cmd := exec.Command(name, args...)

	// If running as root and we have a real user, switch to that user
	if os.Geteuid() == 0 && realUser != nil {
		uid, err := strconv.Atoi(realUser.Uid)
		if err != nil {
			return fmt.Errorf("invalid uid %q: %w", realUser.Uid, err)
		}
		gid, err := strconv.Atoi(realUser.Gid)
		if err != nil {
			return fmt.Errorf("invalid gid %q: %w", realUser.Gid, err)
		}

		cmd.SysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{
				Uid: uint32(uid),
				Gid: uint32(gid),
			},
		}
		cmd.Env = append(os.Environ(),
			"HOME="+realUser.HomeDir,
			"USER="+realUser.Username,
		)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, string(output))
	}
	return nil
}

// NeedsSudo checks if the installation directory requires elevated privileges.
func NeedsSudo(installDir string) bool {
	testFile := filepath.Join(installDir, "."+config.TempFilePrefix+"-test")
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
		config.DefaultInstallDir,
		filepath.Join(homeDir, ".local", "bin"),
		filepath.Join(homeDir, "bin"),
	}

	for _, dir := range candidates {
		if !NeedsSudo(dir) {
			return dir
		}
	}

	return config.DefaultInstallDir
}
