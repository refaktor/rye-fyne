package main

import (
	"log"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func PullGitRepo(path, repoURL, refName string) error {
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
			ReferenceName: plumbing.ReferenceName(refName),
			Progress:      os.Stdout,
		}); err != nil {
			return err
		}
	} else {
		log.Println("Fetching", repoURL)
		_, err := git.PlainClone(path, false, &git.CloneOptions{
			URL:           repoURL,
			ReferenceName: plumbing.ReferenceName(refName),
			Depth:         1,
			Progress:      os.Stdout,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
