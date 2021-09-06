package cli

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/DataDrake/cli-ng/v2/cmd"
)

func init() {
	cmd.Register(&Add)
}

// Add stages a file or set of files for a submission.
var Add = cmd.Sub{
	Name:  "add",
	Alias: "ad",
	Short: "Stage a file for set of files for a submission.",
	Args:  &AddArgs{},
	Run:   AddRun,
}

// AddArgs handles the specific arguments for the add command.
type AddArgs struct {
	Paths []string
}

// AddRun stages a file within the current working directory for a later submission.
func AddRun(r *cmd.Root, c *cmd.Sub) {
	// Setup main application config.
	rFlags := rootInit(r)

	// Check if .ark directory already exists.
	info, err := os.Stat(".ark")

	// If .ark does not exist notify the user to run
	// ark init() first.
	if os.IsNotExist(err) || !info.IsDir() {
		fmt.Printf("This is not an Ark repository! Please run\n\n" +
			"    ark init\n\n" +
			"Before attempting to add any files.\n",
		)
		os.Exit(1)
	}

	// Initialize file cache and paths from args
	fileCache := make(map[string]bool)
	argPaths := c.Args.(*AddArgs).Paths

	// Open previous cache if exists
	f, err := os.Open(AddedFilesPath)
	if err == nil {
		// Import existing cache from file.
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if len(scanner.Text()) > 0 {
				fileCache[scanner.Text()] = true
			}
		}
		f.Close()
	}

	// Iterate through paths to check they exist
	// if adding a dir add all sub files within that dir.
	for _, path := range argPaths {
		stat, err := os.Stat(path)
		checkError(rFlags, err)

		// Walk through a directory and add all children files.
		if stat.IsDir() {
			filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
				if !info.IsDir() {
					fileCache[path] = true
				}
				return nil
			})
		} else {
			fileCache[path] = true
		}
	}

	// Create a string of the keys of the cache map.
	keys := make([]string, 0, len(fileCache))
	for k := range fileCache {
		keys = append(keys, k)
	}

	f, err = os.Create(AddedFilesPath)
	checkError(rFlags, err)
	defer f.Close()

	// Write out cache to file.
	_, err = f.WriteString(strings.Join(keys, "\n") + "\n")
	checkError(rFlags, err)

}
