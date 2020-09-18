package utils

import (
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

func TestIsInRepo(t *testing.T) {
	wd, _ := os.Getwd()
	if filepath.Base(wd) == "utils" {
		_ = os.Chdir("..") //make sure I'm testing from project root
	}
	in, _ := IsInRepo(".")
	assert.True(t, in)
	in, _ = IsInRepo("cli")
	assert.True(t, in)
	in, _ = IsInRepo("utils/utils.g")
	assert.True(t, in)
	in, _ = IsInRepo("../")
	assert.False(t, in)
	in, _ = IsInRepo("../../../../")
	assert.False(t, in)
	in, _ = IsInRepo("/")
	assert.False(t, in)
	in, _ = IsInRepo("cli/../display/../config/../ipfs/../.ait")
	assert.True(t, in)
	in, _ = IsInRepo("cli/../display/../config/../ipfs/../.ait/../")
	assert.True(t, in)
	in, _ = IsInRepo("cli/../display/../config/../ipfs/../.ait/../../")
	assert.False(t, in)
	in, _ = IsInRepo("dirs/that/definitely/exist/../../../../../")
	assert.False(t, in)
	in, _ = IsInRepo("dirs/that/definitely/exist/../../../../")
	assert.True(t, in)
	in, _ = IsInRepo("dirs/that/definitely/exist/../../../../../../")
	assert.False(t, in)
	in, _ = IsInRepo("cli/../../ait/config")
	assert.True(t, in)
	in, _ = IsInRepo("../ait/")
	assert.True(t, in)
}
