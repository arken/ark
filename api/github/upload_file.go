package github

import (
	"context"
	"fmt"
	"github.com/arkenproject/ait/display"
	"github.com/arkenproject/ait/utils"
	"github.com/google/go-github/v32/github"
	"io/ioutil"
	"path/filepath"
)

func UploadFile(url, path string) {
	owner := utils.GetRepoOwner(url)
	repoName := utils.GetRepoName(url)
	Cache.Repo = Repository{
		URL:   url,
		Owner: owner,
		Name:  repoName,
	}
	ctx := context.Background()
	client := getClient()
	app := display.ReadApplication()
	inRepoPath := fmt.Sprintf("%v/%v", app.Category, app.KsName)
	file, err := ioutil.ReadFile(path)
	utils.CheckError(err)
	opts := &github.RepositoryContentFileOptions{
		Message:   github.String(app.Commit),
		Content:   file,
		SHA:       nil,
	}
	_, _, err = client.Repositories.CreateFile(ctx, owner, repoName, inRepoPath, opts)
	utils.CheckError(err)
	//TODO: add update and delete file
}

// path should be the path to the file in the repo, not locally
func getFileSHA(path string) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ctx := context.Background()
	client := getClient()
	opts := &github.RepositoryContentGetOptions{}
	_, contents, _, err := client.Repositories.GetContents(ctx, Cache.Repo.Owner,
		Cache.Repo.Name, dir, opts)
	utils.CheckError(err)
	for _, file := range contents {
		// fetch the metadata of all the files in the keyset file is supposed to
		// go into.
		if *file.Name == base {
			return *file.SHA
		}
	}
	return "" //if the file didn't exist return empty string
}
