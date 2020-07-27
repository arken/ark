package cli

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
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

// GetFileSize returns the size of the file at the given path in bytes plus any
//error encountered by os.Stat()
func GetFileSize(filename string) (int64, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

//IsInSubDir checks if pathToCheck is in a subdirectory of dir.
func IsInSubDir(dir, pathToCheck string) bool {
	pathAbs, _ := filepath.Abs(pathToCheck)
	dirAbs, _ := filepath.Abs(dir)
	return strings.HasPrefix(dirAbs, pathAbs)
}

//PathMatch checks if two paths match using wildcards, but it will also return
//true if path is in a subdirectory of pattern.
func PathMatch(pattern, path string) bool {
	matched, _ := filepath.Match(pattern, path)
	return matched || IsInSubDir(path, pattern)
}

//Splits the given file by newline and adds each line to the given map.
func FillMap(contents map[string]struct{}, file *os.File) {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		if len(scanner.Text()) > 0 {
			contents[scanner.Text()] = struct{}{}
		}
	}
}

//Dumps all keys in the given map to the given file, separated by a newline.
func DumpMap(contents map[string]struct{}, file *os.File) error {
	for line := range contents {
		_, err := file.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}
