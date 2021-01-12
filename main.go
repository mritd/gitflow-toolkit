package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

var (
	version   string
	buildDate string
	commitID  string
)

func main() {
	app := &cli.App{
		Name:    "gitflow-toolkit",
		Usage:   "Git Flow 辅助工具",
		Version: fmt.Sprintf("%s %s %s", version, buildDate, commitID),
		Authors: []*cli.Author{
			{
				Name:  "mritd",
				Email: "mritd@linux.com",
			},
		},
		Copyright:            "Copyright (c) 2020 mritd, All rights reserved.",
		EnableBashCompletion: true,
		Action: func(c *cli.Context) error {
			bin, err := exec.LookPath(os.Args[0])
			if err != nil {
				return err
			}
			binName := filepath.Base(bin)
			for _, cmd := range cmds() {
				for _, ali := range cmd.Aliases {
					if binName == ali {
						return cmd.Run(c)
					}
				}
			}
			return cli.ShowAppHelp(c)
		},
		Commands: cmds(),
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func cmds() []*cli.Command {
	return []*cli.Command{
		newBranchCmd(FEAT),
		newBranchCmd(FIX),
		newBranchCmd(DOCS),
		newBranchCmd(STYLE),
		newBranchCmd(REFACTOR),
		newBranchCmd(TEST),
		newBranchCmd(CHORE),
		newBranchCmd(PERF),
		newBranchCmd(HOTFIX),
		commitCmd(),
		checkMessageCmd(),
		pushCmd(),
		installCmd(),
		uninstallCmd(),
	}
}
