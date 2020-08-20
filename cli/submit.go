package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/arkenproject/ait/keysets"

	"github.com/DataDrake/cli-ng/cmd"
	"github.com/arkenproject/ait/config"
	"github.com/arkenproject/ait/display"
	"github.com/arkenproject/ait/utils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"golang.org/x/crypto/ssh/terminal"
)

// Submit creates and uploads the keyset definition file.
var Submit = cmd.CMD{
	Name:  "submit",
	Short: "Submit your Keyset to a git repository.",
	Args:  &SubmitArgs{},
	Run:   SubmitRun,
}

// SubmitArgs handles the specific arguments for the submit command.
type SubmitArgs struct {
	Args []string
}

// credentials is a simple struct to hold github username and password.
type credentials struct {
	username, password string
}

// isEmpty returns true if both fields are the empty string.
func (c *credentials) isEmpty() bool {
	return len(c.password + c.username) == 0
}

// clear sets both fields to the empty string
func (c *credentials) clear() {
	c.username = ""
	c.password = ""
}

var	ghCreds credentials

//SubmitRun generates a keyset file and then clones the Github repo at the given
//url, adds the keyset file, commits it, and pushes it, and then deletes the repo
//once everything is done or if anything goes wrong before completion. With all
//of those steps, there are MANY possible points of failure. If anything goes
//wrong, the error will be PrintFatal'd and the repo will we deleted from
//its temporary location at .ait/sources. Users are not meant to deal with the
//repos directly at any point so it and the keyset file are basically ephemeral
//and only exist on disk while this command is running.
func SubmitRun(_ *cmd.RootCMD, c *cmd.CMD) {
	args := c.Args.(*SubmitArgs).Args
	if len(args) < 1 {
		utils.FatalPrintln("Not enough arguments, expected repository url")
	}
	if s, _ := utils.GetFileSize(utils.AddedFilesPath); s == 0 {
		utils.FatalPrintln(`No files are currently added, nothing to submit. Use
    ait add <files>...
to add files for submission.`)
	}
	url := args[0]
	repoPath := filepath.Join(".ait", "sources", utils.GetRepoName(url))
	if !utils.FileExists(repoPath) {
		path := filepath.Join(".ait", "sources", utils.GetRepoName(url))
		_, err := keysets.Clone(url, path)
		utils.CheckError(err)
	}
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		Cleanup()
		utils.FatalPrintln(err)
	}
	display.ShowApplication(repoPath)
	ksName := display.ReadApplication().GetKSName()
	category := display.ReadApplication().GetCategory()
	if len(display.ReadApplication().GetTitle()) == 0 ||
		len(display.ReadApplication().GetCommit()) == 0 {
		Cleanup()
		utils.FatalPrintln("Empty commit message or title, submission aborted.")
	}
	ksPath := filepath.Join(repoPath, category, ksName)
	AddKeyset(repo, filepath.Join(category, ksName), ksPath)
	CommitKeyset(repo)
	PushKeyset(repo, url, false)
	Cleanup()
}

//AddKeyset adds the keyset file at the given path to the repo.
//Effectively: git add ksPath
func AddKeyset(repo *git.Repository, ksPathFromRepo, ksPathFromWD string) {
	var choice string
	if utils.FileExists(ksPathFromWD) {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("A file called %v already exists in the cloned repo.\n",
			filepath.Base(ksPathFromWD))
		for choice != "a" && choice != "o" {
			fmt.Print("Would you like to overwrite it (o) or add to it (a)? ")
			choice, _ = reader.ReadString('\n')
			choice = strings.TrimSpace(choice)
		}
	}
	overwrite := choice != "a"
	err := keysets.Generate(ksPathFromWD, overwrite)
	if err != nil {
		Cleanup()
		utils.FatalPrintln(err)
	}
	tree, err := repo.Worktree()
	if err != nil {
		Cleanup()
		utils.FatalPrintln(err)
	}
	_, err = tree.Add(ksPathFromRepo)
	if err != nil {
		Cleanup()
		utils.FatalPrintln(err)
	}
}

