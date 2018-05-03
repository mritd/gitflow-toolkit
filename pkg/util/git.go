package util

import (
	"strings"

	"fmt"

	"github.com/mritd/gitflow-toolkit/pkg/consts"
)

func CheckGitProject() {
	MustExec(consts.GitCmd, "rev-parse", "--show-toplevel")
}

func CheckStagedFiles() bool {
	output := MustExecRtOut(consts.GitCmd, "diff", "--cached", "--name-only")
	return strings.TrimSpace(output) != ""
}

func GetLastCommitInfo() *[]string {
	title := MustExecRtOut(consts.GitCmd, "log", "-1", "--pretty=format:%s")
	desc := MustExecRtOut(consts.GitCmd, "log", "-1", "--pretty=format:%b")

	return &[]string{title, desc}
}

func GetCurrentBranch() string {
	// 1.8+ git symbolic-ref --short HEAD
	return strings.TrimSpace(MustExecRtOut(consts.GitCmd, "rev-parse", "--abbrev-ref", "HEAD"))
}

func Rebase(sourceBranch string, targetBranch string) {
	fmt.Println("checkout branch:", targetBranch)
	MustExec(consts.GitCmd, "checkout", targetBranch)
	fmt.Println("pull origin branch:", targetBranch)
	MustExec(consts.GitCmd, "pull", "origin", targetBranch)
	fmt.Println("checkout branch:", sourceBranch)
	MustExec(consts.GitCmd, "checkout", sourceBranch)
	fmt.Println("exec git rebase")
	MustExec(consts.GitCmd, "rebase", targetBranch)
	fmt.Println("push", sourceBranch, "to origin")
	MustExec(consts.GitCmd, "push", "origin", sourceBranch)
}

func Checkout(prefix consts.CommitType, branch string) {
	MustExec(consts.GitCmd, "checkout", "-b", string(prefix)+"/"+branch)
}

func Push() {
	MustExec(consts.GitCmd, "push", "origin", strings.TrimSpace(GetCurrentBranch()))
}
