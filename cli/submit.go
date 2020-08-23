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

// submitFields is a simple struct to hold github username and password and other
// fields the user has to fill in/choose.
type submitFields struct {
	// ksGenMethod is whether to overwrite or amend to existing keyset files.
	username, password, ksGenMethod string
}

// credsEmpty returns true if both credential fields are the empty string.
func (c *submitFields) credsEmpty() bool {
	return c.username == "" && c.password == ""
}

// clearCreds sets both credential fields to the empty string
func (c *submitFields) clearCreds() {
	c.username = ""
	c.password = ""
}

// doOverwrite returns false if the struct's ksGenMethod is equal to "a" (amend
// or append), false otherwise.
func (c *submitFields) doOverwrite() bool {
	return c.ksGenMethod != "a"
}

var	fields submitFields

// SubmitRun generates a keyset file and then clones the Github repo at the given
// url, adds the keyset file, commits it, and pushes it, and then deletes the repo
// once everything is done or if anything goes wrong before completion. With all
// of those steps, there are MANY possible points of failure. If anything goes
// wrong, the error will be PrintFatal'd and the repo will we deleted from
// its temporary location at .ait/sources. Users are not meant to deal with the
// repos directly at any point so it and the keyset file are basically ephemeral
// and only exist on disk while this command is running.
func SubmitRun(_ *cmd.RootCMD, c *cmd.CMD) {
	args := c.Args.(*SubmitArgs).Args
	if len(args) < 1 {
		utils.FatalPrintln("Not enough arguments, expected repository url")
	}
	var url string
	isPR := false
	for _, arg := range args {
		if strings.HasSuffix(arg, ".git") {
			url = arg
		} else if arg == "p" {
			isPR = true
		}
	}
	if s, _ := utils.GetFileSize(utils.AddedFilesPath); s == 0 {
		utils.FatalPrintln(`No files are currently added, nothing to submit. Use
    ait add <files>...
to add files for submission.`)
	}
	repoPath := filepath.Join(".ait", "sources", utils.GetRepoName(url))
	if utils.FileExists(repoPath) {
		utils.FatalPrintf("A file/folder already exists at %v, " +
			"please delete it and try again\n", repoPath)
	}
	if isPR {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(`You've chosen to start a pull request.
Please enter your GitHub username (not the name of the repo's owner): `)
		username, _ := reader.ReadString('\n')
		fmt.Print("\n")
		username = strings.TrimSpace(username)
		err := PullRequest(url, username)
		if err != nil {
			Cleanup()
			utils.FatalPrintln(err)
		}
	} else {
		repo, err := keysets.Clone(url, repoPath)
		utils.CheckError(err)
		if err != nil {
			Cleanup()
			utils.FatalPrintln(err)
		}
		display.ShowApplication(repoPath)
		app := display.ReadApplication()
		ksName := app.GetKSName()
		category := app.GetCategory()
		if !app.IsValid() {
			Cleanup()
			utils.FatalPrintln("Empty commit message and/or title, submission aborted.")
		}
		ksPath := filepath.Join(repoPath, category, ksName)
		AddKeyset(repo, filepath.Join(category, ksName), ksPath)
		CommitKeyset(repo)
		PushKeyset(repo, url, false)
	}
	Cleanup()
}

// AddKeyset adds the keyset file at the given path to the repo.
// Effectively: git add ksPath
func AddKeyset(repo *git.Repository, ksPathFromRepo, ksPathFromWD string) {
	var choice = &fields.ksGenMethod //want to keep this response saved in the struct
	if utils.FileExists(ksPathFromWD) && *choice == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("A file called %v already exists in the cloned repo.\n",
			filepath.Base(ksPathFromWD))
		for *choice != "a" && *choice != "o" {
			fmt.Print("Would you like to overwrite it (o) or add to it (a)? ")
			*choice, _ = reader.ReadString('\n')
			*choice = strings.TrimSpace(*choice)
		}
		fmt.Print("\n")
	}
	err := keysets.Generate(ksPathFromWD, fields.doOverwrite())
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

// CommitKeyset attempts to commit the file that was previously added. This
// function expects a repo that already has a file added to the worktree.
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
	reader := bufio.NewReader(os.Stdin)
	var err error
	var existingCreds, hasWriteAccess bool
	for choice := "r"; choice == "r"; {
		existingCreds, hasWriteAccess, err = tryPush(repo)
		if err == nil { //push was successful
			break
		}
		printSubmissionPrompt(existingCreds, hasWriteAccess, isPR)
		choice, _ = reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		fmt.Print("\n")
		if choice == "p" && !isPR && existingCreds {
			err = PullRequest(url, fields.username)
			utils.CheckError(err)
			return
		} else if choice == "r" {
			fields.clearCreds()
			continue
		} else {
			fmt.Println("Submission aborted.")
			return
		}
	}
	if err == nil {
		fmt.Println("Submission successful!")
	} else {
		fmt.Println("Submission failed:", err)
	}
}

// tryPush attempts a push on the given repo. This function will prompt for
// credentials if none are currently in fields. In this order, it returns:
//     - whether the attempted credentials belong to an existing account
//     - whether the account has write access to the given repository
//     - any error returned by the push operation, nil if it was successful
// A fully successful push will return (true, true, nil).
func tryPush(repo *git.Repository) (existingCreds bool, hasWriteAccess bool, err error) {
	if fields.credsEmpty() {
		promptCredentials()
	}
	opt := &git.PushOptions{
		Auth: &http.BasicAuth{
			Username: fields.username,
			Password: fields.password,
		},
	}
	err = repo.Push(opt)
	if err == nil {
		return true, true, nil
	} else if err.Error() == "authentication required" {
		existingCreds = false
		hasWriteAccess = false
	} else if err.Error() == "authorization failed" {
		existingCreds = true
		hasWriteAccess = false
	} else {      // if it wasn't one of those ^ errors it was probably file i/o
		Cleanup() // or network related, or repo was already up to date.
		utils.FatalPrintln(err)
	}
	return existingCreds, hasWriteAccess, err
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
	fields.username = strings.TrimSpace(username)
	fields.password = strings.TrimSpace(string(bytePassword))
}

// printSubmissionPrompt takes 3 boolean values and prints the appropriate
// message for a select number of situations. Not all possibilities are covered,
// but if they are not covered it's likely that it's an "impossible" scenario
// (knock on wood). For example, an existingCredits cannot be false while
// hasWriteAccess is true. If the account does not exist, it cannot have write
// access.
// These prompts establish the following inputs as meaning:
//     - "r": retry entering credentials
//     - "p": start a pull request
//     - any other key: abort the submission
func printSubmissionPrompt(existingCreds, hasWriteAccess, isPR bool) {
	if !existingCreds {
		fmt.Print(`
The username/password did not match an existing GitHub account.
Retry (r) entering your credentials or abort submission (any other key)? `)
	} else if existingCreds && !hasWriteAccess && !isPR {
		fmt.Print(`
That account does not have the privileges to write to the requested repo.
Re-enter your credentials (r), submit a pull request (p), or abort (any other key)? `)
	} else if existingCreds && !hasWriteAccess && isPR {
		fmt.Print(`
That account does not have the privileges to write to the requested repo.
Re-enter your credentials (r) or abort (any other key)? `)
	}
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
