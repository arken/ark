package keysets

import (
	"log"
	"os"
	"path/filepath"

	"github.com/arkenproject/ait/utils"
	"github.com/go-git/go-git/v5"
)

// Clone pulls a remote repository to the local instance of AIT.
func Clone(url string) {
    if !utils.FileExists(".ait/sources") {
        err := os.Mkdir(".ait/sources", os.ModePerm)
        if err != nil {
            log.Fatal(err)
        }
    }
    target := filepath.Join(".ait", "sources", utils.GetRepoName(url))
    var opt = &git.CloneOptions {
        URL: url,
    }
    _, err := git.PlainClone(target, false, opt)
    if err != nil {
        _ = os.Remove(target)
        log.Fatal(err)
    }
}
