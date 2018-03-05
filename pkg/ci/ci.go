package ci

import (
	"os/exec"
)

// 检查当前位置是否为 git 项目
func CheckGitProject() bool {
	return exec.Command("git", "rev-parse", "--show-toplevel").Run() == nil
}
