package keysets

import (
	"context"
	"errors"
	"fmt"
	"github.com/arkenproject/ait/utils"
	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v32/github"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/oauth2"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

//PullRequest is root function from which the entire pull request chain is run:
//Fork, Clone the fork, add keyset to fork, commit and push the fork, create
//pull request to upstream repository.
func PullRequest(url string) error {
	owner := utils.GetRepoOwner(url)
	name := utils.GetRepoName(url)
	_, err := fork(owner, name)
	return err
}

//fork uses the github api to create a fork in the user's github account and clone
//that fork into local storage. This is done using oauth2.
func fork(owner, name string) (*git.Repository, error) {
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
		return nil, errors.New(fmt.Sprintf(
			"Something went wrong when trying to fork %v's repo %v:\n%v",
			owner, name, err))
	}
	target := filepath.Join(".ait", "sources", name + "_fork")
	return Clone(remoteRepo.GetHTMLURL(), target)
}