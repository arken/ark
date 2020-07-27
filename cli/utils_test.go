package cli

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
