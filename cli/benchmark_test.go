package cli

import (
	"github.com/DataDrake/cli-ng/cmd"
	"os"
	"testing"
)

func TestBigAdd(t *testing.T) {
	testRoot := "/home/danilo/Documents/"
	_ = os.Chdir(testRoot)
	_ = os.RemoveAll(testRoot + ".ait")
	InitRun(nil, nil) //args are never used in InitRun, this is fine
	addArgs := &cmd.CMD{
		Args: &AddArgs{Patterns: []string{"."}},
	}
	AddRun(nil, addArgs)
	_ = os.RemoveAll(testRoot + ".ait")
}
