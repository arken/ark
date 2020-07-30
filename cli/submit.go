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

	"github.com/DataDrake/cli-ng/cmd"
	"github.com/arkenproject/ait/config"
	"github.com/arkenproject/ait/display"
	"github.com/arkenproject/ait/keysets"
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
	keysetPath := "test.ks"
	err = keysets.Generate(filepath.Join(target, keysetPath))
	if err != nil {
		Cleanup()
		log.Fatal(err)
	}
	add(repo, keysetPath)
	commit(repo)
	push(repo, url)
	Cleanup()
}

//add adds the keyset file at the given path to the repo.
//Effectively: git add keysetPath
func add(repo *git.Repository, keysetPath string) {
	tree, err := repo.Worktree()
	if err != nil {
		Cleanup()
		log.Fatal(err)
	}
	_, err = tree.Add(keysetPath)
	if err != nil {
		Cleanup()
		log.Fatal(err)
	}
}

//commit attempts to commit the file that was previously added.
func commit(repo *git.Repository) {
	tree, err := repo.Worktree()
	if err != nil {
		Cleanup()
		log.Fatal(err)
	}
	msg := display.CollectCommit()
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

//push attempts to push the latest commit to the git repo's default remote.
//Users are prompted for their usernames/passwords for this.
func push(repo *git.Repository, url string) {
	_, err := repo.Worktree()
	if err != nil {
		Cleanup()
		log.Fatal(err)
	}
	opt := &git.PushOptions{
		Auth: &http.BasicAuth{
			Username:"",
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
			   pushErr.Error() == "authorization failed" {
				fmt.Print(`
Those credentials did not give you write access to the repo.
Retry if you think you made a typo, but you might not have the proper permissions.
Re-enter your credentials (r), submit a pull request (p), or abort (any other key)? `)
				choice, _ = reader.ReadString('\n')
				choice = strings.TrimSpace(choice)
				fmt.Print("\n")
			} else { //non-authentication error
				Cleanup()
				log.Fatal(pushErr)
			}

			if choice == "p" {
				pushErr = keysets.PullRequest(url)
				break
			} else if choice == "r" {
				continue
			} else {
				break
			}
		} else {
			break //the push was actually successful
		}
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
