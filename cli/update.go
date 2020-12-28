package cli

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/DataDrake/cli-ng/cmd"
	"github.com/arkenproject/ait/utils"
	"github.com/inconshreveable/go-update"
	"github.com/tcnksm/go-latest"
)

var appVersion string

// Update checks for a new version of the AIT program and updates itself
// if a newer version is found and the user agrees to update.
var Update = cmd.CMD{
	Name:  "update",
	Alias: "upd",
	Short: "Update AIT to the lastest version.",
	Args:  &UpdateArgs{},
	Flags: &UpdateFlags{},
	Run:   UpdateRun,
}

// UpdateArgs handles the specific arguments for the update command.
type UpdateArgs struct {
}

// UpdateFlags handles the specific flags for the update command.
type UpdateFlags struct {
	Yes bool `short:"y" long:"yes" desc:"If a newer version is found update without prompting the user."`
}

// UpdateRun handles the checking and self updating of the AIT program.
func UpdateRun(r *cmd.RootCMD, c *cmd.CMD) {
	fmt.Printf("Current Version: %s\n", appVersion)

	flags := c.Flags.(*UpdateFlags)
	latestVersion := &latest.GithubTag{
		Owner:      "arkenproject",
		Repository: "ait",
	}

	res, _ := latest.Check(latestVersion, appVersion)
	fmt.Printf("Latest Version: %s\n", res.Current)

	if res.Outdated {
		if !flags.Yes {
			fmt.Println("Would you like to update AIT to the newest version? ([y]/n)")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.ToLower(strings.TrimSpace(input))
			if input == "n" {
				return
			}
		}
		url := "https://github.com/arkenproject/ait/releases/download/v" + res.Current + "/ait-v" + res.Current + "-" + runtime.GOOS + "-" + runtime.GOARCH

		resp, err := http.Get(url)
		utils.CheckError(err)

		defer resp.Body.Close()
		err = update.Apply(resp.Body, update.Options{})
		utils.CheckError(err)
	}
}
