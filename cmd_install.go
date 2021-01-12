package main

import "github.com/urfave/cli/v2"

func installCmd() *cli.Command {
	return &cli.Command{
		Name:  "install",
		Usage: "Install gitflow-toolkit",
		Action: func(c *cli.Context) error {
			if c.NArg() != 1 {
				return cli.ShowAppHelp(c)
			}
			return commitMessageCheck(c.Args().First())
		},
	}
}
