package cli

import (
	"errors"
	"fmt"
	"log"
	"os"

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
		if err != nil {
			log.Fatal(err)
		}
	} else if info.IsDir() { //TODO: should this re-initialize the way git does?
		log.Fatal(errors.New("a directory called \".ait\" already exists here, " +
			"suggesting that this is already an ait repo"))
	} else {
		log.Fatal(errors.New("a file called \".ait\" already exists in this " +
			"this directory and it is not itself a directory. Please move or " +
			"rename this file"))
	}
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("New ait repo initiated at %v", wd)
}
