package cli

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/DataDrake/cli-ng/cmd"

	"github.com/arkenproject/ait/config"
	"github.com/arkenproject/ait/utils"
)

// AddRemote allows users to create aliases for GitHub remotes
var AddRemote = cmd.CMD{
	Name:  "remote",
	Alias: "r",
	Short: "Manage saved remote URLs.",
	Flags: &RemoteFlags{},
	Args:  &RemoteArgs{},
	Run:   RemoteRun,
}

// RemoteArgs handles the specific arguments for the add-remote command.
type RemoteArgs struct {
	Args []string
}

type RemoteFlags struct {
	IsAdd bool `short:"a" long:"add" desc:"Add a new remote alias"`
	IsRm bool `short:"d" long:"remove" desc:"Remove a remote alias"`
	IsList bool `short:"l" long:"list" desc:"List your saved aliases"`
}

// RemoteRun handles saving aliases for GitHub remotes.
func RemoteRun(_ *cmd.RootCMD, c *cmd.CMD) {
	flags := c.Flags.(*RemoteFlags)
	validateFlags(flags.IsAdd, flags.IsRm, flags.IsList) //makes sure exactly one flag is present
	args := c.Args.(*RemoteArgs).Args
	if len(args) < 2 && flags.IsAdd {
		utils.FatalPrintln(`Expected an alias and a URL to add:
	ait remote --add MyAlias https://github.com/example-user/example-repo.git`)
	} else if len(args) < 1 && flags.IsRm {
		utils.FatalPrintln(`Expected an alias to remove:
	ait remote -d MyAlias`)
	}

	if flags.IsAdd {
		alias, url := args[0], args[1]
		if aliasIsURL, _ := utils.IsGithubRemote(alias); aliasIsURL {
			utils.FatalPrintln(`It appears that your alias is a URL. The alias should come first:
	ait add-remote MyAlias https://github.com/example-user/example-repo.git`)
		}
		validateURL(url)
		addRemote(alias, url)
	} else if flags.IsRm {
		alias := args[0]
		removeRemote(alias)
	} else if flags.IsList {
		listRemotes()
	}
	config.GenConf(config.Global)
}

func addRemote(alias, url string) {
	if config.Global.Git.Remotes == nil {
		config.Global.Git.Remotes = make(map[string]string)
	}
	oldVal, ok := config.Global.Git.Remotes[alias]
	if ok { // Alias is already a key in the map
		fmt.Printf(`The alias "%v" is already mapped to %v.
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
	fmt.Printf("Alias \"%v\" successfully mapped to %v.\n", alias, url)
}

func removeRemote(alias string) {
	remotes := config.Global.Git.Remotes
	if remotes == nil || len(remotes) == 0 {
		fmt.Println("No aliases are currently saved, nothing was done.")
		return
	}
	oldVal, contains := remotes[alias]
	if !contains {
		fmt.Printf(
			"There are no saved remotes that go by the alias \"%v\". Nothing was done.\n", alias)
	} else {
		delete(remotes, alias)
		fmt.Printf(
			"Alias \"%v\" which mapped to %v has been deleted.\n", alias, oldVal)
	}
}

func listRemotes() {
	remotes := config.Global.Git.Remotes
	if len(remotes) == 0 {
		fmt.Println("No saved remote aliases.")
	} else {
		fmt.Println(len(remotes), "saved aliases:")
		maxLen := 0
		for alias := range config.Global.Git.Remotes {
			maxLen = int(math.Max(float64(maxLen), float64(len(alias))))
		}
		for alias, url := range config.Global.Git.Remotes {
			spaces := maxLen - len(alias)
			fmt.Print("\t\"", alias, "\" = ")
			for i := 0; i < spaces; i++ {
				fmt.Print(" ")
			}
			fmt.Println(url)
		}
	}
}

// validateURL uses utils.IsGithubRemote to detect obvious problems with the
// url. If it sees any, it asks the user if they would like to add the remote
// regardless, and if yes the program continues as expected. If not, the program
// is terminated immediately.
func validateURL(url string) {
	ok, msg := utils.IsGithubRemote(url)
	if !ok {
		if len(msg) > 0 {
			fmt.Printf("There is something wrong with the "+
				"provided URL \"%v\":\n%v\n", url, msg)
		} else {
			fmt.Printf("There is something wrong with the "+
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

func validateFlags(isAdd, isRm, isList bool) {
	if !isAdd && !isRm && !isList { //no flags provided
		utils.FatalPrintln(`Expected a flag to indicate an operation:
	ait remote --add/-a MyAlias https://github.com/example-user/example-repo.git
	ait remote --remove/-d MyAlias
	ait remote --list/-l`)
	} else if (isAdd && isRm) || (isAdd && isList) || (isRm && isList) {
		utils.FatalPrintln(`Too many flags! Please just pick one operation:
	ait remote --add/-a MyAlias https://github.com/example-user/example-repo.git
	ait remote --remove/-d MyAlias
	ait remote --list/-l`)
	}
}
