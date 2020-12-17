package github

import (
	"context"
	"fmt"
	"github.com/arkenproject/ait/display"
	"github.com/arkenproject/ait/utils"
	"github.com/google/go-github/v32/github"
	"io"
	"io/ioutil"
	"path/filepath"
)

func UploadFile(path string) {
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
	_, _, err = client.Repositories.CreateFile(ctx, cache.repo.owner,
		cache.repo.name, inRepoPath, opts)
	utils.CheckError(err)
	//TODO: add update and delete file
}

// path should be the path to the file in the repo, not locally
func getFileSHA(path string) string {
	if cache.keysetSHA != "" {
		return cache.keysetSHA
	}
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ctx := context.Background()
	client := getClient()
	opts := &github.RepositoryContentGetOptions{}
	_, contents, _, err := client.Repositories.GetContents(ctx, cache.repo.owner,
		cache.repo.name, dir, opts)
	utils.CheckError(err)
	for _, file := range contents {
		// fetch the metadata of all the files in the directory the keyset file
		// is supposed to go into.
		if *file.Name == base {
			cache.keysetSHA = *file.SHA
			return *file.SHA
		}
	}
	return "" //if the file didn't exist return empty string
}

func KeysetExistsInRepo(path string) bool {
	return getFileSHA(path) != ""
}

func GetRepoAppTemplates() io.ReadCloser {
	ctx := context.Background()
	client := getClient()
	opts := &github.RepositoryContentGetOptions{}
	reader, err := client.Repositories.DownloadContents(ctx, cache.repo.owner,
		cache.repo.name, "application.md", opts)
	if err != nil {
		return nil
	}
	return reader
}
