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

const InstallBaseDir = "/usr/local/bin"

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

	GitFlowToolKitHome = home + "/.gitflow-toolkit"
	InstallPath = GitFlowToolKitHome + "/gitflow-toolkit"
	HooksPath = GitFlowToolKitHome + "/hooks"
	GitCMHookPath = HooksPath + "/commit-msg"

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
		InstallBaseDir + "/git-ci",
		InstallBaseDir + "/git-feat",
		InstallBaseDir + "/git-fix",
		InstallBaseDir + "/git-docs",
		InstallBaseDir + "/git-style",
		InstallBaseDir + "/git-refactor",
		InstallBaseDir + "/git-test",
		InstallBaseDir + "/git-chore",
		InstallBaseDir + "/git-perf",
		InstallBaseDir + "/git-hotfix",
		InstallBaseDir + "/git-xmr",
		InstallBaseDir + "/git-xpr",
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
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

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
	CheckAndExit(cmd.Run())
	raw, err := ioutil.ReadFile(f.Name())
	CheckAndExit(err)
	input := string(bytes.TrimPrefix(raw, bom))

	return input
}
