package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mritd/bubbles/progressbar"

	"github.com/mitchellh/go-homedir"
)

func install(dir string) error {
	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	toolKitHome := filepath.Join(home, ".gitflow-toolkit")
	toolKitPath := filepath.Join(dir, "gitflow-toolkit")
	//toolKitHooks := filepath.Join(toolKitHome, "hooks")

	m := &progressbar.Model{
		Width:       40,
		InitMessage: "Initializing, please wait...",
		Stages: []progressbar.ProgressFunc{
			func() (string, error) {
				err := os.RemoveAll(toolKitHome)
				if err != nil {
					return "", fmt.Errorf("💔 failed to remove dir: %s: %s", toolKitHome, err)
				}
				return "✔ Clean install dir...", nil
			},
			func() (string, error) {
				for _, link := range linkPath(dir) {
					if _, err := os.Lstat(link); err == nil {
						err := os.RemoveAll(link)
						if err != nil {
							return "", fmt.Errorf("💔 failed to remove symlink: %s: %s", link, err)
						}
					} else if !os.IsNotExist(err) {
						return "", fmt.Errorf("💔 failed to get symlink info: %s: %s", link, err)
					}
				}
				return "✔ Clean symlinks...", nil
			},
			func() (string, error) {
				// ignore unset failed error
				_ = gitCommand(ioutil.Discard, []string{"config", "--global", "--unset", "core.hooksPath"})
				return "✔ Unset commit hooks...", nil
			},
			func() (string, error) {
				err := os.MkdirAll(toolKitHome, 0755)
				if err != nil {
					return "", fmt.Errorf("💔 failed to create toolkit home: %s", err)
				}
				return "✔ Create toolkit home...", nil
			},
			func() (string, error) {
				binPath, err := exec.LookPath(os.Args[0])
				if err != nil {
					return "", fmt.Errorf("💔 failed to get bin file info: %s: %s", os.Args[0], err)
				}

				currentFile, err := os.Open(binPath)
				if err != nil {
					return "", fmt.Errorf("💔 failed to get bin file info: %s: %s", binPath, err)
				}
				defer func() { _ = currentFile.Close() }()

				installFile, err := os.OpenFile(filepath.Join(dir, "gitflow-toolkit"), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
				if err != nil {
					return "", fmt.Errorf("💔 failed to create bin file: %s: %s", filepath.Join(toolKitHome, "gitflow-toolkit"), err)
				}
				defer func() { _ = installFile.Close() }()

				_, err = io.Copy(installFile, currentFile)
				if err != nil {
					return "", fmt.Errorf("💔 failed to copy file: %s: %s", filepath.Join(toolKitHome, "gitflow-toolkit"), err)
				}
				return "✔ Install executable file...", nil
			},
			func() (string, error) {
				for _, link := range linkPath(dir) {
					err := os.Symlink(toolKitPath, link)
					if err != nil {
						return "", fmt.Errorf("💔 failed to create symlink: %s: %s", link, err)
					}
				}
				return "✔ Create symlink...", nil
			},
			//func() (string, error) {
			//	err := os.MkdirAll(toolKitHooks, 0755)
			//	if err != nil {
			//		return "", fmt.Errorf("💔 failed to create hooks dir: %s: %s", toolKitHooks, err)
			//	}
			//	err = os.Symlink(toolKitPath, filepath.Join(toolKitHooks, "commit-msg"))
			//	if err != nil {
			//		return "", fmt.Errorf("💔 failed to create commit hook synlink: %s: %s", filepath.Join(toolKitHooks, "commit-msg"), err)
			//	}
			//	err = gitCommand(ioutil.Discard, []string{"config", "--global", "core.hooksPath", toolKitHooks})
			//	if err != nil {
			//		return "", fmt.Errorf("💔 failed to set commit hooks: %s", err)
			//	}
			//	return "✔ Set commit hooks...", nil
			//},
			func() (string, error) {
				err := gitCommand(ioutil.Discard, []string{"test"})
				if err != nil {
					return "", fmt.Errorf("💔 install failed: %s", err)
				}
				return "✔ Install success...", nil
			},
		},
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

	m := &progressbar.Model{
		Width:       40,
		InitMessage: "Initializing, please wait...",
		Stages: []progressbar.ProgressFunc{
			func() (string, error) {
				err := os.RemoveAll(toolKitHome)
				if err != nil {
					return "", fmt.Errorf("💔 failed to remove dir: %s: %s", toolKitHome, err)
				}
				return "✔ Clean install dir...", nil
			},
			func() (string, error) {
				for _, link := range linkPath(dir) {
					if _, err := os.Lstat(link); err == nil {
						err := os.RemoveAll(link)
						if err != nil {
							return "", fmt.Errorf("💔 failed to remove symlink: %s: %s", link, err)
						}
					} else if !os.IsNotExist(err) {
						return "", fmt.Errorf("💔 failed to get symlink info: %s: %s", link, err)
					}
				}
				return "✔ Clean symlinks...", nil
			},
			func() (string, error) {
				err := os.RemoveAll(toolKitPath)
				if err != nil {
					return "", fmt.Errorf("💔 failed to remove bin file: %s: %s", toolKitPath, err)
				}
				return "✔ Clean bin file...", nil
			},
			func() (string, error) {
				// ignore unset failed error
				_ = gitCommand(ioutil.Discard, []string{"config", "--global", "--unset", "core.hooksPath"})
				return "✔ Unset commit hooks...", nil
			},
			func() (string, error) {
				err := gitCommand(ioutil.Discard, []string{"test"})
				if err == nil {
					return "", fmt.Errorf("💔 uninstall failed: %s", err)
				}
				return "✔ UnInstall success...", nil
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
