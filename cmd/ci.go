package cmd

import (
	"fmt"
	"os"

	"github.com/mritd/gitflow-toolkit/commit"
	"github.com/mritd/gitflow-toolkit/consts"
	"github.com/mritd/gitflow-toolkit/util"
	"github.com/spf13/cobra"
)

var fastCommit = false

func NewCi() *cobra.Command {
	ciCmd := &cobra.Command{
		Use:   "ci",
		Short: "交互式输入 commit message",
		Long: `
交互式输入 git commit message，commit message 格式为

<type>(<scope>): <subject>
<BLANK LINE>
<body>
<BLANK LINE>
<footer>

该格式来源于 Angular 社区提交规范`,
		Aliases: []string{"git-ci"},
		Run: func(cmd *cobra.Command, args []string) {

			util.CheckGitProject()

			if !util.CheckStagedFiles() {
				fmt.Println("No staged any files")
				os.Exit(1)
			}

			cm := &commit.Message{Sob: commit.GenSOB()}

			if fastCommit {
				cm.Type = consts.FEAT
				cm.Scope = "Undefined"
				cm.Subject = commit.InputSubject()
				commit.Commit(cm)
			} else {
				cm.Type = commit.SelectCommitType()
				cm.Scope = commit.InputScope()
				cm.Subject = commit.InputSubject()
				cm.Body = commit.InputBody()
				cm.Footer = commit.InputFooter()
				commit.Commit(cm)
			}
		},
	}

	ciCmd.Flags().BoolVarP(&fastCommit, "fast", "f", false, "快速提交")
	return ciCmd
}
