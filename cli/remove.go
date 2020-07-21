package cli

import (
	"bufio"
	"errors"
	"os"
)

//This method is the reverse of the add method. Given a set of file patterns, it
//un-stages all files that match any of the patterns. It also takes a special arg
//"--all" which will un-stage any and all files currently staged. Currently, this is
//the same behavior as if it was passed "."
//To remove lines from added_files, this function creates a temporary file which
//holds all the lines that will stay from added_files, and at the end, it deletes
//the original added_files and renames the temp file to be the new added_files.
func Remove(args []string) error {
	if !FileExists(AddedFilesPath) || GetFileSize(AddedFilesPath) == 0 {
		return errors.New("no files currently staged, nothing was done")
	} else if len(args) == 0 {
		return errors.New("no files specified, nothing was done")
	}
	if args[0] == "--all" || args[0] == "." {
		file, err := os.OpenFile(AddedFilesPath, os.O_TRUNC | os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		file.Close()
		return nil
	}
	addedFiles, err1 := os.OpenFile(AddedFilesPath, os.O_RDONLY, 0644)
	if err1 != nil {
		return err1
	}
	tempFile, err2 := os.OpenFile(".ait/temp", os.O_WRONLY | os.O_CREATE, 0644)
	if err2 != nil {
		return err2
	}
	scanner := bufio.NewScanner(addedFiles)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		for _, pattern := range args {
			if !PathMatch(pattern, scanner.Text()) {
				_, err := tempFile.WriteString(scanner.Text() + "\n")
				if err != nil {
					return err
				}
			}
		}
	}
	tempFile.Close()
	addedFiles.Close()
	err := os.Remove(AddedFilesPath)
	if err != nil {
		return err
	}
	return os.Rename(".ait/temp", AddedFilesPath)
}
