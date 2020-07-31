package display

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/arkenproject/ait/utils"
)

var commitPrompt = //temporary
`# Provide a name for the keyset file that is about to be created (no file extension, just the name)
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

var application *ApplicationContents

//ShowApplication pulls up our template application, currently stored in the
//string above.
func ShowApplication() {
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
	now := time.Now()
	_ = os.Chtimes(commitPath, now, now)
	//ignored because docs say that if this function an error, it's a PathError,
	//and if commitPath was bad, the program would have already crashed.
}

//ReadApplication reads a text file and puts it into a struct. It keeps track of
//when the last time the commit file was modified, so after one read, this method
//can be called at will without incurring slow file i/o, as long as the file isn't
//modified.
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
	ptr := &application.ksName
	for scanner.Scan() { //fill out the struct with the contents of the file
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") {
			*ptr += line + " \n"
		} else if strings.HasPrefix(line, "# TITLE below") {
			ptr = &application.title
		} else if strings.HasPrefix(line, "# COMMIT below") {
			ptr = &application.commit
		} else if strings.HasPrefix(line, "# PULL REQUEST below") {
			ptr = &application.prBody
		}
	}
	application.TrimFields()
	application.ksName += ".ks"
	application.timeFilled = time.Now()
	return application
}
