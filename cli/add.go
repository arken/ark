package cli

import (
    "errors"
    "github.com/minio/minio/pkg/wildcard"
    "os"
    "path/filepath"
)

func Add(args []string) error {
    if !IsAITRepo() {
        return errors.New("this isn't an ait repository. Run \"ait init\" " +
            "before taking further action")
    }
    if len(args) == 0 {
        return errors.New("no files specified, nothing done")
    }
    var fileFlag int
    const path string = ".ait/added_files"
    if FileExists(path) {
        fileFlag = os.O_APPEND | os.O_RDWR
    } else {
        fileFlag = os.O_CREATE | os.O_WRONLY
    }
    file, err := os.OpenFile(path, fileFlag, 0644)
    if err != nil {
        return err
    }
    defer file.Close()
    for _, token := range args {
        err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
            var writeErr error = nil
            if wildcard.Match(token, path) {
                _, writeErr = file.WriteString(path + "\n")
            }
            return writeErr
        })
        if err != nil {
            return err
        }
    }
    return nil
}

