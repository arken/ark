package cli

import (
    "bufio"
    "fmt"
    "github.com/DataDrake/cli-ng/cmd"
    "github.com/arkenproject/ait/keysets"
    "github.com/arkenproject/ait/utils"
    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing/object"
    "github.com/go-git/go-git/v5/plumbing/transport/http"
    "log"
    "os"
    "path/filepath"
    "time"
)

var Submit = cmd.CMD{
    Name:  "submit",
    Short: "Submit your Keyset to a git repository.",
    Args:  &SubmitArgs{},
    Run:   SubmitRun,
}

type SubmitArgs struct {
    Args []string
}

func SubmitRun(_ *cmd.RootCMD, c *cmd.CMD) {
    args := c.Args.(*SubmitArgs).Args
    if len(args) < 1 {
        log.Fatal("Not enough arguments, expected repository url")
    }
    url := args[0]
    target := filepath.Join(".ait", "sources", utils.GetRepoName(url))
    if !utils.FileExists(target) {
        keysets.Clone(url)
    }
    repo, err := git.PlainOpen(target)
    if err != nil {
        log.Fatal(err)
    }
    keysetPath := "test.ks"
    err = keysets.Generate(filepath.Join(target, keysetPath))
    if err != nil {
        log.Fatal(err)
    }
    add(repo, keysetPath)
    commit(repo)
    push(repo)
    //somewhere in here I'm going to have to add pull request support
}

func add(repo *git.Repository, keysetPath string) {
    tree, err := repo.Worktree()
    if err != nil {
        log.Fatal(err)
    }
    _, err = tree.Add(keysetPath)
    if err != nil {
        log.Fatal(err)
    }
}

//TODO: check for outstanding commits before asking for a new one
func commit(repo *git.Repository) {
    tree, err := repo.Worktree()
    if err != nil {
        log.Fatal(err)
    }
    msg := CollectCommit()
    opt := &git.CommitOptions{
        Author: &object.Signature{
            Name: "name", //get from config whenever that's ready
            Email: "someone@somehwere.com",
            When: time.Now(),
        },
    }
    _, err = tree.Commit(msg, opt)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Commit successful")
}

func push(repo *git.Repository) {
    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Enter your GitHub username: ")
    uname, _ := reader.ReadString('\n')
    fmt.Print("Enter your GitHub password: ")
    password, _ := reader.ReadString('\n')
    opt := &git.PushOptions{
        Progress: os.Stdout,
        Auth: &http.BasicAuth{
            Username: uname[:len(uname) - 1], //remove newline
            Password: password[:len(password) - 1],
        },
    }
    err := repo.Push(opt)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Submission successful!")
}
