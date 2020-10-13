package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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
	Extensions string `short:"e" long:"extension" desc:"Add all files with the given file extension. For multiple extensions, separate each with a comma"`
}

var threads int32 = 0

// AddRun Similar to "git add", this function adds files that match a given list of
// file matching patterns (can include *, ? wildcards) to a file. Currently this
// file is in .ait/added_files, and it contains paths relative to the program's
// working directory. Along the way, the filenames are put into a set, so the
// specific order of the filenames in the file is unpredictable, but users should
// not be directly interacting with files in .ait anyway.
func AddRun(_ *cmd.RootCMD, c *cmd.CMD) {
	runtime.GOMAXPROCS(512) //TODO: assign this number meaningfully
	args, exts := parseAddArgs(c)
	contents := types.NewThreadSafeStringSet()
	file := utils.BasicFileOpen(utils.AddedFilesPath, os.O_CREATE|os.O_RDONLY, 0644)
	utils.FillSet(contents, file)
	origLen := contents.Size()
	file.Close()
	for _, userPath := range args {
		userPath = filepath.Clean(userPath)
		withinRepo, err := utils.IsWithinRepo(userPath)
		utils.CheckError(err)
		if withinRepo {
			addPath(userPath, contents)
		} else {
			fmt.Printf("Will not add files that are not in this ait repo," +
				" skipping %v\n", userPath)
		}
	}
	if exts.Size() > 0 {
		addExtension(contents, exts)
	}
	//completely truncate the file to avoid duplicated filenames
	file = utils.BasicFileOpen(utils.AddedFilesPath, os.O_TRUNC|os.O_WRONLY, 0644)
	defer file.Close()
	//dump the set, which has to have unique values, into the file.
	err := utils.DumpSet(contents, file)
	utils.CheckError(err)
	fmt.Println(contents.Size() - origLen, "file(s) added")
}

// addPath attempts to add the given path to the current collection of added
// files. No attempt will be made if the file doesn't exist or it is already
// in the collection.
func addPath(userPath string, contents *types.ThreadSafeStringSet) {
	info, statErr := os.Stat(userPath)
	if !os.IsNotExist(statErr) && info != nil && !contents.Contains(userPath) {
		// if file exists and isn't already in the set
		if info.IsDir() {
			wg := sync.WaitGroup{}
			wg.Add(1)
			go processDir(userPath, contents, &wg)
			wg.Wait()
		} else {
			contents.Add(userPath)
		}
	} else if os.IsNotExist(statErr) {
		fmt.Printf("Path \"%v\" not found. Continuing...\n", userPath)
	}
}

// processDir walks through the directory at dir and sends the path of all
// regular files back to the main thread via c. If another directory is found,
// another goproc is called to processDir that directory.
func processDir(dir string, contents *types.ThreadSafeStringSet, wg *sync.WaitGroup) {
	defer wg.Done()
	if dir == ".ait" {
		return
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Println("A thread encountered an error:", err)
		return
	}
	for _, info := range files {
		path := filepath.Join(dir, info.Name())
		if info.IsDir() {
			wg.Add(1)
			go processDir(path, contents, wg)
		} else {
			contents.Add(path)
		}
	}
}

// addExtension attempts to add ALL files within the current wd that have the
// extension(s) contained in exts.
func addExtension(contents *types.ThreadSafeStringSet, exts *types.BasicStringSet) {
	atomic.AddInt32(&threads, 1)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go processDirExt(".", exts, contents, &wg)
	wg.Wait()
}

// processDirExt walks through the directory at dir and sends the path of all
// regular files that have the desired file extensions back to the main thread
// via c. If another directory is found, another goproc is called to
// processDirExt that directory.
func processDirExt(dir string, exts *types.BasicStringSet, contents *types.ThreadSafeStringSet, wg *sync.WaitGroup) {
	defer wg.Done()
	if dir == ".ait" {
		return
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Println("A thread encountered an error:", err)
		return
	}
	for _, info := range files {
		path := filepath.Join(dir, info.Name())
		if info.IsDir() {
			wg.Add(1)
			go processDirExt(path, exts, contents, wg)
		} else if exts.Contains(filepath.Ext(info.Name())) {
			contents.Add(path)
		}
	}
}

// parseAddArgs simply does some of the sanitization and extraction required to
// get the desired data structures out of the cmd.CMD object, then returns said
// useful data structures.
func parseAddArgs(c *cmd.CMD) ([]string, *types.BasicStringSet) {
	var args []string
	if c.Args != nil {
		args = c.Args.(*AddArgs).Paths
	}
	var exts = types.NewBasicStringSet()
	ind := utils.IndexOf(os.Args, "-e")
	if c.Flags != nil && ind == -1 {
		//They used the "... -e=png,jpg ..." syntax
		extStr := c.Flags.(*AddFlags).Extensions
		exts = splitExtensions(extStr)
	} else if ind > 0 && ind + 1 < len(os.Args) {
		//They used the "... -e png,jpg ..." syntax
		extStr := os.Args[ind + 1]
		exts = splitExtensions(extStr)
		ind = utils.IndexOf(args, extStr)
		args = append(args[0:ind], args[ind + 1:]...)
		//^remove the extension(s) from what cli-ng thinks is the args
	}
	if exts.Size() == 0 && len(args) == 0 {
		fmt.Println("No files were given to add, please provide arguments")
		os.Exit(0)
	}
	return args, exts
}

// splitExtensions takes a string like "png,pdf,jpg" and returns a sanitized set
// of all extensions with no leading/trailing whitespace and no empty strings.
// They will also have "." appended to them, ie "png,pdf" -> { ".png", ".pdf" }
func splitExtensions(extStr string) *types.BasicStringSet {
	exts := types.NewBasicStringSet()
	for _, extension := range strings.Split(extStr, ",") {
		extension = strings.TrimSpace(extension)
		if len(extension) > 0 {
			if !strings.HasPrefix(extension, ".") {
				extension = "." + extension
			}
			exts.Add(extension)
		}
	}
	return exts
}
