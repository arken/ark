package cli

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"sync"
	"time"

	"github.com/DataDrake/cli-ng/v2/cmd"
	"github.com/arken/ark/config"
)

// AddedFilesPath is the file cache location.
const AddedFilesPath string = ".ark/added_files"

//GlobalFlags contains the flags for commands.
type GlobalFlags struct {
	Config  string `short:"c" long:"config" desc:"Specify a custom config path."`
	Verbose bool   `short:"v" long:"verbose" desc:"Show More Information"`
}

var Root = &cmd.Root{
	Name:    "ark",
	Short:   "The Arken command-line client.",
	Version: config.Version,
	License: "Licensed under the Apache License, Version 2.0",
	Flags:   &GlobalFlags{},
}

// rootInit initializes the main application config
// from the root global flag location.
func rootInit(r *cmd.Root) *GlobalFlags {
	// Parse Root Flags
	rFlags := r.Flags.(*GlobalFlags)

	// Construct config path
	var path string
	if rFlags.Config != "" {
		path = rFlags.Config
	} else {
		user, err := user.Current()
		checkError(rFlags, err)
		path = filepath.Join(user.HomeDir, ".ark", "config.toml")
		rFlags.Config = path
	}

	// Initialize config from path location
	err := config.Init(path)
	checkError(rFlags, err)

	// Return setup root flags
	return rFlags
}

// checkError checks an error and returns either a pretty
// or debug error report based on the verbosity of the
// application.
func checkError(flags *GlobalFlags, err error) {
	if err != nil {
		if flags.Verbose {
			log.Fatal(err)
		}
		fmt.Println(err)
		os.Exit(1)
	}
}

// spinner is an array of the progression of the spinner.
var spinner = []string{"|", "/", "-", "\\"}

// spinnerWait displays a spinner which should be done in a
// separate go routine.
func spinnerWait(done <-chan int, message string, wg *sync.WaitGroup) {
	ticker := time.NewTicker(time.Millisecond * 128)
	frameCounter := 0
	for range ticker.C {
		select {
		case <-done:
			wg.Done()
			ticker.Stop()
			return
		default:
			<-ticker.C
			ind := frameCounter % len(spinner)
			fmt.Printf("\r[%v] "+message, spinner[ind])
			frameCounter++
		}
	}
}
