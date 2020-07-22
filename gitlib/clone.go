package gitlib

import (
    "fmt"
    "github.com/DataDrake/cli-ng/cmd"
    "github.com/arkenproject/ait/utils"
    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing/transport/http"
    "log"
    "os"
    "path/filepath"
)

var Clone = cmd.CMD{
    Name:  "clone",
    Alias: "c",
    Short: "Download a Git repo of Keysets.",
    Args:  &CloneArgs{},
    Run:   PullRun,
}

type CloneArgs struct {
    Args []string
}

func PullRun(_ *cmd.RootCMD, c *cmd.CMD) {
    args := c.Args.(*CloneArgs).Args
    if len(args) == 0 {
        log.Fatal("No Git repo url provided")
    }
    url := args[0]
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
    if len(args) > 2 {
        opt.Auth = &http.BasicAuth{
            Username: args[1],
            Password: args[2],
        }
    }
    _, err := git.PlainClone(target, false, opt)
    if err != nil {
        if err.Error() == "authentication required" {
            fmt.Printf(`You do not have access to the repo %v
Try including your GitHub username and password as arguments to attempt to access the repo:
    "ait clone <url> <username> <password>"`, url)
        }
        _ = os.Remove(target)
    }
}