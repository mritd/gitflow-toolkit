package cmd

import (
	"github.com/mritd/gitflow-toolkit/consts"
	"github.com/mritd/gitflow-toolkit/git"
	"github.com/spf13/cobra"
)

func NewTest() *cobra.Command {
	return &cobra.Command{
		Use:   "test BRANCH_NAME",
		Short: "Create test branch",
		Long: `
Create a branch with a prefix of test.`,
		Aliases: []string{"git-test"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				_ = cmd.Help()
			} else {
				git.Checkout(consts.TEST, args[0])
			}
		},
	}
}
