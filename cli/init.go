package cli

import (
	"fmt"
	"os"

	"github.com/DataDrake/cli-ng/v2/cmd"
)

func init() {
	cmd.Register(&Init)
}

// Init configures Ark's local staging and configuration directory.
var Init = cmd.Sub{
	Name:  "init",
	Alias: "i",
	Short: "Initialize a dataset's local configuration.",
	Args:  &InitArgs{},
	Run:   InitRun,
}

// InitArgs handles the specific arguments for the init command.
type InitArgs struct {
}

// InitRun creates a new ark repo simply by creating a folder called .ark in the working dir.
func InitRun(r *cmd.Root, c *cmd.Sub) {
	// Setup main application config.
	rFlags := rootInit(r)

	// Check if .ark directory already exists.
	info, err := os.Stat(".ark")

	// If .ark does not exist create it.
	if os.IsNotExist(err) {
		err := os.Mkdir(".ark", os.ModePerm)
		checkError(rFlags, err)

		wd, err := os.Getwd()
		checkError(rFlags, err)
		fmt.Printf("New ark repo initiated at %v\n", wd)
		return
	}

	// Check that there wasn't another type of error produced.
	checkError(rFlags, err)

	// If no error was produced it's possible the directory is already an
	// Ark repository.
	if info.IsDir() {
		fmt.Println("a directory called \".ark\" already exists here, " +
			"suggesting that this is already an ark repo")
	} else {
		// If somehow a .ark file exists tell the user about it.
		fmt.Println("a file called \".ark\" already exists in this " +
			"this directory and it is not itself a directory. Please move or " +
			"rename this file")
	}
	os.Exit(1)
}
