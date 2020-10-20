package cli

import (
	"os"

	"github.com/DataDrake/cli-ng/cmd"
	"github.com/arkenproject/ait/utils"
)

//GlobalFlags contains the flags for commands.
type GlobalFlags struct{}

// Root is the main command.
var Root *cmd.RootCMD

// init creates the command interface and registers the possible commands.
func init() {
	isInit := utils.IndexOf(os.Args, "init") > 0
	isTesting := utils.IndexOf(os.Args, "-test.v") > 0 //Don't force init when testing
	if !utils.IsAITRepo() && !isInit && !isTesting {
		utils.FatalPrintln(`This is not an AIT repository! Please run
	ait init
Before issuing any other commands.`)
	}
	Root = &cmd.RootCMD{
		Name:  "ait",
		Short: "Arken Import Tool",
		Flags: &GlobalFlags{},
	}
	Root.RegisterCMD(&cmd.Help)
	Root.RegisterCMD(&Add)
	Root.RegisterCMD(&Init)
	Root.RegisterCMD(&Remove)
	Root.RegisterCMD(&Status)
	Root.RegisterCMD(&Submit)
	Root.RegisterCMD(&Upload)
}
