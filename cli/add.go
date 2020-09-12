package cli

import (
	"os"
	"path/filepath"

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
	Patterns []string
}

// AddRun Similar to "git add", this function adds files that match a given list of
// file matching patterns (can include *, ? wildcards) to a file. Currently this
// file is in .ait/added_files, and it contains paths relative to the program's
// working directory. Along the way, the filenames are put in a hashmap, so the
// specific order of the filenames in the file is unpredictable, but users should
// not be directly interacting with files in .ait anyway.
func AddRun(_ *cmd.RootCMD, c *cmd.CMD) {
	args := c.Args.(*AddArgs)

	contents := make(map[string]struct{}) //basically a set. empty struct has 0 width.
	file := utils.BasicFileOpen(utils.AddedFilesPath, os.O_CREATE|os.O_RDONLY, 0644)
	utils.FillMap(contents, file)
	file.Close()
	//completely truncate the file to avoid duplicated filenames
	file = utils.BasicFileOpen(utils.AddedFilesPath, os.O_TRUNC|os.O_WRONLY, 0644)
	defer file.Close()
	for _, pattern := range args.Patterns {
		pattern = filepath.Clean(pattern)
		if pattern == "*" {
			pattern = "."
			//You would never get here if you wrote "ait rm *" in a shell because
			//the shell should expand that. You'll only get here if you can get
			//the args to this program without going through a shell, like with
			//an IDE. This will have different behavior to going through a shell,
			//namely that hidden files won't be omitted, as they are by some
			//shells (like bash). If not going through a shell, it's best to be
			//more specific with you arguments. Otherwise, let the shell do the work.
		}
		_ = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() && utils.PathMatch(pattern, path) &&
				!utils.IsInSubDir(path, ".ait") {
				contents[path] = struct{}{}
			}
			return nil
		})
	}
	//dump the map's keys, which have to be unique, into the file.
	err := utils.DumpMap(contents, file)
	utils.CheckError(err)
}
