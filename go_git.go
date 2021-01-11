package main

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

func createBranch(name string) error {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return err
	}
	return repo.CreateBranch(&config.Branch{
		Name: name,
	})
}

func push() error {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return err
	}

	return repo.Push(&git.PushOptions{})
}
