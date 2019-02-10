package cmd

import (
	"github.com/mritd/gitflow-toolkit/consts"
	"github.com/mritd/gitflow-toolkit/git"
	"github.com/spf13/cobra"
)

func NewHotFix() *cobra.Command {
	return &cobra.Command{
		Use:   "hotfix BRANCH_NAME",
		Short: "Create hotfix branch",
		Long: `
Create a branch with a prefix of hotfix.`,
		Aliases: []string{"git-hotfix"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				_ = cmd.Help()
			} else {
				git.Checkout(consts.HOTFIX, args[0])
			}
		},
	}
}
