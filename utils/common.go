package utils

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

func init() {

	var err error

	home, err := homedir.Dir()
	CheckAndExit(err)

	GitFlowToolKitHome = filepath.Join(home, ".gitflow-toolkit")
	InstallPath = filepath.Join(GitFlowToolKitHome, "gitflow-toolkit")
	HooksPath = filepath.Join(GitFlowToolKitHome, "hooks")
	GitCMHookPath = filepath.Join(HooksPath, "commit-msg")

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

func MustExecRtOut(name string, arg ...string) string {
	cmd := exec.Command(name, arg...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	b, err := cmd.Output()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	return string(b)
}

func MustExecNoOut(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Stderr = os.Stderr
	CheckAndExit(cmd.Run())
}

func TryExec(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	return cmd.Run()
}

func Root() bool {
	u, err := user.Current()
	CheckAndExit(err)
	return u.Uid == "0" || u.Gid == "0"
}

func OSEditInput() string {

	f, err := ioutil.TempFile("", "gitflow-toolkit")
	CheckAndExit(err)
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
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
	MustExec(editor, f.Name())
	raw, err := ioutil.ReadFile(f.Name())
	CheckAndExit(err)
	input := string(bytes.TrimPrefix(raw, bom))

	return input
}

func CheckOS() {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		fmt.Println("Platform not support!")
		os.Exit(1)
	}
}
