package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DataDrake/cli-ng/v2/cmd"
)

// This function tests the performance of a large remove operation by adding
// everything in your documents folder and removing it. The time the rm itself
// took is printed.
func BenchmarkBigRm(b *testing.B) {
	u, _ := os.UserHomeDir()
	testRoot := filepath.Join(u, "Documents/")
	_ = os.Chdir(testRoot)
	_ = os.RemoveAll(testRoot + "/.ait")
	InitRun(nil, nil)
	addArgs := &cmd.Sub{
		Args: &StageArgs{Paths: []string{"."}},
	}
	StageRun(nil, addArgs)
	files, _ := ioutil.ReadDir(testRoot)
	var fileNames []string
	for _, fi := range files {
		fileNames = append(fileNames, fi.Name())
	}
	rmArgs := &cmd.Sub{
		Args:  &UnstageArgs{Paths: fileNames},
		Flags: &UnstageFlags{All: false},
	}
	start := time.Now()
	UnstageRun(nil, rmArgs)
	fmt.Println("\n\t******** Rm all folders took",
		time.Since(start).Milliseconds(), "ms ********\n ")
	_ = os.RemoveAll(testRoot + "/.ait")
}

func BenchmarkRmExtensions(b *testing.B) {
	u, _ := os.UserHomeDir()
	testRoot := filepath.Join(u, "Documents/")
	_ = os.Chdir(testRoot)
	_ = os.RemoveAll(testRoot + "/.ait")
	InitRun(nil, nil)
	addArgs := &cmd.Sub{
		Args: &StageArgs{Paths: []string{"."}},
	}
	StageRun(nil, addArgs)
	ext := "java,c,json,md,js"
	rmArgs := &cmd.Sub{
		Flags: &UnstageFlags{
			All:        false,
			Extensions: ext,
		},
	}
	start := time.Now()
	UnstageRun(nil, rmArgs)
	fmt.Println("\n\t******** Rm", ext, "took",
		time.Since(start).Milliseconds(), "ms ********\n ")
	_ = os.RemoveAll(testRoot + "/.ait")
}
