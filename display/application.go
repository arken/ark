package display

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/arkenproject/ait/utils"
)

var commitPrompt = //temporary
`# Provide a name for the keyset file that is about to be created
# FILENAME below

# Briefly describe the files you're submitting (preferably <50 characters).
# TITLE below

# An empty commit message will abort the submission.# Describe the files in more detail. 
# Note: lines starting with '#' are excluded from messages
# COMMIT below

# If you will be submitting a pull request, explain why these files should be added
# to the desired repository
# PULL REQUEST below

`

type ApplicationContents struct {
	Title    string
	Commit   string
	PRBody   string
	Category string
	KSPath   string
}

//TrimFields trims the spaces off of all fields.
func (app *ApplicationContents) TrimFields() {
	app.Title    = strings.TrimSpace(app.Title)
	app.Commit   = strings.TrimSpace(app.Commit)
	app.PRBody   = strings.TrimSpace(app.PRBody)
	app.Category = strings.TrimSpace(app.Category)
	app.KSPath   = strings.TrimSpace(app.KSPath)
}


var Appliation *ApplicationContents

// CollectCommit queries the user to fill out the commit template.
func CollectCommit() string {
	editor := "vim" //eventually this will come from the global config struct
	commitPath := filepath.Join(".ait", "commit")
	if s, _ := utils.GetFileSize(commitPath); s == 0 { //commit file is empty
		_ = ioutil.WriteFile(commitPath, []byte(commitPrompt), 0644)
	} //don't overwrite if it already has something in it
	execPath, err := exec.LookPath(editor)
	if err != nil {
		log.Fatalf("%v, your requested editor, could not be found. "+
			"Please make sure it is installed and in your OS's PATH.", editor)
	}
	cmd := exec.Command(execPath, commitPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	res := ReadApplication()
	return res.Title + "\n\n" + res.Commit
}

//ReadApplication reads a text file and puts it into a string, with newlines
//as they appear in the file. Lines that start with '#' are not included.
func ReadApplication() *ApplicationContents {
	commitPath := filepath.Join(".ait", "commit")
	commitFile := utils.BasicFileOpen(commitPath, os.O_RDONLY, 0644)
	defer commitFile.Close()
	scanner := bufio.NewScanner(commitFile)
	scanner.Split(bufio.ScanLines)
	result := &ApplicationContents{}
	ptr := &result.KSPath
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") {
			*ptr += line + " \n"
		} else if strings.HasPrefix(line, "# TITLE below") {
			ptr = &result.Title
		} else if strings.HasPrefix(line, "# COMMIT below") {
			ptr = &result.Commit
		} else if strings.HasPrefix(line, "# PULL REQUEST below") {
			ptr = &result.PRBody
		}
	}
	result.TrimFields()
	return result
}
