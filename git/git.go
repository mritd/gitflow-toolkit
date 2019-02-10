package git

import (
	"strings"

	"github.com/mritd/gitflow-toolkit/utils"

	"fmt"

	"github.com/mritd/gitflow-toolkit/consts"
)

func CheckGitProject() {
	utils.MustExecNoOut(consts.GitCmd, "rev-parse", "--show-toplevel")
}

func CheckStagedFiles() bool {
	output := utils.MustExecRtOut(consts.GitCmd, "diff", "--cached", "--name-only")
	return strings.TrimSpace(output) != ""
}

func GetLastCommitInfo() *[]string {
	title := utils.MustExecRtOut(consts.GitCmd, "log", "-1", "--pretty=format:%s")
	desc := utils.MustExecRtOut(consts.GitCmd, "log", "-1", "--pretty=format:%b")
	return &[]string{title, desc}
}

func GetCurrentBranch() string {
	// 1.8+ git symbolic-ref --short HEAD
	return strings.TrimSpace(utils.MustExecRtOut(consts.GitCmd, "rev-parse", "--abbrev-ref", "HEAD"))
}

func Rebase(sourceBranch string, targetBranch string) {
	fmt.Println("checkout branch:", targetBranch)
	utils.MustExec(consts.GitCmd, "checkout", targetBranch)
	fmt.Println("pull origin branch:", targetBranch)
	utils.MustExec(consts.GitCmd, "pull", "origin", targetBranch)
	fmt.Println("checkout branch:", sourceBranch)
	utils.MustExec(consts.GitCmd, "checkout", sourceBranch)
	fmt.Println("exec git rebase")
	utils.MustExec(consts.GitCmd, "rebase", targetBranch)
	fmt.Println("push", sourceBranch, "to origin")
	utils.MustExec(consts.GitCmd, "push", "origin", sourceBranch)
}

func Checkout(prefix consts.CommitType, branch string) {
	utils.MustExec(consts.GitCmd, "checkout", "-b", string(prefix)+"/"+branch)
}

func Push() {
	utils.MustExec(consts.GitCmd, "push", "origin", strings.TrimSpace(GetCurrentBranch()))
}
