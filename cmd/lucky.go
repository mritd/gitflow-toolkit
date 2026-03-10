package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/creack/pty"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/mritd/gitflow-toolkit/v3/consts"
	"github.com/mritd/gitflow-toolkit/v3/internal/git"
	"github.com/mritd/gitflow-toolkit/v3/internal/ui/common"
)

// luckyCmd represents the lucky commit command.
var luckyCmd = &cobra.Command{
	Use:   consts.CmdLucky,
	Short: "Run lucky commit on the current HEAD",
	Long: `Re-run lucky commit on the current HEAD commit.

This modifies the latest commit's hash to start with the configured prefix.
Useful when lucky commit was skipped or failed during the commit flow.

Configure the prefix with:
  git config gitflow.lucky-commit-prefix <hex-prefix>`,
	RunE: runLucky,
}

func init() {
	rootCmd.AddCommand(luckyCmd)
}

func runLucky(cmd *cobra.Command, _ []string) error {
	// Check lucky commit configuration
	rawPrefix := git.GetLuckyPrefix()
	if rawPrefix == "" {
		return renderError(cmd, "Lucky commit", fmt.Errorf("no prefix configured, set it with: git config gitflow.lucky-commit-prefix <hex>"))
	}

	prefix, err := git.ValidateLuckyPrefix(rawPrefix)
	if err != nil {
		return renderError(cmd, "Lucky commit", err)
	}

	if err := git.CheckLuckyCommit(); err != nil {
		return renderError(cmd, "Lucky commit", err)
	}

	// Run lucky commit on current HEAD
	luckyExec := git.LuckyCommitCmd(prefix)

	if term.IsTerminal(int(os.Stdin.Fd())) {
		// Terminal: run with TUI spinner
		result := common.RunLuckyCommit(prefix, luckyExec, git.GetHeadHash)

		if result.Cancelled {
			r := common.Warning("Lucky commit skipped", "Operation was cancelled by user.")
			fmt.Print(common.RenderResult(r))
			return nil
		}

		if result.Err != nil {
			return renderError(cmd, "Lucky commit failed", result.Err)
		}

		content := common.StyleMuted.Render("Hash: " + result.Hash)
		r := common.Success("Lucky commit done", content)
		fmt.Print(common.RenderResult(r))
		return nil
	}

	// Non-terminal: run with pty so lucky_commit can access /dev/tty
	ptmx, err := pty.Start(luckyExec)
	if err != nil {
		return renderError(cmd, "Lucky commit failed", err)
	}
	_, _ = io.Copy(io.Discard, ptmx)
	if err := luckyExec.Wait(); err != nil {
		_ = ptmx.Close()
		return renderError(cmd, "Lucky commit failed", err)
	}
	_ = ptmx.Close()

	hash, _ := git.GetHeadHash()
	content := common.StyleMuted.Render("Hash: " + hash)
	r := common.Success("Lucky commit done", content)
	fmt.Print(common.RenderResult(r))
	return nil
}
