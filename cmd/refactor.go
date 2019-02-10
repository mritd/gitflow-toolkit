package cmd

import (
	"github.com/mritd/gitflow-toolkit/consts"
	"github.com/mritd/gitflow-toolkit/util"
	"github.com/spf13/cobra"
)

func NewRefactor() *cobra.Command {
	return &cobra.Command{
		Use:   "refactor BRANCH_NAME",
		Short: "Create refactor branch",
		Long: `
Create a branch with a prefix of refactor.`,
		Aliases: []string{"git-refactor"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				_ = cmd.Help()
			} else {
				util.Checkout(consts.REFACTOR, args[0])
			}
		},
	}
}
