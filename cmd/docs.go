package cmd

import (
	"github.com/mritd/gitflow-toolkit/consts"
	"github.com/mritd/gitflow-toolkit/util"
	"github.com/spf13/cobra"
)

func NewDocs() *cobra.Command {
	return &cobra.Command{
		Use:   "docs BRANCH_NAME",
		Short: "Create docs branch",
		Long: `
Create a branch with a prefix of docs.`,
		Aliases: []string{"git-docs"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				_ = cmd.Help()
			} else {
				util.Checkout(consts.DOCS, args[0])
			}
		},
	}
}
