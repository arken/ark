package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/arkenproject/ait/display"
	"github.com/arkenproject/ait/keysets"
	"github.com/arkenproject/ait/utils"

	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v32/github"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/oauth2"
)

// PullRequest is the root function from which the pull request chain is run:
// Fork, Clone the fork, add keyset to fork, commit and push the fork, create
// pull request to upstream repository.
func PullRequest(url, forkOwner string) error {
	upstreamOwner := utils.GetRepoOwner(url)
	upstreamRepo := utils.GetRepoName(url)
	repoPath := filepath.Join(".ait", "sources", upstreamRepo)
	err := os.RemoveAll(repoPath)
	// ^Once a pull request is started, we don't need the old clone of the repo
	if err != nil && utils.FileExists(repoPath) {
		//I don't care if it failed because there was no repo there to begin with
		utils.FatalWithCleanup(utils.SubmissionCleanup,
			"Unable to remove the old clone of " + url + ", please delete the folder at\n" +
			filepath.Join(".ait", "sources", upstreamRepo))
	}
	repo, client, err := fork(upstreamOwner, upstreamRepo)
	utils.CheckErrorWithCleanup(err, utils.SubmissionCleanup)
	if display.ReadApplication() == nil {
		display.ShowApplication(repoPath)
	}
	ksName := display.ReadApplication().KsName // Just the name of the file
	category := display.ReadApplication().Category
	ksPath := filepath.Join(repoPath, category, ksName)
	// Full relative path from repo root ^

	AddKeyset(repo, filepath.Join(category, ksName), ksPath)
	CommitKeyset(repo)
	PushKeyset(repo, url)
	CreatePullRequest(client, upstreamOwner, upstreamRepo, forkOwner)
	return err
}

// fork uses the github api to create a fork in the user's github account and
// clone it into local storage. This is done using oauth2.
func fork(owner, name string) (*git.Repository, *github.Client, error) {
	token := os.Getenv("GITHUB_AUTH_TOKEN")
	if token == "" {
		fmt.Print(
			`You will now need a GitHub Oauth token. If you don't have one, you can make one
by following the steps at https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token
Enter your GitHub Oauth token: `)
		byteToken, _ := terminal.ReadPassword(syscall.Stdin)
		token = string(byteToken)
		fmt.Print("\n\n")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(ctx, tokenSource)
	client := github.NewClient(httpClient)

	fmt.Printf("Attempting to fork %v's repository \"%v\" to your account...\n", owner, name)
	remoteRepo, response, err := client.Repositories.CreateFork(ctx, owner, name, nil)
	// A traditional if err != nil will not work here. See https://godoc.org/github.com/google/go-github/github#RepositoriesService.CreateFork
	status := -1
	if response != nil && remoteRepo != nil {
		status = response.StatusCode
	}

	// 202 means Github is processing the fork request, but this is ok.
	// 202 seems to be the most common non-error status code.
	if remoteRepo == nil || status != 202 && status != 200 {
		var err error
		if response != nil && response.Response.StatusCode == 401 {
			err = fmt.Errorf(
				"Your OAuth token didn't work, make sure you entered it correcty.")
		} else {
			err = fmt.Errorf(
				`Something went wrong when trying to fork %v's repo "%v":\n%v`,
				owner, name, err)
		}
		return nil, nil, err
	}
	fmt.Printf("Fork creation successful. See it at %v\n\n", remoteRepo.GetHTMLURL())
	target := filepath.Join(".ait", "sources", name)
	localRepo, err := keysets.Clone(remoteRepo.GetHTMLURL(), target)
	return localRepo, client, err
}

// CreatePullRequest creates a pull request from the forked repository to the Github Repository.
func CreatePullRequest(client *github.Client, upstreamOwner, upstreamRepo, forkOwner string) {
	head := forkOwner + ":master"
	application := display.ReadApplication()
	pr := &github.NewPullRequest{
		Title:               github.String(application.Title),
		Body:                github.String(application.PRBody),
		Head:                github.String(head),
		Base:                github.String("master"),
		MaintainerCanModify: github.Bool(true),
		Draft:               github.Bool(false),
	}
	fmt.Println("Attempting to create the pull request...")
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	donePR, _, err := client.PullRequests.Create(ctx, upstreamOwner, upstreamRepo, pr)
	utils.CheckErrorWithCleanup(err, utils.SubmissionCleanup)
	fmt.Println("\nYour new pull request can be found at:", donePR.GetHTMLURL())
}
