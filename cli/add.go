package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arkenproject/ait/utils"

	"github.com/DataDrake/cli-ng/cmd"
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

// AddRun Similar to "git add", this function adds files that match a given list of
// file matching patterns (can include *, ? wildcards) to a file. Currently this
// file is in .ait/added_files, and it contains paths relative to the program's
// working directory. Along the way, the filenames are put in a hashmap, so the
// specific order of the filenames in the file is unpredictable, but users should
// not be directly interacting with files in .ait anyway.
func AddRun(_ *cmd.RootCMD, c *cmd.CMD) {
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
		if !alreadyContains && !os.IsNotExist(statErr) && info != nil {
			//if the path isn't already in the map and the file does exist
			if info.IsDir() {
				_ = filepath.Walk(userPath, func(diskPath string, info os.FileInfo, err error) error {
					_, contains := contents[diskPath]
					if !contains && !info.IsDir() && !strings.Contains(diskPath,
						".ait" + string(filepath.Separator)) {
						contents[diskPath] = struct{}{}
						numAdded++
					}
					return nil
				})
			} else {
				contents[userPath] = struct{}{}
				numAdded++
			}
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
