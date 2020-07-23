package cli

import (
	"bufio"
	"fmt"
	"github.com/arkenproject/ait/utils"
	"os"
	"sort"

	"github.com/DataDrake/cli-ng/cmd"
)

// Status prints out what files are currently staged for submission.
var Status = cmd.CMD{
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
func StatusRun(*cmd.RootCMD, *cmd.CMD) {
	file, err := os.OpenFile(utils.AddedFilesPath, os.O_RDONLY, 0644)
	if err == nil {
		defer file.Close()
	}
	lines := make([]string, 0, 10)
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 {
			lines = append(lines, line)
		}
	}
	if len(lines) > 0 {
		sort.Strings(lines)
		fmt.Println("Files currently staged for submission:")
		for _, line := range lines {
			fmt.Println("\t", line)
		}
	} else {
		fmt.Println("No files are currently staged for submission.")
	}
}
