package cli

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/DataDrake/cli-ng/v2/cmd"
	"github.com/arken/ark/config"
	"github.com/arken/ark/ipfs"
	"github.com/arken/ark/manifest"
	"github.com/arken/ark/manifest/upstream"
	"github.com/arken/ark/parser"
	"github.com/schollz/progressbar/v3"
)

func init() {
	cmd.Register(&Submit)
}

// Submit creates a manifest file and uploads it to the destination git repository.
var Submit = cmd.Sub{
	Name:  "submit",
	Alias: "sb",
	Short: "Submit your files to a manifest repository.",
	Args:  &SubmitArgs{},
	Flags: &SubmitFlags{},
	Run:   SubmitRun,
}

// SubmitArgs handles the specific arguments for the submit command.
type SubmitArgs struct {
	Manifest string
}

// SubmitFlags handles the specific flags for the submit command.
type SubmitFlags struct {
	IsPR bool `short:"p" long:"pull-request" desc:"Jump straight into submitting a pull request"`
}

// SubmitRun authenticates the user through our OAuth app and uses that to
// upload a manifest file generated locally, or makes a pull request if necessary.
func SubmitRun(r *cmd.Root, c *cmd.Sub) {
	// +--------------------+
	// |    Setup Command   |
	// +--------------------+

	// Setup main application config.
	rFlags := rootInit(r)

	// Parse upload args
	args := c.Args.(*SubmitArgs)

	// Parse upload args
	flags := c.Flags.(*SubmitFlags)

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

	// Open previous cache if exists
	f, err := os.Open(AddedFilesPath)
	if err != nil && os.IsNotExist(err) {
		fmt.Println("No files are currently added, nothing to submit. Use")
		fmt.Println("    ark add <files>...")
		fmt.Println("to add files for submission.")
		return
	}
	checkError(rFlags, err)
	defer f.Close()

	// Swap out an alias for the corresponding url
	alias, ok := config.Global.Manifest.Aliases[args.Manifest]
	if ok {
		args.Manifest = alias
	}

	// +--------------------+
	// |   Check Git Info   |
	// +--------------------+

	if config.Global.Git.Email == "" || config.Global.Git.Name == "" {
		err = queryUserSaveGitInfo()
		checkError(rFlags, err)

		err = config.WriteFile(rFlags.Config, &config.Global)
		checkError(rFlags, err)
	}

	// +--------------------+
	// |      Load Auth     |
	// +--------------------+

	if config.Global.Git.Token == "" {
		// Give the user a chance to change the account they logged in with
		// if it was incorrect.
		correctUser := false
		var guard upstream.Guard
		for !correctUser {
			// Launch upstream auth workflow if local Git Token is empty.
			guard, err = manifest.Auth(args.Manifest)
			if err != nil && err.Error() == "unknown upstream" {
				fmt.Println("Error: Ark was unable to identify a known upstream")
				fmt.Println("for your repository. Please use,")
				fmt.Println("\t\"ark config git.token YOUR-VALUE\"")
				fmt.Println("to set your git token manually before retrying")
				fmt.Println("your submission without -p")
			}
			checkError(rFlags, err)

			// Print out message to user about device code.
			printAuthCode(guard.GetCode(), guard.GetExpireInterval())

			// Begin polling process for authorization
			interval := time.Duration(guard.GetInterval()) * time.Second
			for {
				wait(interval)
				status, err := guard.CheckStatus()
				checkError(rFlags, err)

				if status == "slow_down" {
					interval = interval + 5*time.Second
					continue
				}
				if status != "authorization_pending" {
					break
				}
			}
			fmt.Print("\r")
			username, err := guard.GetUser()
			checkError(rFlags, err)

			correctUser = queryUserCorrect(username)
			fmt.Println()
		}

		// Save access information to internal memory
		config.Global.Git.Username, _ = guard.GetUser()
		config.Global.Git.Token = guard.GetAccessToken()

		// Ask the user if they would like to save their git credentials
		saveCreds := queryUserSaveCreds()
		if saveCreds {
			err = config.WriteFile(rFlags.Config, &config.Global)
			checkError(rFlags, err)
		}
	}

	// +--------------------+
	// |    Load Manifest   |
	// +--------------------+

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
		manifest.GitOptions{
			Name:     config.Global.Git.Name,
			Username: config.Global.Git.Username,
			Email:    config.Global.Git.Email,
			Token:    config.Global.Git.Token,
		},
	)
	checkError(rFlags, err)

	// +--------------------+
	// |    Display App     |
	// +--------------------+
	var app parser.Application
	overwriteOpt := ""
	for overwriteOpt != "o" && overwriteOpt != "a" {
		// Construct application path
		appPath := filepath.Join(".ark", "commit")

		// Check if an application is already in progress.
		_, err = os.Stat(appPath)
		if err != nil && os.IsNotExist(err) {
			new, err := os.Create(appPath)
			checkError(rFlags, err)

			_, err = new.WriteString(parser.SubmissionTemplate)
			checkError(rFlags, err)
		}

		// Show the user their application in their preferred editor
		cmd := exec.Command(config.Global.Core.Editor, appPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		err = cmd.Run()
		checkError(rFlags, err)

		appFile, err := os.Open(appPath)
		checkError(rFlags, err)

		buf, err := io.ReadAll(appFile)
		checkError(rFlags, err)
		appFile.Close()

		app, err = parser.ParseApplication(string(buf))
		checkError(rFlags, err)

		// Check for existing file
		prevPath := filepath.Join(
			config.Global.Manifest.Path,
			manifestName,
			"manifest",
			app.Category,
			app.Filename,
		)
		_, err = os.Stat(prevPath)
		if err == nil {
			overwriteOpt = queryUserAppendFile(
				filepath.Join(app.Category, app.Filename),
			)
			continue
		}
		overwriteOpt = "o"
	}

	// +--------------------+
	// |   Load IPFS Node   |
	// +--------------------+

	// Create internal IPFS node
	ipfs, err := ipfs.CreateNode(
		filepath.Join(manifestPath, "ipfs"),
		ipfs.NodeConfArgs{
			SwarmKey:       manifest.ClusterKey,
			BootstrapPeers: manifest.BootstrapPeers,
		},
	)
	checkError(rFlags, err)

	// +--------------------+
	// | Generate Manifest  |
	// +--------------------+

	// Count the number of files in the cache
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

	// Create manifest map
	files := make(map[string]string, numFiles)

	if overwriteOpt == "a" {
		// Check for existing file
		prevPath := filepath.Join(
			config.Global.Manifest.Path,
			manifestName,
			"manifest",
			app.Category,
			app.Filename,
		)
		prev, err := os.Open(prevPath)
		if err == nil {
			scanner := bufio.NewScanner(prev)
			for scanner.Scan() {
				data := strings.Fields(scanner.Text())
				files[data[0]] = data[1]
			}
		}
		prev.Close()
	}

	fmt.Println("Building Manifest...")

	// Generate progress bar
	ipfsBar := progressbar.Default(int64(numFiles))
	ipfsBar.RenderBlank()

	// Add files to internal ipfs node
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		cid, err := ipfs.Add(filepath.Join(link, scanner.Text()), true)
		checkError(rFlags, err)

		// Add file to map.
		files[cid] = scanner.Text()
		ipfsBar.Add(1)
	}

	// Remove Symlink to Working Directory
	err = os.Remove(link)
	checkError(rFlags, err)

	// Construct destination manifest path
	manPath := filepath.Join(
		config.Global.Manifest.Path,
		manifestName,
		"manifest",
		app.Category,
		app.Filename,
	)

	// +--------------------+
	// |  Upload Manifest   |
	// +--------------------+
	// Add place holders for PRs to use branches.
	var mainBranchName string
	var newBranchName string

	// Check if we should push direct to
	// the git repository or attempt to create a pull request.
	haveWrite, err := manifest.HaveWriteAccess()
	if err != nil && err.Error() != "unknown upstream" {
		checkError(rFlags, err)
	}

	if !haveWrite || flags.IsPR {
		// Force status to a PR if we don't have
		// write access to the repository.
		flags.IsPR = true

		// Setup a repository fork when creating a PR.
		err = manifest.Fork()
		checkError(rFlags, err)

		// Store main git branch name
		mainBranchName, err = manifest.GetBranchName()
		checkError(rFlags, err)

		// Construct a new branch name
		newBranchName = "submit/" + app.Filename

		// Pull an existing branch to update if possible.
		err = manifest.PullBranch(newBranchName)
		if err != nil {
			if err.Error() == "branch not found" {
				err = manifest.CreateBranch(newBranchName)
			}
			checkError(rFlags, err)
		}

		err = manifest.SwitchBranch(newBranchName)
		checkError(rFlags, err)
	}

	// Make destination manifest path
	err = os.MkdirAll(filepath.Dir(manPath), os.ModePerm)
	checkError(rFlags, err)

	// Create manifest file
	new, err := os.Create(manPath)
	checkError(rFlags, err)
	defer new.Close()

	// Generate manifest content from map.
	out, err := manifest.Generate(files)
	checkError(rFlags, err)

	// Write manifest to file
	_, err = new.WriteString(out)
	checkError(rFlags, err)

	// Close manifest file.
	err = new.Close()
	checkError(rFlags, err)

	// Commit changes to repository.
	err = manifest.Commit(manPath, app.Commit)
	checkError(rFlags, err)

	// Push changes to repository.
	err = manifest.Push()
	checkError(rFlags, err)

	// Open a PR if one doesn't already exist.
	if flags.IsPR {
		err := manifest.SearchPrByBranch(newBranchName)
		if err != nil {
			err = manifest.OpenPR(mainBranchName, app.Title, app.PRBody)
			checkError(rFlags, err)
		}

		// Switch back to the main manifest branch
		err = manifest.SwitchBranch(mainBranchName)
		checkError(rFlags, err)
	}

	fmt.Println("Completed Submission Successfully!")
	f.Close()
	os.Remove(filepath.Join(".ark", "commit"))
}

