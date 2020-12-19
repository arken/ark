package github

import (
	"context"
	"fmt"
	"github.com/arkenproject/ait/utils"
	"github.com/google/go-github/v32/github"
	"time"
)

// fork uses the github api to create a fork in the user's github account and
// clone it into local storage.
func Fork() {
	owner, name := cache.upstream.owner, cache.upstream.name
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	fmt.Printf("Attempting to fork %v's repository \"%v\" to your account...\n", owner, name)
	remoteRepo, response, err := client.Repositories.CreateFork(ctx, owner, name, nil)
	// A traditional if err != nil will not work here. See https://godoc.org/github.com/google/go-github/github#RepositoriesService.CreateFork
	status := -1
	if response != nil {
		status = response.StatusCode
	}

	// 202 means Github is processing the fork request, but this is ok.
	// 202 seems to be the most common non-error status code.
	if remoteRepo == nil || status != 202 && status != 200 {
		if response != nil && response.Response.StatusCode == 401 {
			utils.FatalPrintln("Your personal access token didn't work!")
			// This should never happen anymore since we get the PAT from GH
		} else {
			utils.FatalPrintf(
				`Something went wrong when trying to fork %v's repo "%v":\n%v`,
				owner, name, err)
		}
	}
	fmt.Printf("Fork creation successful. See it at %v\n\n", remoteRepo.GetHTMLURL())
	cache.fork = &Repository{
		url:   remoteRepo.GetHTMLURL(),
		owner: owner,
		name:  name,
	}
}

// CreatePullRequest creates a pull request from the forked repository to the
// upstream repo.
func CreatePullRequest(title, prBody string) {
	branch := getDefaultBranch()
	head := fmt.Sprintf("%v:%v", cache.fork.owner, branch)
	pr := &github.NewPullRequest{
		Title:               github.String(title),
		Body:                github.String(prBody),
		Head:                github.String(head),
		Base:                github.String(branch),
		MaintainerCanModify: github.Bool(true),
		Draft:               github.Bool(false),
	}
	fmt.Println("Attempting to create the pull request...")
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	donePR, _, err := client.PullRequests.Create(ctx, cache.upstream.owner,
		cache.upstream.name, pr)
	utils.CheckErrorWithCleanup(err, utils.SubmissionCleanup)
	fmt.Println("\nYour new pull request can be found at:", donePR.GetHTMLURL())
}

func getDefaultBranch() string {
	repo, _, err := client.Repositories.Get(cache.ctx, cache.upstream.owner, cache.upstream.name)
	if err != nil {
		return "master" //change this as main gets adopted more
	}
	return *repo.DefaultBranch
}

// hasWritePermission checks if the authenticated user has write permissions to
// the repo at the upstream URL
func hasWritePermission() bool {
	perm, resp, err := client.Repositories.GetPermissionLevel(
		cache.ctx, cache.upstream.owner, cache.upstream.name, *cache.user.Login,
	)
	if resp != nil && resp.Response.StatusCode != 200 {
		if resp.Response.StatusCode == 404 {
			utils.FatalPrintln("The repository", cache.upstream.url,
				"doesn't appear to exist.")
		} else {
			utils.FatalPrintln(err)
		}
	}
	return *perm.Permission == "admin" || *perm.Permission == "write"
}
