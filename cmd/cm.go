package cmd

import (
	"github.com/mritd/gitflow-toolkit/commit"
	"github.com/spf13/cobra"
)

func NewCm() *cobra.Command {
	return &cobra.Command{
		Use:   "cm FILE",
		Short: "Check the commit message specification",
		Long: `
Check if the file content meets the Angular Community Specification.`,
		Aliases: []string{"git-cm", "commit-msg"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				_ = cmd.Help()
			} else {
				commit.CheckCommitMessage(args)
			}
		},
	}
}
