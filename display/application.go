package display

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/arkenproject/ait/config"
	"github.com/arkenproject/ait/utils"
)

var application *ApplicationContents

// ShowApplication pulls up our template application, currently stored in the
// string above.
func ShowApplication(repoPath string) {
	commitPath := filepath.Join(".ait", "commit")
	// Don't overwrite the commit file if it already exists.
	if s, _ := utils.GetFileSize(commitPath); s == 0 { 
		//^ if the commit file is empty and/or does not exist, one must be 
		//fetched from the appropriate source
		fetchApplicationTemplate(repoPath, commitPath)
	}
	execPath, err := exec.LookPath(config.Global.General.Editor)
	if err != nil {
		utils.FatalPrintf("%v, your configured editor, could not be found. "+
			"Please make sure it is installed and in your OS's PATH "+
			"or change it in the ~/.ait/ait.config file.\n", config.Global.General.Editor)
	}

	cmd := exec.Command(execPath, commitPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	// Display the editor to the user by running the command.
	err = cmd.Run()
	utils.CheckError(err)
	now := time.Now()
	_ = os.Chtimes(commitPath, now, now)
	// Ignored because docs say that if this function an error, it's a PathError,
	// and if commitPath was bad, the program would have already crashed.
}

// fetchApplicationTemplate fetches the prompt that will be shown to the user.
// It will preferentially choose the cloned repository, but if there is none
// there, the default application template that lives in ~/.ait/application.md
// will be used instead. The appropriate template is deep-copied into
// ./.ait/commit, so this function can cause the program to terminate if i/o
// errors arise
func fetchApplicationTemplate(repoPath, destPath string) {
	fromPath := filepath.Join(repoPath, "application.md")
	//       := ./.ait/sources/<repo-name>/application.md
	if !isValidAppTemplate(fromPath) { // false if the file does not exist
		fromPath = filepath.Join(filepath.Dir(config.Path), "application.md")
		//       = ~/.ait/application.md
		fmt.Println("The application template file found in the cloned " +
			"repo was found to be invalid, using default instead\n ")
	}
	if !isValidAppTemplate(fromPath) {
		utils.FatalPrintf(`Your default application template stored in %v is invalid. 
This means you probably edited it such that it has duplicate labels or
it is missing one or more of the required fields, Commit and Title. 
Please backup %v/ait.config, delete folder %v and rerun ait. 
This will generate a default application template.
In the future, please refrain from editing %v.
`, fromPath, filepath.Dir(fromPath), filepath.Dir(fromPath), fromPath)
	}
	err := utils.CopyFile(fromPath, destPath)
	if err != nil {
		utils.FatalPrintln(err)
	}
}

// isValidAppTemplate makes sure the application template at the given path meets
// the following standards:
//   1. Must not have ANY duplicate labels ("# COMMIT" is an example of a label)
//   2. Must contain at least a title and commit field
// Returns false if any error occurs when opening the file at the given path.
func isValidAppTemplate(path string) bool {
	commitFile, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return false
	}
	defer commitFile.Close()
	scanner := bufio.NewScanner(commitFile)
	scanner.Split(bufio.ScanLines)

	reqs := map[string]bool {
		"# LOCATION": false,
		"# FILENAME": false,
		"# TITLE": false,
		"# COMMIT": false,
		"# PULL REQUEST" : false,
	}

	for scanner.Scan() {
		line := scanner.Text()
		for key := range reqs {
			if strings.HasPrefix(line, key) {
				if reqs[key] {   // this means the label was already found in
					return false // the file, meaning there's a duplicate label
				}
				reqs[key] = true
			}
		}
	}
	if !reqs["# COMMIT"] || !reqs["# TITLE"] {
		return false
	}
	return true
}

// ReadApplication reads a text file and puts it into a struct. It keeps track of
// when the last time the commit file was modified, so after one read, this method
// can be called at will without incurring slow file i/o, as long as the file isn't
// modified.
func ReadApplication() *ApplicationContents {
	commitPath := filepath.Join(".ait", "commit")
	lastMod, err := utils.GetFileModTime(commitPath)

	if application != nil && err == nil && application.timeFilled.After(lastMod) {
		return application
	} else if application == nil {
		application = &ApplicationContents{}
	}
	commitFile := utils.BasicFileOpen(commitPath, os.O_RDONLY, 0644)
	defer commitFile.Close()

	application.Clear()
	scanner := bufio.NewScanner(commitFile)
	scanner.Split(bufio.ScanLines)
	var ptr *string = nil

	// Fill out the struct with the contents of the file
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") && ptr != nil {
			*ptr += line + " \n"
		} else if strings.HasPrefix(line, "# CATEGORY") {
			ptr = &application.category
		} else if strings.HasPrefix(line, "# FILENAME") {
			ptr = &application.ksName
		} else if strings.HasPrefix(line, "# TITLE") {
			ptr = &application.title
		} else if strings.HasPrefix(line, "# COMMIT") {
			ptr = &application.commit
		} else if strings.HasPrefix(line, "# PULL REQUEST") {
			ptr = &application.prBody
		}
	}
	application.TrimFields()

	if !strings.HasSuffix(application.ksName, ".ks") {
		application.ksName += ".ks"
	}
	application.timeFilled = time.Now()
	return application
}
