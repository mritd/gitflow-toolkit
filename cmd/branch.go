package cmd

import (
	"github.com/mritd/gitflow-toolkit/v2/git"
	"github.com/spf13/cobra"
)

// Create feature branch
func NewFeatBranch() *cobra.Command {
	return &cobra.Command{
		Use:   "feat BRANCH_NAME",
		Short: "Create feature branch",
		Long: `
Create a branch with a prefix of feat.`,
		Aliases: []string{"git-feat"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				_ = cmd.Help()
			} else {
				git.Checkout(git.FEAT, args[0])
			}
		},
	}
}

// Create fix branch
func NewFixBranch() *cobra.Command {
	return &cobra.Command{
		Use:   "fix BRANCH_NAME",
		Short: "Create fix branch",
		Long: `
Create a branch with a prefix of fix.`,
		Aliases: []string{"git-fix"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				_ = cmd.Help()
			} else {
				git.Checkout(git.FIX, args[0])
			}
		},
	}
}

// Create hotfix branch
func NewHotFixBranch() *cobra.Command {
	return &cobra.Command{
		Use:   "hotfix BRANCH_NAME",
		Short: "Create hotfix branch",
		Long: `
Create a branch with a prefix of hotfix.`,
		Aliases: []string{"git-hotfix"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				_ = cmd.Help()
			} else {
				git.Checkout(git.HOTFIX, args[0])
			}
		},
	}
}

// Create perf branch
func NewPerfBranch() *cobra.Command {
	return &cobra.Command{
		Use:   "perf BRANCH_NAME",
		Short: "Create perf branch",
		Long: `
Create a branch with a prefix of perf.`,
		Aliases: []string{"git-perf"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				_ = cmd.Help()
			} else {
				git.Checkout(git.PERF, args[0])
			}
		},
	}
}

// Create refactor branch
func NewRefactorBranch() *cobra.Command {
	return &cobra.Command{
		Use:   "refactor BRANCH_NAME",
		Short: "Create refactor branch",
		Long: `
Create a branch with a prefix of refactor.`,
		Aliases: []string{"git-refactor"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				_ = cmd.Help()
			} else {
				git.Checkout(git.REFACTOR, args[0])
			}
		},
	}
}

// Create test branch
func NewTestBranch() *cobra.Command {
	return &cobra.Command{
		Use:   "test BRANCH_NAME",
		Short: "Create test branch",
		Long: `
Create a branch with a prefix of test.`,
		Aliases: []string{"git-test"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				_ = cmd.Help()
			} else {
				git.Checkout(git.TEST, args[0])
			}
		},
	}
}

// Create chore branch
func NewChoreBranch() *cobra.Command {
	return &cobra.Command{
		Use:   "chore BRANCH_NAME",
		Short: "Create chore branch",
		Long: `
Create a branch with a prefix of chore.`,
		Aliases: []string{"git-chore"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				_ = cmd.Help()
			} else {
				git.Checkout(git.CHORE, args[0])
			}

		},
	}
}

// Create style branch
func NewStyleBranch() *cobra.Command {
	return &cobra.Command{
		Use:   "style BRANCH_NAME",
		Short: "Create style branch",
		Long: `
Create a branch with a prefix of style.`,
		Aliases: []string{"git-style"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				_ = cmd.Help()
			} else {
				git.Checkout(git.STYLE, args[0])
			}
		},
	}
}

// Create docs branch
func NewDocsBranch() *cobra.Command {
	return &cobra.Command{
		Use:   "docs BRANCH_NAME",
		Short: "Create docs branch",
		Long: `
Create a branch with a prefix of docs.`,
		Aliases: []string{"git-docs"},
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				_ = cmd.Help()
			} else {
				git.Checkout(git.DOCS, args[0])
			}
		},
	}
}
