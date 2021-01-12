package main

import (
	"github.com/urfave/cli/v2"
)

func pushCmd() *cli.Command {
	return &cli.Command{
		Name:    "push",
		Aliases: []string{"git-push"},
		Usage:   "Push local branch to remote",
		Action: func(c *cli.Context) error {
			if c.NArg() != 0 {
				return cli.ShowAppHelp(c)
			}
			return push()
		},
	}
}
