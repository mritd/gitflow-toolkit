package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func Install(dir string) {

	currentPath, err := exec.LookPath(os.Args[0])
	CheckAndExit(err)
	if !Root() {
		cmd := exec.Command("sudo", currentPath, "install", "--dir", dir)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		CheckAndExit(cmd.Run())
	} else {
		Uninstall(dir)

		fmt.Printf("📥 mkdir %s\n", HooksPath)
		CheckAndExit(os.MkdirAll(HooksPath, 0755))

		fmt.Printf("📥 copy file %s\n", InstallPath)
		currentFile, err := os.Open(CurrentPath)
		CheckAndExit(err)
		defer func() { _ = currentFile.Close() }()

		installFile, err := os.OpenFile(InstallPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
		CheckAndExit(err)
		defer func() { _ = installFile.Close() }()

		_, err = io.Copy(installFile, currentFile)
		CheckAndExit(err)

		var binPaths = []string{
			GitCMHookPath,
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
			fmt.Printf("📥 install symbolic %s\n", binPath)
			CheckAndExit(os.Symlink(InstallPath, binPath))
		}

		fmt.Printf("📥 config set core.hooksPath %s\n", HooksPath)
		MustExec("git", "config", "--global", "core.hooksPath", HooksPath)
	}
}
