package cmd

import (
	"github.com/mritd/gitflow-toolkit/consts"
	"github.com/mritd/gitflow-toolkit/git"
	"github.com/spf13/cobra"
)

func NewFix() *cobra.Command {
	return &cobra.Command{
		Use:   "fix BRANCH_NAME",
		Short: "Create fix branch",
		Long: `
Create a branch with a prefix of fix.`,
		Aliases: []string{"git-fix"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				_ = cmd.Help()
			} else {
				git.Checkout(consts.FIX, args[0])
			}
		},
	}
}
