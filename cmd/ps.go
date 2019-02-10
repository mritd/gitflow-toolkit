package cmd

import (
	"github.com/mritd/gitflow-toolkit/util"
	"github.com/spf13/cobra"
)

func NewPs() *cobra.Command {
	return &cobra.Command{
		Use:   "ps",
		Short: "Push local branch to remote",
		Long: `
Push local branch to remote.`,
		Aliases: []string{"git-ps"},
		Run: func(cmd *cobra.Command, args []string) {
			util.Push()
		},
	}
}
