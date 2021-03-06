package keysets

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/arken/ait/utils"
	"github.com/go-git/go-git/v5"
)

// Clone pulls a remote repository to the local instance of AIT.
func Clone(url, path string) (*git.Repository, error) {
	dir := filepath.Dir(path)
	if !utils.FileExists(dir) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	r, err := git.PlainOpen(path)
	if err != nil && err.Error() == "repository does not exist" {
		r, err = git.PlainClone(path, false, &git.CloneOptions{
			URL:               url,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})

		if err != nil {
			return r, err
		}

	} else {
		if err != nil {
			fmt.Println("The repository", `"`+url+`"`, "was not found. Please double check the URL.")
			return r, err
		}
		w, err := r.Worktree()
		if err != nil {
			return r, err
		}
		err = w.Pull(&git.PullOptions{RemoteName: "origin"})
		if err != nil && err.Error() != "already up-to-date" {
			return r, err
		}
	}

	return r, nil
}
