package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/arken/ait/types"
	"github.com/arken/ait/utils"

	"github.com/DataDrake/cli-ng/v2/cmd"
)

// Unstage is the reverse of the stage method. Given a set of file patterns, it
// un-stages all files that match any of the patterns. It also takes a special arg
// "--all" which will un-stage ALL files currently staged. This is the same
// behavior as "ait un ." Note: this is NOT the same behavior as "ait un *",
// since your shell will probably expand "*" into all non-hidden files (files
// that don't start with "."). So if you've added hidden files, to remove them
// use . or the --all flag.
var Unstage = cmd.Sub{
	Name:  "unstage",
	Alias: "un",
	Short: "Unstage a file or directory from AIT's tracked files.",
	Args:  &UnstageArgs{},
	Flags: &UnstageFlags{},
	Run:   UnstageRun,
}

// UnstageArgs handles the specific arguments for the remove command.
type UnstageArgs struct {
	Paths []string
}

// UnstageFlags handles the specific flags for the remove command.
type UnstageFlags struct {
	All        bool   `long:"all" desc:"unstage all currently staged files"`
	Extensions string `short:"e" long:"extension" desc:"Unstage all files with the given file extension. For multiple extensions, separate each with a comma"`
}

// UnstageRun executes the remove function.
func UnstageRun(_ *cmd.Root, c *cmd.Sub) {
	args, exts, rmAll := parseUnstageArgs(c)
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
	if exts.Size() > 0 && len(args) == 0 {
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

// parseUnstageArgs simply does some of the sanitization and extraction required to
// get the desired data structures out of the cmd.Sub object, then returns said
// useful data structures.
func parseUnstageArgs(c *cmd.Sub) ([]string, *types.BasicStringSet, bool) {
	var args []string
	if c.Args != nil {
		args = c.Args.(*UnstageArgs).Paths
	}
	rmAll := false
	exts := types.NewBasicStringSet()
	ind := utils.IndexOf(os.Args, "-e")
	if c.Flags != nil && ind == -1 {
		//They used the "... -e=png,jpg ..." syntax
		rmAll = c.Flags.(*UnstageFlags).All
		extStr := c.Flags.(*UnstageFlags).Extensions
		exts = splitExtensions(extStr)
	} else if ind > 0 && ind+1 < len(os.Args) {
		//They used the "... -e png,jpg ..." syntax
		extStr := os.Args[ind+1]
		exts = splitExtensions(extStr)
		ind = utils.IndexOf(args, extStr)
		args = append(args[0:ind], args[ind+1:]...)
		//^remove the extension(s) from what cli-ng thinks is the args
	}
	if len(args) == 0 && !rmAll && exts.Size() == 0 {
		utils.FatalPrintln("No arguments provided, nothing was done")
	}
	return args, exts, rmAll
}
