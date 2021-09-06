package cli

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/DataDrake/cli-ng/v2/cmd"
	"github.com/arken/ark/config"
	"github.com/arken/ark/ipfs"
	"github.com/arken/ark/manifest"
	"github.com/schollz/progressbar/v3"
)

func init() {
	cmd.Register(&Upload)
}

// UploadArgs handles the specific arguments for the upload command.
type UploadArgs struct {
	Manifest string
}

// Upload begins seeding your files to an Arken Cluster once your
// submission into the Manifest has been merged into the repository.
var Upload = cmd.Sub{
	Name:  "upload",
	Alias: "up",
	Short: "Upload files to an Arken cluster after an accepted submission.",
	Args:  &UploadArgs{},
	Run:   UploadRun,
}

// UploadRun handles the uploading and display of the upload command.
func UploadRun(r *cmd.Root, c *cmd.Sub) {
	// Setup main application config.
	rFlags := rootInit(r)

	// Parse upload args
	args := c.Args.(*UploadArgs)

	// Check if .ark directory already exists.
	info, err := os.Stat(".ark")

	// If .ark does not exist notify the user to run
	// ark init() first.
	if os.IsNotExist(err) || !info.IsDir() {
		fmt.Printf("This is not an Ark repository! Please run\n\n" +
			"    ark init\n\n" +
			"Before attempting to upload any files.\n",
		)
		os.Exit(1)
	}

	// +--------------------+
	// |    Load Manifest   |
	// +--------------------+

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

	// +--------------------+
	// |   Load IPFS Node   |
	// +--------------------+
	ipfs, err := ipfs.CreateNode(
		filepath.Join(manifestPath, "ipfs"),
		ipfs.NodeConfArgs{
			SwarmKey:       manifest.ClusterKey,
			BootstrapPeers: manifest.BootstrapPeers,
		},
	)
	checkError(rFlags, err)

	// Open previous cache if exists
	f, err := os.Open(AddedFilesPath)
	if err != nil && os.IsNotExist(err) {
		fmt.Println(0, "file(s) currently staged for submission & upload")
		fmt.Println("Are you in the correct directory?")
		return
	}
	checkError(rFlags, err)
	defer f.Close()

	// Count the number of files in the manifest
	numFiles, err := lineCounter(f)
	checkError(rFlags, err)

	_, err = f.Seek(0, 0)
	checkError(rFlags, err)

	// In order to not copy files to ~/.ark/ipfs/
	// we need to create a workdir symlink in .ark
	wd, err := os.Getwd()
	checkError(rFlags, err)

	link := filepath.Join(config.Global.Manifest.Path, manifestName, "workdir")
	err = os.Symlink(wd, link)
	if err != nil && os.IsExist(err) {
		os.Remove(link)
		err = os.Symlink(wd, link)
	}
	checkError(rFlags, err)

	input := make(chan string, numFiles)

	// Add files to internal ipfs node
	go func() {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			cid, err := ipfs.Add(filepath.Join(link, scanner.Text()), false)
			checkError(rFlags, err)

			input <- cid
		}
	}()

	// Display progress bar for uploads.
	fmt.Println("Uploading Files to Cluster")
	ipfsBar := progressbar.Default(int64(numFiles))
	ipfsBar.RenderBlank()

	go func(bar *progressbar.ProgressBar, input chan string) {
		for cid := range input {
			replications, err := ipfs.FindProvs(cid, 20)
			checkError(rFlags, err)
			if rFlags.Verbose {
				fmt.Printf("\nFile: %s is backed up %d time(s)\n", cid, replications)
			}
			if replications > 2 {
				bar.Add(1)
			} else {
				bar.Add(0)
				input <- cid
			}
			if replications == 0 {
				err = ipfs.Pin(cid)
				checkError(rFlags, err)
			}
		}
	}(ipfsBar, input)

	for {
		if ipfsBar.State().CurrentPercent == float64(1) {
			close(input)
			err = os.Remove(link)
			checkError(rFlags, err)
			break
		}
		ipfsBar.Add(0)
		time.Sleep(1000 * time.Millisecond)
	}
}
