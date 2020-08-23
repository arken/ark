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
	isPR bool
}

// credsEmpty returns true if either of the credential fields is empty.
func (c *submitFields) credsEmpty() bool {
	return c.username == "" || c.password == ""
}

// clearCreds sets both credential fields to the empty string
func (c *submitFields) clearCreds() {
	if !c.isPR {
		c.username = ""
	}
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
	for _, arg := range args {
		if strings.HasSuffix(arg, ".git") {
			url = arg
		}
		fields.isPR = fields.isPR || arg == "p"
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
	if fields.isPR { //-p flag was included
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(`You've chosen to start a pull request. Please enter your
GitHub username (make sure this is correct!): `)
		username, _ := reader.ReadString('\n')
		fmt.Print("\n")
		fields.username = strings.TrimSpace(username)
		err := PullRequest(url, fields.username)
		utils.CheckErrorWithCleanup(err, submissionCleanup)
	} else {
		repo, err := keysets.Clone(url, repoPath)
		utils.CheckErrorWithCleanup(err, submissionCleanup)
		display.ShowApplication(repoPath)
		app := display.ReadApplication()
		ksName := app.GetKSName()
		category := app.GetCategory()
		if !app.IsValid() {
			utils.FatalWithCleanup(submissionCleanup,
				"Empty commit message and/or title, submission aborted.")
		}
		ksPath := filepath.Join(repoPath, category, ksName)
		AddKeyset(repo, filepath.Join(category, ksName), ksPath)
		CommitKeyset(repo)
		PushKeyset(repo, url)
	}
	submissionCleanup()
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
	utils.CheckErrorWithCleanup(err, submissionCleanup)
	tree, err := repo.Worktree()
	utils.CheckErrorWithCleanup(err, submissionCleanup)
	_, err = tree.Add(ksPathFromRepo)
	utils.CheckErrorWithCleanup(err, submissionCleanup)
}

// CommitKeyset attempts to commit the file that was previously added. This
// function expects a repo that already has a file added to the worktree.
func CommitKeyset(repo *git.Repository) {
	tree, err := repo.Worktree()
	utils.CheckErrorWithCleanup(err, submissionCleanup)
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
	utils.CheckErrorWithCleanup(err, submissionCleanup)
}

// PushKeyset attempts to push the latest commit to the git repo's default remote.
// Users are prompted for their usernames/passwords for this.
func PushKeyset(repo *git.Repository, url string) {
	reader := bufio.NewReader(os.Stdin)
	var err error
	var existingCreds, hasWriteAccess bool
	for choice := "r"; choice == "r"; {
		existingCreds, hasWriteAccess, err = tryPush(repo)
		if err == nil { //push was successful
			return
		}
		printSubmissionPrompt(existingCreds, hasWriteAccess)
		choice, _ = reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		fmt.Print("\n")
		if choice == "p" && !fields.isPR && existingCreds {
			fields.isPR = true
			err = PullRequest(url, fields.username)
			utils.CheckError(err)
			return
		} else if choice == "r" {
			fields.clearCreds()
			continue
		} else {
			utils.FatalWithCleanup(submissionCleanup, "Submission aborted.")
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
	} else { // if it wasn't one of those ^ errors it was probably file i/o
		     // or network related, or repo was already up to date.
		utils.FatalWithCleanup(submissionCleanup, err)
	}
	return existingCreds, hasWriteAccess, err
}

// promptCredentials gets the user's github username and password. When the user
// types their password, it does not appear on screen by use of the terminal
// package.
func promptCredentials() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your GitHub username: ")
	if len(fields.username) == 0 {
		username, _ := reader.ReadString('\n')
		fields.username = strings.TrimSpace(username)
	} else {
		fmt.Println(fields.username)
	}
	fmt.Print("Enter your GitHub password: ")
	bytePassword, err := terminal.ReadPassword(syscall.Stdin)
	if err != nil {
		utils.FatalWithCleanup(submissionCleanup,
			"\nSomething went wrong when collecting your password:", err.Error())
	}
	fmt.Print("\n") //necessary
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
func printSubmissionPrompt(existingCreds, hasWriteAccess bool) {
	if !existingCreds {
		fmt.Print(`
The username/password did not match an existing GitHub account.
Retry (r) entering your credentials or abort submission (any other key)? `)
	} else if existingCreds && !hasWriteAccess && !fields.isPR {
		fmt.Print(`
That account does not have the privileges to write to the requested repo.
Re-enter your credentials (r), submit a pull request (p), or abort (any other key)? `)
	} else if existingCreds && !hasWriteAccess && fields.isPR {
		fmt.Print(`
That account does not have the privileges to write to the requested repo.
Re-enter your credentials (r) or abort (any other key)? `)
	}
}

// submissionCleanup attempts to delete the sources and commit file. Nothing
// is done if either of those operations is unsuccessful
func submissionCleanup() {
	_ = os.RemoveAll(filepath.Join(".ait", "sources"))
	_ = os.Remove(".ait/commit")
}
