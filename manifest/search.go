package manifest

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func (m *Manifest) Search(path string) (map[string][]string, error) {
	// Create hashes map to hold results
	hashes := make(map[string][]string)

	// Define the file's category to improve search times.
	category := filepath.Dir(path)
	base := filepath.Base(path)

	err := filepath.Walk(m.path, func(path string, info os.FileInfo, err error) error {
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

				if matched, _ := filepath.Match(base, data[1]); matched {
					if hashes[data[1]] == nil {
						hashes[data[1]] = []string{}
					}
					hashes[data[1]] = append(hashes[data[1]], data[0])
				}
			}

		}
		return nil
	})
	return hashes, err
}
