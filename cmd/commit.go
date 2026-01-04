package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mritd/gitflow-toolkit/v3/consts"
	"github.com/mritd/gitflow-toolkit/v3/internal/git"
	"github.com/mritd/gitflow-toolkit/v3/internal/ui/commit"
	"github.com/mritd/gitflow-toolkit/v3/internal/ui/common"
)

// commitCmd represents the commit command.
var commitCmd = &cobra.Command{
	Use:     consts.CmdCommit,
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

func runCommit(cmd *cobra.Command, _ []string) error {
	// Check if there are staged files
	if err := git.HasStagedFiles(); err != nil {
		return renderError(cmd, "No staged files", err)
	}

	// Check lucky commit configuration at startup
	var luckyPrefix string
	if rawPrefix := git.GetLuckyPrefix(); rawPrefix != "" {
		// Validate prefix format
		prefix, err := git.ValidateLuckyPrefix(rawPrefix)
		if err != nil {
			return renderError(cmd, "Lucky commit", err)
		}

		// Check lucky_commit executable exists
		if err := git.CheckLuckyCommit(); err != nil {
			return renderError(cmd, "Lucky commit", err)
		}

		luckyPrefix = prefix
	}

	// Run the interactive commit flow (pass luckyPrefix)
	result := commit.Run(luckyPrefix)

	if result.Cancelled {
		r := common.Warning("Commit cancelled", "Operation was cancelled by user.")
		fmt.Print(common.RenderResult(r))
		return nil
	}

	if result.Err != nil {
		return renderError(cmd, "Commit failed", result.Err)
	}

	// Build success message with proper width for wrapping
	msg := result.Message
	// GetContentWidth returns at least 40, minus 4 for border/padding = at least 36
	contentWidth := common.GetContentWidth(0) - 4
	content := common.FormatCommitMessage(common.CommitMessageContent{
		Type:    msg.Type,
		Scope:   msg.Scope,
		Subject: msg.Subject,
		Body:    msg.Body,
		Footer:  msg.Footer,
		SOB:     msg.SOB,
	}, contentWidth)

	// Add hash info
	if result.Hash != "" {
		content += "\n\n" + common.StyleMuted.Render("Hash: "+result.Hash)
	}

	// Handle lucky commit results
	if result.LuckySkipped {
		r := common.Warning("Commit created (lucky skipped)", content)
		r.Note = "lucky commit skipped, original commit preserved"
		fmt.Print(common.RenderResult(r))
		return nil
	}

	if result.LuckyFailed != nil {
		r := common.Warning("Commit created (lucky failed)", content)
		r.Note = fmt.Sprintf("lucky commit failed: %s, original commit preserved", result.LuckyFailed)
		fmt.Print(common.RenderResult(r))
		return nil
	}

	r := common.Success("Commit created", content)
	r.Note = "Always code as if the guy who ends up maintaining your code will be a violent psychopath who knows where you live."
	fmt.Print(common.RenderResult(r))
	return nil
}
