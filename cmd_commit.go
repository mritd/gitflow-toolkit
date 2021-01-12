package main

import (
	"github.com/urfave/cli/v2"
)

func commitCmd() *cli.Command {
	return &cli.Command{
		Name:    "ci",
		Aliases: []string{"git-ci"},
		Usage:   "Commit",
		Action: func(c *cli.Context) error {
			if c.NArg() != 0 {
				return cli.ShowAppHelp(c)
			}
			return commit()
		},
	}
}

func checkMessageCmd() *cli.Command {
	return &cli.Command{
		Name:      "check",
		Aliases:   []string{"git-cm"},
		Usage:     "Check commit message",
		UsageText: "check FILE",
		Action: func(c *cli.Context) error {
			if c.NArg() != 1 {
				return cli.ShowAppHelp(c)
			}
			return commitMessageCheck(c.Args().First())
		},
	}
}
