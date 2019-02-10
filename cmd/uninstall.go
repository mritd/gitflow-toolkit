package cmd

import (
	"github.com/mritd/gitflow-toolkit/utils"
	"github.com/spf13/cobra"
)

func NewUninstall() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall gitflow-toolkit",
		Long: `
Uninstall gitflow-toolkit`,
		Aliases: []string{"uninstall"},
		Run: func(cmd *cobra.Command, args []string) {
			utils.Uninstall()
		},
	}
}
