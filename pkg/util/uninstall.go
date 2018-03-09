package util

import (
	"fmt"
	"os"
	"os/exec"
)

func Uninstall() {

	fmt.Println("Uninstall gitflow-toolkit")
	os.RemoveAll(GitFlowToolKitHome)
	os.Remove(GitCIPath)
	exec.Command("git", "config", "--global", "--unset", "core.hooksPath").Run()
}
