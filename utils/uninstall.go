package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Uninstall(dir string) {

	CheckOS()

	currentPath, err := exec.LookPath(os.Args[0])
	CheckAndExit(err)
	if !Root() {
		cmd := exec.Command("sudo", currentPath, "uninstall", "--dir", dir)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		CheckAndExit(cmd.Run())
	} else {
		fmt.Printf("👉 remove %s\n", GitFlowToolKitHome)
		_ = os.RemoveAll(GitFlowToolKitHome)

		var binPaths = []string{
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

		for _, binPath := range binPaths {
			fmt.Printf("👉 remove %s\n", binPath)
			_ = os.Remove(binPath)
		}
		_ = TryExec("git", "config", "--global", "--unset", "core.hooksPath")
	}

}
