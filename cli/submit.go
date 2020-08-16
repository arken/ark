package cli

import (
	"bufio"
	"fmt"
	"log"
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

//SpecialRepos is a place to put special repo "aliases"
var SpecialRepos = map[string]string{
	"core": "https://github.com/arkenproject/core-keyset.git",
}

//SubmitRun generates a keyset file and then clones the Github repo at the given
//url, adds the keyset file, commits it, and pushes it, and then deletes the repo
//once everything is done or if anything goes wrong before completion. With all
//of those steps, there are MANY possible points of failure. If anything goes
//wrong, the error will be log.Fatal'd and the repo will we deleted from
//its temporary location at .ait/sources. Users are not meant to deal with the
//repos directly at any point so it and the keyset file are basically ephemeral
//and only exist on disk while this command is running.
func SubmitRun(_ *cmd.RootCMD, c *cmd.CMD) {
	args := c.Args.(*SubmitArgs).Args
	if len(args) < 1 {
		log.Fatal("Not enough arguments, expected repository url")
	}
	if s, _ := utils.GetFileSize(utils.AddedFilesPath); s == 0 {
		log.Fatal(`No files are currently added, nothing to submit. Use
    ait add <files>...
to add files for submission.`)
	}
	url, contains := SpecialRepos[args[0]]
	if !contains {
		url = args[0]
	}
	target := filepath.Join(".ait", "sources", utils.GetRepoName(url))
	if !utils.FileExists(target) {
		path := filepath.Join(".ait", "sources", utils.GetRepoName(url))
		_, err := keysets.Clone(url, path)
		if err != nil {
			log.Fatal(err)
		}
	}
	repo, err := git.PlainOpen(target)
	if err != nil {
		Cleanup()
		log.Fatal(err)
	}
	display.ShowApplication()
	ksName := display.ReadApplication().GetKSName()
	category := display.ReadApplication().GetCategory()
	err = keysets.Generate(filepath.Join(target, category, ksName))
	if err != nil {
		Cleanup()
		log.Fatal(err)
	}
	AddKeyset(repo, filepath.Join(category, ksName))
	CommitKeyset(repo)
	PushKeyset(repo, url, false)
	Cleanup()
}

//AddKeyset adds the keyset file at the given path to the repo.
//Effectively: git add ksPath
func AddKeyset(repo *git.Repository, ksPath string) {
	tree, err := repo.Worktree()
	if err != nil {
		Cleanup()
		log.Fatal(err)
	}
	_, err = tree.Add(ksPath)
	if err != nil {
		Cleanup()
		log.Fatal(err)
	}
}

//CommitKeyset attempts to commit the file that was previously added. This
//function expects a repo that already has a file added to the worktree.
func CommitKeyset(repo *git.Repository) {
	tree, err := repo.Worktree()
	if err != nil {
		Cleanup()
		log.Fatal(err)
	}
	app := display.ReadApplication()
	msg := app.GetTitle() + "\n\n" + app.GetCommit()
	if len(strings.TrimSpace(msg)) == 0 {
		Cleanup()
		log.Fatal("Empty commit message, submission aborted.")
	}
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
		log.Fatal(err)
	}
}

//PushKeyset attempts to push the latest commit to the git repo's default remote.
//Users are prompted for their usernames/passwords for this.
func PushKeyset(repo *git.Repository, url string, isPR bool) {
	_, err := repo.Worktree()
	if err != nil {
		Cleanup()
		log.Fatal(err)
	}
	opt := &git.PushOptions{
		Auth: &http.BasicAuth{
			Username: "",
			Password: "",
		},
	}
	reader := bufio.NewReader(os.Stdin)
	var pushErr error
	fmt.Print("\n")
	for choice := "r"; choice == "r"; {
		username, password := promptCredentials()
		opt.Auth = &http.BasicAuth{
			Username: username,
			Password: password,
		}
		pushErr = repo.Push(opt)
		if pushErr != nil {
			if pushErr.Error() == "authentication required" ||
				pushErr.Error() == "authorization failed" { //TODO: give specific error messages for both of these errors
				fmt.Print(getCredentialPrompt(isPR))
				choice, _ = reader.ReadString('\n')
				choice = strings.TrimSpace(choice)
				fmt.Print("\n")
			} else { //non-authentication error
				Cleanup()
				log.Fatal(pushErr)
			}

			if choice == "p" && !isPR { //start pull request process
				pushErr = PullRequest(url, username)
				break
			} else if choice == "r" { //retry credentials
				continue
			} else { //any other key
				Cleanup()
				log.Fatal("Submission aborted.")
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

//promptCredentials gets the user's github username and password. When the user
//types their password, it does not appear on screen by use of the terminal
//package.
func promptCredentials() (string, string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your GitHub username: ")
	username, _ := reader.ReadString('\n')
	fmt.Print("Enter your GitHub password: ")
	bytePassword, err := terminal.ReadPassword(syscall.Stdin)
	if err != nil {
		log.Fatal("\nSomething went wrong when collecting your password: ", err)
	}
	fmt.Print("\n") //necessary
	return strings.TrimSpace(username), strings.TrimSpace(string(bytePassword))
}

//Cleanup deletes the folder at the given path and prints a message if it fails.
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
