package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func newBranchCmd(ct CommitType) *cli.Command {
	return &cli.Command{
		Name:      string(ct),
		Usage:     fmt.Sprintf("Create %s branch", ct),
		UsageText: fmt.Sprintf("%s NAME", ct),
		Action: func(c *cli.Context) error {
			if c.NArg() != 1 {
				return cli.ShowAppHelp(c)
			}
			err := createBranch(fmt.Sprintf("%s/%s", ct, c.Args().First()))
			if err != nil {
				return fmt.Errorf("failed to create branch %s/%s: %s", ct, c.Args().First(), err)
			}

			return nil
		},
	}
}
