package cli

import (
	"os"
)

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
