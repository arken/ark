package cli

import (
	"fmt"
	"github.com/arkenproject/ait/types"
	"github.com/arkenproject/ait/utils"
	"os"
	"path/filepath"

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
	All        bool   `long:"all" desc:"remove all staged files."`
	Extensions string `short:"e" long:"extension" desc:"Add all files with the given file extension. For multiple extensions, separate each with a comma"`
}

// RemoveRun executes the remove function.
func RemoveRun(_ *cmd.RootCMD, c *cmd.CMD) {
	args, exts, rmAll := parseRmArgs(c)
	size, _ := utils.GetFileSize(utils.AddedFilesPath)
	if !utils.FileExists(utils.AddedFilesPath) || size == 0 {
		utils.FatalPrintln("No files currently staged, nothing was done")
	} else if rmAll || (len(args) > 0 && args[0] == ".") {
		file := utils.BasicFileOpen(utils.AddedFilesPath, os.O_TRUNC|os.O_WRONLY, 0644)
		file.Close()
		fmt.Println("All files unstaged")
		return
	}
	contents := types.NewBasicStringSet()
	file := utils.BasicFileOpen(utils.AddedFilesPath, os.O_RDONLY, 0644)
	utils.FillSet(contents, file)
	file.Close()
	numRMd := 0
	if exts.Size() >0 && len(args) == 0 {
		args = append(args, ".")
	}
	for _, userPath := range args {
		userPath = filepath.Clean(userPath)
		_ = contents.ForEach(func(addedPath string) error {
			if utils.IsInSubDir(addedPath, userPath) || exts.Contains(filepath.Ext(addedPath)) {
				contents.Delete(addedPath)
				numRMd++
			}
			return nil
		})
	}
	file = utils.BasicFileOpen(utils.AddedFilesPath, os.O_WRONLY|os.O_TRUNC, 0644)
	err := utils.DumpSet(contents, file)
	file.Close()
	utils.CheckError(err)
	fmt.Println(numRMd, "file(s) unstaged")
}

// parseRmArgs simply does some of the sanitization and extraction required to
// get the desired data structures out of the cmd.CMD object, then returns said
// useful data structures.
func parseRmArgs(c *cmd.CMD) ([]string, *types.BasicStringSet, bool) {
	var args []string
	if c.Args != nil {
		args = c.Args.(*RemoveArgs).Paths
	}
	rmAll := false
	exts := types.NewBasicStringSet()
	ind := utils.IndexOf(os.Args, "-e")
	if c.Flags != nil && ind == -1 {
		//They used the "... -e=png,jpg ..." syntax
		rmAll = c.Flags.(*RemoveFlags).All
		extStr := c.Flags.(*RemoveFlags).Extensions
		exts = splitExtensions(extStr)
	} else if ind > 0 && ind + 1 < len(os.Args) {
		//They used the "... -e png,jpg ..." syntax
		extStr := os.Args[ind + 1]
		exts = splitExtensions(extStr)
		ind = utils.IndexOf(args, extStr)
		args = append(args[0:ind], args[ind + 1:]...)
		//^remove the extension(s) from what cli-ng thinks is the args
	}
	if len(args) == 0 && !rmAll && exts.Size() == 0 {
		utils.FatalPrintln("No arguments provided, nothing was done")
	}
	return args, exts, rmAll
}
