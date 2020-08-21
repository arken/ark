package utils

import (
	"fmt"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathMatch(t *testing.T) {
	//note: wd for this test is ait/ci, not cli, thus the cd ..'s
	assert.True(t, IsInSubDir("../cli/*.go", "../cli"))
	assert.False(t, IsInSubDir("../.ait/added_files", "../cli"))
	assert.False(t, IsInSubDir("../cli", "../.ait/added_files"))
}

func TestGetRepoName(t *testing.T) {
	assert.Equal(t, "ait", GetRepoName("https://github.com/arkenproject/ait.git"))
	assert.Equal(t, "ait", GetRepoName("git@github.com:arkenproject/ait.git"))
	assert.Equal(t, "", GetRepoName(""))
	assert.Equal(t, "", GetRepoName("/"))
}

func TestGetRepoOwner(t *testing.T) {
	assert.Equal(t, "arkenproject", GetRepoOwner("https://github.com/arkenproject/ait.git"))
	assert.Equal(t, "google", GetRepoOwner("https://github.com/google/go-github.git"))
	assert.Equal(t, "go-git", GetRepoOwner("https://github.com/go-git/go-git.git"))
	assert.Equal(t, "torvalds", GetRepoOwner("https://github.com/torvalds/linux.git"))
	assert.Equal(t, "", GetRepoOwner("https://github.com//linux.git"))
	assert.Equal(t, "", GetRepoOwner(""))
	assert.Equal(t, "a", GetRepoOwner("123456789012345678/a/"))
}

func TestCopyFile(t *testing.T) {
	u, err := user.Current()
	CheckError(err)
	// Create expected config path.
	path := filepath.Join(u.HomeDir, ".ait", "application.md")
	err = CopyFile(path, "commit.md")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
}
