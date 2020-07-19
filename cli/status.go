package cli

import (
    "bufio"
    "errors"
    "fmt"
    "os"
    "sort"
)

//Simply prints out what files are currently staged for submission. Will return
//an error if the working directory is not an ait repo.
func Status() error {
    if !IsAITRepo() {
        return errors.New("this isn't an ait repository. Run \"ait init\"" +
            " before taking further action")
    }
    file, err := os.OpenFile(AddedFilesPath, os.O_RDONLY, 0644)
    if err == nil {
        defer file.Close()
    }
    lines := make([]string, 0, 10)
    scanner := bufio.NewScanner(file)
    scanner.Split(bufio.ScanLines)
    for scanner.Scan() {
        line := scanner.Text()
        if len(line) > 0 {
            lines = append(lines, line)
        }
    }
    if len(lines) > 0 {
        sort.Strings(lines)
        fmt.Println("Files currently staged for submission:")
        for _, line := range lines {
            fmt.Println("\t", line)
        }
    } else {
        fmt.Println("No files are currently staged for submission.")
    }
    return nil
}