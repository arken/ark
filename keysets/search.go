package keysets

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Search checks for the existance of a file in a keyset and returns
// its' coorisponding CID hash if found.
func Search(keysetPath, filePath string) (hashes map[string][]string, err error) {
	hashes = make(map[string][]string)
	filedata := strings.Split(filePath, "/")
	category := filedata[0]

	filename, err := regexp.Compile(filedata[1])
	if err != nil {
		return hashes, err
	}

	err = filepath.Walk(keysetPath, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, category+".ks") {
			file, err := os.Open(path)
			if err != nil {
				return err
			}

			scanner := bufio.NewScanner(file)

			// Scan through the lines in the file.
			for scanner.Scan() {
				// Split data on white space.
				data := strings.Fields(scanner.Text())

				if filename.MatchString(data[1]) {
					if hashes[data[1]] == nil {
						hashes[data[1]] = []string{}
					}
					hashes[data[1]] = append(hashes[data[1]], data[0])
				}
			}

		}
		return nil
	})

	return hashes, nil
}
