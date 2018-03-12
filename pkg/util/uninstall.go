package util

import (
	"fmt"
	"os"
)

func Uninstall() {

	CheckRoot()

	fmt.Println("Uninstall gitflow-toolkit")
	os.RemoveAll(GitFlowToolKitHome)
	for _, binPath := range *BinPaths() {
		os.Remove(binPath)
	}
	TryExec("git", "config", "--global", "--unset", "core.hooksPath")
}
