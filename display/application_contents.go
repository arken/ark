package display

import (
	"strings"
	"time"
)

type ApplicationContents struct {
	title      string
	commit     string
	prBody     string
	category   string
	ksName     string
	timeFilled time.Time //when the application struct was last filled in
}

//TrimFields trims the spaces off of all fields.
func (app *ApplicationContents) TrimFields() {
	app.title = strings.TrimSpace(app.title)
	app.commit = strings.TrimSpace(app.commit)
	app.prBody = strings.TrimSpace(app.prBody)
	app.category = strings.TrimSpace(app.category)
	app.ksName = strings.TrimSpace(app.ksName)
}

func (app *ApplicationContents) GetTitle() string {
	return app.title
}

func (app *ApplicationContents) GetCommit() string {
	return app.commit
}

func (app *ApplicationContents) GetPRBody() string {
	return app.prBody
}

func (app *ApplicationContents) GetCategory() string {
	return app.category
}

func (app *ApplicationContents) GetKSName() string {
	return app.ksName
}

func (app *ApplicationContents) IsEmpty() bool {
	return len(app.title) + len(app.commit) + len(app.prBody) +
		   len(app.category) + len(app.ksName) == 0
}

func (app *ApplicationContents) Clear() {
	app.title    = ""
	app.commit   = ""
	app.prBody   = ""
	app.category = ""
	app.ksName = ""
}
