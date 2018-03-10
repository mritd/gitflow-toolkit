package util

import (
	"strings"

	"github.com/tsuyoshiwada/go-gitcmd"
)

func Git() gitcmd.Client {
	return gitcmd.New(nil)
}

// 检查当前位置是否为 git 项目
func CheckGitProject() bool {
	_, err := Git().Exec("rev-parse", "--show-toplevel")
	return err == nil
}

// 检测暂存区是否有文件
func CheckStagedFiles() bool {
	output, _ := Git().Exec("diff", "--cached", "--name-only")
	return strings.Replace(output, " ", "", -1) != ""
}
