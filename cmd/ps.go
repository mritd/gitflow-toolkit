package cmd

import (
	"github.com/mritd/gitflow-toolkit/git"
	"github.com/spf13/cobra"
)

// Push local branch to remote git server
func NewPs() *cobra.Command {
	return &cobra.Command{
		Use:   "ps",
		Short: "Push local branch to remote git server",
		Long: `
Push local branch to remote git server.`,
		Aliases: []string{"git-ps"},
		Run: func(cmd *cobra.Command, args []string) {
			git.Push()
		},
	}
}
