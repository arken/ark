package types

import (
	"strings"
	"time"
)

// ApplicationContents is a struct holding the application fields.
type ApplicationContents struct {
	Title      string
	Commit     string
	PRBody     string
	Category   string
	KsName     string
	TimeFilled time.Time //when the application struct was last filled in
}

// TrimFields trims the spaces off of all fields.
func (app *ApplicationContents) TrimFields() {
	app.Title = strings.TrimSpace(app.Title)
	app.Commit = strings.TrimSpace(app.Commit)
	app.PRBody = strings.TrimSpace(app.PRBody)
	app.Category = strings.TrimSpace(app.Category)
	app.KsName = strings.TrimSpace(app.KsName)
}

// IsEmpty checks if the Commit file is empty.
func (app *ApplicationContents) IsEmpty() bool {
	return len(app.Title)+len(app.Commit)+len(app.PRBody)+
		len(app.Category)+len(app.KsName) == 0
}

// Clear empties the values of the struct.
func (app *ApplicationContents) Clear() {
	app.Title = ""
	app.Commit = ""
	app.PRBody = ""
	app.Category = ""
	app.KsName = ""
}

// IsValid return true only if both the Title and Commit fields are not empty.
func (app *ApplicationContents) IsValid() bool {
	return len(app.Title) != 0 && len(app.Commit) != 0
}
