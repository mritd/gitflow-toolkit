// Package git provides Git command execution utilities.
package git

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/mritd/gitflow-toolkit/v3/config"
)

// Common errors.
var (
	// ErrNoStagedFiles is returned when there are no staged files.
	ErrNoStagedFiles = errors.New("there is no file to commit, please execute the `git add` command to add the commit file")

	// ErrNotGitRepo is returned when the current directory is not a git repository.
	ErrNotGitRepo = errors.New("not a git repository")
)

// Run executes a git command with the given arguments.
func Run(args ...string) (string, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("git.exe", args...)
	} else {
		cmd = exec.Command("git", args...)
	}

	// Disable strict host key checking unless explicitly enabled via gitconfig
	if !config.GetBool(config.GitConfigSSHStrictHost, false) {
		cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=no")
	}

	bs, err := cmd.CombinedOutput()
	if err != nil {
		if bs != nil {
			return "", errors.New(strings.TrimSpace(string(bs)))
		}
		return "", err
	}

	return strings.TrimSpace(string(bs)), nil
}

// RepoCheck checks if the current directory is a git repository.
func RepoCheck() error {
	_, err := Run("rev-parse", "--show-toplevel")
	if err != nil {
		return ErrNotGitRepo
	}
	return nil
}

// CurrentBranch returns the current branch name.
func CurrentBranch() (string, error) {
	return Run("symbolic-ref", "--short", "HEAD")
}

// Author returns the git user name and email.
// Returns empty strings if not configured.
func Author() (name, email string) {
	if cfg, err := Run("config", "user.name"); err == nil {
		name = cfg
	}
	if cfg, err := Run("config", "user.email"); err == nil {
		email = cfg
	}
	return name, email
}

// HasStagedFiles checks if there are any staged files.
func HasStagedFiles() error {
	msg, err := Run("diff", "--cached", "--name-only")
	if err != nil {
		return err
	}
	if msg == "" {
		return ErrNoStagedFiles
	}
	return nil
}
