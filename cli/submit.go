package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	//vv to differentiate between go-github and our github package
	aitgh "github.com/arkenproject/ait/apis/github"
	"github.com/arkenproject/ait/config"
	"github.com/arkenproject/ait/display"
	"github.com/arkenproject/ait/ipfs"
	"github.com/arkenproject/ait/keysets"
	"github.com/arkenproject/ait/utils"

	"github.com/DataDrake/cli-ng/cmd"
)

// Submit creates and uploads the keyset definition file.
var Submit = cmd.CMD{
	Name:  "submit",
	Alias: "sm",
	Short: "Submit your Keyset to a git repository.",
	Args:  &SubmitArgs{},
	Flags: &SubmitFlags{},
	Run:   SubmitRun,
}

// SubmitArgs handles the specific arguments for the submit command.
type SubmitArgs struct {
	Args []string
}

// SubmitFlags handles the specific flags for the submit command.
type SubmitFlags struct {
	IsPR bool `short:"p" long:"pull-request" desc:"Jump straight into submitting a pull request"`
}

// TODO: rewrite
func SubmitRun(_ *cmd.RootCMD, c *cmd.CMD) {
	url, isPR := parseSubmitArgs(c)
	fmt.Print("Initiating IPFS...\n\n")
	ipfs.Init(false)
	hasWritePerm := aitgh.Init(url, isPR)
	if config.Global.Git.Name == "" || config.Global.Git.Email == "" {
		promptNameEmail()
	}
	if !hasWritePerm && !isPR {
		isPR = promptDoPullRequest(url)
	}
	if isPR {
		fmt.Println("You chose to submit via pull request.")
		aitgh.CreateFork()
	}
	display.ShowApplication()
	overwrite := true
	app := display.ReadApplication()
	fileExists := aitgh.KeysetExistsInRepo(app.FullPath())
	for fileExists {
		var resolved bool
		overwrite, resolved = promptOverwriteConflict(app.FullPath())
		if resolved { break }
		app = display.ReadApplication()
		fileExists = aitgh.KeysetExistsInRepo(app.FullPath())
	}
	ksPath := filepath.Join(".ait", "keysets", "generated.ks")
	utils.CheckError(keysets.Generate(ksPath, overwrite))
	if !fileExists {
		aitgh.CreateFile(ksPath, app.FullPath(), app.Commit, isPR)
	} else {
		if overwrite {
			aitgh.ReplaceFile(ksPath, app.FullPath(), app.Commit, isPR)
		} else {
			aitgh.UpdateFile(ksPath, app.FullPath(), app.Commit, isPR)
		}
	}
	utils.SubmissionCleanup()
	if isPR {
		aitgh.CreatePullRequest(app.Title, app.PRBody)
	}
	fmt.Println("Submission successful!")
	if config.Global.Git.PAT == "" {
		promptSaveToken()
	}
}

func promptDoPullRequest(url string) bool {
	fmt.Printf(
`You don't appear to have write permissions for the %v.
Do you want to submit a pull request to the repository instead? This is the 
only way to continue the submission. (y/[n]) `, url)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	return input == "y"
}

func promptOverwriteConflict(path string) (bool, bool) {
	fmt.Printf(
`A file already exists at %v in the repo. Do you want to 
overwrite it (o), append to it (a), rename yours (r), or abort (any other key)? `, path)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "o" {
		return true, true
	} else if input == "a" {
		localPath := filepath.Join(".ait", "keysets", "generated.ks")
		utils.CheckError(aitgh.DownloadFile(path, localPath))
		return false, true
	} else if input == "r" {
		display.ShowApplication()
	} else {
		utils.FatalPrintln("Submission aborted.")
	}
	return true, false
}

// promptNameEmail asks the user to enter their name and email for git purposes.
// this is saved into the file at ~/.ait/ait.config
func promptNameEmail() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("We don't appear to have an identity saved for you."+
				  "Please enter your name (spaces are ok): ")
	input, _ := reader.ReadString('\n')
	config.Global.Git.Name = strings.TrimSpace(input)
	fmt.Print("Please enter your email: ")
	input, _ = reader.ReadString('\n')
	config.Global.Git.Email = strings.TrimSpace(input)
	config.GenConf(config.Global)
}

func promptSaveToken() {
	fmt.Print("\nWould you like to save your personal access token for future submissions? (y/[n]) ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "y" {
		fmt.Print(`Please note that the token will be stored in plain text. It can be utilized by a 
savvy attacker to modify your GitHub account and take actions on your behalf.
Saving the token is not recommended if you share this computer with other people.
Are you sure you want to save it? (y/[n]) `)
		input, _ = reader.ReadString('\n')
		input = strings.ToLower(strings.TrimSpace(input))
		if input == "y" {
			aitgh.SaveToken()
		}
	}
}

// parseSubmitArgs simply does some of the sanitization and extraction required to
// get the desired data structures out of the cmd.CMD object, then returns said
// useful data structures.
func parseSubmitArgs(c *cmd.CMD) (string, bool) {
	args := c.Args.(*SubmitArgs).Args
	if len(args) < 1 {
		utils.FatalPrintln("Not enough arguments, expected repository url")
	}
	url := config.GetRemote(args[0])
	if url != args[0] {
		fmt.Printf("Submitting to the remote at %v\n", url)
	}
	if s, _ := utils.GetFileSize(utils.AddedFilesPath); s == 0 {
		utils.FatalPrintln(`No files are currently added, nothing to submit. Use
    ait add <files>...
to add files for submission.`)
	}
	return url, c.Flags.(*SubmitFlags).IsPR
}
