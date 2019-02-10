package git

import (
	"strings"

	"github.com/mritd/gitflow-toolkit/utils"

	"fmt"
)

type CommitType string

type RepoType string

const Cmd = "git"

const (
	FEAT     CommitType = "feat"
	FIX      CommitType = "fix"
	DOCS     CommitType = "docs"
	STYLE    CommitType = "style"
	REFACTOR CommitType = "refactor"
	TEST     CommitType = "test"
	CHORE    CommitType = "chore"
	PERF     CommitType = "perf"
	HOTFIX   CommitType = "hotfix"
	EXIT     CommitType = "exit"
)

const CommitTpl = `{{ .Type }}({{ .Scope }}): {{ .Subject }}

{{ .Body }}

{{ .Footer }}

{{ .Sob }}
`

const CommitMessagePattern = `^(?:fixup!\s*)?(\w*)(\(([\w\$\.\*/-].*)\))?\: (.*)|^Merge\ branch(.*)`

func CheckGitProject() {
	utils.MustExecNoOut(Cmd, "rev-parse", "--show-toplevel")
}

func CheckStagedFiles() bool {
	output := utils.MustExecRtOut(Cmd, "diff", "--cached", "--name-only")
	return strings.TrimSpace(output) != ""
}

func GetLastCommitInfo() *[]string {
	title := utils.MustExecRtOut(Cmd, "log", "-1", "--pretty=format:%s")
	desc := utils.MustExecRtOut(Cmd, "log", "-1", "--pretty=format:%b")
	return &[]string{title, desc}
}

func GetCurrentBranch() string {
	// 1.8+ git symbolic-ref --short HEAD
	return strings.TrimSpace(utils.MustExecRtOut(Cmd, "rev-parse", "--abbrev-ref", "HEAD"))
}

func Rebase(sourceBranch string, targetBranch string) {
	fmt.Println("checkout branch:", targetBranch)
	utils.MustExec(Cmd, "checkout", targetBranch)
	fmt.Println("pull origin branch:", targetBranch)
	utils.MustExec(Cmd, "pull", "origin", targetBranch)
	fmt.Println("checkout branch:", sourceBranch)
	utils.MustExec(Cmd, "checkout", sourceBranch)
	fmt.Println("exec git rebase")
	utils.MustExec(Cmd, "rebase", targetBranch)
	fmt.Println("push", sourceBranch, "to origin")
	utils.MustExec(Cmd, "push", "origin", sourceBranch)
}

func Checkout(prefix CommitType, branch string) {
	utils.MustExec(Cmd, "checkout", "-b", string(prefix)+"/"+branch)
}

func Push() {
	utils.MustExec(Cmd, "push", "origin", strings.TrimSpace(GetCurrentBranch()))
}
