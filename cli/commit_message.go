package cli

import (
    "io/ioutil"
    "log"
    "os"
    "os/exec"
    "path/filepath"
)

func CollectCommit() string {
    editor := "vim" //eventually this will come from the global config struct
    commitFile := filepath.Join(".ait", "commit")
    _ = os.Remove(commitFile) //just in case there's one there already
    execPath, err := exec.LookPath(editor)
    if err != nil {
        log.Fatal(err)
    }
    cmd := exec.Command(execPath, commitFile)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    err = cmd.Run()
    if err != nil {
        log.Fatal(err)
    }
    commitMsg, err := ioutil.ReadFile(commitFile)
    if err != nil {
        log.Fatal(err)
    }
    _ = os.Remove(commitFile)
    return string(commitMsg)
}