// printAuthCode prints the user's code in a pretty format.
func printAuthCode(code string, expiry int) {
	now := time.Now()
	expireTime := now.Add(time.Duration(expiry) * time.Second)
	minutes := math.Round(float64(expiry) / 60.0)
	fmt.Printf(
		`Go to https://github.com/login/device and enter the following code. You should
see a request to authorize "Ark GitHub Worker". Please authorize this request, but 
not if it's from anyone other than Ark GitHub Worker by Arken!
=================================================================
                            %v
=================================================================
This code will expire in about %v minutes at %v.

`, code, int(minutes), expireTime.Format("3:04 PM"))
}

// wait prints a pretty little animation while Ark waits for the user's to
// authenticate the app for an Upstream.
func wait(length time.Duration) {
	seconds := int(length.Seconds())
	ticker := time.NewTicker(time.Second)
	for i := 0; i < seconds; i++ {
		fmt.Printf("\r[%v] Checking in %v second(s)...",
			spinner[i%len(spinner)], seconds-i)
		<-ticker.C
	}
	ticker.Stop()
}

func queryUserCorrect(user string) bool {
	fmt.Println("Successfully authenticated as user", user)
	fmt.Printf("Is this correct? ([y]/n) ")

	// Collect input from user.
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	// Validate user accepted logged in status.
	return input != "n" && input != "no"
}

func queryUserSaveCreds() bool {
	fmt.Print("\nWould you like to save your access token for future submissions? (y/[n]) ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	return input == "y" || input == "yes"
}

func queryUserAppendFile(filePath string) string {
	fmt.Printf("\nA file already exists at %v in the repo.\n", filePath)
	fmt.Println("Do you want to overwrite it (o), append to it (a), rename yours (r),")
	fmt.Print("or abort (any other key)? ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	return input
}

func queryUserSaveGitInfo() error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("You don't appear to have an identity saved.\n" +
		"Please enter your name (spaces are ok): ")
	input, _ := reader.ReadString('\n')
	config.Global.Git.Name = strings.TrimSpace(input)
	fmt.Print("Please enter your email: ")
	input, _ = reader.ReadString('\n')
	config.Global.Git.Email = strings.TrimSpace(input)
	return nil
}
