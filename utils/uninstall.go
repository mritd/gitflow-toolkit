package utils

import (
	"fmt"
	"os"
)

func Uninstall() {

	CheckOS()
	CheckRoot()

	fmt.Println("Uninstall gitflow-toolkit")
	_ = os.RemoveAll(GitFlowToolKitHome)
	for _, binPath := range BinPaths() {
		_ = os.Remove(binPath)
	}
	_ = TryExec("git", "config", "--global", "--unset", "core.hooksPath")
}
