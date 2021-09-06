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
	cmd.Register(&Remove)
}

// Add stages a file or set of files for a submission.
var Remove = cmd.Sub{
	Name:  "remove",
	Alias: "rm",
	Short: "Remove a file from the internal submission cache.",
	Args:  &RemoveArgs{},
	Run:   RemoveRun,
}

// RemoveArgs handles the specific arguments for the remove command.
type RemoveArgs struct {
	Paths []string
}

// RemoveRun removes file from the submission cache.
func RemoveRun(r *cmd.Root, c *cmd.Sub) {
	// Setup main application config.
	rFlags := rootInit(r)

	// Check if .ark directory already exists.
	info, err := os.Stat(".ark")

	// If .ark does not exist notify the user to run
	// ark init() first.
	if os.IsNotExist(err) || !info.IsDir() {
		fmt.Printf("This is not an Ark repository! Please run\n\n" +
			"    ark init\n\n" +
			"Before attempting to remove any files.\n",
		)
		os.Exit(1)
	}

	// Initialize file cache and paths from args
	fileCache := make(map[string]bool)
	argPaths := c.Args.(*RemoveArgs).Paths

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
					delete(fileCache, path)
				}
				return nil
			})
		} else {
			delete(fileCache, path)
		}
	}

	// Create a string of the keys of the cache map.
	keys := make([]string, 0, len(fileCache))
	for k := range fileCache {
		keys = append(keys, k)
	}

	if len(keys) > 0 {
		f, err = os.Create(AddedFilesPath)
		checkError(rFlags, err)
		defer f.Close()

		// Write out cache to file.
		_, err = f.WriteString(strings.Join(keys, "\n") + "\n")
		checkError(rFlags, err)
	} else {
		err = os.Remove(AddedFilesPath)
		checkError(rFlags, err)
	}
}
