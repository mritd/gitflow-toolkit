package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/mritd/gitflow-toolkit/v2/internal/config"
	"github.com/mritd/gitflow-toolkit/v2/internal/ui/common"
	"github.com/mritd/gitflow-toolkit/v2/internal/ui/install"
)

var (
	installDir  string
	installHook bool
)

// installCmd represents the install command.
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install gitflow-toolkit to your system",
	Long: `Install gitflow-toolkit and create symlinks for git subcommands.

This will:
  1. Copy the binary to the install directory
  2. Create symlinks for all commands (git-ci, git-ps, git-feat, etc.)
  3. Optionally configure global git hooks

After installation, you can use commands like:
  git ci      - Interactive commit
  git ps      - Push current branch
  git feat    - Create feature branch
  git fix     - Create fix branch
  ...`,
	RunE: runInstall,
}

// uninstallCmd represents the uninstall command.
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall gitflow-toolkit from your system",
	Long: `Uninstall gitflow-toolkit and remove all symlinks.

This will:
  1. Remove all git command symlinks
  2. Remove the binary
  3. Remove configuration and hooks`,
	RunE: runUninstall,
}

func init() {
	installCmd.Flags().StringVarP(&installDir, "dir", "d", config.DefaultInstallDir, "Installation directory")
	installCmd.Flags().BoolVar(&installHook, "hook", false, "Install global commit-msg hook")

	uninstallCmd.Flags().StringVarP(&installDir, "dir", "d", config.DefaultInstallDir, "Installation directory")

	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
}

// isInteractive checks if we're running in an interactive terminal.
func isInteractive() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func runInstall(cmd *cobra.Command, args []string) error {
	paths, err := install.NewPaths(installDir)
	if err != nil {
		return err
	}

	tasks := install.InstallTasks(paths, installHook)

	// If not interactive (e.g., in Docker/CI), run tasks directly
	if !isInteractive() {
		return runTasksNonInteractive("Installing gitflow-toolkit", tasks)
	}

	// Interactive mode with TUI
	model := common.NewMultiTaskModel("Installing gitflow-toolkit", tasks)
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	m, ok := finalModel.(common.MultiTaskModel)
	if !ok {
		return fmt.Errorf("unexpected model type")
	}

	if m.HasError() {
		return fmt.Errorf("installation failed")
	}

	return nil
}

func runUninstall(cmd *cobra.Command, args []string) error {
	paths, err := install.NewPaths(installDir)
	if err != nil {
		return err
	}

	tasks := install.UninstallTasks(paths)

	// If not interactive, run tasks directly
	if !isInteractive() {
		return runTasksNonInteractive("Uninstalling gitflow-toolkit", tasks)
	}

	// Interactive mode with TUI
	model := common.NewMultiTaskModel("Uninstalling gitflow-toolkit", tasks)
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("uninstallation failed: %w", err)
	}

	m, ok := finalModel.(common.MultiTaskModel)
	if !ok {
		return fmt.Errorf("unexpected model type")
	}

	if m.HasError() {
		return fmt.Errorf("uninstallation failed")
	}

	return nil
}

// runTasksNonInteractive runs tasks without TUI (for CI/Docker).
func runTasksNonInteractive(title string, tasks []common.Task) error {
	fmt.Println(common.StyleTitle.Render(title))
	fmt.Println()

	hasError := false
	for _, task := range tasks {
		fmt.Printf("  %s %s... ", common.SymbolRunning, task.Name)

		err := task.Run()
		if err != nil {
			if common.IsWarnErr(err) {
				fmt.Println(common.StyleWarning.Render(common.SymbolWarning + " " + err.Error()))
			} else {
				fmt.Println(common.StyleError.Render(common.SymbolError + " " + err.Error()))
				hasError = true
			}
		} else {
			fmt.Println(common.StyleSuccess.Render(common.SymbolSuccess))
		}
	}

	fmt.Println()
	if hasError {
		fmt.Println(common.StyleError.Render("Some tasks failed."))
		return fmt.Errorf("some tasks failed")
	}
	fmt.Println(common.StyleSuccess.Render("All tasks completed successfully."))
	return nil
}
