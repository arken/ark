package keysets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arkenproject/ait/ipfs"
	"github.com/arkenproject/ait/utils"
)

const delimiter string = "\t\t"

func Generate(path string, overwrite bool) error {
	if overwrite {
		return createNew(path)
	} else {
		return amendExisting(path)
	}
}

// createNew creates a keyset file with the given path. Path should not be the
// desired directory, rather it should be a full path to a file which does not
// exist yet (will be truncated if it does exist), and the file should end in
// ".ks" The resultant keyset files contains the name (not path) of the file and
// an IPFS cid hash, separated by a space.
func createNew(path string) error {
	os.MkdirAll(filepath.Dir(path), os.ModePerm)

	keySetFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	addedFiles, err := os.OpenFile(utils.AddedFilesPath, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	contents := make(map[string]struct{})
	utils.FillMap(contents, addedFiles)
	addedFiles.Close()
	var maxTitle int
	for filePath := range contents {
		filename := strings.Join(strings.Fields(filepath.Base(filePath)), "-")
		if len(filename) > maxTitle {
			maxTitle = len(filename)
		}
	}
	for filePath := range contents {
		cid, err := ipfs.Add(filePath)
		if err != nil {
			cleanup(keySetFile)
			return err
		}
		// Scrub filename for spaces and replace with dashes.
		filename := strings.Join(strings.Fields(filepath.Base(filePath)), "-")
		line := fmt.Sprintf("%v%v%v\n", filename, strings.Repeat(" ", maxTitle+4-len(filename)), cid)
		_, err = keySetFile.WriteString(line)
		if err != nil {
			cleanup(keySetFile)
			return err
		}
	}
	return keySetFile.Close()
}

// amendExisting looks at current files in added_files and adds any that aren't
// already in the keyset file to the keyset files. The keyset file in question
// should be at path.
func amendExisting(path string) error {
	keySetFile, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer keySetFile.Close()
	addedFiles, err := os.OpenFile(utils.AddedFilesPath, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer addedFiles.Close()
	addedFilesContents := make(map[string]struct{})
	utils.FillMap(addedFilesContents, addedFiles) //full of just filenames
	ksContents := make(map[string]struct{})
	utils.FillMap(ksContents, keySetFile) //full of filenames and cid's
	//TODO finish this
	return nil
}

func cleanup(file *os.File) {
	path := file.Name()
	file.Close()
	_ = os.Remove(path)
}
