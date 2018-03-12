package prmr

import (
	"errors"

	"strings"

	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/mritd/gitflow-toolkit/pkg/consts"
	"github.com/mritd/gitflow-toolkit/pkg/util"
	"github.com/mritd/promptui"
	"github.com/spf13/viper"
	"github.com/xanzy/go-gitlab"
)

type Repository struct {
	Contributor string
	Token       string
	Project     string
	Type        consts.RepoType
	Address     string
	Path        string
	AutoRebase  bool
}

func ConfigRepository() *Repository {

	repo := Repository{}

	validate := func(input string) error {
		if strings.TrimSpace(input) == "" {
			return errors.New("subject is blank")
		}
		return nil
	}

	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}

	prompt := promptui.Prompt{
		Label:     "❯ Project(eg: namespace/project):",
		Templates: templates,
		Validate:  validate,
	}

	project, err := prompt.Run()
	util.CheckAndExit(err)
	repo.Project = strings.ToLower(project)

	prompt = promptui.Prompt{
		Label:     "❯ Contributor(eg: mritd):",
		Templates: templates,
		Validate:  validate,
	}

	contributor, err := prompt.Run()
	util.CheckAndExit(err)
	repo.Contributor = contributor

	prompt = promptui.Prompt{
		Label:     "❯ Token(eg: q6C49FK47WhU68ofb):",
		Templates: templates,
		Validate:  validate,
	}

	token, err := prompt.Run()
	util.CheckAndExit(err)
	repo.Token = token

	prompt = promptui.Prompt{
		Label:     "❯ Address(eg: https://github.com):",
		Templates: templates,
		Validate:  validate,
	}

	address, err := prompt.Run()
	util.CheckAndExit(err)
	repo.Address = strings.ToLower(address)

	if strings.Contains(repo.Address, "github.com") {
		repo.Type = consts.GitHubRepo
	} else {
		repo.Type = consts.GitLabRepo
	}

	validate = func(input string) error {
		if strings.TrimSpace(input) == "" {
			return errors.New("subject is blank")
		}
		if strings.ToLower(strings.TrimSpace(input)) != "y" && strings.ToLower(strings.TrimSpace(input)) != "n" {
			return errors.New("only enter y or n")
		}
		return nil
	}

	prompt = promptui.Prompt{
		Label:     "❯ Automatic rebase before merge request?(y/n)",
		Templates: templates,
		Validate:  validate,
		IsVimMode: true,
	}

	autoRebase, err := prompt.Run()
	util.CheckAndExit(err)

	if strings.ToLower(strings.TrimSpace(autoRebase)) == "y" {
		repo.AutoRebase = true
	} else {
		repo.AutoRebase = false
	}

	repo.Path = util.WorkingDir

	return &repo
}

func (repo *Repository) SaveRepository() {

	home, err := homedir.Dir()
	util.CheckAndExit(err)
	cfgPath := home + string(filepath.Separator) + ".gitflow-toolkit.yaml"

	if _, err = os.Stat(cfgPath); err != nil {
		os.Create(cfgPath)
	}

	var repositories []*Repository

	util.CheckAndExit(viper.UnmarshalKey("repositories", &repositories))
	repositories = append(repositories, repo)
	viper.Set("repositories", repositories)
	util.CheckAndExit(viper.WriteConfig())

}

func GetRepoInfo() *Repository {

	var repositories []*Repository
	util.CheckAndExit(viper.UnmarshalKey("repositories", &repositories))

	for _, repo := range repositories {
		if repo.Path == util.WorkingDir {
			return repo
		}
	}
	return nil
}

func (repo *Repository) Mr() {
	git := gitlab.NewClient(nil, repo.Token)
	git.SetBaseURL(repo.Address + "/api/v3")
	p, _, err := git.Projects.GetProject(repo.Project)
	util.CheckAndExit(err)

	lastCommitInfo := *util.GetLastCommitInfo()

	validate := func(input string) error {
		if strings.TrimSpace(input) == "" {
			return errors.New("target branch is blank")
		}
		return nil
	}

	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}

	prompt := promptui.Prompt{
		Label:     "❯ Title(eg: fix(v0.0.1):fix a bug):",
		Templates: templates,
	}

	title, err := prompt.Run()
	util.CheckAndExit(err)
	if strings.TrimSpace(title) == "" {
		title = lastCommitInfo[0]
	}

	prompt = promptui.Prompt{
		Label:     "❯ Description(eg: Rollback etcd server version to 3.1.11):",
		Templates: templates,
	}

	desc, err := prompt.Run()
	util.CheckAndExit(err)

	if strings.TrimSpace(desc) == "big" {
		desc = util.OSEditInput()
	}

	if strings.TrimSpace(desc) == "" {
		desc = lastCommitInfo[1]
	}

	prompt = promptui.Prompt{
		Label:     "❯ Target Branch(eg: develop):",
		Templates: templates,
		Validate:  validate,
	}

	targetBranch, err := prompt.Run()
	util.CheckAndExit(err)

	sourceBranch := util.GetCurrentBranch()

	if repo.AutoRebase {
		util.Rebase(sourceBranch, targetBranch)
	}

	opt := &gitlab.CreateMergeRequestOptions{
		Title:        gitlab.String(title),
		Description:  gitlab.String(desc),
		SourceBranch: gitlab.String(sourceBranch),
		TargetBranch: gitlab.String(targetBranch),
	}
	_, _, err = git.MergeRequests.CreateMergeRequest(p.ID, opt)
	util.CheckAndExit(err)
}
