package main

import (
	"github.com/urfave/cli/v2"
)

func pushRepo() *cli.Command {
	return &cli.Command{
		Name:  "push",
		Usage: "Push local branch to remote",
		Action: func(c *cli.Context) error {
			if c.NArg() != 0 {
				return cli.ShowAppHelp(c)
			}
			// TODO: Check commit message file
			return nil
		},
	}
}
