package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var subApps = []*cli.App{
	newBranchApp(FEAT),
	newBranchApp(FIX),
	newBranchApp(DOCS),
	newBranchApp(STYLE),
	newBranchApp(REFACTOR),
	newBranchApp(TEST),
	newBranchApp(CHORE),
	newBranchApp(PERF),
	newBranchApp(HOTFIX),
	commitApp(),
	checkMessageApp(),
	pushApp(),
}

func newBranchApp(ct CommitType) *cli.App {
	return &cli.App{
		Name:                 "git-" + string(ct),
		Usage:                fmt.Sprintf("Create %s branch", ct),
		UsageText:            fmt.Sprintf("git %s BRANCH", ct),
		Version:              fmt.Sprintf("%s %s %s", version, buildDate, commitID),
		Authors:              []*cli.Author{{Name: "mritd", Email: "mritd@linux.com"}},
		Copyright:            "Copyright (c) 2020 mritd, All rights reserved.",
		EnableBashCompletion: true,
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

func commitApp() *cli.App {
	return &cli.App{
		Name:                 "git-ci",
		Usage:                "Interactive commit",
		UsageText:            "git ci",
		Version:              fmt.Sprintf("%s %s %s", version, buildDate, commitID),
		Authors:              []*cli.Author{{Name: "mritd", Email: "mritd@linux.com"}},
		Copyright:            "Copyright (c) 2020 mritd, All rights reserved.",
		EnableBashCompletion: true,
		Action: func(c *cli.Context) error {
			if c.NArg() != 0 {
				return cli.ShowAppHelp(c)
			}
			return commit()
		},
	}
}

func checkMessageApp() *cli.App {
	return &cli.App{
		Name:                 "commit-msg",
		Usage:                "Commit message hook",
		UsageText:            "commit-msg FILE",
		Version:              fmt.Sprintf("%s %s %s", version, buildDate, commitID),
		Authors:              []*cli.Author{{Name: "mritd", Email: "mritd@linux.com"}},
		Copyright:            "Copyright (c) 2020 mritd, All rights reserved.",
		EnableBashCompletion: true,
		Action: func(c *cli.Context) error {
			if c.NArg() != 1 {
				return cli.ShowAppHelp(c)
			}
			return commitMessageCheck(c.Args().First())
		},
	}
}

func pushApp() *cli.App {
	return &cli.App{
		Name:                 "git-ps",
		Usage:                "Push local branch to remote",
		UsageText:            "git ps",
		Version:              fmt.Sprintf("%s %s %s", version, buildDate, commitID),
		Authors:              []*cli.Author{{Name: "mritd", Email: "mritd@linux.com"}},
		Copyright:            "Copyright (c) 2020 mritd, All rights reserved.",
		EnableBashCompletion: true,
		Action: func(c *cli.Context) error {
			if c.NArg() != 0 {
				return cli.ShowAppHelp(c)
			}
			return push()
		},
	}
}
