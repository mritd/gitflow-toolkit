package cmd

import (
	"github.com/mitchellh/go-homedir"
	"github.com/mritd/gitflow-toolkit/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	// init config
	cobra.OnInitialize(initConfig)

	// add flags
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gitflow-toolkit.yaml)")

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
	RootCmd.AddCommand(NewInstall())
	RootCmd.AddCommand(NewUninstall())
	RootCmd.AddCommand(NewPs())
	RootCmd.AddCommand(NewVersion())
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		utils.CheckAndExit(err)
		viper.AddConfigPath(home)
		viper.SetConfigName(".gitflow-toolkit")
	}
	viper.AutomaticEnv()
	viper.ReadInConfig()
}
