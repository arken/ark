package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/DataDrake/cli-ng/cmd"
	"github.com/arkenproject/ait/config"
	"github.com/arkenproject/ait/ipfs"
	"github.com/arkenproject/ait/keysets"
	"github.com/arkenproject/ait/utils"
	files "github.com/ipfs/go-ipfs-files"
)

// Pull downloads files from the Arken cluster.
var Pull = cmd.CMD{
	Name:  "pull",
	Alias: "pl",
	Short: "Pull a file from the Arken Cluster.",
	Args:  &PullArgs{},
	Run:   PullRun,
}

// PullArgs handles the specific arguments for the pull command.
type PullArgs struct {
	Keyset    string
	Filepaths []string
}

// PullRun handles pulling and saving a file from the Arken cluster.
func PullRun(r *cmd.RootCMD, c *cmd.CMD) {
	args := c.Args.(*PullArgs)
	currentwd, err := os.Getwd()
	if err != nil {
		utils.FatalPrintln(err.Error())
	}

	user, err := user.Current()
	if err != nil {
		utils.FatalPrintln(err.Error())
	}

	// Initialize the IPFS subsystem without confirming the node is
	// reachable from the rest of the cluster.
	ipfs.Init(false)

	// Convert/Check URL against known alaises.
	url := config.GetRemote(args.Keyset)
	repoPath := filepath.Join(user.HomeDir, ".ait", "sources", utils.GetRepoName(url))

	// Clone/Update the keyset locally
	_, err = keysets.Clone(url, repoPath)
	if err != nil {
		utils.FatalPrintln(err.Error())
	}

	for pathNum := range args.Filepaths {
		results, err := keysets.Search(repoPath, args.Filepaths[pathNum])
		if err != nil {
			utils.FatalPrintln(err.Error())
		}

		for filename, cids := range results {
			i := 0
			if len(cids) > 1 {
				fmt.Printf("There is more than 1 file with the name: %s\n"+
					"Which version would you like to download?\n", filename)

				fmt.Printf("Select a number between 0 - %d\n", len(cids)-1)
				for i, hash := range cids {
					fmt.Printf("  | %d - %s", i, hash)
				}

				reader := bufio.NewReader(os.Stdin)
				for {
					text, err := reader.ReadString('\n')
					if strings.ToLower(text) == "exit" {
						return
					}
					i, err = strconv.Atoi(text)
					if err == nil && i >= 0 && i < len(cids) {
						break
					}
					fmt.Printf("Select a number between 0 - %d\n", len(cids)-1)
				}
			}

			// Display Spinner when pulling a file.
			doneChan := make(chan int, 1)
			wg := sync.WaitGroup{}
			wg.Add(1)

			go utils.SpinnerWait(doneChan, "Pulling "+filename+"...", &wg)
			file, err := ipfs.Pull(cids[i])
			err = files.WriteTo(file, filepath.Join(currentwd, filename))
			if err != nil {
				panic(fmt.Errorf("Could not write out the fetched CID: %s", err))
			}

			doneChan <- 0
			wg.Wait()

			fmt.Println()
			close(doneChan)
		}
	}

}
