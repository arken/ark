package cli

import (
	"github.com/DataDrake/cli-ng/cmd"
	"os"
	"path/filepath"
	"testing"
)

func TestBigRm(t *testing.T) {
	u, _ := os.UserHomeDir()
	testRoot := filepath.Join(u, "Documents/")
	_ = os.Chdir(testRoot)
	_ = os.RemoveAll(testRoot + "/.ait")
	InitRun(nil, nil) //args are never used in InitRun, this is fine
	addArgs := &cmd.CMD{
		Args: &AddArgs{Paths: []string{"."}},
	}
	AddRun(nil, addArgs)
	rmArgs := &cmd.CMD{						//add your own folders/files here
		Args: &RemoveArgs{Paths: []string{}},
		Flags: &RemoveFlags{All: false},
	}
	RemoveRun(nil, rmArgs)
	_ = os.RemoveAll(testRoot + "/.ait")
}
