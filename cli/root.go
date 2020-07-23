package cli

import (
	"github.com/DataDrake/cli-ng/cmd"
	"github.com/arkenproject/ait/gitlib"
)

//GlobalFlags contains the flags for commands.
type GlobalFlags struct{}

// Root is the main command.
var Root *cmd.RootCMD

// init creates the command interface and registers the possible commands.
func init() {
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
	Root.RegisterCMD(&gitlib.Clone)
	Root.RegisterCMD(&gitlib.Submit)
}
