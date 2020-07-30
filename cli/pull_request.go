package cli

import (
	"context"
	"errors"
	"fmt"
	"github.com/arkenproject/ait/display"
	"github.com/arkenproject/ait/keysets"
	"github.com/arkenproject/ait/utils"
	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v32/github"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/oauth2"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

//PullRequest is the root function from which the pull request chain is run:
//Fork, Clone the fork, add keyset to fork, commit and push the fork, create
//pull request to upstream repository.
func PullRequest(url, forkOwner string) error {
	upstreamOwner := utils.GetRepoOwner(url)
	upstreamRepo := utils.GetRepoName(url)
	repo, client, err := fork(upstreamOwner, upstreamRepo)
	if err != nil {
		Cleanup()
		log.Fatal(err)
	}
	ksPath := filepath.Join(".ait", "sources", upstreamRepo + "_fork", "test_f.ks")
	err = keysets.Generate(ksPath)
	if err != nil {
		Cleanup()
		log.Fatal(err)
	}
	AddKeyset(repo, filepath.Base(ksPath))
	CommitKeyset(repo)
	PushKeyset(repo, url, true)
	CreatePullRequest(client, upstreamOwner, upstreamRepo, forkOwner)
	return err
}

//fork uses the github api to create a fork in the user's github account and clone
//that fork into local storage. This is done using oauth2.
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
	ctx, cancel := context.WithTimeout(context.Background(), 8 * time.Second)
	defer cancel()
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(ctx, tokenSource)
	client := github.NewClient(httpClient)

	remoteRepo, response, err := client.Repositories.CreateFork(ctx, owner, name,nil)
	//A traditional if err != nil will not work here. See https://godoc.org/github.com/google/go-github/github#RepositoriesService.CreateFork
	status := -1
	if response != nil && remoteRepo != nil {
		status = response.StatusCode
	}
	//202 means Github is processing the fork request, but this is ok.
	//202 seems to be the most common non-error status code.
	if remoteRepo == nil || status != 202 && status != 200 {
		return nil, nil, errors.New(fmt.Sprintf(
			"Something went wrong when trying to fork %v's repo %v:\n%v",
			owner, name, err))
	}
	target := filepath.Join(".ait", "sources", name + "_fork")
	localRepo, err := keysets.Clone(remoteRepo.GetHTMLURL(), target)
	return localRepo, client, err
}

func CreatePullRequest(client *github.Client, upstreamOwner, upstreamRepo, forkOwner string) {
	head := forkOwner + ":master"
	application := display.ReadApplication()
	pr := &github.NewPullRequest {
		Title:               github.String(application.Title),
		Body:                github.String(application.PRBody),
		Head:                github.String(head),
		Base:                github.String("master"),
		MaintainerCanModify: github.Bool(true),
		Draft:               github.Bool(false),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8 * time.Second)
	defer cancel()
	donePR, _, err := client.PullRequests.Create(ctx, upstreamOwner, upstreamRepo, pr)
	if err != nil {
		Cleanup()
		log.Fatal(err)
	}
	fmt.Println(donePR.GetHTMLURL())
}
