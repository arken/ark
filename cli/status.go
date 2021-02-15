package cli

import (
	"fmt"
	"os"

	"github.com/arken/ait/types"
	"github.com/arken/ait/utils"

	"github.com/DataDrake/cli-ng/v2/cmd"
)

// Status prints out what files are currently staged for submission.
var Status = cmd.Sub{
	Name:  "status",
	Alias: "s",
	Short: "View what files are currently staged for submission.",
	Args:  &StatusArgs{},
	Run:   StatusRun,
}

// StatusArgs handles the specific arguments for the status command.
type StatusArgs struct {
}

// StatusRun executes the status function.
func StatusRun(*cmd.Root, *cmd.Sub) {
	file, err := os.OpenFile(utils.AddedFilesPath, os.O_RDONLY, 0644)
	if err == nil {
		defer file.Close()
	}
	lines := types.NewSortedStringSet()
	utils.FillSet(lines, file)
	if lines.Size() > 0 {
		fmt.Println(lines.Size(), "file(s) currently staged for submission:")
		_ = lines.ForEach(func(line string) error {
			fmt.Println("\t", line)
			return nil
		})
	} else {
		fmt.Println("No files are currently staged for submission.")
	}
}
