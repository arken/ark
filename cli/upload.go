package cli

import (
	"os"
	"runtime"
	"time"

	"github.com/DataDrake/cli-ng/cmd"
	"github.com/arkenproject/ait/ipfs"
	"github.com/arkenproject/ait/utils"
	"github.com/schollz/progressbar/v3"
)

// Upload begins seeding your files to the Arken Cluster once your
// submission into the Keyset has been merged into the repository.
var Upload = cmd.CMD{
	Name:  "upload",
	Short: "After Submitting Your Files you can use AIT to Upload Them to the Arken Cluster.",
	Args:  &UploadArgs{},
	Run:   UploadRun,
}

// UploadArgs handles the specific arguments for the upload command.
type UploadArgs struct {
}

// UploadRun handles the uploading and display of the upload command.
func UploadRun(r *cmd.RootCMD, c *cmd.CMD) {
	contents := make(map[string]struct{}) // basically a set. empty struct has 0 width.
	file := utils.BasicFileOpen(utils.AddedFilesPath, os.O_CREATE|os.O_RDONLY, 0644)
	utils.FillMap(contents, file)
	file.Close()

	bar := progressbar.Default(int64(len(contents)))
	bar.RenderBlank()

	workers := genNumWorkers()

	input := make(chan string, len(contents))
	for path := range contents {
		cid, err := ipfs.Add(path)
		utils.CheckError(err)
		input <- cid
	}
	for i := 0; i < workers; i++ {
		go func(bar *progressbar.ProgressBar, input chan string) {
			for cid := range input {
				replications, err := ipfs.FindProvs(cid, 20)
				utils.CheckError(err)
				if replications >= 3 {
					bar.Add(1)
				} else {
					bar.Add(0)
					input <- cid
				}
			}
		}(bar, input)
	}

	for {
		if bar.State().CurrentPercent == float64(1) {
			close(input)
			return
		}
		bar.Add(0)
		time.Sleep(100 * time.Millisecond)
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
