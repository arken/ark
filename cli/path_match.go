package cli

import "github.com/minio/minio/pkg/wildcard"

//This function will need to have an algorithm for matching a path to a pattern that
//goes beyond what wildcard.Match() can do.
//Examples of things that wildcard.Match() will not cover but should:
//  "./file" should match "file" if it's in the same directory
//  "aDirectory" should be treated as "aDirectory/*", thus
//  "aDirectory" should not be added as a file, only its contents
func PathMatch(pattern, path string) bool {
	return wildcard.Match(pattern, path)
}