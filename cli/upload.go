package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/DataDrake/cli-ng/cmd"
	"github.com/arkenproject/ait/config"
	"github.com/arkenproject/ait/ipfs"
	"github.com/arkenproject/ait/types"
	"github.com/arkenproject/ait/utils"
	"github.com/schollz/progressbar/v3"
)

// Upload begins seeding your files to the Arken Cluster once your
// submission into the Keyset has been merged into the repository.
var Upload = cmd.CMD{
	Name:  "upload",
	Alias: "up",
	Short: "After Submitting Your Files you can use AIT to Upload Them to the Arken Cluster.",
	Args:  &UploadArgs{},
	Flags: &UploadFlags{},
	Run:   UploadRun,
}

// UploadArgs handles the specific arguments for the upload command.
type UploadArgs struct {
}

// UploadFlags handles the specific flags for the upload command.
type UploadFlags struct {
	Debug bool `short:"d" long:"debug" desc:"Print Debug information to the console."`
}

// UploadRun handles the uploading and display of the upload command.
func UploadRun(r *cmd.RootCMD, c *cmd.CMD) {
	flags := c.Flags.(*UploadFlags)
	contents := types.NewBasicStringSet()
	file := utils.BasicFileOpen(utils.AddedFilesPath, os.O_CREATE|os.O_RDONLY, 0644)
	utils.FillSet(contents, file)
	file.Close()

	// In order to not copy files to ~/.ait/ipfs/ we need to create a workdir symlink
	// in .ait
	wd, err := os.Getwd()
	if err != nil {
		utils.FatalWithCleanup(utils.SubmissionCleanup, err.Error())
	}
	link := filepath.Join(filepath.Dir(config.Global.IPFS.Path), "workdir")
	err = os.Symlink(wd, link)
	defer os.Remove(link)

	workers := genNumWorkers()

	doneChan := make(chan int, 1)
	wg := sync.WaitGroup{}
	wg.Add(1)

	// Display Spinner on IPFS Init.
	go utils.SpinnerWait(doneChan, "Initializing IPFS...", &wg)
	ipfs.Init(true)
	doneChan <- 0
	wg.Wait()

	fmt.Print("\rInitializing IPFS: Done!")
	fmt.Println()
	close(doneChan)

	fmt.Println("Adding Files to IPFS Store")
	addBar := progressbar.Default(int64(contents.Size()))
	addBar.RenderBlank()

	input := make(chan string, contents.Size())
	_ = contents.ForEach(func(path string) error {
		cid, err := ipfs.Add(filepath.Join(link, path))
		utils.CheckError(err)

		addBar.Add(1)
		input <- cid
		return nil
	})

	fmt.Println("Uploading Files to Cluster")
	ipfsBar := progressbar.Default(int64(contents.Size()))
	ipfsBar.RenderBlank()

	for i := 0; i < workers; i++ {
		go func(bar *progressbar.ProgressBar, input chan string) {
			for cid := range input {
				replications, err := ipfs.FindProvs(cid, 20)
				if flags.Debug {
					fmt.Printf("File: %s is backed up %d time(s)\n", cid, replications)
				}
				if replications > 2 {
					bar.Add(1)
				} else {
					bar.Add(0)
					input <- cid
				}
				if replications == 0 {
					err = ipfs.Pin(cid)
					utils.CheckError(err)
				}
			}
		}(ipfsBar, input)
	}

	for {
		if ipfsBar.State().CurrentPercent == float64(1) {
			close(input)
			return
		}
		ipfsBar.Add(0)
		time.Sleep(1000 * time.Millisecond)
	}
}

// Generate the number of worker processes to optimize efficiency.
// Subtract 2 from the number of cores because of the main thread and the GetAll function.
func genNumWorkers() int {
	if runtime.NumCPU() > 2 {
		return runtime.NumCPU() - 1
	}
	return 1
}
