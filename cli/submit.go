package cli

import (
	"bufio"
	"fmt"
	"github.com/arkenproject/ait/display"
	"github.com/arkenproject/ait/keysets"
	"os"
	"path/filepath"
	"strings"

	aitgh "github.com/arkenproject/ait/apis/github"
	//^ to differentiate between go-github and our github package
	"github.com/arkenproject/ait/config"
	"github.com/arkenproject/ait/ipfs"
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

// submitFields is a simple struct to hold github username and password and other
// fields the user has to fill in/choose.
type submitFields struct {
	// ksGenMethod is whether to overwrite or amend to existing keyset files.
	ksGenMethod string
}

// doOverwrite returns false if the struct's ksGenMethod is equal to "a" (amend
// or append), false otherwise.
func (c *submitFields) doOverwrite() bool {
	return c.ksGenMethod != "a"
}

var fields submitFields

// SubmitRun generates a keyset file and then clones the Github repo at the given
// url, adds the keyset file, commits it, and pushes it, and then deletes the repo
// once everything is done or if anything goes wrong before completion. With all
// of those steps, there are MANY possible points of failure. If anything goes
// wrong, the error will be PrintFatal'd and the repo will we deleted from
// its temporary location at .ait/sources. Users are not meant to deal with the
// repos directly at any point so it and the keyset file are basically ephemeral
// and only exist on disk while this command is running.
func SubmitRun(_ *cmd.RootCMD, c *cmd.CMD) {
	url, isPR := parseSubmitArgs(c)
	fmt.Println("Initiating IPFS...")
	ipfs.Init(false)
	aitgh.Init(url, isPR)
	if config.Global.Git.Name == "" || config.Global.Git.Email == "" {
		promptNameEmail()
	}
	if isPR {
		fmt.Println("You chose to submit via pull request.")
		aitgh.Fork()
	}
	display.ShowApplication()
	overwrite := true
	app := display.ReadApplication()
	fileExists := aitgh.KeysetExistsInRepo(app.FullPath())
	for fileExists {
		overwrite = promptOverwriteConflict(app.FullPath())
		// User can rewrite application in ^ so I have to refresh the app
		app = display.ReadApplication()
	}
	ksPath := filepath.Join(".ait", "keysets", "generated.ks")
	if !fileExists {
		utils.CheckError(keysets.Generate(ksPath, overwrite))
		aitgh.CreateFile(ksPath, app.FullPath(), app.Commit, isPR)
	} else {
		if overwrite {
			aitgh.UpdateFile(ksPath, app.FullPath(), app.Commit, isPR)
		} else {
			aitgh.ReplaceFile(ksPath, app.FullPath(), app.Commit, isPR)
		}
	}
	utils.SubmissionCleanup()
	fmt.Println("Submission successful!")
	if config.Global.Git.PAT == "" {
		promptSaveToken()
	}
}

func promptSaveToken() {
	fmt.Println("Would you like to save your personal access token for future submits? (y/[n])")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "y" {
		aitgh.SaveToken()
	}
}

func promptOverwriteConflict(path string) bool {
	fmt.Printf(
`A file already exists at %v in the repo. Do you want to 
overwrite it (o), append to it (a), rename yours (r), or abort (any other key)? `, path)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "o" {
		return true
	} else if input == "u" {
		localPath := filepath.Join(".ait", "keysets", filepath.Base(path))
		aitgh.DownloadFile(path, localPath)
		return false
	} else if input == "r" {
		display.ShowApplication()
	} else {
		utils.FatalPrintln("Submission aborted.")
	}
	return false
}

// promptNameEmail asks the user to enter their name and email for git purposes.
// this is saved into the file at ~/.ait/ait.config
func promptNameEmail() {
	fmt.Println("We don't appear to have an identity saved for you.")
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please enter your name (spaces are ok): ")
	input, _ := reader.ReadString('\n')
	config.Global.Git.Name = strings.TrimSpace(input)
	fmt.Print("Please enter your email: ")
	input, _ = reader.ReadString('\n')
	config.Global.Git.Email = strings.TrimSpace(input)
	config.GenConf(config.Global)
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
