package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/mritd/gitflow-toolkit/v3/consts"
	"github.com/mritd/gitflow-toolkit/v3/internal/ui/common"
	"github.com/mritd/gitflow-toolkit/v3/internal/ui/install"
)

var installDir string

// installCmd represents the install command.
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install gitflow-toolkit to your system",
	Long: `Install gitflow-toolkit and create symlinks for git subcommands.

This will:
  1. Copy the binary to the install directory
  2. Create symlinks for all commands (git-ci, git-ps, git-feat, etc.)

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
  2. Remove the binary`,
	RunE: runUninstall,
}

func init() {
	installCmd.Flags().StringVarP(&installDir, "dir", "d", consts.DefaultInstallDir, "Installation directory")
	uninstallCmd.Flags().StringVarP(&installDir, "dir", "d", consts.DefaultInstallDir, "Installation directory")

	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
}

// isInteractive checks if we're running in an interactive terminal.
func isInteractive() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func runInstall(cmd *cobra.Command, _ []string) error {
	// Check if we have write permission to the install directory
	if install.NeedsSudo(installDir) {
		return renderError(cmd, "Permission denied",
			fmt.Errorf("cannot write to %s, please run with sudo", installDir))
	}

	paths, err := install.NewPaths(installDir)
	if err != nil {
		return renderError(cmd, "Installation failed", err)
	}

	tasks := install.InstallTasks(paths)

	// If not interactive (e.g., in Docker/CI), run tasks directly
	if !isInteractive() {
		return runTasksNonInteractive(cmd, "Installing gitflow-toolkit", tasks)
	}

	// Interactive mode with TUI
	model := common.NewMultiTaskModel("Installing gitflow-toolkit", tasks)
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return renderError(cmd, "Installation failed", err)
	}

	m, ok := finalModel.(common.MultiTaskModel)
	if !ok {
		return renderError(cmd, "Installation failed", fmt.Errorf("unexpected model type"))
	}

	if m.HasError() {
		return renderError(cmd, "Installation failed", fmt.Errorf("some tasks failed"))
	}

	return nil
}

func runUninstall(cmd *cobra.Command, _ []string) error {
	// Check if we have write permission to the install directory
	if install.NeedsSudo(installDir) {
		return renderError(cmd, "Permission denied",
			fmt.Errorf("cannot write to %s, please run with sudo", installDir))
	}

	paths, err := install.NewPaths(installDir)
	if err != nil {
		return renderError(cmd, "Uninstallation failed", err)
	}

	tasks := install.UninstallTasks(paths)

	// If not interactive, run tasks directly
	if !isInteractive() {
		return runTasksNonInteractive(cmd, "Uninstalling gitflow-toolkit", tasks)
	}

	// Interactive mode with TUI
	model := common.NewMultiTaskModel("Uninstalling gitflow-toolkit", tasks)
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return renderError(cmd, "Uninstallation failed", err)
	}

	m, ok := finalModel.(common.MultiTaskModel)
	if !ok {
		return renderError(cmd, "Uninstallation failed", fmt.Errorf("unexpected model type"))
	}

	if m.HasError() {
		return renderError(cmd, "Uninstallation failed", fmt.Errorf("some tasks failed"))
	}

	return nil
}

// runTasksNonInteractive runs tasks without TUI (for CI/Docker).
func runTasksNonInteractive(cmd *cobra.Command, title string, tasks []common.Task) error {
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
		r := common.Error("Task execution failed", "Some tasks failed.")
		fmt.Print(common.RenderResult(r))
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		return fmt.Errorf("some tasks failed")
	}
	r := common.Success("All tasks completed", "All tasks completed successfully.")
	fmt.Print(common.RenderResult(r))
	return nil
}
