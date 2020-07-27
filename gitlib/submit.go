package gitlib

import (
    "bufio"
    "fmt"
    "github.com/DataDrake/cli-ng/cmd"
    "github.com/arkenproject/ait/utils"
    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing/object"
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
    if len(args) < 2 {
        log.Fatal("Not enough arguments, expected repository name and path to Keyset")
    }
    repoName, keysetPath := args[0], args[1]
    repo := add(repoName, keysetPath)
    commit(repo)
    push(repo)
}

func add(repoName, keysetPath string) *git.Repository {
    //vvvv TODO: move this stuff into SubmitRun vvvvv
    target := filepath.Join(".ait", "sources", repoName)
    if !utils.FileExists(target) {
        log.Fatal("Specified repo does not exist")
    } else if !utils.FileExists(filepath.Join(target, keysetPath)) {
        log.Fatal("Specified Keyset file does not exist")
    }
    repo, err := git.PlainOpen(target)
    if err != nil {
        log.Fatal(err)
    }
    //                  ^^^^^^^
    var tree *git.Worktree
    tree, err = repo.Worktree()
    if err != nil {
        log.Fatal(err)
    }
    _, err = tree.Add(keysetPath)
    if err != nil {
        log.Fatal(err)
    }
    return repo
}

//TODO: check for outstanding commits before asking for a new one
func commit(repo *git.Repository) {
    tree, err := repo.Worktree()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Enter a message explaining why this submission is important:")
    reader := bufio.NewReader(os.Stdin)
    msg, _ := reader.ReadString('\n')
    opt := &git.CommitOptions{
        Author: &object.Signature{
            Name: "name", //get from config whenever that's ready
            Email: "someone@somehwere.com",
            When: time.Now(),
        },
    }
    obj, err := tree.Commit(msg, opt)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(obj, "Commit successful")
}

func push(repo *git.Repository) {
    opt := &git.PushOptions{
        Progress: os.Stdout,
    }
    err := repo.Push(opt)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Submission successful!")
}
