package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunPush_NonTerminal(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake Git executable uses a POSIX shell")
	}
	for _, tc := range []struct {
		name string
		fail string
		want string
	}{
		{"success", "", "Push to origin/feature/test success."},
		{"failure", "1", "remote rejected push"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			fakeGit := `#!/bin/sh
case "$*" in
  'config --get gitflow.ssh-strict-host-key') echo true ;;
  'rev-parse --show-toplevel') echo /fake/repo ;;
  'symbolic-ref --short HEAD') echo feature/test ;;
  'push origin feature/test')
    echo called > "$PUSH_TEST_CALL"
    if [ "$PUSH_TEST_FAIL" = 1 ]; then
      echo 'remote rejected push' >&2
      exit 1
    fi
    echo 'Everything up-to-date'
    ;;
  *) echo "unexpected Git arguments: $*" >&2; exit 2 ;;
esac
`
			if err := os.WriteFile(filepath.Join(dir, "git"), []byte(fakeGit), 0o700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", dir)
			t.Setenv("PUSH_TEST_FAIL", tc.fail)
			callFile := filepath.Join(dir, "called")
			t.Setenv("PUSH_TEST_CALL", callFile)
			input, err := os.Open(os.DevNull)
			if err != nil {
				t.Fatal(err)
			}
			output, err := os.CreateTemp(dir, "output")
			if err != nil {
				_ = input.Close()
				t.Fatal(err)
			}
			oldIn, oldOut := os.Stdin, os.Stdout
			os.Stdin, os.Stdout = input, output
			t.Cleanup(func() {
				os.Stdin, os.Stdout = oldIn, oldOut
				_ = input.Close()
				_ = output.Close()
			})
			cmd := &cobra.Command{}
			err = runPush(cmd, nil)
			if tc.fail == "" && err != nil {
				t.Fatalf("runPush() = %v", err)
			}
			if tc.fail != "" && (err == nil || !strings.Contains(err.Error(), tc.want)) {
				t.Fatalf("runPush() = %v, want %q", err, tc.want)
			}
			if _, err := os.Stat(callFile); err != nil {
				t.Fatalf("push was not invoked: %v", err)
			}
			data, err := os.ReadFile(output.Name())
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(data), tc.want) {
				t.Errorf("output = %q, want %q", data, tc.want)
			}
			if tc.fail != "" && (!cmd.SilenceErrors || !cmd.SilenceUsage) {
				t.Error("push failure should suppress duplicate Cobra diagnostics")
			}
		})
	}
}
