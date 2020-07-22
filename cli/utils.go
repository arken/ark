package cli

import (
	"os"

	"github.com/minio/minio/pkg/wildcard"
)

// IsAITRepo is a trivial check to see if the program's working dir is an ait repo.
func IsAITRepo() bool {
	return FileExists(".ait")
}

// FileExists is a test to check the existence of a file.
func FileExists(filename string) bool {
	_, statErr := os.Stat(filename)
	return !os.IsNotExist(statErr)
}

func GetFileSize(filename string) int64 {
	info, err := os.Stat(filename)
	if err != nil {
		return 0
	}
	return info.Size()
}

// PathMatch will need to have an algorithm for matching a path to a pattern that
//goes beyond what wildcard.Match() can do.
//Examples of things that wildcard.Match() will not cover but should:
//  "./file" should match "file" if it's in the same directory
//  "aDirectory" should be treated as "aDirectory/*", thus
//  "aDirectory" should not be added as a file, only its contents
func PathMatch(pattern, path string) bool {
	return wildcard.Match(pattern, path)
}
