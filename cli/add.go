package cli

import (
	"fmt"
	"github.com/DataDrake/cli-ng/cmd"
	"github.com/arkenproject/ait/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync/atomic"
)

// Add imports a file or directory to AIT's local staging file.
var Add = cmd.CMD{
	Name:  "add",
	Alias: "a",
	Short: "Add a file or directory to AIT's tracked files.",
	Args:  &AddArgs{},
	Run:   AddRun,
}

// AddArgs handles the specific arguments for the add command.
type AddArgs struct {
	Paths []string
}

var threads int32 = 0

// AddRun Similar to "git add", this function adds files that match a given list of
// file matching patterns (can include *, ? wildcards) to a file. Currently this
// file is in .ait/added_files, and it contains paths relative to the program's
// working directory. Along the way, the filenames are put in a hashmap, so the
// specific order of the filenames in the file is unpredictable, but users should
// not be directly interacting with files in .ait anyway.
func AddRun(_ *cmd.RootCMD, c *cmd.CMD) {
	//runtime.GOMAXPROCS(2000) //macOS?
	args := c.Args.(*AddArgs).Paths
	contents := make(map[string]struct{}) //basically a set. empty struct has 0 width.
	file := utils.BasicFileOpen(utils.AddedFilesPath, os.O_CREATE|os.O_RDONLY, 0644)
	utils.FillMap(contents, file)
	file.Close()
	numAdded := 0
	for _, userPath := range args {
		userPath = filepath.Clean(userPath)
		info, statErr := os.Stat(userPath)
		_, alreadyContains := contents[userPath]
		if !os.IsNotExist(statErr) && info != nil && !alreadyContains {
			if !info.IsDir() {
				contents[userPath] = struct{}{}
			} else {
				c := make(chan string)
				atomic.AddInt32(&threads, 1)
				go processDir(userPath, c)
				for msg := range c {
					contents[msg] = struct{}{}
					numAdded++
				}
			}
		} else if os.IsNotExist(statErr) {
			fmt.Printf("Path \"%v\" found. Continuing...\n", userPath)
		}
	}
	//completely truncate the file to avoid duplicated filenames
	file = utils.BasicFileOpen(utils.AddedFilesPath, os.O_TRUNC|os.O_WRONLY, 0644)
	defer file.Close()
	//dump the map's keys, which have to be unique, into the file.
	err := utils.DumpMap(contents, file)
	utils.CheckError(err)
	fmt.Println(numAdded, "file(s) added")
}

func processDir(dir string, c chan string) {
	files, err := ioutil.ReadDir(dir)
	utils.CheckError(err)
	for _, info := range files {
		if info.IsDir() {
			atomic.AddInt32(&threads, 1)
			go processDir(filepath.Join(dir, info.Name()), c)
		} else {
			c <- info.Name()
		}
	}
	atomic.AddInt32(&threads, -1)
	if atomic.LoadInt32(&threads) <= 0 {
		close(c)
	}
}
