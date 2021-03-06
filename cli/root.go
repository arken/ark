package cli

import (
	"os"

	"github.com/arken/ait/utils"

	"github.com/DataDrake/cli-ng/v2/cmd"
)

//GlobalFlags contains the flags for commands.
type GlobalFlags struct{}

// Root is the main command.
var Root *cmd.Root

// init creates the command interface and registers the possible commands.
func init() {
	isHelp := len(os.Args) < 2 || utils.IndexOf(os.Args, "help") > 0
	isInit := utils.IndexOf(os.Args, "init") > 0 || utils.IndexOf(os.Args, "i") > 0
	isPull := utils.IndexOf(os.Args, "pull") > 0
	isRemote := utils.IndexOf(os.Args, "remote") > 0
	isUpdate := utils.IndexOf(os.Args, "update") > 0
	isTesting := utils.IndexOf(os.Args, "-test.v") > 0 //Don't force init when testing
	if !utils.IsAITRepo() && !isInit && !isTesting && !isPull && !isRemote && !isHelp && !isUpdate {
		utils.FatalPrintln(`This is not an AIT repository! Please run
	ait init
Before issuing any other commands.`)
	}
	Root = &cmd.Root{
		Name:  "ait",
		Short: "Arken Import Tool",
		Flags: &GlobalFlags{},
	}
	cmd.Register(&cmd.Help)
	cmd.Register(&Stage)
	cmd.Register(&AddRemote)
	cmd.Register(&Init)
	cmd.Register(&Unstage)
	cmd.Register(&Status)
	cmd.Register(&Submit)
	cmd.Register(&Upload)
	cmd.Register(&Pull)
	cmd.Register(&Update)
}
