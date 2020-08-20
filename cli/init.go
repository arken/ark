package cli

import (
	"fmt"
	"os"

	"github.com/arkenproject/ait/utils"

	"github.com/DataDrake/cli-ng/cmd"
)

// Init configures AIT's local staging and configuration directory.
var Init = cmd.CMD{
	Name:  "init",
	Alias: "i",
	Short: "Initialize a dataset's local configuration.",
	Args:  &InitArgs{},
	Run:   InitRun,
}

// InitArgs handles the specific arguments for the init command.
type InitArgs struct {
}

//InitRun creates a new ait repo simply by creating a folder called .ait in the working dir.
func InitRun(r *cmd.RootCMD, c *cmd.CMD) {
	info, err := os.Stat(".ait")
	if os.IsNotExist(err) {
		err := os.Mkdir(".ait", os.ModePerm)
		utils.CheckError(err)
	} else if info.IsDir() {
		utils.FatalPrintln("a directory called \".ait\" already exists here, " +
			"suggesting that this is already an ait repo")
	} else {
		utils.FatalPrintln("a file called \".ait\" already exists in this " +
			"this directory and it is not itself a directory. Please move or " +
			"rename this file")
	}
	wd, err := os.Getwd()
	utils.CheckError(err)
	fmt.Printf("New ait repo initiated at %v\n", wd)
}
