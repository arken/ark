package cli

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/arken/ait/config"
	"github.com/arken/ait/utils"

	"github.com/DataDrake/cli-ng/v2/cmd"
)

// AddRemote allows users to create aliases for GitHub remotes
var AddRemote = cmd.Sub{
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

// RemoteFlags handles the specific flags for the add-remote command.
type RemoteFlags struct {
	IsAdd   bool `short:"a" long:"add" desc:"Stage a new remote alias"`
	IsRm    bool `short:"d" long:"delete" desc:"Unstage a remote alias"`
	IsRmAll bool `short:"D" long:"delete-all" desc:"Unstage all remote aliases"`
	IsList  bool `short:"l" long:"list" desc:"List your saved aliases"`
}

const usageEx = `	ait remote --add/-a MyAlias https://github.com/example-user/example-repo.git  # Saves an alias/URL pair for use later
	ait remote --delete/-d MyAlias      # Removes an alias/URL pair
	ait remote --delete-all/-D MyAlias  # Removes all alias/URL pairs
	ait remote --list/-l                # See all your saved alias/URL pairs`

// RemoteRun handles managing aliases for GitHub remotes.
func RemoteRun(_ *cmd.Root, c *cmd.Sub) {
	flags := c.Flags.(*RemoteFlags)
	validateFlags(flags.IsAdd, flags.IsRm, flags.IsList, flags.IsRmAll)
	// ^ makes sure exactly one flag is present
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
	ait remote --add MyAlias https://github.com/example-user/example-repo.git`)
		}
		validateURL(url)
		addRemote(alias, url)
	} else if flags.IsRm {
		alias := args[0]
		deleteRemote(alias)
	} else if flags.IsList {
		listRemotes()
	} else { //remove all
		deleteAllRemotes()
	}
	config.GenConf(config.Global)
}

// addRemote adds the given alias and url to the map config.Global.Git.Remotes.
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

// deleteRemote tries to delete the given alias and url from the the map
// config.Global.Git.Remotes.
func deleteRemote(alias string) {
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

// deleteAllRemotes clears the map config.Global.Git.Remotes.
func deleteAllRemotes() {
	remotes := config.Global.Git.Remotes
	if remotes == nil || len(remotes) == 0 {
		fmt.Println("No aliases are currently saved, nothing was done.")
		return
	}
	oLen := len(remotes)
	config.Global.Git.Remotes = make(map[string]string)
	fmt.Println(oLen, "alias(es) removed.")
}

// listRemotes lists aliases/url pairs in the map config.Global.Git.Remotes.
func listRemotes() {
	remotes := config.Global.Git.Remotes
	if len(remotes) == 0 {
		fmt.Println("No saved remote aliases.")
	} else {
		fmt.Println(len(remotes), "saved alias(es):")
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

// validateFlags makes sure that exactly one of the flags was provided. If none
// or more than one flag was provided, the program will terminate with an error
// message.
func validateFlags(flags ...bool) {
	oneTrue := false //At least one is true
	for i, bool1 := range flags {
		oneTrue = oneTrue || bool1
		for j := i; j < len(flags); j++ {
			if i != j && bool1 == true && flags[j] == true { //two flags are true
				utils.FatalPrintln("Too many flags! Please pick only one operation:\n" +
					usageEx)
			}
		}
	}
	if !oneTrue {
		utils.FatalPrintln("Expected one flag to indicate an operation:\n" +
			usageEx)
	}
}
