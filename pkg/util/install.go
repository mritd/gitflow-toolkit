package util

import (
	"fmt"
	"io"
	"os"
	"runtime"
)

func Install() {

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {

		Uninstall()

		fmt.Println("Install gitflow-toolkit")
		fmt.Println("Create install home dir")
		CheckAndExit(os.MkdirAll(GitFlowToolKitHome, 0755))

		fmt.Println("Copy file to install home")
		currentFile, err := os.Open(CurrentPath)
		defer currentFile.Close()
		CheckAndExit(err)

		installFile, err := os.OpenFile(InstallPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
		defer installFile.Close()
		CheckAndExit(err)
		_, err = io.Copy(installFile, currentFile)
		CheckAndExit(err)

		fmt.Println("Create symbolic file")
		CheckAndExit(os.MkdirAll(HooksPath, 0755))
		CheckAndExit(os.Symlink(InstallPath, GitCIPath))
		CheckAndExit(os.Symlink(InstallPath, GitCMHookPath))

		fmt.Println("Config git")
		MustExec("git", "config", "--global", "core.hooksPath", HooksPath)

	} else {
		fmt.Println("Platform not support!")
		os.Exit(1)
	}
}
