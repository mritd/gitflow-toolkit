package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

func createBranch(name string) error {
	err := repoCheck()
	if err != nil {
		return fmt.Errorf("the current directory is not a git repository")
	}
	return gitCommand(os.Stdout, "checkout", "-b", name)
}

func hasStagedFiles() error {
	var buf bytes.Buffer
	err := gitCommand(&buf, "diff", "--cached", "--name-only")
	if err != nil {
		return err
	}
	if strings.TrimSpace(buf.String()) == "" {
		return errors.New("There is no file to commit, please execute the `git add` command to add the commit file.")
	}
	return nil
}

func currentBranch() (string, error) {
	var buf bytes.Buffer
	err := gitCommand(&buf, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}

func push() error {
	err := repoCheck()
	if err != nil {
		return fmt.Errorf("the current directory is not a git repository")
	}

	branch, err := currentBranch()
	if err != nil {
		return err
	}
	return gitCommand(os.Stdout, "push", "origin", branch)
}

func commitMessageCheck(f string) error {
	reg := regexp.MustCompile(commitMessagePattern)
	bs, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	msgs := reg.FindStringSubmatch(string(bs))
	if len(msgs) != 4 {
		return fmt.Errorf(commitMessageCheckFailedTpl)
	}

	return nil
}

func commit(msg commitMsg) error {
	if err := hasStagedFiles(); err != nil {
		return err
	}

	if msg.Body == "" {
		msg.Body = msg.Subject
	}

	f, err := ioutil.TempFile("", "gitflow-commit")
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()

	_, err = fmt.Fprintf(f, "%s(%s): %s\n\n%s\n\n%s\n\n%s\n", msg.Type, msg.Scope, msg.Subject, msg.Body, msg.Footer, msg.SOB)
	if err != nil {
		return err
	}

	var errBuf bytes.Buffer
	err = gitCommand(&errBuf, "commit", "-F", f.Name())
	if err != nil {
		return errors.New(strings.TrimSpace(errBuf.String()))
	}

	return nil
}

func createSOB() (string, error) {
	name, email, err := gitAuthor()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Signed-off-by: %s %s", name, email), nil
}

func gitAuthor() (string, string, error) {
	name := "Undefined"
	email := "Undefined"

	var buf bytes.Buffer
	err := gitCommand(&buf, "var", "GIT_AUTHOR_IDENT")
	if err != nil {
		return "", "", err
	}

	authorInfo := strings.Fields(buf.String())
	if len(authorInfo) > 1 && authorInfo[0] != "" {
		name = authorInfo[0]
	}
	if len(authorInfo) > 2 && authorInfo[1] != "" {
		email = authorInfo[1]
	}
	return name, email, nil
}

func repoCheck() error {
	var buf bytes.Buffer
	err := gitCommand(&buf, "rev-parse", "--show-toplevel")
	if err != nil {
		return errors.New(strings.TrimSpace(buf.String()))
	}
	return nil
}

func gitCommand(out io.Writer, cmds ...string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("git.exe", cmds...)
	case "linux", "darwin":
		cmd = exec.Command("git", cmds...)
	default:
		return fmt.Errorf("unsupported platform")
	}

	cmd.Stdin = os.Stdin
	if out != nil {
		cmd.Stdout = out
		cmd.Stderr = out
	}

	return cmd.Run()
}
