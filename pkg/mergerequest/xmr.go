package mergerequest

import (
	"errors"

	"strings"

	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/mritd/gitflow-toolkit/pkg/consts"
	"github.com/mritd/gitflow-toolkit/pkg/util"
	"github.com/mritd/promptx"
	"github.com/spf13/viper"
	"github.com/xanzy/go-gitlab"
)

const EmptyErrorMassage = "Input is empty!"

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

	// namespace/project
	p := promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) == "" {
			return errors.New(EmptyErrorMassage)
		} else {
			return nil
		}
	}, "Project(eg: namespace/project):")

	project := p.Run()
	repo.Project = strings.ToLower(project)

	// Contributor
	p = promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) == "" {
			return errors.New(EmptyErrorMassage)
		} else {
			return nil
		}
	}, "Contributor(eg: mritd):")

	contributor := p.Run()
	repo.Contributor = contributor

	// Token
	p = promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) == "" {
			return errors.New(EmptyErrorMassage)
		} else {
			return nil
		}
	}, "Token(eg: q6C49FK47WhU68ofb):")

	token := p.Run()
	repo.Token = token

	// Address
	p = promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) == "" {
			return errors.New(EmptyErrorMassage)
		} else {
			return nil
		}
	}, "Address(eg: https://github.com):")

	address := p.Run()
	repo.Address = strings.ToLower(address)

	if strings.Contains(repo.Address, "github.com") {
		repo.Type = consts.GitHubRepo
	} else {
		repo.Type = consts.GitLabRepo
	}

	// Automatic rebase
	p = promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) == "" {
			return errors.New("subject is blank")
		}
		if strings.ToLower(strings.TrimSpace(string(line))) != "y" && strings.ToLower(strings.TrimSpace(string(line))) != "n" {
			return errors.New("only enter y or n")
		}
		return nil
	}, "Automatic rebase before merge request?(y/n)")

	autoRebase := p.Run()

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

func (repo *Repository) XMr() {
	git := gitlab.NewClient(nil, repo.Token)
	git.SetBaseURL(repo.Address + "/api/v3")
	p, _, err := git.Projects.GetProject(repo.Project)
	util.CheckAndExit(err)

	lastCommitInfo := *util.GetLastCommitInfo()

	// Title
	prompt := promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) == "" {
			return errors.New("Title is blank!")
		}
		return nil
	}, "Title(eg: fix(v0.0.1):fix a bug):")

	title := prompt.Run()
	if strings.TrimSpace(title) == "" {
		title = lastCommitInfo[0]
	}

	// Description
	prompt = promptx.NewDefaultPrompt(func(line []rune) error {
		return nil
	}, "Description(eg: Rollback etcd server version to 3.1.11):")

	desc := prompt.Run()

	if strings.TrimSpace(desc) == "big" {
		desc = util.OSEditInput()
	}

	if strings.TrimSpace(desc) == "" {
		desc = lastCommitInfo[1]
	}

	// Target Branch
	prompt = promptx.NewDefaultPrompt(func(line []rune) error {
		if strings.TrimSpace(string(line)) == "" {
			return errors.New("Target branch is blank!")
		}
		return nil
	}, "Target Branch(eg: develop)")

	targetBranch := prompt.Run()
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
