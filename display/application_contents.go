package display

import (
	"strings"
	"time"
)

// ApplicationContents is a read-only struct holding the application fields.
type ApplicationContents struct {
	title      string
	commit     string
	prBody     string
	category   string
	ksName     string
	timeFilled time.Time //when the application struct was last filled in
}

// TrimFields trims the spaces off of all fields.
func (app *ApplicationContents) TrimFields() {
	app.title = strings.TrimSpace(app.title)
	app.commit = strings.TrimSpace(app.commit)
	app.prBody = strings.TrimSpace(app.prBody)
	app.category = strings.TrimSpace(app.category)
	app.ksName = strings.TrimSpace(app.ksName)
}

// GetTitle returns the name of the title.
func (app *ApplicationContents) GetTitle() string {
	return app.title
}

// GetCommit returns the commit message.
func (app *ApplicationContents) GetCommit() string {
	return app.commit
}

// GetPRBody returns the Pull Request Message.
func (app *ApplicationContents) GetPRBody() string {
	return app.prBody
}

// GetCategory returns the Category of the KeySet File.
func (app *ApplicationContents) GetCategory() string {
	return app.category
}

// GetKSName returns the name of the KeySet File.
func (app *ApplicationContents) GetKSName() string {
	return app.ksName
}

// IsEmpty checks if the commit file is empty.
func (app *ApplicationContents) IsEmpty() bool {
	return len(app.title)+len(app.commit)+len(app.prBody)+
		len(app.category)+len(app.ksName) == 0
}

// Clear empties the values of the struct.
func (app *ApplicationContents) Clear() {
	app.title = ""
	app.commit = ""
	app.prBody = ""
	app.category = ""
	app.ksName = ""
}

// IsValid return true only if both the title and commit fields are not empty.
func (app *ApplicationContents) IsValid() bool {
	return len(app.title) != 0 && len(app.commit) != 0
}
