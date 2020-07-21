package cli

import (
	"log"
	"os"
)

func Init() {
	info, statErr := os.Stat(".ait")
	if os.IsNotExist(statErr) {
		dirErr := os.Mkdir(".ait", os.ModeDir)
		if dirErr != nil {
			log.Fatal(dirErr)
		}
	} else if !info.IsDir() { //a non-dir file called ".ait" already exists

	}
}

//trivial check to see if the program's working dir is an ait repo.
func IsAITRepo() bool {
	return FileExists(".ait")
}

func FileExists(filename string) bool {
	_, statErr := os.Stat(filename)
	return !os.IsNotExist(statErr)
}

func GetFileSize(filename string) int64 {
	info, err := os.Stat(filename)
	if err != nil {
		return 0
	} else {
		return info.Size()
	}
}