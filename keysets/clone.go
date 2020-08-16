package keysets

import (
	"os"
	"path/filepath"

	"github.com/arkenproject/ait/utils"
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
    var opt = &git.CloneOptions {
        URL: url,
    }
    repo, err := git.PlainClone(path, false, opt)
    if err != nil {
        _ = os.Remove(path)
        return nil, err
    }
    return repo, nil
}
