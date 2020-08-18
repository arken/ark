package display

import (
	"bufio"
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
func ShowApplication() {
	commitPath := filepath.Join(".ait", "commit")
	// Don't overwrite the commit file if it already exists.
	if s, _ := utils.GetFileSize(commitPath); s == 0 {
		fromPath := filepath.Join(filepath.Dir(config.Path), "application.md")
		err := utils.CopyFile(fromPath, commitPath)
		utils.CheckError(err)
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

// ReadApplication reads a text file and puts it into a struct. It keeps track of
// when the last time the commit file was modified, so after one read, this method
// can be called at will without incurring slow file i/o, as long as the file isn't
// modified.
func ReadApplication() *ApplicationContents {
	commitPath := filepath.Join(".ait", "commit")
	commitFile := utils.BasicFileOpen(commitPath, os.O_RDONLY, 0644)

	defer commitFile.Close()

	scanner := bufio.NewScanner(commitFile)
	scanner.Split(bufio.ScanLines)

	lastMod, err := utils.GetFileModTime(commitPath)
	if application != nil && err == nil && application.timeFilled.After(lastMod) {
		return application
	} else if application == nil {
		application = &ApplicationContents{}
	}

	application.Clear()
	ptr := &application.category

	// Fill out the struct with the contents of the file
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") {
			*ptr += line + " \n"
		} else if strings.HasPrefix(line, "# FILENAME below") {
			ptr = &application.ksName
		} else if strings.HasPrefix(line, "# TITLE below") {
			ptr = &application.title
		} else if strings.HasPrefix(line, "# COMMIT below") {
			ptr = &application.commit
		} else if strings.HasPrefix(line, "# PULL REQUEST below") {
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
