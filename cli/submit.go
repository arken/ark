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
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
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

//A place to put special repo "aliases"
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
		keysets.Clone(url)
	}
	repo, err := git.PlainOpen(target)
	if err != nil {
		cleanup(target)
		log.Fatal(err)
	}
	keysetPath := "test.ks"
	err = keysets.Generate(filepath.Join(target, keysetPath))
	if err != nil {
		cleanup(target)
		log.Fatal(err)
	}
	add(repo, keysetPath, target)
	commit(repo, target)
	push(repo, target)
	cleanup(target)
	//somewhere in here I'm going to have to add pull request support
}

//add adds the keyset file at the given path to the repo.
//Effectively: git add keysetPath
func add(repo *git.Repository, keysetPath, repoPath string) {
	tree, err := repo.Worktree()
	if err != nil {
		cleanup(repoPath)
		log.Fatal(err)
	}
	_, err = tree.Add(keysetPath)
	if err != nil {
		cleanup(repoPath)
		log.Fatal(err)
	}
}

//commit attempts to commit the file that was previously added.
func commit(repo *git.Repository, repoPath string) {
	tree, err := repo.Worktree()
	if err != nil {
		cleanup(repoPath)
		log.Fatal(err)
	}
	msg := CollectCommit()
	opt := &git.CommitOptions{
		Author: &object.Signature{
			Name:  "name", //get from config whenever that's ready
			Email: "someone@somehwere.com",
			When:  time.Now(),
		},
	}
	_, err = tree.Commit(msg, opt)
	if err != nil {
		cleanup(repoPath)
		log.Fatal(err)
	}
}

//push attempts to push the latest commit to the git repo's default remote.
//Users are prompted for their usernames/passwords for this.
func push(repo *git.Repository, repoPath string) {
	_, err := repo.Worktree()
	if err != nil {
		cleanup(repoPath)
		log.Fatal(err)
	}
	username, password := promptCredentials()
	opt := &git.PushOptions{
		Progress: os.Stdout,
		Auth: &http.BasicAuth{
			Username: username,
			Password: password,
		},
	}
	pushErr := repo.Push(opt)
	if pushErr != nil {
		cleanup(repoPath)
		if pushErr.Error() == "authentication required" {
			log.Fatal(`The given username/password did not give you access to the requested repo.
You can retry submitting if you think you made a typo, or you may not have the proper permissions.`)
		} else {
			log.Fatal(pushErr)
		}
	}
	fmt.Println("Submission successful!")
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

//cleanup deletes the folder at the given path and prints a message if it fails.
func cleanup(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf(`Unable to remove the repo which was temporarily cloned to %v.
It is advisable that you delete it.\n`, path)
	}
}
