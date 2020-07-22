package cli

import (
	"errors"
	"github.com/DataDrake/cli-ng/cmd"
	"log"
	"os"
	"path/filepath"
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
	Patterns []string
}

// RemoveFlags handles the specific flags for the remove command.
type RemoveFlags struct {
	All bool `long:"all" desc:"remove all staged files."`
}

// RemoveRun executes the remove function.
func RemoveRun(_ *cmd.RootCMD, c *cmd.CMD) {
	flags := c.Flags.(*RemoveFlags)
	args := c.Args.(*RemoveArgs)
	size, _ := GetFileSize(addedFilesPath)
	if !FileExists(addedFilesPath) || size == 0 {
		log.Fatal(errors.New("no files currently staged, nothing was done"))
	}
	if flags.All {
		file, err := os.OpenFile(addedFilesPath, os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		file.Close()
		return
	}
	file, err := os.OpenFile(addedFilesPath, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	contents := make(map[string]struct{})
	FillMap(contents, file)
	for _, pattern := range args.Patterns {
		if pattern == "*" {
			pattern = "." //see AddRun for a description of why this is done
		}
		for path := range contents {
			pattern = filepath.Clean(pattern)
			if PathMatch(pattern, path) {
				delete(contents, path)
			}
		}
	}
	file.Close()
	file, err = os.OpenFile(addedFilesPath, os.O_WRONLY | os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	err = DumpMap(contents, file)
	if err != nil {
		log.Fatal(err)
	}
}
