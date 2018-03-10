package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"os/user"

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

func MustExec(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	CheckAndExit(cmd.Run())
}

func TryExec(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func MustExecRtOut(name string, arg ...string) string {
	cmd := exec.Command(name, arg...)
	b, err := cmd.CombinedOutput()
	CheckAndExit(err)
	return string(b)
}

func CheckRoot() {
	u, err := user.Current()
	CheckAndExit(err)

	if u.Uid != "0" || u.Gid != "0" {
		fmt.Println("This command must be run as root! (sudo)")
		os.Exit(1)
	}
}
