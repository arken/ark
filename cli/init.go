package cli

import (
	"errors"
	"fmt"
	"os"
)

//Creates a new ait repo simply by creating a folder called .ait in the working dir.
func Init() error {
	info, statErr := os.Stat(".ait")
	if os.IsNotExist(statErr) {
		dirErr := os.Mkdir(".ait", os.ModeDir)
		if dirErr != nil {
			return dirErr
		}
	} else if info.IsDir() { //TODO: should this re-initialize the way git does?
		return errors.New("a directory called \".ait\" already exists here, " +
			"suggesting that this is already an ait repo")
	} else {
		return errors.New("a file called \".ait\" already exists in this " +
			"this directory and it is not itself a directory. Please move or " +
			"rename this file")
	}
	wd, _ := os.Getwd()
	fmt.Printf("New ait repo initiated at %v", wd)
	return nil
}
