package main

import (
	"github.com/urfave/cli/v2"
)

func checkMessage() *cli.Command {
	return &cli.Command{
		Name:      "check",
		Usage:     "Check commit message",
		UsageText: "check FILE",
		Action: func(c *cli.Context) error {
			if c.NArg() != 1 {
				return cli.ShowAppHelp(c)
			}
			// TODO: Check commit message file
			return nil
		},
	}
}
