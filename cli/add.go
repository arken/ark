package cli

import (
    "bufio"
    "errors"
    "github.com/minio/minio/pkg/wildcard"
    "os"
    "path/filepath"
)

//Similar to "git add", this function adds files that match a given list of
//file matching patterns (can include *, ? wildcards) to a file. Currently this
//file is in .ait/added_files, and it contains paths relative to the program's
//working directory. Along the way, the filenames are put in a hashmap, so the
//specific order of the filenames in the file is unpredictable, but users should
//not be directly interacting with files in .ait anyway.
func Add(args []string) error {
    if !IsAITRepo() {
        return errors.New("this isn't an ait repository. Run \"ait init\"" +
            " before taking further action")
    }
    if len(args) == 0 {
        return errors.New("no files specified, nothing was done")
    }
    const PATH string = ".ait/added_files" //can later be put somewhere more central
    file, err := os.OpenFile(PATH, os.O_CREATE | os.O_RDONLY, 0644)
    if err != nil { //open it for reading its contents
        return err
    }
    contents := make(map[string]struct{}) //basically a set. empty struct has 0 width.
    fillMap(contents, file)
    file.Close()
    file, err = os.OpenFile(PATH, os.O_TRUNC | os.O_WRONLY, 0644)
    //completely truncate the file to avoid duplicated filenames
    if err != nil {
        return err
    }
    defer file.Close()
    for _, token := range args {
        err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
            if wildcard.Match(token, path) {
                contents[path] = struct{}{}
            }
            return nil
        })
        if err != nil {
            return err
        }
    }
    //dump the map's keys, which have to be unique, into the file.
    return dumpMap(contents, file)
}

//Splits the given file by newline and adds each line to the given map.
func fillMap(contents map[string]struct{}, file *os.File) {
    scanner := bufio.NewScanner(file)
    scanner.Split(bufio.ScanLines)
    for scanner.Scan() {
        line := scanner.Text()
        if len(line) > 0 {
            contents[line] = struct{}{}
        }
    }
}

//Dumps all keys in the given map to the given file, separated by a newline.
func dumpMap(contents map[string]struct{}, file *os.File) error {
    for line := range contents {
        _, err := file.WriteString(line + "\n")
        if err != nil {
            return err
        }
    }
    return nil
}

