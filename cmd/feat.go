package cmd

import (
	"github.com/mritd/gitflow-toolkit/consts"
	"github.com/mritd/gitflow-toolkit/git"
	"github.com/spf13/cobra"
)

func NewFeat() *cobra.Command {
	return &cobra.Command{
		Use:   "feat BRANCH_NAME",
		Short: "Create feature branch",
		Long: `
Create a branch with a prefix of feat.`,
		Aliases: []string{"git-feat"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				_ = cmd.Help()
			} else {
				git.Checkout(consts.FEAT, args[0])
			}
		},
	}
}
