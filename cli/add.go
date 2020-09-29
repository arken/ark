package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"

	"github.com/DataDrake/cli-ng/cmd"
	"github.com/arkenproject/ait/types"
	"github.com/arkenproject/ait/utils"
)

// Add imports a file or directory to AIT's local staging file.
var Add = cmd.CMD{
	Name:  "add",
	Alias: "a",
	Short: "Add a file or directory to AIT's tracked files.",
	Args:  &AddArgs{},
	Flags: &AddFlags{},
	Run:   AddRun,
}

// AddArgs handles the specific arguments for the add command.
type AddArgs struct {
	Paths []string
}

type AddFlags struct {
	Extension string `short:"e" long:"extension" desc:"Add all files with the given file extension. For multiple extensions, separate each with a comma"`
}

var threads int32 = 0

// AddRun Similar to "git add", this function adds files that match a given list of
// file matching patterns (can include *, ? wildcards) to a file. Currently this
// file is in .ait/added_files, and it contains paths relative to the program's
// working directory. Along the way, the filenames are put in a hashmap, so the
// specific order of the filenames in the file is unpredictable, but users should
// not be directly interacting with files in .ait anyway.
func AddRun(_ *cmd.RootCMD, c *cmd.CMD) {
	runtime.GOMAXPROCS(512) //TODO: assign this number meaningfully
	args, exts := parseAddArgs(c)
	contents := make(map[string]struct{}) //basically a set. empty struct has 0 width.
	file := utils.BasicFileOpen(utils.AddedFilesPath, os.O_CREATE|os.O_RDONLY, 0644)
	utils.FillMap(contents, file)
	origLen := len(contents)
	file.Close()
	for _, userPath := range args {
		userPath = filepath.Clean(userPath)
		withinRepo, err := utils.IsWithinRepo(userPath)
		utils.CheckError(err)
		if withinRepo {
			addPath(userPath, contents)
		} else {
			fmt.Printf("Will not add files that are not in this ait repo," +
				" skipping %v", userPath)
		}
	}
	if exts.Size() > 0 {
		addExtension(contents, exts)
	}
	//completely truncate the file to avoid duplicated filenames
	file = utils.BasicFileOpen(utils.AddedFilesPath, os.O_TRUNC|os.O_WRONLY, 0644)
	defer file.Close()
	//dump the map's keys, which have to be unique, into the file.
	err := utils.DumpMap(contents, file)
	utils.CheckError(err)
	fmt.Println(len(contents) - origLen, "file(s) added")
}

// addPath attempts to add the given path to the current collection of added
// files. No attempt will be made if the file doesn't exist or it is already
// in the collection.
func addPath(userPath string, contents map[string]struct{}) {
	info, statErr := os.Stat(userPath)
	_, alreadyContains := contents[userPath]
	if !os.IsNotExist(statErr) && info != nil && !alreadyContains {
		// if file exists and isn't already in the map
		if info.IsDir() {
			c := make(chan string)
			atomic.AddInt32(&threads, 1)
			go processDir(userPath, c)
			for msg := range c {
				contents[msg] = struct{}{}
			}
		} else {
			contents[userPath] = struct{}{}
		}
	} else if os.IsNotExist(statErr) {
		fmt.Printf("Path \"%v\" not found. Continuing...\n", userPath)
	}
}

// processDir walks through the directory at dir and sends the path of all
// regular files back to the main thread via c. If another directory is found,
// another goproc is called to processDir that directory.
func processDir(dir string, c chan string) {
	defer func() {
		atomic.AddInt32(&threads, -1)
		if atomic.LoadInt32(&threads) <= 0 {
			close(c)
		}
	}()
	if dir == ".ait" {
		return
	}
	files, err := ioutil.ReadDir(dir)
	utils.CheckError(err)
	for _, info := range files {
		if info.IsDir() {
			atomic.AddInt32(&threads, 1)
			go processDir(filepath.Join(dir, info.Name()), c)
		} else {
			c <- filepath.Join(dir, info.Name())
		}
	}
}

// addExtension attempts to add ALL files within the current wd that have the
// extension(s) contained in exts.
func addExtension(contents map[string]struct{}, exts *types.StringSet) {
	c := make(chan string)
	atomic.AddInt32(&threads, 1)
	go processDirExt(".", c, exts)
	for msg := range c {
		contents[msg] = struct{}{}
	}
}

// processDirExt walks through the directory at dir and sends the path of all
// regular files that have the desired file extensions back to the main thread
// via c. If another directory is found, another goproc is called to
// processDirExt that directory.
func processDirExt(dir string, c chan string, exts *types.StringSet) {
	defer func() {
		atomic.AddInt32(&threads, -1)
		if atomic.LoadInt32(&threads) <= 0 {
			close(c)
		}
	}()
	if dir == ".ait" {
		return
	}
	files, err := ioutil.ReadDir(dir)
	utils.CheckError(err)
	for _, info := range files {
		if info.IsDir() {
			atomic.AddInt32(&threads, 1)
			go processDirExt(filepath.Join(dir, info.Name()), c, exts)
		} else if exts.Contains(filepath.Ext(info.Name())) {
			c <- filepath.Join(dir, info.Name())
		}
	}
}

// parseAddArgs simply does some of the sanitization and extraction required to
// get the desired data structures out of the cmd.CMD object, then returns said
// useful data structures.
func parseAddArgs(c *cmd.CMD) ([]string, *types.StringSet) {
	var exts = types.NewThreadSafeSet()
	extStr := ""
	if c.Flags != nil {
		extStr = c.Flags.(*AddFlags).Extension
		for _, extension := range strings.Split(extStr, ",") {
			extension = strings.TrimSpace(extension)
			if len(extension) > 0 {
				if !strings.HasPrefix(extension, ".") {
					extension = "." + extension
				}
				exts.Add(extension)
			}
		}
	}
	var args []string
	if c.Args != nil {
		args = c.Args.(*AddArgs).Paths
	}
	if exts.Size() == 0 && len(args) == 0 {
		fmt.Println("No files were given to add, please provide arguments")
		os.Exit(0)
	}
	return args, exts
}
