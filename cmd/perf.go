package cmd

import (
	"github.com/mritd/gitflow-toolkit/consts"
	"github.com/mritd/gitflow-toolkit/util"
	"github.com/spf13/cobra"
)

func NewPerf() *cobra.Command {
	return &cobra.Command{
		Use:   "perf BRANCH_NAME",
		Short: "Create perf branch",
		Long: `
Create a branch with a prefix of perf.`,
		Aliases: []string{"git-perf"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				_ = cmd.Help()
			} else {
				util.Checkout(consts.PERF, args[0])
			}
		},
	}
}
