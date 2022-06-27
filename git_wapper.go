package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

func createBranch(name string) (string, error) {
	err := repoCheck()
	if err != nil {
		return "", err
	}

	return git("switch", "-c", name)
}

func commit(msg commitMsg) error {
	if err := hasStagedFiles(); err != nil {
		return err
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

	_, err = git("commit", "-F", f.Name())
	if err != nil {
		return err
	}

	return luckyCommit()
}

func push() (string, error) {
	err := repoCheck()
	if err != nil {
		return "", err
	}

	branch, err := currentBranch()
	if err != nil {
		return "", err
	}

	msg, err := git("push", "origin", branch)
	if err == nil {
		msg = fmt.Sprintf("Push to origin/%s success.\n\n%s", branch, msg)
	}
	return msg, err
}

func hasStagedFiles() error {
	msg, err := git("diff", "--cached", "--name-only")
	if err != nil {
		return err
	}
	if msg == "" {
		return errors.New("There is no file to commit, please execute the `git add` command to add the commit file.")
	}
	return nil
}

func currentBranch() (string, error) {
	return git("symbolic-ref", "--short", "HEAD")
}

func commitMessageCheck(f string) error {
	reg := regexp.MustCompile(commitMessageCheckPattern)
	bs, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	msgs := reg.FindStringSubmatch(string(bs))
	if len(msgs) != 4 {
		return fmt.Errorf(commitMessageCheckFailedMsg)
	}

	return nil
}

func createSOB() (string, error) {
	name, email, err := gitAuthor()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Signed-off-by: %s <%s>", name, email), nil
}

func gitAuthor() (string, string, error) {
	name := ""
	email := ""

	if cfg, err := git("config", "user.name"); err == nil {
		name = cfg
	}
	if cfg, err := git("config", "user.email"); err == nil {
		email = cfg
	}

	return name, email, nil
}

func repoCheck() error {
	_, err := git("rev-parse", "--show-toplevel")
	return err
}

func git(cmds ...string) (string, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("git.exe", cmds...)
	} else {
		cmd = exec.Command("git", cmds...)
	}

	luckyPrefix := os.Getenv(strictHostKey)
	if luckyPrefix != "true" {
		cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=no")
	}

	bs, err := cmd.CombinedOutput()
	if err != nil {
		if bs != nil {
			return "", errors.New(strings.TrimSpace(string(bs)))
		}
		return "", err
	}

	return strings.TrimSpace(string(bs)), nil
}

func luckyCommit() error {
	luckyPrefix := os.Getenv(luckyCommitEnv)
	if luckyPrefix == "" {
		return nil
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("lucky_commit.exe", luckyPrefix)
	} else {
		cmd = exec.Command("lucky_commit", luckyPrefix)
	}

	bs, err := cmd.CombinedOutput()
	if err != nil {
		if bs != nil {
			return errors.New(strings.TrimSpace(string(bs)))
		}
		return err
	}

	return nil
}
