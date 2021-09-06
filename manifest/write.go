package manifest

import (
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Commit performs a git commit on the repository.
func (m *Manifest) Commit(path, commitMessage string) (err error) {
	w, err := m.r.Worktree()
	if err != nil {
		return err
	}

	_, err = w.Add(".")
	if err != nil {
		return err
	}

	commit, err := w.Commit(commitMessage, &git.CommitOptions{
		Author: &object.Signature{
			Name:  m.gitOpts.Name,
			Email: m.gitOpts.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	_, err = m.r.CommitObject(commit)
	if err != nil {
		return err
	}

	return nil
}
