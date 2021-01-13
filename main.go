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
	bin, err := exec.LookPath(os.Args[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	binName := filepath.Base(bin)
	for _, app := range subApps {
		if binName == app.Name {
			err = app.Run(os.Args)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			return
		}
	}

	mainApp := &cli.App{
		Name:                 "gitflow-toolkit",
		Usage:                "Git Flow ToolKit",
		Version:              fmt.Sprintf("%s %s %s", version, buildDate, commitID),
		Authors:              []*cli.Author{{Name: "mritd", Email: "mritd@linux.com"}},
		Copyright:            "Copyright (c) 2020 mritd, All rights reserved.",
		EnableBashCompletion: true,
		Action: func(c *cli.Context) error {
			return cli.ShowAppHelp(c)
		},
		Commands: []*cli.Command{
			installCmd(),
			uninstallCmd(),
		},
	}

	err = mainApp.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
