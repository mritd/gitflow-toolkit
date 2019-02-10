package cmd

import (
	"fmt"
	"os"

	"github.com/mritd/gitflow-toolkit/git"

	"github.com/mritd/gitflow-toolkit/consts"
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

			git.CheckGitProject()

			if !git.CheckStagedFiles() {
				fmt.Println("No staged any files")
				os.Exit(1)
			}

			cm := &git.Message{Sob: git.GenSOB()}

			if fastCommit {
				cm.Type = consts.FEAT
				cm.Scope = "Undefined"
				cm.Subject = git.InputSubject()
				git.Commit(cm)
			} else {
				cm.Type = git.SelectCommitType()
				cm.Scope = git.InputScope()
				cm.Subject = git.InputSubject()
				cm.Body = git.InputBody()
				cm.Footer = git.InputFooter()
				git.Commit(cm)
			}
		},
	}

	ciCmd.Flags().BoolVarP(&fastCommit, "fast", "f", false, "快速提交")
	return ciCmd
}
