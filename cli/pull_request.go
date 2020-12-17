package cli
/*
import (
	"os"
	"path/filepath"

	"github.com/arkenproject/ait/display"
	"github.com/arkenproject/ait/utils"
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
			"Unable to remove the old clone of "+url+", please delete the folder at\n"+
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
}*/
