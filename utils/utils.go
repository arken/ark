package utils

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/arkenproject/ait/types"
)

const AddedFilesPath string = ".ait/added_files" //can later be put somewhere more central

// IsAITRepo is a trivial check to see if the program's working dir is an ait repo.
func IsAITRepo() bool {
	fs, err := os.Stat(".ait")
	if err != nil {
		return false
	}
	return fs.IsDir()
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
	return strings.HasPrefix(dir, pathToCheck)
}

// FillSet splits the given file by newline and adds each line to the given set.
func FillSet(contents types.StringSet, file *os.File) {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		if len(scanner.Text()) > 0 {
			contents.Add(scanner.Text())
		}
	}
}

// DumpSet dumps all values in the given set into the given file, separated by
// newlines.
func DumpSet(contents types.StringSet, file *os.File) error {
	toDump := make([]byte, 0, 256)
	_ = contents.ForEach(func(line string) error {
		bLine := []byte(line)
		for i := 0; i < len(bLine); i++ {
			toDump = append(toDump, bLine[i])
		}
		toDump = append(toDump, '\n')
		return nil
	})
	_, err := file.Write(toDump)
	return err
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

// SubmissionCleanup attempts to delete the sources and commit file. Nothing
// is done if either of those operations is unsuccessful
func SubmissionCleanup() {
	_ = os.RemoveAll(filepath.Join(".ait", "sources"))
	_ = os.Remove(".ait/commit")
}

// IsWithinRepo tests if the given path is within this current repo.
func IsWithinRepo(path string) (bool, error) {
	var err error
	var wd string
	path, err = filepath.Abs(path)
	if err != nil {
		return false, err
	}
	wd, err = os.Getwd()
	if err != nil {
		return false, err
	}
	return strings.HasPrefix(path, wd), nil
}

// IndexOf returns the index of key in slice, or -1 if it doesn't exist
func IndexOf(slice []string, key string) int {
	for i, s := range slice {
		if s == key {
			return i
		}
	}
	return -1
}

// IsGithubRemote lexically checks if the given url appears to be an HTTPS url to a
// Github repository. It returns true and an empty string if it is, and false
// and an explanation of the problem if it is not.
func IsGithubRemote(url string) (bool, string) {
	//https://github.com/a/a.git
	re := regexp.MustCompile(`https://github\.com/([a-zA-Z1-9\-_]+)/([a-zA-Z1-9\-_]+)\.git`)
	// This is a naive check for remote url validity. It is not 100% correct,
	// for example names can't start with -, there can't be two dashes in a row,
	// and others. But it's good enough to catch common mistakes.
	var msg string
	if re.MatchString(url) {
		return true, ""
	}
	// Go through some common mistakes
	if len(url) < 26 { //It's shorter than "https://github.com/a/a.git"
		msg += "The URL is not long enough to possibly be a full HTTPS Github remote.\n"
	} else {
		if !strings.Contains(url, "github.com") {
			msg += "The URL does not contain \"github.com\". Currently, we only " +
				"support GitHub remotes, but support for others may be added in the future.\n"
		}
		if strings.HasPrefix(url, "git@") {
			msg += "The URL uses the SSH protocol which we do not support at the moment.\n"
		}
		if !strings.HasSuffix(url, ".git") {
			msg += "The URL does not end in .git."
		}
	}
	if strings.HasSuffix(msg, "\n") {
		msg = msg[0:len(msg) - 1] //cut off the newline.
	}
	return false, msg
}
