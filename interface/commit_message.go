package _interface

import (
    "bufio"
    "github.com/arkenproject/ait/utils"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

const commitPrompt = //temporary
`# Describe the files you're submitting below (preferably <50 characters).


# Describe the files in more detail below. Note: lines starting with '#' are excluded.
# Leave the message empty to abort the submission.

`

func CollectCommit() string {
    editor := "vim" //eventually this will come from the global config struct
    commitPath := filepath.Join(".ait", "commit")
    _ = os.Remove(commitPath) //just in case there's one there already
    commitFile := utils.BasicFileOpen(commitPath, os.O_CREATE|os.O_WRONLY, 0644)
    _, _ = commitFile.WriteString(commitPrompt) //not the end of the world if the prompt isn't written
    commitFile.Close()
    execPath, err := exec.LookPath(editor)
    if err != nil {
        log.Fatalf("%v, your requested editor, could not be found. " +
            "Please make sure it is installed and in your OS's PATH.", editor)
    }
    cmd := exec.Command(execPath, commitPath)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    err = cmd.Run()
    if err != nil {
        log.Fatal(err)
    }
    commitFile = utils.BasicFileOpen(commitPath, os.O_RDONLY, 0644)
    commitMsg := readCommit(commitFile)
    commitFile.Close()
    _ = os.Remove(commitPath)
    return commitMsg
}

func readCommit(file *os.File) string {
    scanner := bufio.NewScanner(file)
    scanner.Split(bufio.ScanLines)
    var result string
    for scanner.Scan() {
        line := scanner.Text()
        if !strings.HasPrefix(line, "#") {
            result += line + "\n"
        }
    }
    return result
}
