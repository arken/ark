package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/DataDrake/cli-ng/cmd"

	"github.com/arkenproject/ait/config"
	"github.com/arkenproject/ait/utils"
)

// AddRemote allows users to create aliases for GitHub remotes
var AddRemote = cmd.CMD{
	Name:  "add-remote",
	Alias: "ar",
	Short: "Save a remote URL to submit to later.",
	Args:  &AddRemoteArgs{},
	Run:   AddRemoteRun,
}

// AddRemoteArgs handles the specific arguments for the add-remote command.
type AddRemoteArgs struct {
	Args []string
}

// AddRemoteRun handles saving aliases for GitHub remotes.
func AddRemoteRun(_ *cmd.RootCMD, c *cmd.CMD) {
	args := c.Args.(*AddRemoteArgs).Args
	if len(args) < 2 {
		utils.FatalPrintln(`Expected a URL and an alias:
	ait add-remote alias https://github.com/...`)
	}
	alias, url := args[0], args[1]
	validateURL(url)
	if config.Global.Git.Remotes == nil {
		config.Global.Git.Remotes = make(map[string]string)
	}
	oldVal, ok := config.Global.Git.Remotes[alias]
	if ok { // Alias is already a key in the map
		fmt.Printf(`The alias "%v" is already mapped to "%v".
Would you like to proceed regardless (y) or abort (any other key)? `,
			alias, oldVal)
		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		if !strings.EqualFold(strings.TrimSpace(choice), "y") {
			utils.FatalPrintln("Aborting.")
		}
	}
	config.Global.Git.Remotes[alias] = url
	config.GenConf(config.Global)
	fmt.Printf("Alias \"%v\" successfully mapped to \"%v\".\n", alias, url)
}

// validateURL uses utils.IsGithubRemote to detect obvious problems with the
// url. If it sees any, it asks the user if they would like to add the remote
// regardless, and if yes the program continues as expected. If not, the program
// is terminated immediately.
func validateURL(url string) {
	ok, msg := utils.IsGithubRemote(url)
	if !ok {
		if len(msg) > 0 {
			fmt.Printf("There is something wrong with the " +
				"provided URL \"%v\":\n%v\n", url, msg)
		} else {
			fmt.Printf("There is something wrong with the " +
				"provided URL \"%v\".", url)
		}
		fmt.Print("\nWould you like to proceed regardless (y) or abort (any other key)? ")
		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		if !strings.EqualFold(strings.TrimSpace(choice), "y") {
			utils.FatalPrintln("Aborting.")
		}
	}
}
