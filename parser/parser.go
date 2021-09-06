package parser

import (
	"bufio"
	"errors"
	"path/filepath"
	"strings"
)

// Application is a struct holding the application fields.
type Application struct {
	Title    string
	Commit   string
	PRBody   string
	Category string
	Filename string
}

func ParseApplication(input string) (Application, error) {
	app := Application{}

	scanner := bufio.NewScanner(strings.NewReader(input))
	var ptr *string = nil

	// Fill out the struct with the contents of the file
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") && ptr != nil {
			*ptr += line + " \n"
		} else if strings.HasPrefix(line, "# CATEGORY") {
			ptr = &app.Category
		} else if strings.HasPrefix(line, "# FILENAME") {
			ptr = &app.Filename
		} else if strings.HasPrefix(line, "# TITLE") {
			ptr = &app.Title
		} else if strings.HasPrefix(line, "# COMMIT") {
			ptr = &app.Commit
		} else if strings.HasPrefix(line, "# PULL REQUEST") {
			ptr = &app.PRBody
		}
	}

	// Trim whitespace from around fields
	app.Category = strings.TrimSpace(filepath.Clean(app.Category))
	app.Commit = strings.TrimSpace(app.Commit)
	app.Filename = strings.TrimSpace(app.Filename)
	app.PRBody = strings.TrimSpace(app.PRBody)
	app.Title = strings.TrimSpace(app.Title)

	if !strings.HasSuffix(app.Filename, ".ks") {
		app.Filename += ".ks"
	}

	if strings.Contains(app.Category, "..") {
		return app, errors.New("path backtracking (\"..\") is not allowed in the category")
	}
	return app, nil
}
