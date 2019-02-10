package cmd

import (
	"github.com/mritd/gitflow-toolkit/consts"
	"github.com/mritd/gitflow-toolkit/git"
	"github.com/spf13/cobra"
)

func NewChore() *cobra.Command {
	return &cobra.Command{
		Use:   "chore BRANCH_NAME",
		Short: "Create chore branch",
		Long: `
Create a branch with a prefix of chore.`,
		Aliases: []string{"git-chore"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				_ = cmd.Help()
			} else {
				git.Checkout(consts.CHORE, args[0])
			}

		},
	}
}
