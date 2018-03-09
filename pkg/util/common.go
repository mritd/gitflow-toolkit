package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

var GitFlowToolKitHome string
var InstallPath string
var HooksPath string
var GitCMHookPath string
var CurrentPath string
var GitCIPath = "/usr/local/bin/git-ci"

func init() {

	var err error

	home, err := homedir.Dir()
	CheckAndExit(err)

	GitFlowToolKitHome = home + string(filepath.Separator) + ".gitflow-toolkit"
	InstallPath = GitFlowToolKitHome + string(filepath.Separator) + "gitflow-toolkit"
	HooksPath = GitFlowToolKitHome + string(filepath.Separator) + "hooks"
	GitCMHookPath = HooksPath + string(filepath.Separator) + "commit-msg"

	CurrentPath, err = exec.LookPath(os.Args[0])
	CheckAndExit(err)
}

func CheckErr(err error) bool {
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func CheckAndExit(err error) {
	if !CheckErr(err) {
		os.Exit(1)
	}
}
