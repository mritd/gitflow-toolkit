package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"text/template"

	tea "github.com/charmbracelet/bubbletea"
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

func hasStagedFiles() (bool, error) {
	var buf bytes.Buffer
	err := gitCommand(&buf, []string{"diff", "--cached", "--name-only"})
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(buf.String()) != "", nil
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

	ok, err := hasStagedFiles()
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("\nPlease execute the `git add` command to add files before commit.\n")
	}

	m := model{
		selectorModel: newSelectorModel(),
		inputsModel:   newInputsModel(),
	}

	return tea.NewProgram(&m).Start()
}

func execCommit(m *model) error {
	sob, err := createSOB()
	if err != nil {
		fmt.Printf("ERROR(SOB): %v\n", err)
		os.Exit(1)
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

	tpl, _ := template.New("").Parse(commitMessageTpl)
	err = tpl.Execute(f, msg)
	if err != nil {
		return err
	}

	return gitCommand(os.Stdout, []string{"commit", "-F", f.Name()})
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
