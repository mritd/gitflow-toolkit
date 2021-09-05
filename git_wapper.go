package main

import (
	"bytes"
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

type CommitMessage struct {
	Type    string
	Scope   string
	Subject string
	Body    string
	Footer  string
	Sob     string
}

func createBranch(name string) error {
	err := repoCheck()
	if err != nil {
		return fmt.Errorf("the current directory is not a git repository")
	}
	return gitCommand(os.Stdout, []string{"checkout", "-b", name})
}

func hasStagedFiles() error {
	var buf bytes.Buffer
	err := gitCommand(&buf, []string{"diff", "--cached", "--name-only"})
	if err != nil {
		return errors.New(buf.String())
	}
	return nil
}

func currentBranch() (string, error) {
	var buf bytes.Buffer
	err := gitCommand(&buf, []string{"symbolic-ref", "--short", "HEAD"})
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
	return gitCommand(os.Stdout, []string{"push", "origin", branch})
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

func commit() error {
	err := repoCheck()
	if err != nil {
		return fmt.Errorf("the current directory is not a git repository")
	}

	m := model{
		selector: newSelectorModel(),
		inputs:   newInputsModel(),
		spinner:  newSpinnerModel(),
	}

	return tea.NewProgram(&m).Start()
}

func execCommit(m *model) error {
	if err := hasStagedFiles(); err != nil {
		return err
	}

	sob, err := createSOB()
	if err != nil {
		return fmt.Errorf("failed to create SOB: %v", err)
	}

	msg := CommitMessage{
		Type:    m.cType,
		Scope:   m.cScope,
		Subject: m.cSubject,
		Body:    m.cBody,
		Footer:  m.cFooter,
		Sob:     sob,
	}
	if msg.Body == "" {
		msg.Body = m.cSubject
	}

	f, err := ioutil.TempFile("", "git-commit")
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}()

	_, err = fmt.Fprintf(f, "%s(%s): %s\n\n%s\n\n%s\n\n%s\n", msg.Type, msg.Scope, msg.Subject, msg.Body, msg.Footer, msg.Sob)
	if err != nil {
		return err
	}

	var errBuf bytes.Buffer
	err = gitCommand(&errBuf, []string{"commit", "-F", f.Name()})
	if err != nil {
		return errors.New(errBuf.String())
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
	err := gitCommand(&buf, []string{"var", "GIT_AUTHOR_IDENT"})
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
	return gitCommand(ioutil.Discard, []string{"rev-parse", "--show-toplevel"})
}
