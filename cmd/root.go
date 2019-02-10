package cmd

import (
	"github.com/mitchellh/go-homedir"
	"github.com/mritd/gitflow-toolkit/util"
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
	RootCmd.AddCommand(NewFeat())
	RootCmd.AddCommand(NewFix())
	RootCmd.AddCommand(NewDocs())
	RootCmd.AddCommand(NewStyle())
	RootCmd.AddCommand(NewRefactor())
	RootCmd.AddCommand(NewPerf())
	RootCmd.AddCommand(NewHotFix())
	RootCmd.AddCommand(NewTest())
	RootCmd.AddCommand(NewChore())
	RootCmd.AddCommand(NewInstall())
	RootCmd.AddCommand(NewUninstall())
	RootCmd.AddCommand(NewPs())
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		util.CheckAndExit(err)
		viper.AddConfigPath(home)
		viper.SetConfigName(".gitflow-toolkit")
	}
	viper.AutomaticEnv()
	viper.ReadInConfig()
}
