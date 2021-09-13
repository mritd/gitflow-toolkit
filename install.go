package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mitchellh/go-homedir"
)

func install(dir string) error {
	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	toolKitHome := filepath.Join(home, ".gitflow-toolkit")
	toolKitPath := filepath.Join(dir, "gitflow-toolkit")
	toolKitHooks := filepath.Join(toolKitHome, "hooks")
	links := linkPath(dir)

	m := stageModel{
		stages: []stage{
			{
				title: "Clean install dir...",
				f:     func() error { return os.RemoveAll(toolKitHome) },
			},
			{
				title: "Clean symlinks...",
				f: func() error {
					for _, link := range links {
						if _, err := os.Lstat(link); err == nil {
							err := os.RemoveAll(link)
							if err != nil {
								return fmt.Errorf("💔 failed to remove symlink: %s: %s", link, err)
							}
						} else if !os.IsNotExist(err) {
							return fmt.Errorf("💔 failed to get symlink info: %s: %s", link, err)
						}
					}
					return nil
				},
			},
			{
				title: "Unset commit hooks...",
				f: func() error {
					_, _ = git("config", "--global", "--unset", "core.hooksPath")
					return nil
				},
			},
			{
				title: "Create toolkit home...",
				f: func() error {
					return os.MkdirAll(toolKitHome, 0755)
				},
			},
			{
				title: "Install executable file...",
				f: func() error {
					binPath, err := exec.LookPath(os.Args[0])
					if err != nil {
						return fmt.Errorf("💔 failed to get bin file info: %s: %s", os.Args[0], err)
					}

					currentFile, err := os.Open(binPath)
					if err != nil {
						return fmt.Errorf("💔 failed to get bin file info: %s: %s", binPath, err)
					}
					defer func() { _ = currentFile.Close() }()

					installFile, err := os.OpenFile(filepath.Join(dir, "gitflow-toolkit"), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
					if err != nil {
						return fmt.Errorf("💔 failed to create bin file: %s: %s", filepath.Join(toolKitHome, "gitflow-toolkit"), err)
					}
					defer func() { _ = installFile.Close() }()

					_, err = io.Copy(installFile, currentFile)
					if err != nil {
						return fmt.Errorf("💔 failed to copy file: %s: %s", filepath.Join(toolKitHome, "gitflow-toolkit"), err)
					}
					return nil
				},
			},
			{
				title: "Create symlink...",
				f: func() error {
					for _, link := range links {
						err := os.Symlink(toolKitPath, link)
						if err != nil {
							return fmt.Errorf("💔 failed to create symlink: %s: %s", link, err)
						}
					}
					return nil
				},
			},
			{
				title: "Set commit hooks...",
				f: func() error {
					err := os.MkdirAll(toolKitHooks, 0755)
					if err != nil {
						return fmt.Errorf("💔 failed to create hooks dir: %s: %s", toolKitHooks, err)
					}
					err = os.Symlink(toolKitPath, filepath.Join(toolKitHooks, "commit-msg"))
					if err != nil {
						return fmt.Errorf("💔 failed to create commit hook synlink: %s: %s", filepath.Join(toolKitHooks, "commit-msg"), err)
					}
					_, _ = git("config", "--global", "core.hooksPath", toolKitHooks)
					return nil
				},
			},
			{
				title: "Install success...",
				f:     func() error { return nil },
			},
		},
		spinner: stageSpinner,
	}

	return tea.NewProgram(m).Start()
}

func uninstall(dir string) error {
	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	toolKitHome := filepath.Join(home, ".gitflow-toolkit")
	toolKitPath := filepath.Join(dir, "gitflow-toolkit")
	links := linkPath(dir)

	m := stageModel{
		spinner: stageSpinner,
		stages: []stage{
			{
				title: "Clean install dir...",
				f:     func() error { return os.RemoveAll(toolKitHome) },
			},
			{
				title: "Clean symlinks...",
				f: func() error {
					for _, link := range links {
						if _, err := os.Lstat(link); err == nil {
							err := os.RemoveAll(link)
							if err != nil {
								return fmt.Errorf("💔 failed to remove symlink: %s: %s", link, err)
							}
						} else if !os.IsNotExist(err) {
							return fmt.Errorf("💔 failed to get symlink info: %s: %s", link, err)
						}
					}
					return nil
				},
			},
			{
				title: "Clean bin file...",
				f: func() error {
					return os.Remove(toolKitPath)
				},
			},
			{
				title: "Unset commit hooks...",
				f: func() error {
					_, _ = git("config", "--global", "--unset", "core.hooksPath")
					return nil
				},
			},
			{
				title: "UnInstall success...",
				f: func() error {
					return nil
				},
			},
		},
	}
	return tea.NewProgram(m).Start()
}

func linkPath(dir string) []string {
	return []string{
		filepath.Join(dir, "git-ci"),
		filepath.Join(dir, "git-feat"),
		filepath.Join(dir, "git-fix"),
		filepath.Join(dir, "git-docs"),
		filepath.Join(dir, "git-style"),
		filepath.Join(dir, "git-refactor"),
		filepath.Join(dir, "git-test"),
		filepath.Join(dir, "git-chore"),
		filepath.Join(dir, "git-perf"),
		filepath.Join(dir, "git-hotfix"),
		filepath.Join(dir, "git-ps"),
	}
}
