package util

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func Install() {

	home, err := homedir.Dir()
	CheckAndExit(err)

	toolHome := home + string(filepath.Separator) + ".gitflow-toolkit"
	dstPath := toolHome + string(filepath.Separator) + "gitflow-toolkit"
	hooksPath := toolHome + string(filepath.Separator) + "hooks"
	commitMessageHookPath := hooksPath + string(filepath.Separator) + "commit-msg"
	ciPath := "/usr/local/bin/git-ci"

	currentPath, err := exec.LookPath(os.Args[0])
	CheckAndExit(err)

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {

		fmt.Println("Clean old files")
		CheckAndExit(os.RemoveAll(toolHome))
		CheckAndExit(os.Remove(ciPath))

		fmt.Println("Create install home dir")
		CheckAndExit(os.MkdirAll(toolHome, 0755))

		fmt.Println("Copy file to home dir")
		currentFile, err := os.Open(currentPath)
		CheckAndExit(err)
		dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
		CheckAndExit(err)
		io.Copy(dstFile, currentFile)

		fmt.Println("Create symbolic file")
		CheckAndExit(os.MkdirAll(hooksPath, 0755))
		CheckAndExit(os.Symlink(dstPath, ciPath))
		CheckAndExit(os.Symlink(dstPath, commitMessageHookPath))

		fmt.Println("Config git")
		CheckAndExit(exec.Command("git", "config", "--global", "--unset", "core.hooksPath").Run())
		CheckAndExit(exec.Command("git", "config", "--global", "core.hooksPath", hooksPath).Run())

		fmt.Println("Well done.")

	}
}
