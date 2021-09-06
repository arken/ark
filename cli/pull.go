package cli

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/DataDrake/cli-ng/v2/cmd"
	"github.com/arken/ark/config"
	"github.com/arken/ark/ipfs"
	"github.com/arken/ark/manifest"
	files "github.com/ipfs/go-ipfs-files"
)

func init() {
	cmd.Register(&Pull)
}

// Pull downloads files from an Arken cluster.
var Pull = cmd.Sub{
	Name:  "pull",
	Alias: "pl",
	Short: "Pull a file from an Arken Cluster.",
	Args:  &PullArgs{},
	Run:   PullRun,
}

// PullArgs handles the specific arguments for the pull command.
type PullArgs struct {
	Manifest  string
	Filepaths []string
}

// PullRun handles pulling and saving a file from an Arken cluster.
func PullRun(r *cmd.Root, c *cmd.Sub) {
	// Setup main application config.
	rFlags := rootInit(r)

	// Parse command arguments.
	args := c.Args.(*PullArgs)

	// Get current working directory
	currentwd, err := os.Getwd()
	checkError(rFlags, err)

	// Swap out an alias for the corresponding url
	alias, ok := config.Global.Manifest.Aliases[args.Manifest]
	if ok {
		args.Manifest = alias
	}

	// Parse manifest url
	urlPath, err := url.Parse(args.Manifest)
	checkError(rFlags, err)

	// Extract manifest name from url
	manifestName := filepath.Base(urlPath.Path)

	// Generate internal manifest path from name
	manifestPath := filepath.Join(config.Global.Manifest.Path, manifestName)

	// Initialize Manifest
	manifest, err := manifest.Init(
		filepath.Join(manifestPath, "manifest"),
		args.Manifest,
		manifest.GitOptions{},
	)
	checkError(rFlags, err)

	// Create internal IPFS node for manifest
	ipfs, err := ipfs.CreateNode(
		filepath.Join(manifestPath, "ipfs"),
		ipfs.NodeConfArgs{
			SwarmKey:       manifest.ClusterKey,
			BootstrapPeers: manifest.BootstrapPeers,
		},
	)
	checkError(rFlags, err)

	for _, path := range args.Filepaths {
		results, err := manifest.Search(path)
		checkError(rFlags, err)

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
					checkError(rFlags, err)

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
			go spinnerWait(doneChan, "Pulling "+filename+"...", &wg)

			// Pull file over IPFS
			file, err := ipfs.Get(cids[i])
			checkError(rFlags, err)

			// Write IPFS file out to filesystem
			err = files.WriteTo(file, filepath.Join(currentwd, filename))
			if err != nil {
				fmt.Printf("Could not write out the fetched CID: %s", err)
				os.Exit(1)
			}

			doneChan <- 0
			wg.Wait()

			fmt.Println()
			close(doneChan)
		}
	}

}
