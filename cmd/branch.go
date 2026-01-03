package cmd

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/mritd/gitflow-toolkit/v2/internal/config"
	"github.com/mritd/gitflow-toolkit/v2/internal/ui/branch"
)

func init() {
	// Add a command for each commit type
	for _, ct := range config.CommitTypes {
		cmd := createBranchCommand(ct.Name, ct.Description)
		rootCmd.AddCommand(cmd)
	}
}

// createBranchCommand creates a branch command for a specific commit type.
func createBranchCommand(commitType, description string) *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("%s <name>", commitType),
		Short: fmt.Sprintf("Create a %s branch (%s)", commitType, description),
		Long: fmt.Sprintf(`Create a new branch with the %s/ prefix.

This will create a branch named %s/<name> and switch to it.

Example:
  gitflow-toolkit %s my-feature
  # Creates and switches to branch: %s/my-feature`, commitType, commitType, commitType, commitType),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBranch(cmd, commitType, args[0])
		},
	}
}

func runBranch(cmd *cobra.Command, commitType, name string) error {
	model := branch.NewModel(commitType, name)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return renderError(cmd, "Branch creation failed", fmt.Errorf("error running branch UI: %w", err))
	}

	m, ok := finalModel.(branch.Model)
	if !ok {
		return renderError(cmd, "Branch creation failed", errors.New("unexpected model type"))
	}

	// Error already rendered by branch UI View()
	if m.Error() != nil {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		return m.Error()
	}

	return nil
}
