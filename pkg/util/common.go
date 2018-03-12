package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"os/user"

	"github.com/mitchellh/go-homedir"
)

var GitFlowToolKitHome string
var InstallPath string
var HooksPath string
var GitCMHookPath string
var CurrentPath string
var CurrentDir string
var WorkingDir string

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
	CurrentDir, err = filepath.Abs(filepath.Dir(os.Args[0]))
	CheckAndExit(err)
	WorkingDir, err = os.Getwd()
	CheckAndExit(err)
}

func BinPaths() *[]string {
	return &[]string{
		GitCMHookPath,
		"/usr/local/bin/git-ci",
		"/usr/local/bin/git-feat",
		"/usr/local/bin/git-fix",
		"/usr/local/bin/git-docs",
		"/usr/local/bin/git-style",
		"/usr/local/bin/git-refactor",
		"/usr/local/bin/git-test",
		"/usr/local/bin/git-chore",
		"/usr/local/bin/git-perf",
		"/usr/local/bin/git-xmr",
		"/usr/local/bin/git-xpr",
	}
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
	b, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(b))
		os.Exit(1)
	}
}

func MustExecRtOut(name string, arg ...string) string {
	cmd := exec.Command(name, arg...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(b))
		os.Exit(1)
	}
	return string(b)
}

func MustExecNoOut(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	CheckAndExit(cmd.Run())
}

func TryExec(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func CheckRoot() {
	u, err := user.Current()
	CheckAndExit(err)

	if u.Uid != "0" || u.Gid != "0" {
		fmt.Println("This command must be run as root! (sudo)")
		os.Exit(1)
	}
}

func OSEditInput() string {

	f, err := ioutil.TempFile("", "gitflow-toolkit")
	CheckAndExit(err)
	defer os.Remove(f.Name())

	// write utf8 bom
	bom := []byte{0xef, 0xbb, 0xbf}
	_, err = f.Write(bom)
	CheckAndExit(err)

	// 获取系统编辑器
	editor := "vim"
	if runtime.GOOS == "windows" {
		editor = "notepad"
	}
	if v := os.Getenv("VISUAL"); v != "" {
		editor = v
	} else if e := os.Getenv("EDITOR"); e != "" {
		editor = e
	}

	// 执行编辑文件
	cmd := exec.Command(editor, f.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	CheckAndExit(cmd.Run())
	raw, err := ioutil.ReadFile(f.Name())
	CheckAndExit(err)
	input := string(bytes.TrimPrefix(raw, bom))

	return input
}
