package main

import "github.com/urfave/cli/v2"

func installCmd() *cli.Command {
	return &cli.Command{
		Name:    "install",
		Aliases: []string{"git-install"},
		Usage:   "Install gitflow-toolkit",
		Action: func(c *cli.Context) error {
			if c.NArg() != 0 {
				return cli.ShowAppHelp(c)
			}
			return install(c.String("dir"))
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "dir",
				Aliases: []string{"d"},
				Usage:   "Install dir",
				Value:   "/usr/local/bin",
			},
		},
	}
}

func uninstallCmd() *cli.Command {
	return &cli.Command{
		Name:    "uninstall",
		Aliases: []string{"git-uninstall"},
		Usage:   "Uninstall gitflow-toolkit",
		Action: func(c *cli.Context) error {
			if c.NArg() != 0 {
				return cli.ShowAppHelp(c)
			}
			return uninstall(c.String("dir"))
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "dir",
				Aliases: []string{"d"},
				Usage:   "Install dir",
				Value:   "/usr/local/bin",
			},
		},
	}
}
