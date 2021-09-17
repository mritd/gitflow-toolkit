package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"
)

var mainApp = &cli.App{
	Name:                 "gitflow-toolkit",
	Usage:                "Git Flow ToolKit",
	Version:              fmt.Sprintf("%s %s %s", version, buildDate, commitID),
	Authors:              []*cli.Author{{Name: "mritd", Email: "mritd@linux.com"}},
	Copyright:            "Copyright (c) " + time.Now().Format("2006") + " mritd, All rights reserved.",
	EnableBashCompletion: true,
	Action: func(c *cli.Context) error {
		return cli.ShowAppHelp(c)
	},
	Commands: []*cli.Command{
		installCmd(),
		uninstallCmd(),
	},
}

var subApps = []*cli.App{
	newBranchApp(feat),
	newBranchApp(fix),
	newBranchApp(docs),
	newBranchApp(style),
	newBranchApp(refactor),
	newBranchApp(test),
	newBranchApp(chore),
	newBranchApp(perf),
	newBranchApp(hotfix),
	commitApp(),
	checkMessageApp(),
	pushApp(),
}

func newBranchApp(ct string) *cli.App {
	return &cli.App{
		Name:                 "git-" + string(ct),
		Usage:                fmt.Sprintf("Create %s branch", ct),
		UsageText:            fmt.Sprintf("git %s BRANCH", ct),
		Version:              fmt.Sprintf("%s %s %s", version, buildDate, commitID),
		Authors:              []*cli.Author{{Name: "mritd", Email: "mritd@linux.com"}},
		Copyright:            "Copyright (c) " + time.Now().Format("2006") + " mritd, All rights reserved.",
		EnableBashCompletion: true,
		Action: func(c *cli.Context) error {
			if c.NArg() != 1 {
				return cli.ShowAppHelp(c)
			}

			m := newBranchModel(fmt.Sprintf("%s/%s", ct, c.Args().First()))

			return tea.NewProgram(&m).Start()
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
		Copyright:            "Copyright (c) " + time.Now().Format("2006") + " mritd, All rights reserved.",
		EnableBashCompletion: true,
		Action: func(c *cli.Context) error {
			if c.NArg() != 0 {
				return cli.ShowAppHelp(c)
			}

			m := commitModel{
				views: []tea.Model{
					newSelectorModel(),
					newInputsModel(),
					newCommittingModel(),
					newErrorModel(),
				},
			}

			return tea.NewProgram(&m).Start()
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
		Copyright:            "Copyright (c) " + time.Now().Format("2006") + " mritd, All rights reserved.",
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
		Copyright:            "Copyright (c) " + time.Now().Format("2006") + " mritd, All rights reserved.",
		EnableBashCompletion: true,
		Action: func(c *cli.Context) error {
			if c.NArg() != 0 {
				return cli.ShowAppHelp(c)
			}

			m := newPushModel()

			return tea.NewProgram(&m).Start()
		},
	}
}
