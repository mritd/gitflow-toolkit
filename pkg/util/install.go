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

	currentPath, err := exec.LookPath(os.Args[0])
	CheckAndExit(err)

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {

		ciPath := "/usr/local/bin/git-ci"

		fmt.Println("Clean old files")
		os.RemoveAll(toolHome)
		os.Remove(ciPath)

		fmt.Println("Create install home dir")
		CheckAndExit(os.MkdirAll(toolHome, 0755))

		fmt.Println("Copy file to install home")
		currentFile, err := os.Open(currentPath)
		defer currentFile.Close()
		CheckAndExit(err)

		dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
		defer dstFile.Close()
		CheckAndExit(err)
		_, err = io.Copy(dstFile, currentFile)
		CheckAndExit(err)

		fmt.Println("Create symbolic file")
		CheckAndExit(os.MkdirAll(hooksPath, 0755))
		CheckAndExit(os.Symlink(dstPath, ciPath))
		CheckAndExit(os.Symlink(dstPath, commitMessageHookPath))

		fmt.Println("Config git")
		exec.Command("git", "config", "--global", "--unset", "core.hooksPath").Run()
		CheckAndExit(exec.Command("git", "config", "--global", "core.hooksPath", hooksPath).Run())

		fmt.Println("Well done.")

	} else if runtime.GOOS == "windows" {

		ciPath := toolHome + string(filepath.Separator) + "git-ci.exe"

		fmt.Println("Clean old files")
		os.RemoveAll(toolHome)
		os.Remove(ciPath)

		fmt.Println("Create install home dir")
		CheckAndExit(os.MkdirAll(toolHome, 0755))

		fmt.Println("Copy file to install home")
		currentFile, err := os.Open(currentPath)
		defer currentFile.Close()
		CheckAndExit(err)

		dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
		defer dstFile.Close()
		CheckAndExit(err)
		_, err = io.Copy(dstFile, currentFile)
		CheckAndExit(err)
		currentFile.Seek(0, 0)

		CheckAndExit(os.MkdirAll(hooksPath, 0755))
		commitMessageHookFile, err := os.OpenFile(commitMessageHookPath+".exe", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
		defer commitMessageHookFile.Close()
		_, err = io.Copy(commitMessageHookFile, currentFile)
		CheckAndExit(err)
		currentFile.Seek(0, 0)

		ciFile, err := os.OpenFile(ciPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
		CheckAndExit(err)
		defer ciFile.Close()
		_, err = io.Copy(ciFile, currentFile)
		CheckAndExit(err)
		currentFile.Seek(0, 0)

		fmt.Println("Config git")
		exec.Command("git", "config", "--global", "--unset", "core.hooksPath").Run()
		CheckAndExit(exec.Command("git", "config", "--global", "core.hooksPath", hooksPath).Run())

		fmt.Println("Config env")
		winPath := os.Getenv("Path")
		newPath := toolHome
		if winPath != "" {
			newPath += ";" + winPath
		}
		CheckAndExit(exec.Command("SETX", "Path", newPath).Run())

		fmt.Println("Well done.")
	}
}
