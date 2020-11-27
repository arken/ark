package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/arkenproject/ait/config"
	"github.com/arkenproject/ait/display"
	"github.com/arkenproject/ait/ipfs"
	"github.com/arkenproject/ait/keysets"
	"github.com/arkenproject/ait/utils"

	"github.com/DataDrake/cli-ng/cmd"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"golang.org/x/crypto/ssh/terminal"
)

// Submit creates and uploads the keyset definition file.
var Submit = cmd.CMD{
	Name:  "submit",
	Alias: "sm",
	Short: "Submit your Keyset to a git repository.",
	Args:  &SubmitArgs{},
	Flags: &SubmitFlags{},
	Run:   SubmitRun,
}

// SubmitArgs handles the specific arguments for the submit command.
type SubmitArgs struct {
	Args []string
}

// SubmitFlags handles the specific flags for the submit command.
type SubmitFlags struct {
	IsPR bool `short:"p" long:"pull-request" desc:"Jump straight into submitting a pull request"`
}

// submitFields is a simple struct to hold github username and password and other
// fields the user has to fill in/choose.
type submitFields struct {
	// ksGenMethod is whether to overwrite or amend to existing keyset files.
	username, password, ksGenMethod string
	isPR                            bool
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

var fields submitFields

// SubmitRun generates a keyset file and then clones the Github repo at the given
// url, adds the keyset file, commits it, and pushes it, and then deletes the repo
// once everything is done or if anything goes wrong before completion. With all
// of those steps, there are MANY possible points of failure. If anything goes
// wrong, the error will be PrintFatal'd and the repo will we deleted from
// its temporary location at .ait/sources. Users are not meant to deal with the
// repos directly at any point so it and the keyset file are basically ephemeral
// and only exist on disk while this command is running.
func SubmitRun(_ *cmd.RootCMD, c *cmd.CMD) {
	var url string
	url, fields.isPR = parseSubmitArgs(c)
	ipfs.Init(false)
	repoPath := filepath.Join(".ait", "sources", utils.GetRepoName(url))
	if utils.FileExists(repoPath) {
		err := os.RemoveAll(repoPath)
		if err != nil {
			utils.FatalPrintln("Removing Previous Local Keyset Failed." +
				"Do you have permission to access it?")
		}
	}
	if config.Global.Git.Name == "" || config.Global.Git.Email == "" {
		fmt.Println("You have not defined your name and email in", config.Path)
		getNameEmail()
	}
	if fields.isPR { //-p flag was included
		fmt.Println("You have chosen to create a pull request.")
		collectUsername()
		err := PullRequest(url, fields.username)
		utils.CheckErrorWithCleanup(err, utils.SubmissionCleanup)
	} else {
		repo, err := keysets.Clone(url, repoPath)
		utils.CheckErrorWithCleanup(err, utils.SubmissionCleanup)
		display.ShowApplication(repoPath)
		app := display.ReadApplication()
		ksName := app.KsName
		category := app.Category
		if !app.IsValid() {
			utils.FatalWithCleanup(utils.SubmissionCleanup,
				"Empty commit message and/or title, submission aborted.")
		}
		ksPath := filepath.Join(repoPath, category, ksName)
		AddKeyset(repo, filepath.Join(category, ksName), ksPath)
		CommitKeyset(repo)
		PushKeyset(repo, url)
	}
	utils.SubmissionCleanup()
	fmt.Println("Submission successful!")
}

// AddKeyset adds the keyset file at the given path to the repo.
// Effectively: git add ksPath
func AddKeyset(repo *git.Repository, ksPathFromRepo, ksPathFromWD string) {
	fmt.Println("Adding keyset file to worktree...")
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
	utils.CheckErrorWithCleanup(err, utils.SubmissionCleanup)
	tree, err := repo.Worktree()
	utils.CheckErrorWithCleanup(err, utils.SubmissionCleanup)
	_, err = tree.Add(ksPathFromRepo)
	utils.CheckErrorWithCleanup(err, utils.SubmissionCleanup)
}

// CommitKeyset attempts to commit the file that was previously added. This
// function expects a repo that already has a file added to the worktree.
func CommitKeyset(repo *git.Repository) {
	fmt.Println("Committing keyset file...")
	tree, err := repo.Worktree()
	utils.CheckErrorWithCleanup(err, utils.SubmissionCleanup)
	app := display.ReadApplication()
	msg := app.Title + "\n\n" + app.Commit
	opt := &git.CommitOptions{
		Author: &object.Signature{
			Name:  config.Global.Git.Name,
			Email: config.Global.Git.Email,
			When:  time.Now(),
		},
	}
	_, err = tree.Commit(msg, opt)
	utils.CheckErrorWithCleanup(err, utils.SubmissionCleanup)
}

// PushKeyset attempts to push the latest commit to the git repo's default remote.
// Users are prompted for their usernames/passwords for this.
func PushKeyset(repo *git.Repository, url string) {
	reader := bufio.NewReader(os.Stdin)
	var err error
	var existingCreds, hasWriteAccess bool
	for choice := "r"; choice == "r"; {
		fmt.Printf("Attempting to push to %v...\n\n", url)
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
			fmt.Println("You have chosen to create a pull request.")
			err = PullRequest(url, fields.username)
			utils.CheckError(err)
			return
		} else if choice == "r" {
			fields.clearCreds()
			continue
		} else {
			utils.FatalWithCleanup(utils.SubmissionCleanup, "Submission aborted.")
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
		utils.FatalWithCleanup(utils.SubmissionCleanup, err)
	}
	return existingCreds, hasWriteAccess, err
}

// promptCredentials gets the user's github username and password. When the user
// types their password, it does not appear on screen by use of the terminal
// package.
func promptCredentials() {
	fmt.Print("Enter your GitHub username: ")
	reader := bufio.NewReader(os.Stdin)
	if len(fields.username) == 0 {
		username, _ := reader.ReadString('\n')
		fields.username = strings.TrimSpace(username)
	} else {
		fmt.Println(fields.username)
	}
	fmt.Print("Enter your GitHub password: ")
	bytePassword, err := terminal.ReadPassword(syscall.Stdin)
	if err != nil {
		utils.FatalWithCleanup(utils.SubmissionCleanup,
			"\nSomething went wrong when collecting your password:", err.Error())
	}
	fmt.Print("\n") //necessary
	fields.password = strings.TrimSpace(string(bytePassword))
}

// printSubmissionPrompt takes 2 boolean values and prints the appropriate
// message for a select number of situations. Not all possibilities are covered,
// but if they are not covered it's likely that it's an "impossible" scenario
// (knock on wood). For example, existingCreds cannot be false while
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

// collectUsername prompts the user for their github username and puts it in the
// fields struct. It asks the user to verify that their username is entered
// correctly, because the rest of the pull request chain depends on it being an
// actual account that exists.
func collectUsername() {
	reader := bufio.NewReader(os.Stdin)
	var choice string
	for {
		fmt.Print("Please enter your GitHub username: ")
		username, _ := reader.ReadString('\n')
		fields.username = strings.TrimSpace(username)
		fmt.Printf(`You entered "%v", Is this correct? If so, continue (y), if
not, try again (n), or abort (any other key). y/[n]: `,
			fields.username)
		choice, _ = reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		if choice == "y" {
			break
		} else if choice != "n" && len(choice) > 0 {
			utils.FatalPrintln("Submission aborted.")
		}
		fmt.Print("\n")
	}
	fmt.Print("\n")
}

// getNameEmail asks the user to enter their name and email for git purposes.
// this is saved into the file at ~/.ait/ait.config
func getNameEmail() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please enter your name (spaces are ok): ")
	input, _ := reader.ReadString('\n')
	config.Global.Git.Name = strings.TrimSpace(input)
	fmt.Print("Please enter your email: ")
	input, _ = reader.ReadString('\n')
	config.Global.Git.Email = strings.TrimSpace(input)
	config.GenConf(config.Global)
}

// parseSubmitArgs simply does some of the sanitization and extraction required to
// get the desired data structures out of the cmd.CMD object, then returns said
// useful data structures.
func parseSubmitArgs(c *cmd.CMD) (string, bool) {
	args := c.Args.(*SubmitArgs).Args
	if len(args) < 1 {
		utils.FatalPrintln("Not enough arguments, expected repository url")
	}
	url := config.GetRemote(args[0])
	if url != args[0] {
		fmt.Printf("Submitting to the remote at %v\n", url)
	}
	fields.isPR = c.Flags.(*SubmitFlags).IsPR
	if s, _ := utils.GetFileSize(utils.AddedFilesPath); s == 0 {
		utils.FatalPrintln(`No files are currently added, nothing to submit. Use
    ait add <files>...
to add files for submission.`)
	}
	return url, c.Flags.(*SubmitFlags).IsPR
}
