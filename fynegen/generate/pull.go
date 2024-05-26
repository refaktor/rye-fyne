package main

import (
	"log"
	"os"

	"github.com/go-git/go-git/v5"
)

func PullGitRepo(path, repoURL string) error {
	if _, err := os.Stat(path); err == nil {
		log.Println("Updating", repoURL)
		repo, err := git.PlainOpen(path)
		if err != nil {
			return err
		}
		wt, err := repo.Worktree()
		if err != nil {
			return err
		}
		if err := wt.Pull(&git.PullOptions{
			Progress: os.Stdout,
		}); err != nil {
			return err
		}
	} else {
		log.Println("Fetching", repoURL)
		_, err := git.PlainClone(path, false, &git.CloneOptions{
			URL:      repoURL,
			Depth:    1,
			Progress: os.Stdout,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
