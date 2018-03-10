package prmr

import (
	"errors"

	"strings"

	"fmt"

	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/mritd/gitflow-toolkit/pkg/consts"
	"github.com/mritd/gitflow-toolkit/pkg/util"
	"github.com/mritd/promptui"
	"github.com/spf13/viper"
)

type Repository struct {
	Contributor string
	Project     string
	Type        consts.RepoType
	Address     string
	Path        string
}

func GetRepository() *Repository {

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
		Label:     "❯ Project(eg: gitflow-toolkit):",
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
		Label:     "❯ Address(eg: https://github.com/mritd/idgen):",
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
	repo.Path = util.CurrentDir

	return &repo
}

func SaveRepository(repository *Repository) {

	home, err := homedir.Dir()
	util.CheckAndExit(err)
	cfgPath := home + string(filepath.Separator) + ".gitflow-toolkit.yaml"

	if _, err = os.Stat(cfgPath); err != nil {
		os.Create(cfgPath)
	}

	repositories := []*Repository{}

	util.CheckAndExit(viper.UnmarshalKey("repositories", &repositories))
	repositories = append(repositories, repository)
	viper.Set("repositories", repositories)
	fmt.Println(viper.WriteConfig())

}
