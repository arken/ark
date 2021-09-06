package manifest

import (
	"errors"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

// CreateBranch creates a new branch within the input
// repository.
func (m *Manifest) CreateBranch(branchName string) (err error) {
	h, err := m.r.Head()
	if err != nil {
		return err
	}

	ref := plumbing.NewHashReference(plumbing.NewBranchReferenceName(branchName), h.Hash())
	err = m.r.Storer.SetReference(ref)

	return err

}

// GetBranchName returns the name of the current branch.
func (m *Manifest) GetBranchName() (name string, err error) {
	h, err := m.r.Head()
	if err != nil {
		return name, err
	}
	name = strings.TrimPrefix(h.Name().String(), "refs/heads/")

	return name, nil
}

// SwitchBranch switches from the current branch to the
// one with the name provided.
func (m *Manifest) SwitchBranch(branchName string) (err error) {
	w, err := m.r.Worktree()
	if err != nil {
		return err
	}

	branchRef := plumbing.NewBranchReferenceName(branchName)
	opts := &git.CheckoutOptions{Branch: branchRef}

	err = w.Checkout(opts)
	return err
}

// PullBranch attempts to pull the branch from the git origin fork.
func (m *Manifest) PullBranch(branchName string) (err error) {
	localBranchReferenceName := plumbing.NewBranchReferenceName(branchName)
	remoteReferenceName := plumbing.NewRemoteReferenceName("fork", branchName)

	rem, err := m.r.Remote("fork")
	if err != nil {
		return err
	}

	refs, err := rem.List(&git.ListOptions{})
	if err != nil {
		return err
	}

	found := false
	for _, ref := range refs {
		if ref.Name().IsBranch() && ref.Name() == localBranchReferenceName {
			found = true
		}
	}

	if !found {
		return errors.New("branch not found")
	}

	err = m.r.CreateBranch(&config.Branch{Name: branchName, Remote: "origin", Merge: localBranchReferenceName})
	if err != nil {
		return err
	}
	newReference := plumbing.NewSymbolicReference(localBranchReferenceName, remoteReferenceName)
	err = m.r.Storer.SetReference(newReference)
	return err
}