//CommitKeyset attempts to commit the file that was previously added. This
//function expects a repo that already has a file added to the worktree.
func CommitKeyset(repo *git.Repository) {
	tree, err := repo.Worktree()
	if err != nil {
		Cleanup()
		utils.FatalPrintln(err)
	}
	app := display.ReadApplication()
	msg := app.GetTitle() + "\n\n" + app.GetCommit()
	opt := &git.CommitOptions{
		Author: &object.Signature{
			Name:  config.Global.Git.Name,
			Email: config.Global.Git.Email,
			When:  time.Now(),
		},
	}
	_, err = tree.Commit(msg, opt)
	if err != nil {
		Cleanup()
		utils.FatalPrintln(err)
	}
}

// PushKeyset attempts to push the latest commit to the git repo's default remote.
// Users are prompted for their usernames/passwords for this.
func PushKeyset(repo *git.Repository, url string, isPR bool) {
	_, err := repo.Worktree()
	if err != nil {
		Cleanup()
		utils.FatalPrintln(err)
	}
	opt := &git.PushOptions{}
	reader := bufio.NewReader(os.Stdin)
	var pushErr error
	fmt.Print("\n")
	for choice := "r"; choice == "r"; {
		if ghCreds.isEmpty() {
			promptCredentials()
		}
		opt.Auth = &http.BasicAuth{
			Username: ghCreds.username,
			Password: ghCreds.password,
		}
		if isPR {
			ghCreds.clear() //don't need these anymore
		}
		pushErr = repo.Push(opt)
		if pushErr != nil {
			correctCreds := true
			if pushErr.Error() == "authentication required" {
				fmt.Print("\nThe username/password did not match a GitHub account.\n" +
					"Retry (r) or abort submission (any other key)? ")
				correctCreds = false
			} else if pushErr.Error() == "authorization failed" {
				fmt.Print("\nThat account does not have the privileges to write to the requested repo.\n" +
					"Retry entering your credentials (r), start a pull request (p), or abort submission (any other key)? ")
			} else { //non-authentication error
				Cleanup()
				utils.FatalPrintln(pushErr)
			}
			choice, _ = reader.ReadString('\n')
			choice = strings.TrimSpace(choice)
			fmt.Print("\n")
			if choice == "p" && !isPR && correctCreds { //start pull request process
				pushErr = PullRequest(url, ghCreds.username)
				break
			} else if choice == "r" { //retry credentials
				ghCreds.clear()
				continue
			} else { //any other key
				Cleanup()
				utils.FatalPrintln("Submission aborted.")
			}
		} else { //the push was actually successful
			break
		}
	}
	if isPR {
		return
	}
	if pushErr == nil {
		fmt.Println("Submission successful!")
	} else {
		fmt.Println("Submission failed: ", pushErr)
	}
}

// promptCredentials gets the user's github username and password. When the user
// types their password, it does not appear on screen by use of the terminal
// package.
func promptCredentials() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your GitHub username: ")
	username, _ := reader.ReadString('\n')
	fmt.Print("Enter your GitHub password: ")
	bytePassword, err := terminal.ReadPassword(syscall.Stdin)
	if err != nil {
		utils.FatalPrintf("\nSomething went wrong when collecting your password: %v\n", err.Error())
	}
	fmt.Print("\n") //necessary
	ghCreds.username = strings.TrimSpace(username)
	ghCreds.password = strings.TrimSpace(string(bytePassword))
}

// Cleanup deletes the folder at the given path and prints a message if it fails.
func Cleanup() {
	path := filepath.Join(".ait", "sources")
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf(`Unable to remove the repo which was temporarily cloned to %v.
It is advisable that you delete it.\n`, path)
	}
	_ = os.Remove(".ait/commit")
}

func getCredentialPrompt(isPR bool) string {
	if isPR {
		return `
Those credentials did not give you write access to the repo. Retry if you 
think you made a typo. Re-enter your credentials (r) or abort (any other key)? `
	}
	return `
Those credentials did not give you write access to the repo.
Retry if you think you made a typo, but you might not have the proper permissions.
Re-enter your credentials (r), submit a pull request (p), or abort (any other key)? `
}
