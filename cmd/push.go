package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/mritd/gitflow-toolkit/v2/internal/config"
	"github.com/mritd/gitflow-toolkit/v2/internal/ui/push"
)

// pushCmd represents the push command.
var pushCmd = &cobra.Command{
	Use:     config.CmdPush,
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

func runPush(cmd *cobra.Command, args []string) error {
	model := push.NewModel()
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running push UI: %w", err)
	}

	m, ok := finalModel.(push.Model)
	if !ok {
		return fmt.Errorf("unexpected model type")
	}

	if m.Error() != nil {
		return m.Error()
	}

	return nil
}
