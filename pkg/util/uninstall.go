package util

import (
	"fmt"
	"os"
)

func Uninstall() {

	fmt.Println("Uninstall gitflow-toolkit")
	os.RemoveAll(GitFlowToolKitHome)
	os.Remove(GitCIPath)
	TryExec("git", "config", "--global", "--unset", "core.hooksPath")
}
