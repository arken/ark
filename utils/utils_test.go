package utils

import (
    "github.com/stretchr/testify/assert"
    "testing"
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
