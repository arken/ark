package display

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	aitgh "github.com/arkenproject/ait/apis/github"
	"github.com/arkenproject/ait/config"
	"github.com/arkenproject/ait/types"
	"github.com/arkenproject/ait/utils"
)

var application *types.ApplicationContents

// ShowApplication pulls up our template application, currently stored in the
// string above.
func ShowApplication() {
	appPath := filepath.Join(".ait", "commit")
	// Don't overwrite the commit file if it already exists.
	if s, _ := utils.GetFileSize(appPath); s == 0 {
		//^ if the commit file is empty and/or does not exist, one must be
		//fetched from the appropriate source
		fetchApplicationTemplate(appPath)
	}
	execPath, err := exec.LookPath(config.Global.General.Editor)
	if err != nil {
		utils.FatalPrintf("%v, your configured editor, could not be found. "+
			"Please make sure it is installed and in your OS's PATH "+
			"or change it in the ~/.ait/ait.config file.\n", config.Global.General.Editor)
	}

	cmd := exec.Command(execPath, appPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	// Display the editor to the user by running the command.
	err = cmd.Run()
	utils.CheckError(err)
	now := time.Now()
	_ = os.Chtimes(appPath, now, now)
	// Ignored because docs say that if this function an error, it's a PathError,
	// and if appPath was bad, the program would have already crashed.
}

// fetchApplicationTemplate fetches the prompt that will be shown to the user.
// It will preferentially choose the cloned repository, but if there is none
// there, the default application template that lives in ~/.ait/application.md
// will be used instead. The appropriate template is deep-copied into
// ./.ait/commit, so this function can cause the program to terminate if i/o
// errors arise
func fetchApplicationTemplate(destPath string) {
	fromPath, err := aitgh.DownloadRepoAppTemplate()
	// downloads the file into fromPath if it existed in the repo.
	if err == nil && fileIsValidTemplate(fromPath) { // false if the file does not exist

		_ = os.Remove(destPath)
		_ = os.Rename(fromPath, destPath)
		return
	}
	// application template in repo was invalid/missing, use default instead
	fromPath = filepath.Join(filepath.Dir(config.Path), "application.md")
	//       = ~/.ait/application.md
	if fileIsValidTemplate(fromPath) {
		err = utils.CopyFile(fromPath, destPath)
		utils.CheckError(err)
	} else {
		_ = os.Remove(destPath)
		utils.FatalPrintf(`Your default application template stored in %v is invalid. 
This means you probably edited it such that it has duplicate labels or
it is missing one or more of the required fields, Commit and Title. 
Please backup %v/ait.config, delete folder %v and rerun ait. 
This will generate a default application template.
In the future, please refrain from editing %v.
`, fromPath, filepath.Dir(fromPath), filepath.Dir(fromPath), fromPath)
	}
}

func fileIsValidTemplate(path string) bool {
	commitFile, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return false
	}
	return validateTemplate(commitFile)
}

// validateTemplate makes sure the application template at the given path meets
// the following standards:
//   1. Must not have ANY duplicate labels ("# COMMIT" is an example of a label)
//   2. Must contain at least a title and commit field
// Returns false if any error occurs when opening the file at the given path.
func validateTemplate(reader io.ReadCloser) bool {
	defer reader.Close()
	reqs := map[string]bool{
		"# LOCATION":     false,
		"# FILENAME":     false,
		"# TITLE":        false,
		"# COMMIT":       false,
		"# PULL REQUEST": false,
	}
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
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
// modified. Return nil if the file does not exist.
func ReadApplication() *types.ApplicationContents {
	appPath := filepath.Join(".ait", "commit")
	if !utils.FileExists(appPath) {
		return nil
	}
	lastMod, err := utils.GetFileModTime(appPath)

	if application != nil && err == nil && application.TimeFilled.After(lastMod) {
		return application
	} else if application == nil {
		application = &types.ApplicationContents{}
	}
	appFile := utils.BasicFileOpen(appPath, os.O_RDONLY, 0644)
	defer appFile.Close()

	application.Clear()
	scanner := bufio.NewScanner(appFile)
	scanner.Split(bufio.ScanLines)
	var ptr *string = nil

	// Fill out the struct with the contents of the file
	var isComment bool
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "<!--") {
			isComment = true
		}
		if !strings.HasPrefix(line, "#") && !isComment && ptr != nil {
			*ptr += line + " \n"
		} else if strings.HasPrefix(line, "# CATEGORY") {
			ptr = &application.Category
		} else if strings.HasPrefix(line, "# FILENAME") {
			ptr = &application.KsName
		} else if strings.HasPrefix(line, "# TITLE") {
			ptr = &application.Title
		} else if strings.HasPrefix(line, "# COMMIT") {
			ptr = &application.Commit
		} else if strings.HasPrefix(line, "# PULL REQUEST") {
			ptr = &application.PRBody
		}
		if strings.Contains(line, "-->") {
			isComment = false
		}
	}
	application.TrimFields()
	sanitizeCategory()
	if !strings.HasSuffix(application.KsName, ".ks") {
		application.KsName += ".ks"
	}
	application.TimeFilled = time.Now()
	return application
}

func sanitizeCategory() {
	category := &application.Category
	if strings.HasPrefix(*category, string(filepath.Separator)) {
		*category = (*category)[1:]
	}
	*category = filepath.Clean(*category)
	if strings.Contains(*category, "..") {
		utils.FatalWithCleanup(utils.SubmissionCleanup,
			"Path backtracking (\"..\") is not allowed in the Category.")
	}
}
