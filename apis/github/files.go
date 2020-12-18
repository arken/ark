package github

import (
	"io/ioutil"
	"path/filepath"

	"github.com/arkenproject/ait/utils"

	"github.com/google/go-github/v32/github"
)

func CreateFile(localPath, repoPath, commit string, isPR bool) {
	file, err := ioutil.ReadFile(localPath)
	utils.CheckError(err)
	opts := &github.RepositoryContentFileOptions{
		Message:   github.String(commit),
		Content:   file,
	}
	owner := cache.upstream.owner
	if isPR {
		owner = *cache.user.Login
		// if it's a PR, the repo belongs to our user and not what we pulled out
		// of the original URL.
	}
	_, _, err = client.Repositories.CreateFile(cache.ctx, owner, cache.upstream.name,
		repoPath, opts)
	utils.CheckError(err)
}

func UpdateFile(localPath, repoPath, commit string, isPR bool) {
	file, err := ioutil.ReadFile(localPath)
	utils.CheckError(err)
	opts := &github.RepositoryContentFileOptions{
		Message:   github.String(commit),
		Content:   file,
		SHA: 	   github.String(getKeysetSHA(repoPath)),
	}
	owner := cache.upstream.owner
	if isPR {
		owner = *cache.user.Name
	}
	_, _, err = client.Repositories.UpdateFile(cache.ctx, owner, cache.upstream.name,
		repoPath, opts)
	utils.CheckError(err)
}

func ReplaceFile(localPath, repoPath, commit string, isPR bool) {
	file, err := ioutil.ReadFile(localPath)
	utils.CheckError(err)
	opts := &github.RepositoryContentFileOptions{
		Message:   github.String(commit),
		Content:   file,
		SHA: 	   github.String(getKeysetSHA(repoPath)),
	}
	owner := cache.upstream.owner
	if isPR {
		owner = *cache.user.Name
	}
	_, _, err = client.Repositories.DeleteFile(cache.ctx, owner, cache.upstream.name,
		repoPath, opts)
	utils.CheckError(err)
	opts.SHA = nil
	_, _, err = client.Repositories.CreateFile(cache.ctx, owner, cache.upstream.name,
		repoPath, opts)
}

// path should be the path to the file in the fork, not locally
func getKeysetSHA(ksPath string) string {
	if cache.keysetSHA != "" {
		return cache.keysetSHA
	}
	dir := filepath.Dir(ksPath)
	base := filepath.Base(ksPath)
	opts := &github.RepositoryContentGetOptions{}
	_, contents, _, err := client.Repositories.GetContents(cache.ctx, cache.upstream.owner,
		cache.upstream.name, dir, opts)
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
	return getKeysetSHA(path) != ""
}

func DownloadRepoAppTemplate() (string, error) {
	path := filepath.Join(".ait", cache.upstream.name + "_application.md")
	return path, DownloadFile("application.md", path)
}

// DownloadFile downloads the file at repoPath from the upstream repository to
// the given localPath
func DownloadFile(repoPath, localPath string) error {
	opts := &github.RepositoryContentGetOptions{}
	reader, err := client.Repositories.DownloadContents(cache.ctx, cache.upstream.owner,
		cache.upstream.name, repoPath, opts)
	if err != nil {
		return err //probably means the file didn't exist in the fork
	}
	data, err := ioutil.ReadAll(reader)
	utils.CheckError(err)
	err = ioutil.WriteFile(localPath, data, 0644)
	reader.Close()
	return nil
}
