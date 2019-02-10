package cmd

import (
	"github.com/mritd/gitflow-toolkit/consts"
	"github.com/mritd/gitflow-toolkit/git"
	"github.com/spf13/cobra"
)

func NewStyle() *cobra.Command {
	return &cobra.Command{
		Use:   "style BRANCH_NAME",
		Short: "Create style branch",
		Long: `
Create a branch with a prefix of style.`,
		Aliases: []string{"git-style"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				_ = cmd.Help()
			} else {
				git.Checkout(consts.STYLE, args[0])
			}
		},
	}
}
