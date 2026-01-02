package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mritd/gitflow-toolkit/v2/internal/config"
	"github.com/mritd/gitflow-toolkit/v2/internal/git"
	"github.com/mritd/gitflow-toolkit/v2/internal/ui/common"
)

// hookCmd represents the hook command group.
var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Git hook commands",
	Long:  `Commands for git hooks. These are typically called by git itself.`,
}

// commitMsgCmd represents the commit-msg hook command.
var commitMsgCmd = &cobra.Command{
	Use:    "commit-msg <file>",
	Short:  "Validate commit message (git hook)",
	Long:   `Validate that a commit message follows the conventional commit format.`,
	Args:   cobra.ExactArgs(1),
	Hidden: true, // Hide from help as this is called by git
	RunE:   runCommitMsgHook,
}

func init() {
	hookCmd.AddCommand(commitMsgCmd)
	rootCmd.AddCommand(hookCmd)
}

func runCommitMsgHook(cmd *cobra.Command, args []string) error {
	msgFile := args[0]

	if err := git.ValidateCommitMessage(msgFile); err != nil {
		if errors.Is(err, git.ErrInvalidCommitMessage) {
			content := fmt.Sprintf("The commit message must match the pattern:\n\n%s", config.CommitMessagePattern)
			r := common.Error("Invalid commit message", content)
			fmt.Print(common.RenderResult(r))
		} else {
			r := common.Error("Commit message validation failed", err.Error())
			fmt.Print(common.RenderResult(r))
		}
		return err
	}

	return nil
}
