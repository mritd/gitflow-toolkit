package cmd

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/mritd/gitflow-toolkit/v3/consts"
	"github.com/mritd/gitflow-toolkit/v3/internal/git"
	"github.com/mritd/gitflow-toolkit/v3/internal/ui/common"
	"github.com/mritd/gitflow-toolkit/v3/internal/ui/push"
)

// pushCmd represents the push command.
var pushCmd = &cobra.Command{
	Use:     consts.CmdPush,
	Aliases: []string{"push"},
	Short:   "Push current branch to origin",
	Long: `Push the current branch to the origin remote.

This is equivalent to:
  git push origin <current-branch>`,
	RunE: runPush,
}

func init() {
	rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, _ []string) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		result, err := git.Push()
		if err != nil {
			return renderError(cmd, "Push failed", err)
		}
		fmt.Print(common.RenderResult(common.Success("Push completed", result)))
		return nil
	}

	model := push.NewModel()
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return renderError(cmd, "Push failed", fmt.Errorf("error running push UI: %w", err))
	}

	m, ok := finalModel.(push.Model)
	if !ok {
		return renderError(cmd, "Push failed", errors.New("unexpected model type"))
	}

	// Error already rendered by push UI View()
	if m.Error() != nil {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		return m.Error()
	}

	return nil
}
