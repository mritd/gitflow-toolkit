package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mritd/gitflow-toolkit/v2/internal/config"
	"github.com/mritd/gitflow-toolkit/v2/internal/git"
	"github.com/mritd/gitflow-toolkit/v2/internal/ui/commit"
	"github.com/mritd/gitflow-toolkit/v2/internal/ui/common"
)

// commitCmd represents the commit command.
var commitCmd = &cobra.Command{
	Use:     config.CmdCommit,
	Aliases: []string{"commit"},
	Short:   "Interactive commit with conventional commit format",
	Long: `Create a commit following the conventional commit format.

This command provides an interactive TUI to help you create
properly formatted commit messages with type, scope, subject,
body, and footer.

The commit message format follows the Angular specification:
  type(scope): subject

  body

  footer

  Signed-off-by: Name <email>`,
	RunE: runCommit,
}

func init() {
	rootCmd.AddCommand(commitCmd)
}

func runCommit(cmd *cobra.Command, args []string) error {
	// Check if there are staged files
	if err := git.HasStagedFiles(); err != nil {
		return renderError(cmd, "No staged files", err)
	}

	// Run the interactive commit flow
	result := commit.Run()

	if result.Cancelled {
		r := common.Warning("Commit cancelled", "Operation was cancelled by user.")
		fmt.Print(common.RenderResult(r))
		return nil
	}

	if result.Err != nil {
		return renderError(cmd, "Commit failed", result.Err)
	}

	msg := result.Message
	content := common.FormatCommitMessage(common.CommitMessageContent{
		Type:    msg.Type,
		Scope:   msg.Scope,
		Subject: msg.Subject,
		Body:    msg.Body,
		Footer:  msg.Footer,
		SOB:     msg.SOB,
	})
	r := common.Success("Commit created", content)
	fmt.Print(common.RenderResult(r))
	return nil
}
