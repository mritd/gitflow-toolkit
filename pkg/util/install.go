package util

import (
	"fmt"
	"io"
	"os"
)

func Install() {

	Uninstall()

	fmt.Println("Install gitflow-toolkit")
	fmt.Println("Create install home dir")
	CheckAndExit(os.MkdirAll(GitFlowToolKitHome, 0755))

	fmt.Println("Copy file to install home")
	currentFile, err := os.Open(CurrentPath)
	CheckAndExit(err)
	defer currentFile.Close()

	installFile, err := os.OpenFile(InstallPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	CheckAndExit(err)
	defer installFile.Close()

	_, err = io.Copy(installFile, currentFile)
	CheckAndExit(err)

	fmt.Println("Create symbolic file")
	CheckAndExit(os.MkdirAll(HooksPath, 0755))

	for _, binPath := range *BinPaths() {
		CheckAndExit(os.Symlink(InstallPath, binPath))
	}

	fmt.Println("Config git")
	MustExec("git", "config", "--global", "core.hooksPath", HooksPath)

}
