package keysets

import (
    "github.com/arkenproject/ait/utils"
    "github.com/go-git/go-git/v5"
    "log"
    "os"
    "path/filepath"
)

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
        Progress: os.Stdout,
    }
    _, err := git.PlainClone(target, false, opt)
    if err != nil {
        _ = os.Remove(target)
        log.Fatal(err)
    }
}
