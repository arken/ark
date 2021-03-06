package utils

import (
	"fmt"
	"os"
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
	assert.Equal(t, "ait", GetRepoName("https://github.com/arken/ait.git"))
	assert.Equal(t, "ait", GetRepoName("git@github.com:arken/ait.git"))
	assert.Equal(t, "", GetRepoName(""))
	assert.Equal(t, "", GetRepoName("/"))
	assert.Equal(t, "core-keyset", GetRepoName("https://github.com/arken/core-keyset"))
	assert.Equal(t, "core-keyset", GetRepoName("https://github.com/arken/core-keyset.git"))
}

func TestGetRepoOwner(t *testing.T) {
	assert.Equal(t, "arken", GetRepoOwner("https://github.com/arken/ait.git"))
	assert.Equal(t, "arken", GetRepoOwner("https://github.com/arken/ait"))
	assert.Equal(t, "google", GetRepoOwner("https://github.com/google/go-github.git"))
	assert.Equal(t, "go-git", GetRepoOwner("https://github.com/go-git/go-git.git"))
	assert.Equal(t, "go-git", GetRepoOwner("https://github.com/go-git/go-git"))
	assert.Equal(t, "torvalds", GetRepoOwner("https://github.com/torvalds/linux.git"))
	assert.Equal(t, "torvalds", GetRepoOwner("https://github.com/torvalds/linux"))
	assert.Equal(t, "", GetRepoOwner("https://github.com//linux.git"))
	assert.Equal(t, "", GetRepoOwner(""))
	assert.Equal(t, "a", GetRepoOwner("123456789012345678/a/"))
}

func TestIsInRepo(t *testing.T) {
	wd, _ := os.Getwd()
	if filepath.Base(wd) == "utils" {
		_ = os.Chdir("..") //make sure I'm testing from project root
	}
	in, _ := IsWithinRepo(".")
	assert.True(t, in)
	in, _ = IsWithinRepo("cli")
	assert.True(t, in)
	in, _ = IsWithinRepo("utils/utils.g")
	assert.True(t, in)
	in, _ = IsWithinRepo("../")
	assert.False(t, in)
	in, _ = IsWithinRepo("../../../../")
	assert.False(t, in)
	in, _ = IsWithinRepo("/")
	assert.False(t, in)
	in, _ = IsWithinRepo("cli/../display/../config/../ipfs/../.ait")
	assert.True(t, in)
	in, _ = IsWithinRepo("cli/../display/../config/../ipfs/../.ait/../")
	assert.True(t, in)
	in, _ = IsWithinRepo("cli/../display/../config/../ipfs/../.ait/../../")
	assert.False(t, in)
	in, _ = IsWithinRepo("dirs/that/definitely/exist/../../../../../")
	assert.False(t, in)
	in, _ = IsWithinRepo("dirs/that/definitely/exist/../../../../")
	assert.True(t, in)
	in, _ = IsWithinRepo("dirs/that/definitely/exist/../../../../../../")
	assert.False(t, in)
	in, _ = IsWithinRepo("cli/../../ait/config")
	assert.True(t, in)
	in, _ = IsWithinRepo("../ait/")
	assert.True(t, in)
}

func TestIsGithubRemote(t *testing.T) {
	ok, _ := IsGithubRemote("https://github.com/arken/ait.git")
	assert.True(t, ok)
	_, msg := IsGithubRemote("")
	fmt.Println(msg)
	fmt.Println()
	_, msg = IsGithubRemote("git@github.com:arken/ait.git")
	fmt.Println(msg)
}
