package cmd

import (
	"github.com/mritd/gitflow-toolkit/utils"
	"github.com/spf13/cobra"
)

var installDir string

func NewInstall() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install gitflow-toolkit",
		Long: `
Install gitflow-toolkit(only support *Unix).`,
		Aliases: []string{"install"},
		Run: func(cmd *cobra.Command, args []string) {
			utils.Install(installDir)
		},
	}
}
