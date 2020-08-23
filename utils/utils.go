package utils

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const AddedFilesPath string = ".ait/added_files" //can later be put somewhere more central

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

// PathMatch checks if two paths match using wildcards, but it will also return
// true if path is in a subdirectory of pattern.
func PathMatch(pattern, path string) bool {
	matched, _ := filepath.Match(pattern, path)
	return matched || IsInSubDir(path, pattern)
}

// FillMap splits the given file by newline and adds each line to the given map.
func FillMap(contents map[string]struct{}, file *os.File) {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		if len(scanner.Text()) > 0 {
			contents[scanner.Text()] = struct{}{}
		}
	}
}

// Dumps all keys in the given map to the given file, separated by a newline.
func DumpMap(contents map[string]struct{}, file *os.File) error {
	for line := range contents {
		_, err := file.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}

// GetRepoName returns the name of a repo given its HTTPS or SSH address. If no
// name was found, the empty string is returned.
func GetRepoName(url string) string {
	index := strings.LastIndex(url, "/") + 1
	if index < 0 || len(url)-4 < 0 || index > len(url)-4 {
		return ""
	} else {
		return url[index : len(url)-4]
	}
}

// GetRepoOwner returns the owner of a repo given its HTTPS or SSH address. If
// no name was found, the empty string is returned.
func GetRepoOwner(url string) string {
	if len(url) < 19 {
		return ""
	}
	start := 19 // == len("https://github.com/")
	end := strings.Index(url[start:], "/") + start
	if end < start {
		return ""
	}
	return url[start:end]
}

// BasicFileOpen just opens a file and log.Fatal's any error that arises
func BasicFileOpen(path string, flag int, bits os.FileMode) *os.File {
	file, err := os.OpenFile(path, flag, bits)
	CheckError(err)
	return file
}

// GetFileModTime returns the file at the path's last time of modification
// according to the OS. If there is an error, it returns time.Now() and the error.
func GetFileModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Now(), err
	}
	return info.ModTime(), nil
}

// FatalPrintln Println's the given arguments and then exits with exit code 1.
func FatalPrintln(a ...interface{}) {
	if a != nil {
		fmt.Println(a...)
	}
	os.Exit(1)
}

// FatalPrintf Printf's the given arguments and then exits with exit code 1.
func FatalPrintf(format string, a ...interface{}) {
	if a != nil {
		fmt.Printf(format, a...)
	} else {
		fmt.Printf(format)
	}
	os.Exit(1)
}

// CheckError checks if the given error is nil, and if not it FatalPrintln's the
// error.
func CheckError(err error) {
	if err != nil {
		FatalPrintln(err)
	}
}

// CopyFile performs a depp copy of the path at fromPath to toPath. Returns any
// returns any i/o errors that arise.
func CopyFile(fromPath, toPath string) error {
	fromBytes, err := ioutil.ReadFile(fromPath)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(toPath, fromBytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

// CheckErrorWithCleanup checks the given error for nil and calls the given
// cleanup function if it is not nil. Optionally add a custom message to be
// printed, if none is provided, the error is printed.
func CheckErrorWithCleanup(err error, cleanup func(), a ...interface{}) {
	if err != nil {
		cleanup()
		if a == nil {
			FatalPrintln(err)
		}
		FatalPrintln(a...)
	}
}

// FatalWithCleanup calls the given function then calls FatalPrintln with the
// other arg(s)
func FatalWithCleanup(cleanup func(), a ...interface{}) {
	cleanup()
	FatalPrintln(a...)
}

// submissionCleanup attempts to delete the sources and commit file. Nothing
// is done if either of those operations is unsuccessful
func SubmissionCleanup() {
	_ = os.RemoveAll(filepath.Join(".ait", "sources"))
	_ = os.Remove(".ait/commit")
}
