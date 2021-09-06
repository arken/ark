package manifest

import (
	"github.com/go-git/go-git/v5"
)

// Pull Performs a git pull on the repository
func (m *Manifest) Pull() error {
	// Checkout the repository worktree
	w, err := m.r.Worktree()
	if err != nil {
		return err
	}

	// Check for updates to the Manifest Repository
	err = w.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	return nil
}
