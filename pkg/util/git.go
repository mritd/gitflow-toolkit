package util

import (
	"strings"

	"github.com/tsuyoshiwada/go-gitcmd"
)

func Git() gitcmd.Client {
	return gitcmd.New(nil)
}

func CheckGitProject() bool {
	_, err := Git().Exec("rev-parse", "--show-toplevel")
	return err == nil
}

func CheckStagedFiles() bool {
	output, _ := Git().Exec("diff", "--cached", "--name-only")
	return strings.Replace(output, " ", "", -1) != ""
}

func GetLastCommitInfo() *[]string {
	title, _ := Git().Exec("log", "-1", "--pretty=format:%s")
	desc, _ := Git().Exec("log", "-1", "--pretty=format:%b")

	return &[]string{title, desc}
}

func GetCurrentBranch() string {
	// 1.8+ git symbolic-ref --short HEAD
	branch, _ := Git().Exec("rev-parse", "--abbrev-ref", "HEAD")
	return branch
}

func Rebase(sourceBranch string, targetBranch string) {
	_, err := Git().Exec("checkout", targetBranch)
	CheckAndExit(err)
	_, err = Git().Exec("pull", "origin", targetBranch)
	CheckAndExit(err)
	_, err = Git().Exec("checkout", sourceBranch)
	CheckAndExit(err)
	_, err = Git().Exec("rebase", targetBranch)
	CheckAndExit(err)
}
