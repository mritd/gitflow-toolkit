package cmd

import (
	"github.com/spf13/cobra"
)

var cfgFile string

var RootCmd = &cobra.Command{
	Use:   "gitflow-toolkit",
	Short: "Git Flow 辅助工具",
	Long: `
一个用于 CI/CD 实施的 Git Flow 辅助工具`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func init() {

	installCmd := NewInstall()
	installCmd.PersistentFlags().StringVar(&installDir, "dir", "/usr/local/bin", "install dir")

	uninstallCmd := NewUninstall()
	uninstallCmd.PersistentFlags().StringVar(&installDir, "dir", "/usr/local/bin", "install dir")

	RootCmd.AddCommand(installCmd)
	RootCmd.AddCommand(uninstallCmd)

	// add sub cmd
	RootCmd.AddCommand(NewCi())
	RootCmd.AddCommand(NewCm())
	RootCmd.AddCommand(NewFeatBranch())
	RootCmd.AddCommand(NewFixBranch())
	RootCmd.AddCommand(NewDocsBranch())
	RootCmd.AddCommand(NewStyleBranch())
	RootCmd.AddCommand(NewRefactorBranch())
	RootCmd.AddCommand(NewPerfBranch())
	RootCmd.AddCommand(NewHotFixBranch())
	RootCmd.AddCommand(NewTestBranch())
	RootCmd.AddCommand(NewChoreBranch())
	RootCmd.AddCommand(NewPs())
	RootCmd.AddCommand(NewVersion())
}
