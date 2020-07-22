package cli

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"

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

const addedFilesPath string = ".ait/added_files" //can later be put somewhere more central

// AddRun Similar to "git add", this function adds files that match a given list of
// file matching patterns (can include *, ? wildcards) to a file. Currently this
// file is in .ait/added_files, and it contains paths relative to the program's
// working directory. Along the way, the filenames are put in a hashmap, so the
// specific order of the filenames in the file is unpredictable, but users should
// not be directly interacting with files in .ait anyway.
// TODO: prevent addition of files outside of the repo
func AddRun(_ *cmd.RootCMD, c *cmd.CMD) {
	args := c.Args.(*AddArgs)

	file, err := os.OpenFile(addedFilesPath, os.O_CREATE | os.O_RDONLY, 0644)
	if err != nil { //open it for reading its contents
		log.Fatal(err)
	}
	contents := make(map[string]struct{}) //basically a set. empty struct has 0 width.
	fillMap(contents, file)
	file.Close()
	file, err = os.OpenFile(addedFilesPath, os.O_TRUNC|os.O_WRONLY, 0644)
	//completely truncate the file to avoid duplicated filenames
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	for _, pattern := range args.Patterns {
		fmt.Println(args.Patterns)
		_ = filepath.Walk(".", func(fPath string, info os.FileInfo, err error) error {
			if PathMatch(pattern, fPath) {
				contents[fPath] = struct{}{}
			}
			return nil
		})
	}
	//dump the map's keys, which have to be unique, into the file.
	err = dumpMap(contents, file)
	if err != nil {
		log.Fatal(err)
	}
}

//Splits the given file by newline and adds each line to the given map.
func fillMap(contents map[string]struct{}, file *os.File) {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		if len(scanner.Text()) > 0 {
			contents[scanner.Text()] = struct{}{}
		}
	}
}

//Dumps all keys in the given map to the given file, separated by a newline.
func dumpMap(contents map[string]struct{}, file *os.File) error {
	for line := range contents {
		_, err := file.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}
