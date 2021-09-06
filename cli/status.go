package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/DataDrake/cli-ng/v2/cmd"
)

func init() {
	cmd.Register(&Status)
}

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

// StatusRun handles the execution of the status command.
func StatusRun(r *cmd.Root, c *cmd.Sub) {
	// Setup main application config.
	rFlags := rootInit(r)

	// Check if .ark directory already exists.
	info, err := os.Stat(".ark")

	// If .ark does not exist notify the user to run
	// ark init() first.
	if os.IsNotExist(err) || !info.IsDir() {
		fmt.Printf("This is not an Ark repository! Please run\n\n" +
			"    ark init\n\n" +
			"Before attempting to and any files.\n",
		)
		os.Exit(1)
	}

	// Open previous cache if exists
	f, err := os.Open(AddedFilesPath)
	if err != nil && os.IsNotExist(err) {
		fmt.Println(0, "file(s) currently staged for submission")
		return
	}
	checkError(rFlags, err)
	defer f.Close()

	lines, err := lineCounter(f)
	checkError(rFlags, err)

	_, err = f.Seek(0, 0)
	checkError(rFlags, err)

	fmt.Println(lines, "file(s) currently staged for submission")
	if lines <= 50 {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if len(scanner.Text()) > 0 {
				fmt.Println("\t", scanner.Text())
			}
		}
	}
}

// efficient lineCounter function from
// https://stackoverflow.com/a/24563853
func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}
