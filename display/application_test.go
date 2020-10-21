package display

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/arkenproject/ait/types"
	"github.com/stretchr/testify/assert"
)

func TestReadApplicationModTiming(t *testing.T) {
	correctWd := filepath.Join(build.Default.GOPATH, "src", "ait")
	_ = os.Chdir(correctWd)
	commitPath := filepath.Join(".ait", "commit")
	ioutil.WriteFile(commitPath, commitTestPrompt, 0644)
	app := ReadApplication() //read the file
	ReadApplication()        //should use the struct in memory
	time.Sleep(15 * time.Second)
	//modifying the file with os.OpenFile does not actually change what the os
	//thinks is the file's last modified time. However, if you go open up vim
	//and modify it during these 15 seconds, you'll see the desired behavior.
	//run with coverage to see.
	app = ReadApplication() //should re-read the file if you modify it somehow
	printApp(app)           //any changes made should be reflected here
}

func TestIsValidApplication(t *testing.T) {
	_ = ioutil.WriteFile("temp", commitTestPrompt, 0644)
	assert.True(t, isValidAppTemplate("temp"))
	_ = ioutil.WriteFile("temp", []byte(""), 0644)
	assert.False(t, isValidAppTemplate("temp"))
	_ = ioutil.WriteFile("temp", []byte(`# COMMIT
#TITLE`), 0644)
	assert.False(t, isValidAppTemplate("temp"))
	_ = ioutil.WriteFile("temp", []byte(`# COMMIT
# TITLE`), 0644)
	assert.True(t, isValidAppTemplate("temp"))
	os.Remove("temp")
}

func printApp(app *types.ApplicationContents) {
	fmt.Print(app.Title, "\n\n", app.Commit, "\n\n", app.PRBody, "\n\n", app.KsName, "\n")
}

var commitTestPrompt = []byte(
	`# Provide a name for the keyset file that is about to be created
# FILENAME below
testing_filename
# Briefly describe the files you're submitting (preferably <50 characters).
# TITLE below
This is a title
# An empty commit message will abort the submission.# Describe the files in more detail. 
# Note: lines starting with '#' are excluded from messages
# COMMIT below
this is a commit message.
# If you will be submitting a pull request, explain why these files should be added
# to the desired repository
# PULL REQUEST below
This is pull request body message, and it should be many lines long. 
`)
