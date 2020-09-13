package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/arkenproject/ait/utils"

	"github.com/DataDrake/cli-ng/cmd"
)

// Remove is the reverse of the add method. Given a set of file patterns, it
// un-stages all files that match any of the patterns. It also takes a special arg
// "--all" which will un-stage ALL files currently staged. This is the same
//behavior as "ait rm ." Note: this is NOT the same behavior as "ait rm *",
//since your shell will probably expand "*" into all non-hidden files (files
//that don't start with "."). So if you've added hidden files, to remove them
//use . or the --all flag.
var Remove = cmd.CMD{
	Name:  "remove",
	Alias: "rm",
	Short: "Remove a file or directory from AIT's tracked files.",
	Args:  &RemoveArgs{},
	Flags: &RemoveFlags{},
	Run:   RemoveRun,
}

// RemoveArgs handles the specific arguments for the remove command.
type RemoveArgs struct {
	Paths []string
}

// RemoveFlags handles the specific flags for the remove command.
type RemoveFlags struct {
	All bool `long:"all" desc:"remove all staged files."`
}

// RemoveRun executes the remove function.
func RemoveRun(_ *cmd.RootCMD, c *cmd.CMD) {
	flags := c.Flags.(*RemoveFlags)
	args := c.Args.(*RemoveArgs).Paths
	size, _ := utils.GetFileSize(utils.AddedFilesPath)
	if !utils.FileExists(utils.AddedFilesPath) || size == 0 {
		utils.FatalPrintln("no files currently staged, nothing was done")
	}
	if flags.All {
		file := utils.BasicFileOpen(utils.AddedFilesPath, os.O_TRUNC|os.O_WRONLY, 0644)
		file.Close()
		return
	}
	contents := make(map[string]struct{})
	file := utils.BasicFileOpen(utils.AddedFilesPath, os.O_RDONLY, 0644)
	utils.FillMap(contents, file)
	file.Close()
	numRMd := 0
	for _, pattern := range args {
		pattern = filepath.Clean(pattern)
		for path := range contents {
			if utils.PathMatch(pattern, path) {
				delete(contents, path)
				numRMd++
			}
		}
	}
	file = utils.BasicFileOpen(utils.AddedFilesPath, os.O_WRONLY|os.O_TRUNC, 0644)
	err := utils.DumpMap(contents, file)
	file.Close()
	utils.CheckError(err)
	fmt.Println(numRMd, "file(s) unstaged.")
}
